package common

import (
	"errors"
	"fmt"
	"log"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/corpix/uarand"
	"github.com/valyala/fasthttp"
)

var Status503Error = errors.New("Server returned 503 status response")
var Status500Error = errors.New("Server returned 500 status response. (Slow down)")

// WebArchive and Common Crawl (index.commoncrawl.org) CDX API Response structure from
type CdxResponse struct {
	Urlkey       string `json:"urlkey,omitempty"`
	Timestamp    string `json:"timestamp,omitempty"`
	Charset      string `json:"charset,omitempty"`
	MimeType     string `json:"mime,omitempty"`
	Languages    string `json:"languages,omitempty"`
	MimeDetected string `json:"mimedetected,omitempty"`
	Digest       string `json:"digest,omitempty"`
	Offset       string `json:"offset,omitempty"`
	Original     string `json:"url,omitempty"` // Original URL
	Length       string `json:"length,omitempty"`
	StatusCode   string `json:"status,omitempty"`
	Filename     string `json:"filename,omitempty"`
	Source       Source
}

// Source of web archive data
type Source interface {
	Name() string
	ParseResponse(resp []byte) ([]*CdxResponse, error)
	GetNumPages(url string) (int, error)
	GetPages(config RequestConfig) ([]*CdxResponse, error)
	FetchPages(config RequestConfig, results chan []*CdxResponse, errors chan error)
	GetFile(*CdxResponse) ([]byte, error)
}

type RequestConfig struct {
	URL            string   // Url to parse
	Filters        []string // Extenstion to search
	Limit          uint     // Max number of results per page
	CollapseColumn string   // Which column to use to collapse results
	SinglePage     bool     // Get results only from 1st page (mostly used for tests)
	FromDate       string   // Filter results from Date
	ToDate         string   // Filter results to Date
}

// GetUrlFromConfig ... Compose URL with CDX server request parameters
func (config RequestConfig) GetUrl(serverURL string, page int) string {
	reqURL := fmt.Sprintf("%v?url=%v&output=json", serverURL, config.URL)

	if config.Limit != 0 {
		reqURL = fmt.Sprintf("%v&limit=%v", reqURL, config.Limit)
	}

	if config.CollapseColumn != "" {
		reqURL = fmt.Sprintf("%v&collapse=%v", reqURL, config.CollapseColumn)
	}

	for _, filter := range config.Filters {
		if filter != "" {
			reqURL = fmt.Sprintf("%v&filter=%v", reqURL, filter)
		}
	}

	if config.FromDate != "" {
		reqURL = fmt.Sprintf("%v&from=%v", reqURL, config.FromDate)
	}

	if config.ToDate != "" {
		reqURL = fmt.Sprintf("%v&to=%v", reqURL, config.ToDate)
	}

	if !config.SinglePage {
		reqURL = fmt.Sprintf("%v&page=%v", reqURL, page)
	}
	return reqURL
}

func DoRequest(url string, timeout int, headers map[string]string) ([]byte, error) {
	timeoutDuration := time.Second * time.Duration(timeout)

	req := fasthttp.AcquireRequest()
	req.SetRequestURI(url)
	req.Header.SetMethod(fasthttp.MethodGet)
	req.Header.Set(fasthttp.HeaderUserAgent, uarand.GetRandom())
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	defer fasthttp.ReleaseRequest(req)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	client := &fasthttp.Client{}
	client.ReadTimeout = timeoutDuration
	err := client.DoTimeout(req, resp, timeoutDuration)
	if err != nil {
		return nil, fmt.Errorf("[GetRequest] Error making request: %v", err)
	}

	switch resp.StatusCode() {
	case 500:
		return nil, Status500Error
	case 503:
		return resp.Body(), Status503Error
	}

	if len(resp.Body()) > 0 {
		return resp.Body(), nil
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("[GetRequest] Got %v status response", resp.StatusCode())
	}

	if resp.Body() == nil {
		return nil, fmt.Errorf("[GetRequest] Response body is empty")
	}

	return resp.Body(), nil
}

// Get ... Performs HTTP GET request and returns response bytes
func Get(url string, timeout int, maxRetries int) ([]byte, error) {
	var err error
	var responseBytes []byte

	for i := maxRetries; i != 0; i-- {
		log.Printf("GET [t=%v] [r=%v]: %v", timeout, maxRetries, url)

		responseBytes, err = DoRequest(url, timeout, nil)
		if err == nil {
			return responseBytes, nil
		}

		if err == Status503Error || err == Status500Error {
			time.Sleep(time.Duration(timeout * int(time.Second)))
		}
	}

	return nil, fmt.Errorf("Perfomed max retries, no result: %v", err)
}

// Save data using file fullpath
func SaveFile(data []byte, path string) error {
	err := os.WriteFile(path, data, 0644)
	if err != nil {
		return err
	}

	return nil
}

// Save files from CDX Response channel into output directory
func SaveFiles(results <-chan []*CdxResponse, outputDir string, errors chan error, downloadRate float32) {
	log.Println("[SaveFiles] worker started:", outputDir)

	for {
		select {
		case resBatch, ok := <-results:
			if ok {
				for _, res := range resBatch {
					data, err := res.Source.GetFile(res)
					if err != nil {
						errors <- err
						continue
					}

					exts, _ := mime.ExtensionsByType(res.MimeType)
					if exts == nil {
						exts = []string{""}
					}

					filename := fmt.Sprintf("%v-%v-%v%v", res.Original, res.Timestamp, res.Source.Name(), exts[0])
					escapedFilename := url.QueryEscape(filename)
					fullPath := filepath.Join(outputDir, escapedFilename)

					err = SaveFile(data, fullPath)
					if err != nil {
						errors <- err
					}

					time.Sleep(time.Second * time.Duration(downloadRate))
				}
			}
		}
	}

}

func GetFileExtenstion(file *[]byte) (string, error) {
	contentType := http.DetectContentType(*file)
	contentType = strings.Split(contentType, ";")[0]
	exts, err := mime.ExtensionsByType(contentType)
	if err != nil || len(exts) == 0 {
		return "", fmt.Errorf("Cannot get extension from file")
	}

	return exts[0], nil
}
