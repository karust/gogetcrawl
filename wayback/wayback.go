package wayback

import (
	"fmt"
	"strconv"

	jsoniter "github.com/json-iterator/go"
	common "github.com/karust/goGetCrawl/common"
)

const INDEX_SERVER = "https://web.archive.org/cdx/search/cdx"
const CRAWL_STORAGE = "https://web.archive.org/web"

type Wayback struct {
	MaxTimeout int // Request timeout
	MaxRetries int // Max number of request retries if timeouted
}

func New(timeout, retries int) (*Wayback, error) {
	source := &Wayback{MaxTimeout: timeout, MaxRetries: retries}
	return source, nil
}

func (Wayback) Name() string {
	return "Wayback"
}

// Return the number of pages located in WebArchive for given url
func (as *Wayback) GetNumPages(url string) (int, error) {

	requestURI := fmt.Sprintf("%v?url=%v&showNumPages=true", INDEX_SERVER, url)
	response, err := common.Get(requestURI, as.MaxTimeout, as.MaxRetries)
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

// Parse response from https://web.archive.org/cdx/search/cdx CDX server
func (as *Wayback) ParseResponse(resp []byte) ([]*common.CdxResponse, error) {
	var results [][]string

	err := jsoniter.Unmarshal(resp, &results)
	if err != nil {
		return nil, fmt.Errorf("[ParseResponse] Failed to decode Wayback results '%v'", err)
	}

	parsedResults := []*common.CdxResponse{}
	for i, entry := range results {
		// Skip header
		if i == 0 {
			continue
		}

		parsed := common.CdxResponse{
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

// GetPages ... Makes request to WebArchive CDX API to gather all url observations
func (as *Wayback) GetPages(config common.RequestConfig) ([]*common.CdxResponse, error) {
	var pages int
	var err error

	if config.SinglePage {
		pages = 1
	} else {
		pages, err = as.GetNumPages(config.URL)
		if err != nil {
			return nil, err
		}
	}

	var results []*common.CdxResponse
	numResults := 0

	for page := 0; page < pages; page++ {
		reqURL := common.GetUrlFromConfig(INDEX_SERVER, config, page)

		response, err := common.Get(reqURL, as.MaxTimeout, as.MaxRetries)
		if err != nil {
			return results, fmt.Errorf("[GetPages] Request error: %v", err)
		}

		parsedResponse, err := as.ParseResponse(response)
		if err != nil {
			return results, fmt.Errorf("[GetPages] Cannot parse response: %v", err)
		}
		results = append(results, parsedResponse...)
		numResults += len(parsedResponse)

		if config.Limit != 0 && uint(numResults) >= config.Limit {
			break
		}
	}

	return results, nil
}

// FetchPages ... Concurrent way to GetPages.
// Makes request to WebArchive CDX API and return observations in a channel.
func (as *Wayback) FetchPages(config common.RequestConfig, results chan []*common.CdxResponse, errors chan error) {
	var pages int
	var err error

	if config.SinglePage {
		pages = 1
	} else {
		pages, err = as.GetNumPages(config.URL)
		if err != nil {
			errors <- err
		}
	}

	numResults := 0

	for page := 0; page < pages; page++ {
		reqURL := common.GetUrlFromConfig(INDEX_SERVER, config, page)

		response, err := common.Get(reqURL, as.MaxTimeout, as.MaxRetries)
		if err != nil {
			errors <- fmt.Errorf("[FetchPages] Request error: %v", err)
		}

		parsedResponse, err := as.ParseResponse(response)
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

// Download file from WebArchive using a link from CDX response
func (as *Wayback) GetFile(page *common.CdxResponse) ([]byte, error) {
	requestURI := fmt.Sprintf("%v/%vid_/%v", CRAWL_STORAGE, page.Timestamp, page.Original)
	response, err := common.Get(requestURI, as.MaxTimeout, as.MaxRetries)
	if err != nil {
		return nil, fmt.Errorf("[GetFile] Request error: %v", err)
	}
	return response, nil
}
