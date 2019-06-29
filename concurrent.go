package commoncrawl

import (
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"log"
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

// Config ... Holds configurations of crawler
type Config struct {
	ResultChan chan Result
	Timeout    int
	CrawlDB    string
	WaitMS     int
	Extensions []string
	MaxAmount  int
	//RandomizeName bool
	//HashFilter    bool
}

func saveContent(pages []IndexAPI, saveTo string, timeout int, config Config) {
	defer func() {
		if r := recover(); r != nil {
			config.ResultChan <- Result{Error: fmt.Errorf("saveContent recover: %v", r)}
		}
	}()

	downloaded := 0
	client := http.Client{Timeout: time.Duration(timeout) * time.Second}

	for i, page := range pages {
		if downloaded >= config.MaxAmount {
			return
		}

		offset, _ := strconv.Atoi(page.Offset)
		length, _ := strconv.Atoi(page.Length)
		offsetEnd := offset + length + 1

		req, _ := http.NewRequest("GET", crawlStorage+page.Filename, nil)
		req.Header.Set("Range", fmt.Sprintf("bytes=%v-%v", offset, offsetEnd))
		req.Header.Set("User-Agent", randomOption(userAgents))

		resp, err := client.Do(req)
		if err != nil {
			config.ResultChan <- Result{Error: fmt.Errorf("saveContent request error: %v", err)}
			return
		}
		defer resp.Body.Close()

		//Deflate response and split the WARC, HEADER, HTML from it
		reader, err := gzip.NewReader(resp.Body)
		if err != nil {
			config.ResultChan <- Result{Error: fmt.Errorf("saveContent error deflating response 1: %v", err)}
			//continue
		}
		if reader == nil {
			continue
		}
		defer reader.Close()

		b, err := ioutil.ReadAll(reader)
		if err != nil {
			config.ResultChan <- Result{Error: fmt.Errorf("saveContent error deflating response 2: %v", err)}
			//continue
		}
		//fmt.Println(string(b))
		splitted := strings.Split(string(b), "\r\n\r\n")
		warc := splitted[0]
		response := splitted[2]

		// Return if extension of file is not in allowed
		ext := ExtensionByContent([]byte(response))
		if !IsExtensionExist(config.Extensions, ext) {
			continue
		}

		startURL := strings.Index(warc, "WARC-Target-URI:") + 17
		endURL := strings.Index(warc, "\r\nWARC-Payload-Digest")

		link := warc[startURL:endURL]
		linkEscaped := EscapeURL(link)

		// Save extracted file and write progess to channel
		err = ioutil.WriteFile(saveTo+"/"+linkEscaped+ext, []byte(response), 0644)
		if err != nil {
			config.ResultChan <- Result{Error: fmt.Errorf("saveContent writing file error: %v", err)}
			continue
		}
		config.ResultChan <- Result{URL: link, Progress: i + 1, Total: len(pages)}
		time.Sleep(time.Millisecond * time.Duration(config.WaitMS))
		downloaded++
	}
}

// FetchURLData ... Fetches pages located on the given URL from Common Crawl archive using Index API and saves them to the pointed location
// Set `crawlDB` argument empty if not sure what it is
func FetchURLData(url string, saveto string, config Config) {
	crawlDB, timeout := "CC-MAIN-2019-22", 30
	if config.ResultChan == nil {
		log.Fatalln("[FetchURLData] No Result channel provided")
	}
	if config.Timeout != 0 {
		timeout = config.Timeout
	}
	if config.CrawlDB != "" {
		crawlDB = config.CrawlDB
	}

	// Create directory if not exists
	if _, err := os.Stat(saveto); os.IsNotExist(err) {
		err := os.Mkdir(saveto, os.ModeDir)
		if err != nil {
			config.ResultChan <- Result{Error: err}
			return
		}
	}

	urlSite := "*." + url // Clutch :(

	// Get info about URL from Index server
	pages, err := GetPagesInfo(crawlDB, urlSite, timeout)
	if err != nil {
		config.ResultChan <- Result{Error: err}
		return
	}

	// Retrieve found pages from Amazon S3 storage and save them
	saveContent(pages, saveto, timeout, config)
	config.ResultChan <- Result{Done: true, URL: url}
}
