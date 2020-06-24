# goCommonCrawl
**goCommonCrawl** extracts web data from [Common Crawl](http://commoncrawl.org) Web archive, that is located on Amazon S3 storage, using their [index API server](http://index.commoncrawl.org/).

## Release
Compiled version available at https://github.com/karust/goCommonCrawl/releases/tag/1

## Installation
```
go get -u github.com/karust/goCommonCrawl
```

## Usage
If you need to fetch some URL without concurrency:
```go
package main

import (
	"log"
	cc "github.com/karust/gocommoncrawl"
)

func main() {
    // Get information about `example.com/` URL from  `CC-MAIN-2019-22` archive
	pages, err := cc.GetPagesInfo("CC-MAIN-2019-22", "example.com/", 45)
	if err != nil {
		log.Fatalln(err)
	}

    // Parse retrieved pages information and save it in `./data`
	cc.SaveContent(pages, "./data", 45)
}
```

Concurrent way to do things:
```go
func main() {
	// Create channel for results
	resChan := make(chan cc.Result)

	// Some URLs to fetch pages from
	sites := []string{"medium.com/", "example.com/", "tutorialspoint.com/"}

	// Make save folder and start goroutine for each URL
	for _, url := range sites {
		// Configure request
		commonConfig := cc.Config{
			ResultChan: resChan,
			Timeout:    30,
			// Version of archive
			CrawlDB: "CC-MAIN-2019-22",
			// Wait time between AWS S3 downloads in milliseconds
			WaitMS: 53,
			// Extensions to save
			Extensions: []string{".html", ".pdf", ".doc", ".txt"},
			// Max amount of files to save
			MaxAmount: 20,
		}

		saveFolder := "./data/" + cc.EscapeURL(url)
		go cc.FetchURLData(url, saveFolder, commonConfig)
	}

	// Listen for results from goroutines
	for r := range resChan {
		if r.Error != nil {
			fmt.Printf("Error occured: %v\n", r.Error)
		} else if r.Progress > 0 {
			fmt.Printf("Progress %v: %v/%v\n", r.URL, r.Progress, r.Total)
		}
	}
}
```
In the result, you should get folders with files (mostly HTML and robot.txt) that belong to given URLs. 
