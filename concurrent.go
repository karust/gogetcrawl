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
	Done     bool
}

func saveContent(pages []IndexAPI, saveTo string, res chan Result, timeout int, waitMS int) {
	// defer func() {
	// 	if r := recover(); r != nil {
	// 		res <- Result{Error: fmt.Errorf("saveContent recover: %v", r)}
	// 	}
	// }()

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
		req.Header.Set("User-Agent", randomOption(userAgents))

		resp, err := client.Do(req)
		if err != nil {
			res <- Result{Error: fmt.Errorf("saveContent request error: %v", err)}
			return
		}
		defer resp.Body.Close()

		//Deflate response and split the WARC, HEADER, HTML from it
		reader, err := gzip.NewReader(resp.Body)
		if err != nil {
			res <- Result{Error: fmt.Errorf("saveContent error deflating response 1: %v", err)}
			//continue
		}
		if reader == nil {
			continue
		}
		defer reader.Close()

		b, err := ioutil.ReadAll(reader)
		if err != nil {
			res <- Result{Error: fmt.Errorf("saveContent error deflating response 2: %v", err)}
			//continue
		}
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
			continue
		}
		res <- Result{URL: url, Progress: i + 1, Total: len(pages)}
		time.Sleep(time.Millisecond * time.Duration(waitMS))
	}
}

// FetchURLData ... Fetches pages located on the given URL from Common Crawl archive using Index API and saves them to the pointed location
// Set `crawlDB` argument empty if not sure what it is
func FetchURLData(url string, saveto string, res chan Result, timeout int, crawlDB string, waitMS int) {
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

	urlSite := "*." + url // KOSTYL
	// Get info about URL from Index server
	pages, err := GetPagesInfo(crawlDB, urlSite, timeout)
	if err != nil {
		res <- Result{Error: err}
		return
	}

	// Retrieve found pages from Amazon S3 storage and save them
	saveContent(pages, saveto, res, timeout, waitMS)
	res <- Result{Done: true, URL: url}
}
