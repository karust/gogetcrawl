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
)

var indexServer = "http://index.commoncrawl.org/"
var crawlStorage = "https://commoncrawl.s3.amazonaws.com/"

// IndexAPI ... API Response structure from `index.commoncrawl.org`
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

// GetPagesInfo ... Makes request to commoncrawl index API to gather all offsets that contain pointed URL
//   crawl: Crawl a database which should be used, e.g 'CC-MAIN-2019-22';
//   url: URL of a site, offsets and other info of which should be returned.
// Returns a list of JSON objects with information about each file offset and other data.
func GetPagesInfo(crawl string, url string) ([]IndexAPI, error) {
	// Build request
	client := http.Client{Timeout: 30 * time.Second}
	req, _ := http.NewRequest("GET", indexServer+crawl+"-index", nil)

	// Add request params and do it
	q := req.URL.Query()
	q.Add("url", url)
	q.Add("output", "json")
	req.URL.RawQuery = q.Encode()
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("getIndex response read error: %v", err)
	}

	// Read response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("getIndex response read error: %v", err)
	}

	// Parse the response that contains JSON objects seperated with new line
	rawPages := strings.Split(string(body), "\n")
	pages := []IndexAPI{}
	for _, p := range rawPages {
		val := &IndexAPI{}
		err := json.NewDecoder(strings.NewReader(p)).Decode(&val)
		if err != nil {
			fmt.Println(err)
			continue
		}
		pages = append(pages, *val)
	}
	return pages, nil
}

// SaveContent ... Saves pages or text that were found in Common Crawl to choosen folder
//   pages: info about found web pages from `getIndex` function
//   saveTo: destination fodler, where save fetched web data
func SaveContent(pages []IndexAPI, saveTo string) error {
	client := http.Client{Timeout: 30 * time.Second}

	for i, page := range pages {
		offset, _ := strconv.Atoi(page.Offset)
		length, _ := strconv.Atoi(page.Length)
		offsetEnd := offset + length + 1

		req, _ := http.NewRequest("GET", crawlStorage+page.Filename, nil)
		req.Header.Set("Range", fmt.Sprintf("bytes=%v-%v", offset, offsetEnd))

		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("saveContent response read error: %v", err)
		}

		// Read response
		if err != nil {
			return fmt.Errorf("saveContent response read error: %v", err)
		}
		// Deflate response and split the WARC, HEADER, HTML from it
		reader, _ := gzip.NewReader(resp.Body)
		b, err := ioutil.ReadAll(reader)
		splitted := strings.Split(string(b), "\r\n\r\n")
		warc := splitted[0]
		//header := splitted[1]
		response := splitted[2]

		ext := extensionByContent([]byte(response))

		startURL := strings.Index(warc, "WARC-Target-URI:") + 17
		endURL := strings.Index(warc, "\r\nWARC-Payload-Digest")
		url := escapeURL(warc[startURL:endURL])

		// Write extracted HTML and show progess
		ioutil.WriteFile(saveTo+"/"+url+ext, []byte(response), 0644)
		fmt.Printf("Page %v/%v\n", i+1, len(pages))
	}
	return nil
}

// ChangeIndexServer ... changes default address `http://index.commoncrawl.org/`
func ChangeIndexServer(server string) {
	indexServer = server
}

func main() {
	pages, err := GetPagesInfo("CC-MAIN-2019-22", "example.com/")
	if err != nil {
		fmt.Println(err)
	}

	SaveContent(pages, "./data")
}
