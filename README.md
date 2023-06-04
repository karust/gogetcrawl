# Go Get Crawl
[![Go Report Card](https://goreportcard.com/badge/github.com/karust/goGetCrawl)](https://goreportcard.com/report/github.com/karust/gogetcrawl)
[![Go Reference](https://pkg.go.dev/badge/github.com/karust/gogetcrawl.svg)](https://pkg.go.dev/github.com/karust/gogetcrawl)

**gogetcrawl** is a tool and package that helps you download URLs and Files from popular Web Archives like [Common Crawl](http://commoncrawl.org) and [Wayback Machine](https://web.archive.org/). You can use it as a command line tool or import the solution into your Go project. 

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
Check out the latest release [here](https://github.com/karust/gogetcrawl/releases).

## Usage
### Docker
```
docker run uranusq/gogetcrawl url *.tutorialspoint.com/* --ext pdf --limit 5
```
### Docker compose
```
docker-compose up --build
```
### CLI usage
* See commands and flags:
```
gogetcrawl -h
```

#### Get URLs

* You can get multiple-domain archive data, flags will be applied to each. By default, you will get all results displayed in your terminal (use `--collapse` to get **unique** results):
```
gogetcrawl url *.example.com *.tutorialspoint.com/* --collapse
```

* To **limit** the number of results, enable output to a file and select only Wayback as a **source** you can:
```
gogetcrawl url *.tutorialspoint.com/* --limit 10 --sources wb -o ./urls.txt
```

* Set **date range**:
```
gogetcrawl url *.tutorialspoint.com/* --limit 10 --from 20140131 --to 20231231
```
#### Download files
* Download 5 `PDF` files to `./test` directory with 3 **workers**:
```
gogetcrawl download *.cia.gov/* --limit 5 -w 3 -d ./test -f "mimetype:application/pdf"
```

### Package usage
```
go get github.com/karust/gogetcrawl
```
For both Wayback and Common crawl you can use `concurrent` and `non-concurrent` ways to interract with archives: 
#### Wayback
* **Get urls**
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

	// Set request timout and retries
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

#### CommonCrawl
*To use CommonCrawl you just need to replace `wayback` module with `commoncrawl`. Let's use Common Crawl concurretly*

* **Get urls**
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