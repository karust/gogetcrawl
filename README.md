# Go Get Crawl
[![Go Report Card](https://goreportcard.com/badge/github.com/karust/goGetCrawl)](https://goreportcard.com/report/github.com/karust/gogetcrawl)
[![Go Reference](https://pkg.go.dev/badge/github.com/karust/gogetcrawl.svg)](https://pkg.go.dev/github.com/karust/gogetcrawl)

**gogetcrawl** is a tool and package which help you download URLs and Files from popular Web Archives like [Common Crawl](http://commoncrawl.org) and [Wayback Machine](https://web.archive.org/). You can use it as a command line tool or import the solution into your Go project. 

## Installation
### Source
```
go install github.com/karust/gogetcrawl@latest
```

### Docker
```
docker build -t gogetcrawl .
docker run gogetcrawl --help
```

### Binary
Check out the latest release if you need binary [here](https://github.com/karust/gogetcrawl/releases).


## Usage
### Docker
```
	docker run uranusq/gogetcrawl url *.tutorialspoint.com/* --ext pdf
```
### Cmd usage
* See commands and flags:
```
gogetcrawl -h
```

#### Get URLs

* You can fetch multiple domains archive data, the flags will be applied to each. By default you'll get all results displayed in your terminal:
```
gogetcrawl url *.example.com kamaloff.ru 
```

* To limit number of results, output to file and select only Wayback as source you can:
```
gogetcrawl url *.example.com kamaloff.ru --limit 10 --sources wb -o ./urls.txt
```

#### Download files
* To download 10 `PDF` files to `./test` directory with 3 workers:
```
gogetcrawl download *.cia.gov/* --limit 10 -w 3 -d ./test -f "mimetype:application/pdf"
```

### Package usage
```
go get github.com/karust/gogetcrawl
```
*For both Wayback and Common crawl you can use `concurrent` and `non-concurrent` ways to interract with archives*
#### Wayback
* **Get urls:**
```go
package main

import (
	"fmt"

	"github.com/karust/gogetcrawl/common"
	"github.com/karust/gogetcrawl/wayback"
)

func main() {
	// Get only 10 status:200 pages
	config := common.RequestConfig{
		URL:     "*.example.com/*",
		Filters: []string{"statuscode:200"},
		Limit:   10,
	}

	// Set requests timout and retries
	wb, _ := wayback.New(15, 2)

	// Use config to obtain all CDX server responses
	results, _ := wb.GetPages(config)

	for _, r := range results {
		fmt.Println(r.Urlkey, r.Original, r.MimeType)
	}
}
```

* **Get files:**
```go
// Get all status:200 HTML files 
config := common.RequestConfig{
	URL:     "*.tutorialspoint.com/*",
	Filters: []string{"statuscode:200", "mimetype:text/html"},
}

wb, _ := wayback.New(15, 2)
results, _ := wb.GetPages(config)

// Get first file from CDX response
file, err := wb.GetFile(results[0])

fmt.Println(string(file))
```

#### Common Crawl
*To use Common Crawl you just need to replace `wayback` module with `commoncrawl`. Let's use Common Crawl concurretly*

* **Get urls:**
```go
cc, _ := commoncrawl.New(30, 3)

config1 := common.RequestConfig{
	URL:        "*.tutorialspoint.com/*",
	Filters:    []string{"statuscode:200", "mimetype:text/html"},
	Limit:      6,
}

config2 := common.RequestConfig{
	URL:        "example.com/*",
	Filters:    []string{"statuscode:200", "mimetype:text/html"},
	Limit:      6,
}

resultsChan := make(chan []*common.CdxResponse)
errorsChan := make(chan error)

go func() {
	cc.FetchPages(config1, resultsChan, errorsChan)
}()

go func() {
	cc.FetchPages(config2, resultsChan, errorsChan)
}()

for {
	select {
	case err := <-errorsChan:
		fmt.Printf("FetchPages goroutine failed: %v", err)
	case res, ok := <-resultsChan:
		if ok {
			fmt.Println(res)
		}
	}
}
```

* **Get files:**
```go
config := common.RequestConfig{
	URL:     "kamaloff.ru/*",
	Filters: []string{"statuscode:200", "mimetype:text/html"},
}

cc, _ := commoncrawl.New(15, 2)
results, _ := wb.GetPages(config)
file, err := cc.GetFile(results[0])
```

## Bugs + Features
If you have some issues/bugs or feature request, feel free to open an issue.