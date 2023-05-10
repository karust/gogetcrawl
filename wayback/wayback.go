package wayback

import (
	"fmt"
	"strconv"

	jsoniter "github.com/json-iterator/go"
	common "github.com/karust/goCommonCrawl/common"
)

const INDEX_SERVER = "https://web.archive.org/cdx/search/cdx"
const CRAWL_STORAGE = "https://web.archive.org/web"

var STD_TIMEOUT = 30
var STD_RETRIES = 3

// WebArchive API Response structure
type IndexAPI struct {
	Urlkey     string
	Timestamp  string
	Original   string
	MimeType   string
	StatusCode string
	Digest     string
	Length     string
}

// FetchPages ... Concurrent way to GetPages.
// Makes request to WebArchive index API and return observations in a channel.
func FetchPages(config common.RequestConfig, results chan []*IndexAPI, errors chan error) {
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
		reqURL := common.GetUrlFromConfig(INDEX_SERVER, config)
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

// GetPages ... Makes request to WebArchive index API to gather all url observations
func GetPages(config common.RequestConfig) ([]*IndexAPI, error) {
	var pages int
	var err error

	if config.SinglePage {
		pages = 1
	} else {
		pages, err = GetNumPages(config.URL)
		if err != nil {
			return nil, err
		}
	}

	var results []*IndexAPI

	for page := 0; page < pages; page++ {
		reqURL := common.GetUrlFromConfig(INDEX_SERVER, config)
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

// Parse response from https://web.archive.org/cdx/search/cdx index server
func ParseResponse(resp []byte) ([]*IndexAPI, error) {
	var results [][]string

	err := jsoniter.Unmarshal(resp, &results)
	if err != nil {
		return nil, fmt.Errorf("[ParseResponse] Failed to decode wayback results '%v'", err)
	}

	parsedResults := []*IndexAPI{}
	for i, entry := range results {
		if i == 0 {
			continue
		}
		parsed := IndexAPI{
			Urlkey:     entry[0],
			Timestamp:  entry[1],
			Original:   entry[2],
			MimeType:   entry[3],
			StatusCode: entry[4],
			Digest:     entry[5],
			Length:     entry[6],
		}

		parsedResults = append(parsedResults, &parsed)
	}

	return parsedResults, nil
}

// Return the number of pages located in WebArchive for given url
func GetNumPages(url string) (int, error) {

	requestURI := fmt.Sprintf("%v?url=%v&showNumPages=true", INDEX_SERVER, url)
	response, err := common.Get(requestURI, STD_TIMEOUT, STD_RETRIES)
	if err != nil {
		return 0, fmt.Errorf("[GetNumPages] Request error: %v", err)
	}

	// Remove return and convert to integer
	strRes := string(response[:len(response)-1])
	res, err := strconv.Atoi(strRes)
	if err != nil {
		return 0, fmt.Errorf("[GetNumPages] Cannot convert response value: %v", response)
	}

	return res, nil
}

// Download file from WebArchive using link from Index API
func GetFile(url, timestamp string) ([]byte, error) {

	requestURI := fmt.Sprintf("%v/%vid_/%v", CRAWL_STORAGE, timestamp, url)
	response, err := common.Get(requestURI, STD_TIMEOUT, STD_RETRIES)
	if err != nil {
		return nil, fmt.Errorf("[GetNumPages] Request error: %v", err)
	}
	return response, nil
}
