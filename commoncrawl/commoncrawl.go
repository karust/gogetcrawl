package commoncrawl

import (
	"bytes"
	"fmt"
	"io"
	"strconv"

	jsoniter "github.com/json-iterator/go"
	common "github.com/karust/goCommonCrawl/common"
	"github.com/slyrz/warc"
)

const INDEX_SERVER = "https://index.commoncrawl.org/"
const CRAWL_STORAGE = "https://data.commoncrawl.org/" //https://commoncrawl.s3.amazonaws.com/
var STD_TIMEOUT = 30
var STD_RETRIES = 3

var latestIndexCache = []LatestIndex{}

// API Response structure from index.commoncrawl.org
type IndexAPI struct {
	Urlkey       string `json:"urlkey,omitempty"`
	Timestamp    string `json:"timestamp,omitempty"`
	Charset      string `json:"charset,omitempty"`
	Mime         string `json:"mime,omitempty"`
	Languages    string `json:"languages,omitempty"`
	MimeDetected string `json:"mimedetected,omitempty"`
	Digest       string `json:"digest,omitempty"`
	Offset       string `json:"offset,omitempty"`
	URL          string `json:"url,omitempty"`
	Length       string `json:"length,omitempty"`
	Status       string `json:"status,omitempty"`
	Filename     string `json:"filename,omitempty"`
}

// JSON response containing latest index name at http://index.commoncrawl.org/collinfo.json
type LatestIndex struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	Timegate string `json:"timegate"`
	CdxAPI   string `json:"cdx-api"`
}

// ex: http://index.commoncrawl.org/CC-MAIN-2015-11-index?url=*.wikipedia.org/&showNumPages=true
type NumPagesResponse struct {
	Pages    int `json:"pages"`
	PageSize int `json:"pageSize"`
	Blocks   int `json:"blocks"`
}

func GetIndexIDs() ([]LatestIndex, error) {
	if len(latestIndexCache) != 0 {
		return latestIndexCache, nil
	}

	response, err := common.Get(INDEX_SERVER+"/collinfo.json", STD_TIMEOUT, STD_RETRIES)
	if err != nil {
		return nil, fmt.Errorf("[GetIndexIDs] response read error: %v", err)
	}

	latestIndexes := []LatestIndex{}
	err = jsoniter.Unmarshal(response, &latestIndexes)
	if err != nil {
		return latestIndexes, fmt.Errorf("[GetIndexIDs] Cannot get latest index ID: %v", err)
	}

	latestIndexCache = latestIndexes
	return latestIndexes, nil
}

// Returns the number of pages located in CommonCrawl for given url
//	index: needs to be set manually here like "CC-MAIN-2023-14"
func GetNumPagesIndex(url, index string) (int, error) {
	requestURI := fmt.Sprintf("%v%v-index?url=%v&showNumPages=true", INDEX_SERVER, index, url)

	response, err := common.Get(requestURI, STD_TIMEOUT, STD_RETRIES)
	if err != nil {
		return 0, fmt.Errorf("[GetNumPagesIndex] Request error: %v", err)
	}

	numPagesResp := NumPagesResponse{}
	err = jsoniter.Unmarshal(response, &numPagesResp)
	if err != nil {
		return 0, fmt.Errorf("[GetNumPagesIndex] JSON decode error: %v", err)
	}

	return numPagesResp.Pages, nil
}

// Returns the number of pages located in CommonCrawl for given url
// Use latest index from http://index.commoncrawl.org/collinfo.json
func GetNumPages(url string) (int, error) {
	indexIDs, err := GetIndexIDs()
	if err != nil {
		return 0, err
	}
	return GetNumPagesIndex(url, indexIDs[0].Id)
}

// Parse response from http://index.commoncrawl.org/*****-index index server
func ParseResponse(resp []byte) ([]*IndexAPI, error) {
	pages := []*IndexAPI{}

	// Parse the response that contains JSON objects seperated with new line
	for _, line := range bytes.Split(resp[:len(resp)-1], []byte{'\n'}) {
		var indexVal IndexAPI
		if err := jsoniter.Unmarshal(line, &indexVal); err != nil {
			return nil, fmt.Errorf("[ParseResponse] Cannot decode JSON line: %v. Response: %v", err, string(line))
		}
		pages = append(pages, &indexVal)
	}

	return pages, nil
}

// GetPagesIndex ... Makes request to WebArchive index API to gather all url observations
//	index: needs to be set manually here like "CC-MAIN-2023-14"
func GetPagesIndex(config common.RequestConfig, index string) ([]*IndexAPI, error) {
	var pages int
	var err error

	if config.SinglePage {
		pages = 1
	} else {
		pages, err = GetNumPagesIndex(config.URL, index)
		if err != nil {
			return nil, err
		}
	}

	var results []*IndexAPI

	for page := 0; page < pages; page++ {
		indexURL := fmt.Sprintf("%v%v-index", INDEX_SERVER, index)

		reqURL := common.GetUrlFromConfig(indexURL, config)
		if !config.SinglePage {
			reqURL = fmt.Sprintf("%v&page=%v", reqURL, page)
		}

		response, err := common.Get(reqURL, config.Timeout, config.MaxRetries)
		if err != nil {
			return results, fmt.Errorf("[GetPages] Request error: %v", err)
		}

		parsedResponse, err := ParseResponse(response)
		if err != nil {
			return results, fmt.Errorf("[GetPages] Cannot parse response: %v", err)
		}
		results = append(results, parsedResponse...)
	}

	return results, nil
}

// Makes request to the Commoncrawl index API to gather all offsets that contain chosen URL.
//	Uses the latest CommonCrawl index.
func GetPages(config common.RequestConfig) ([]*IndexAPI, error) {
	indexes, err := GetIndexIDs()
	if err != nil {
		return nil, err
	}

	return GetPagesIndex(config, indexes[0].Id)
}

// FetchPages is a concurrent way to GetPages.
// Makes request to CommonCrawl index API and returns observations in a channel.
//	index: needs to be set manually here
func FetchPages(config common.RequestConfig, index string, results chan []*IndexAPI, errors chan error) {
	var pages int
	var err error

	if config.SinglePage {
		pages = 1
	} else {
		pages, err = GetNumPages(config.URL)
		if err != nil {
			errors <- err
		}
	}

	for page := 0; page < pages; page++ {
		indexURL := fmt.Sprintf("%v%v-index", INDEX_SERVER, index)
		reqURL := common.GetUrlFromConfig(indexURL, config)
		if !config.SinglePage {
			reqURL = fmt.Sprintf("%v&page=%v", reqURL, page)
		}

		response, err := common.Get(reqURL, config.Timeout, config.MaxRetries)
		if err != nil {
			errors <- fmt.Errorf("[GetPages] Request error: %v", err)
		}

		parsedResponse, err := ParseResponse(response)
		if err != nil {
			errors <- fmt.Errorf("[GetPages] Cannot parse response: %v", err)
		}

		results <- parsedResponse
	}
}

// Gets files from CommonCrawl storage using info from IndexAPI server
//   page: info about found web page in IndexAPI
//   timeout: timeout in seconds
func GetFile(page *IndexAPI, timeout int) ([]byte, error) {
	offset, _ := strconv.Atoi(page.Offset)
	length, _ := strconv.Atoi(page.Length)
	offsetEnd := offset + length + 1

	headers := map[string]string{
		"Range": fmt.Sprintf("bytes=%v-%v", page.Offset, offsetEnd),
	}
	resp, err := common.DoRequest(CRAWL_STORAGE+page.Filename, STD_TIMEOUT, headers)
	if err != nil {
		return nil, fmt.Errorf("[GetFile] Request error: %v", err)
	}

	reader, err := warc.NewReader(bytes.NewReader(resp))
	if err != nil {
		return nil, fmt.Errorf("[GetFile] Cannot decode WARC: %v", err)
	}
	defer reader.Close()

	for {
		record, err := reader.ReadRecord()
		if err != nil {
			return nil, fmt.Errorf("[GetFile] Cannot decode WARC: %v", err)
		}

		var buf bytes.Buffer
		io.Copy(&buf, record.Content)
		return buf.Bytes(), nil
	}
}
