package cmd

import (
	"fmt"
	"io"
	"log"
	"mime"
	"os"
	"strings"

	"github.com/karust/gogetcrawl/common"
	"github.com/karust/gogetcrawl/commoncrawl"
	"github.com/karust/gogetcrawl/wayback"
	"github.com/spf13/cobra"
)

const version = "1.1.2"

var (
	filters      []string
	isCollapse   bool
	isSuccessful bool
	isLogging    bool
	isVerbose    bool
	maxTimeout   int
	maxRetries   int
	maxResults   uint
	maxWorkers   uint
	extensions   []string
	sourceNames  []string
)

var rootCmd = &cobra.Command{
	Use:     "gogetcrawl",
	Version: version,
	Short:   "gogetcrawl - helps you to collect URLs and Files from web archives",
	Long: `gogetcrawl is a tool that collects URLs or downloads files 
from 2 different web archive sources - Wayback Mahine and Common Crawl.
You can use different filters and arguments to solve your task more effectively.`,
}

var sources []common.Source
var results = make(chan []*common.CdxResponse)
var errors = make(chan error)

func initSources() {
	for _, s := range sourceNames {
		if s == "cc" {
			log.Println("Initializing CommonCrawl")
			cc, err := commoncrawl.New(maxTimeout, maxRetries)
			if err != nil {
				log.Fatalf("Cannot initialize CommonCrawl source: %v", err)
			}
			sources = append(sources, cc)
		}

		if s == "wb" {
			log.Println("Initializing Wayback")
			wb, err := wayback.New(maxTimeout, maxRetries)
			if err != nil {
				log.Fatalf("Cannot initialize Wayback source: %v", err)
			}
			sources = append(sources, wb)
		}
	}

	if len(sources) == 0 {
		log.Fatalf("No archive sources provided.")
	}
}

// Prepare arvhive request configs
func getRequestConfigs(args []string) chan common.RequestConfig {
	confChan := make(chan common.RequestConfig, len(args))

	if len(extensions) != 0 {
		for _, ext := range extensions {
			extRaw := mime.TypeByExtension("." + ext)
			extMime := strings.Split(extRaw, ";")[0]

			if extMime == "" {
				log.Fatalln(fmt.Sprintf("No MIME type found for '%v', please use '--filter' with correlated MIME.", ext))
			}

			filters = append(filters, "mimetype:"+extMime)
		}
	}

	if isSuccessful {
		filters = append(filters, "statuscode:200")
	}

	for _, domain := range args {
		config := common.RequestConfig{
			URL:     domain,
			Filters: filters,
			Limit:   maxResults,
		}

		if isCollapse {
			config.CollapseColumn = "urlkey"
		}
		confChan <- config
	}
	return confChan
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "There was an error while executing CLI: '%s'", err)
		os.Exit(1)
	}
}

func initArgs() {
	writers := []io.Writer{}

	if isLogging {
		file, err := os.OpenFile("./logs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			log.Fatal(err)
		}

		writers = append(writers, file)
	}

	if isVerbose {
		writers = append(writers, os.Stdout)
	}

	multi := io.MultiWriter(writers...)
	log.SetOutput(multi)
}

func init() {
	cobra.OnInitialize(initArgs)
	rootCmd.PersistentFlags().StringSliceVarP(&filters, "filter", "f", []string{}, `Filters to use. You can use multiple. Example: --filter "mimetype:application/pdf"`)
	rootCmd.PersistentFlags().BoolVarP(&isCollapse, "collapse", "c", false, `Get only unique URLs.`)
	rootCmd.PersistentFlags().BoolVarP(&isSuccessful, "successful", "", false, `Get only status 200 response items.`)
	rootCmd.PersistentFlags().IntVarP(&maxTimeout, "timeout", "t", 30, `Max timeout of requests.`)
	rootCmd.PersistentFlags().IntVarP(&maxRetries, "retries", "r", 3, `Max request retries."`)
	rootCmd.PersistentFlags().UintVarP(&maxResults, "limit", "l", 0, `Max number of results to fetch."`)
	rootCmd.PersistentFlags().UintVarP(&maxWorkers, "workers", "w", 4, `Max number of workers (threads) to use. URL consumes 1 worker"`)
	rootCmd.PersistentFlags().StringSliceVarP(&extensions, "ext", "e", []string{}, `Which extensions to collect. Example: --ext "pdf,xml,jpeg"`)
	rootCmd.PersistentFlags().StringSliceVarP(&sourceNames, "sources", "s", []string{"wb", "cc"}, `Web archive sources to use. Example: --sources "wb" to use only the Wayback`)
	rootCmd.PersistentFlags().BoolVarP(&isVerbose, "verbose", "v", false, `Use verbose output.`)
	rootCmd.PersistentFlags().BoolVarP(&isLogging, "log", "", false, `Print logs to ./logs.txt.`)
}
