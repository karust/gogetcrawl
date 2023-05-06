package commoncrawl

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/corpix/uarand"
)

const INDEX_SERVER = "http://index.commoncrawl.org/"

const CRAWL_STORAGE = "https://commoncrawl.s3.amazonaws.com/"

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

// GetPages ... Makes request to commoncrawl index API to gather all offsets that contain pointed URL
//   crawl: Crawl a database which should be used, e.g 'CC-MAIN-2019-22';
//   url: URL of a site, offsets and other info of which should be returned.
//   timeout: timeout in seconds, default 30
// Returns a list of JSON objects with information about each file offset and other data.
func GetPages(indexName, url, extension string, limit, timeout int, collapse bool) (*[]IndexAPI, error) {
	// Build request
	client := http.Client{Timeout: time.Duration(timeout) * time.Second}
	req, _ := http.NewRequest("GET", INDEX_SERVER+indexName+"-index", nil)
	req.Header.Set("User-Agent", uarand.GetRandom())

	// Add request params and do it
	q := req.URL.Query()
	q.Add("url", url)
	q.Add("output", "json")

	if extension != "" {
		q.Add("filter", "mime:"+extension)
	}

	if limit != 0 {
		q.Add("limit", fmt.Sprint(limit))
	}

	if collapse {
		q.Add("collapse", "urlkey")
	}

	req.URL.RawQuery = q.Encode()
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("[GetPages] Request error: %v", err)
	}

	// Read response
	defer resp.Body.Close()
	// body, err := ioutil.ReadAll(resp.Body)
	// if err != nil {
	// 	return nil, fmt.Errorf("[GetPages] Response read error: %v", err)
	// }

	// Parse the response that contains JSON objects seperated with new line
	//rawPages := strings.Split(string(body), "\n")
	pages := []IndexAPI{}
	err = json.NewDecoder(resp.Body).Decode(&pages)
	if err != nil {
		return &pages, fmt.Errorf("getIndex JSON decode error: %v", err)
	}

	// for _, p := range rawPages {
	// 	val := &IndexAPI{}
	// 	err := json.NewDecoder(strings.NewReader(p)).Decode(&val)
	// 	if err != nil {
	// 		//fmt.Errorf("getIndex JSON decode error: %v", err)
	// 		continue
	// 	}
	// 	pages = append(pages, *val)
	// }
	return &pages, nil
}

// SaveContent ... Saves pages or text that were found in Common Crawl to choosen folder
//   pages: info about found web pages from `getIndex` function
//   saveTo: destination fodler, where save fetched web data
//   timeout: timeout in seconds, default 30
func SaveContent(pages []IndexAPI, saveTo string, timeout int) error {
	if timeout == 0 {
		timeout = 30
	}

	client := http.Client{Timeout: time.Duration(timeout) * time.Second}

	for i, page := range pages {
		offset, _ := strconv.Atoi(page.Offset)
		length, _ := strconv.Atoi(page.Length)
		offsetEnd := offset + length + 1

		req, _ := http.NewRequest("GET", CRAWL_STORAGE+page.Filename, nil)
		req.Header.Set("Range", fmt.Sprintf("bytes=%v-%v", offset, offsetEnd))
		req.Header.Set("User-Agent", uarand.GetRandom())

		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("saveContent request error: %v", err)
		}

		// Deflate response and split the WARC, HEADER, HTML from it
		reader, err := gzip.NewReader(resp.Body)
		if err != nil {
			//return fmt.Errorf("saveContent error deflating response: %v", err)
		}
		b, err := ioutil.ReadAll(reader)
		if err != nil {
			//return fmt.Errorf("saveContent error deflating response: %v", err)
		}

		splitted := strings.Split(string(b), "\r\n\r\n")
		warc := splitted[0]
		//header := splitted[1]
		response := splitted[2]

		ext := ExtensionByContent([]byte(response))

		startURL := strings.Index(warc, "WARC-Target-URI:") + 17
		endURL := strings.Index(warc, "\r\nWARC-Payload-Digest")
		url := EscapeURL(warc[startURL:endURL])

		// Write extracted HTML and show progess
		err = ioutil.WriteFile(saveTo+"/"+url+ext, []byte(response), 0644)
		if err != nil {
			return fmt.Errorf("saveContent writing file error: %v", err)
		}
		fmt.Printf("Page %v/%v\n", i+1, len(pages))
	}
	return nil
}

// ChangeIndexServer ... changes default address `http://index.commoncrawl.org/`
// func ChangeIndexServer(server string) {
// 	INDEX_SERVER = server
// }

func GetIndexIDs() ([]LatestIndex, error) {
	latestIndexes := []LatestIndex{}
	client := http.Client{Timeout: time.Duration(10) * time.Second}
	req, _ := http.NewRequest("GET", INDEX_SERVER+"/collinfo.json", nil)
	//req.Header.Set("User-Agent", randomOption(userAgents))

	resp, err := client.Do(req)
	//body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return latestIndexes, fmt.Errorf("[GetPagesInfo] response read error: %v", err)
	}

	err = json.NewDecoder(resp.Body).Decode(&latestIndexes)
	if err != nil {
		return latestIndexes, fmt.Errorf("Cannot get latest index ID: %v", err)
	}

	return latestIndexes, nil
}
