package main

import (
	"log"

	common "github.com/karust/goCommonCrawl/common"
	wayback "github.com/karust/goCommonCrawl/wayback"
)

func main() {
	// index := flag.String("i", "CC-MAIN-2020-24", "Define Index File")
	// url := flag.String("u", "", "URL to crawl")
	// dest := flag.String("d", "./data", "Destination to save downloaded content")

	// flag.Parse()

	// if *url == "" {
	// 	println("You should provide URL '-u {URL}'")
	// 	println("	'-h' for help")
	// 	return
	// }

	// err := os.MkdirAll(*dest, os.ModePerm)
	// if err != nil {
	// 	log.Fatalln(err)
	// }

	// pages, err := cc.GetPages(*index, *url, "", 100, 45, true)
	// if err != nil {
	// 	log.Fatalln(err)
	// }

	// err = cc.SaveContent(*pages, *dest, 45)
	// if err != nil {
	// 	log.Fatalln(err)
	// }

	config := common.RequestConfig{
		URL:        "cia.gov/*",
		Filters:    []string{"statuscode:200", "mimetype:application/pdf"},
		Limit:      6,
		SinglePage: true,
		Timeout:    10,
		MaxRetries: 2,
	}

	pages, err := wayback.GetPages(config)
	if err != nil {
		log.Fatalln(err)
	}

	for _, p := range pages {
		data, err := wayback.GetFile(p.Original, p.Timestamp)
		if err != nil {
			log.Fatalln(err)
		}
		common.SaveFile(data, "F:\\Projects\\goCommonCrawl\\data\\"+p.Digest+".pdf")
	}
}
