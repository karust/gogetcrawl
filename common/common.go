package common

import (
	"errors"
	"fmt"
	"log"
	"mime"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/corpix/uarand"
	"github.com/valyala/fasthttp"
)

var Status503Error = errors.New("Server returned 503 status response")
var Status500Error = errors.New("Server returned 500 status response. (Slow down)")

type RequestConfig struct {
	URL        string   // Url to parse
	Filters    []string // Extenstion to search
	Limit      int      // Max number of results
	Collapse   bool     // Collapse results by similar URLkeys
	Timeout    int      // Request timeout
	MaxRetries int      // Max number of request retries if timeouted
	SinglePage bool
}

// GetUrlFromConfig ... Compose URI with request parameters for Index server
func GetUrlFromConfig(serverURL string, config RequestConfig) string {
	reqURL := fmt.Sprintf("%v?url=%v&output=json", serverURL, config.URL)

	if config.Limit != 0 {
		reqURL = fmt.Sprintf("%v&limit=%v", reqURL, config.Limit)
	}

	if config.Collapse {
		reqURL = fmt.Sprintf("%v&collapse=urlkey", reqURL)
	}

	for _, filter := range config.Filters {
		if filter != "" {
			reqURL = fmt.Sprintf("%v&filter=%v", reqURL, filter)
		}
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
	// File from CommonCrawl storage
	case 206:
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
		log.Println("Get: ", timeout, maxRetries, url)

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

func SaveFile(data []byte, path string) error {
	err := os.WriteFile(path, data, 0644)
	if err != nil {
		return err
	}

	return nil
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
