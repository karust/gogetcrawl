package commoncrawl

import (
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// Result ... of `FetchURLData` function execution
type Result struct {
	URL      string
	Progress int
	Total    int
	Error    error
}

func saveContent(pages []IndexAPI, saveTo string, res chan Result, timeout int) {
	if timeout == 0 {
		timeout = 30
	}
	client := http.Client{Timeout: time.Duration(timeout) * time.Second}

	for i, page := range pages {
		offset, _ := strconv.Atoi(page.Offset)
		length, _ := strconv.Atoi(page.Length)
		offsetEnd := offset + length + 1

		req, _ := http.NewRequest("GET", crawlStorage+page.Filename, nil)
		req.Header.Set("Range", fmt.Sprintf("bytes=%v-%v", offset, offsetEnd))

		resp, err := client.Do(req)
		if err != nil {
			res <- Result{Error: fmt.Errorf("saveContent response read error: %v", err)}
			return
		}

		// Deflate response and split the WARC, HEADER, HTML from it
		reader, _ := gzip.NewReader(resp.Body)
		b, err := ioutil.ReadAll(reader)
		splitted := strings.Split(string(b), "\r\n\r\n")
		warc := splitted[0]
		response := splitted[2]

		ext := ExtensionByContent([]byte(response))

		startURL := strings.Index(warc, "WARC-Target-URI:") + 17
		endURL := strings.Index(warc, "\r\nWARC-Payload-Digest")
		url := warc[startURL:endURL]
		urlEsc := EscapeURL(url)

		// Write extracted HTML and show progess
		err = ioutil.WriteFile(saveTo+"/"+urlEsc+ext, []byte(response), 0644)
		if err != nil {
			res <- Result{Error: fmt.Errorf("saveContent writing file error: %v", err)}
			return
		}
		res <- Result{URL: url, Progress: i + 1, Total: len(pages)}
	}
}

// FetchURLData ... Fetches pages located on the given URL from Common Crawl archive using Index API and saves them to the pointed location
// Set `crawlDB` argument empty if not sure what it is
func FetchURLData(url string, saveto string, res chan Result, timeout int, crawlDB string) {
	if crawlDB == "" {
		crawlDB = "CC-MAIN-2019-22"
	}
	if timeout == 0 {
		timeout = 30
	}
	// Create directory if not exists
	if _, err := os.Stat(saveto); os.IsNotExist(err) {
		err := os.Mkdir(saveto, os.ModeDir)
		if err != nil {
			res <- Result{Error: err}
			return
		}
	}

	// Get info about URL from Index server
	pages, err := GetPagesInfo(crawlDB, url, timeout)
	if err != nil {
		res <- Result{Error: err}
		return
	}

	// Retrieve found pages from Amazon S3 storage and save them
	saveContent(pages, saveto, res, timeout)
}
