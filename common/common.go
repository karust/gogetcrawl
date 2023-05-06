package common

import (
	"fmt"
	"os"
	"time"

	"github.com/corpix/uarand"
	"github.com/valyala/fasthttp"
)

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

func getRequest(url string, timeout int) ([]byte, error) {
	req := fasthttp.AcquireRequest()
	req.SetRequestURI(url)
	req.Header.SetMethod(fasthttp.MethodGet)
	req.Header.Set(fasthttp.HeaderUserAgent, uarand.GetRandom())

	resp := fasthttp.AcquireResponse()
	client := &fasthttp.Client{}
	err := client.DoTimeout(req, resp, time.Duration(timeout)*time.Second)
	if err != nil {
		return nil, fmt.Errorf("[GetRequest] Error making request: %v", err)
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

	for i := maxRetries; i >= 0; i-- {
		responseBytes, err := getRequest(url, timeout)
		if err == nil {
			return responseBytes, nil
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
