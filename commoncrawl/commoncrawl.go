package commoncrawl

import (
	"bytes"
	"fmt"
	"io"
	"strconv"

	jsoniter "github.com/json-iterator/go"
	common "github.com/karust/goGetCrawl/common"
	"github.com/slyrz/warc"
)

const INDEX_SERVER = "https://index.commoncrawl.org/"
const CRAWL_STORAGE = "https://data.commoncrawl.org/" // https://commoncrawl.s3.amazonaws.com/

// JSON response containing latest index name at http://index.commoncrawl.org/collinfo.json
type latestIndex struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	Timegate string `json:"timegate"`
	CdxAPI   string `json:"cdx-api"`
}

// ex: http://index.commoncrawl.org/CC-MAIN-2015-11-index?url=*.wikipedia.org/&showNumPages=true
type numPagesResponse struct {
	Pages    int `json:"pages"`
	PageSize int `json:"pageSize"`
	Blocks   int `json:"blocks"`
}

type CommonCrawl struct {
	MaxTimeout int           // Request timeout
	MaxRetries int           // Max number of request retries if timeouted
	indexes    []latestIndex // CDX Indexes versions cache
}

func New(timeout, retries int) (*CommonCrawl, error) {
	source := &CommonCrawl{MaxTimeout: timeout, MaxRetries: retries}

	// Cache latest indexes to not overload the server
	var err error
	source.indexes, err = source.GetIndexes()
	if err != nil {
		return nil, err
	}

	return source, nil
}

func (CommonCrawl) Name() string {
	return "CommonCrawl"
}

// Get latest CDX indexes from http://index.commoncrawl.org/collinfo.json
func (cc *CommonCrawl) GetIndexes() ([]latestIndex, error) {
	response, err := common.Get(INDEX_SERVER+"/collinfo.json", cc.MaxTimeout, cc.MaxRetries)
	if err != nil {
		return nil, fmt.Errorf("[GetIndexIDs] response read error: %v", err)
	}

	latestIndexes := []latestIndex{}
	err = jsoniter.Unmarshal(response, &latestIndexes)
	if err != nil {
		return latestIndexes, fmt.Errorf("[GetIndexIDs] Cannot get latest index ID: %v", err)
	}

	return latestIndexes, nil
}

// Returns the number of pages located in CommonCrawl for given url
//	index: needs to be set manually here like "CC-MAIN-2023-14"
func (cc *CommonCrawl) GetNumPagesIndex(url, index string) (int, error) {
	requestURI := fmt.Sprintf("%v%v-index?url=%v&showNumPages=true", INDEX_SERVER, index, url)

	response, err := common.Get(requestURI, cc.MaxTimeout, cc.MaxRetries)
	if err != nil {
		return 0, fmt.Errorf("[GetNumPagesIndex] Request error: %v", err)
	}

	numPagesResp := numPagesResponse{}
	err = jsoniter.Unmarshal(response, &numPagesResp)
	if err != nil {
		return 0, fmt.Errorf("[GetNumPagesIndex] JSON decode error: %v", err)
	}

	return numPagesResp.Pages, nil
}

// Returns the number of pages located in CommonCrawl for given url
// Use latest index from http://index.commoncrawl.org/collinfo.json
func (cc *CommonCrawl) GetNumPages(url string) (int, error) {
	return cc.GetNumPagesIndex(url, cc.indexes[0].Id)
}

// Parse response from http://index.commoncrawl.org/[Index Version]-index index server
func (cc *CommonCrawl) ParseResponse(resp []byte) ([]*common.CdxResponse, error) {
	pages := []*common.CdxResponse{}

	if len(resp) == 0 {
		return nil, fmt.Errorf("Empty response provided")
	}

	// Parse the response that contains JSON objects seperated with new line
	for _, line := range bytes.Split(resp[:len(resp)-1], []byte{'\n'}) {
		var indexVal common.CdxResponse
		if err := jsoniter.Unmarshal(line, &indexVal); err != nil {
			return nil, fmt.Errorf("[ParseResponse] Cannot decode JSON line: %v. Response: %v", err, string(line))
		}
		indexVal.Source = cc
		pages = append(pages, &indexVal)
	}

	return pages, nil
}

// GetPagesIndex ... Makes request to WebArchive index API to gather all url observations
//	index: needs to be set manually here like "CC-MAIN-2023-14"
func (cc *CommonCrawl) GetPagesIndex(config common.RequestConfig, index string) ([]*common.CdxResponse, error) {
	var pages int
	var err error

	if config.SinglePage {
		pages = 1
	} else {
		pages, err = cc.GetNumPagesIndex(config.URL, index)
		if err != nil {
			return nil, err
		}
	}

	var results []*common.CdxResponse
	numResults := 0

	for page := 0; page < pages; page++ {
		indexURL := fmt.Sprintf("%v%v-index", INDEX_SERVER, index)
		reqURL := common.GetUrlFromConfig(indexURL, config, page)

		response, err := common.Get(reqURL, cc.MaxTimeout, cc.MaxRetries)
		if err != nil {
			return results, fmt.Errorf("[GetPagesIndex] Request error: %v", err)
		}

		parsedResponse, err := cc.ParseResponse(response)
		if err != nil {
			return results, fmt.Errorf("[GetPagesIndex] Cannot parse response: %v", err)
		}
		results = append(results, parsedResponse...)
		numResults += len(parsedResponse)

		if config.Limit != 0 && uint(numResults) >= config.Limit {
			break
		}
	}

	return results, nil
}

// Makes request to the Commoncrawl index API to gather all offsets that contain chosen URL.
//	Uses the latest CommonCrawl index.
func (cc *CommonCrawl) GetPages(config common.RequestConfig) ([]*common.CdxResponse, error) {
	return cc.GetPagesIndex(config, cc.indexes[0].Id)
}

// FetchPages is a concurrent way to GetPages.
// Makes request to CommonCrawl index API and returns observations in a channel.
//	index: needs to be set manually here
func (cc *CommonCrawl) FetchPages(config common.RequestConfig, results chan []*common.CdxResponse, errors chan error) {
	var pages int
	var err error

	if config.SinglePage {
		pages = 1
	} else {
		pages, err = cc.GetNumPages(config.URL)
		if err != nil {
			errors <- err
		}
	}
	numResults := 0

	for page := 0; page < pages; page++ {
		indexURL := fmt.Sprintf("%v%v-index", INDEX_SERVER, cc.indexes[0].Id)
		reqURL := common.GetUrlFromConfig(indexURL, config, page)

		response, err := common.Get(reqURL, cc.MaxTimeout, cc.MaxRetries)
		if err != nil {
			errors <- fmt.Errorf("[FetchPages] Request error: %v", err)
		}

		parsedResponse, err := cc.ParseResponse(response)
		if err != nil {
			errors <- fmt.Errorf("[FetchPages] Cannot parse response: %v", err)
		}
		numResults += len(parsedResponse)
		results <- parsedResponse

		if config.Limit != 0 && uint(numResults) >= config.Limit {
			return
		}
	}
}

// Gets files from CommonCrawl storage using info from CdxResponse server
//   page: info about found web page in CdxResponse
//   timeout: timeout in seconds
func (cc *CommonCrawl) GetFile(page *common.CdxResponse) ([]byte, error) {
	offset, _ := strconv.Atoi(page.Offset)
	length, _ := strconv.Atoi(page.Length)
	offsetEnd := offset + length + 1

	headers := map[string]string{
		"Range": fmt.Sprintf("bytes=%v-%v", page.Offset, offsetEnd),
	}
	resp, err := common.DoRequest(CRAWL_STORAGE+page.Filename, cc.MaxTimeout, headers)
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
