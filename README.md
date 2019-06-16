# goCommonCrawl
Helps to extract web data (mostly HTML pages) from Common Crawl Web archive using their index API server.

## Installation
```
go get github.com/karust/goCommonCrawl
```

## Usage
```go
package main

import (
	"log"
	cc "github.com/karust/gocommoncrawl"
)

func main() {
    // Get information about `example.com/` URL from  `CC-MAIN-2019-22` archive
	pages, err := cc.GetPagesInfo("CC-MAIN-2019-22", "example.com/")
	if err != nil {
		log.Fatalln(err)
	}

    // Parse retrieved information about pages and save it in `./data`
	cc.SaveContent(pages, "./data")
}
```