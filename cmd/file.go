package cmd

import (
	"fmt"
	"log"
	"mime"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/karust/goGetCrawl/common"
	"github.com/spf13/cobra"
)

type fileScenario struct {
	finishedWorkers uint
	outputDir       string
	downloadRate    uint
}

var fileScn = fileScenario{}

var fileCMD = &cobra.Command{
	Use:     "file",
	Aliases: []string{"download"},
	Short:   "Download files located in web arhives for desired domains",
	Run:     fileScn.spawnWorkers,
}

func (fs *fileScenario) worker(configs chan common.RequestConfig) {
	for {
		select {
		case config, ok := <-configs:
			if ok {
				var wg sync.WaitGroup
				for _, s := range sources {
					wg.Add(1)
					go func(s common.Source) {
						defer wg.Done()
						pages, err := s.GetPages(config)
						if err != nil {
							errors <- err
							return
						}
						for _, p := range pages {
							fs.saveFile(s, p)
							time.Sleep(time.Second * time.Duration(fs.downloadRate))
						}
					}(s)
				}
				wg.Wait()
			} else {
				fs.finishedWorkers += 1
				return
			}
		}
	}
}

func (fs *fileScenario) spawnWorkers(cmd *cobra.Command, args []string) {
	err := os.MkdirAll(fs.outputDir, os.ModePerm)
	if err != nil {
		log.Fatalf("Cannot get access to '%v' dir: %v", fileScn.outputDir, err)
	}

	initSources()
	configs := getRequestConfigs(args)

	var wg sync.WaitGroup

	// Spawn Workers
	go func() {
		for i := uint(0); i < maxWorkers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				fs.worker(configs)
			}()
		}
	}()

	close(configs)

	// Read Results and errors
	for fs.finishedWorkers != maxWorkers {
		select {
		case err, ok := <-errors:
			if ok {
				log.Printf("ERROR: %v\n", err)
			}
		// Unblock if no errors produced
		case <-time.After(time.Second * 3):
			continue
		}
	}

	wg.Wait()
	close(errors)
	close(results)
}

func (fs *fileScenario) saveFile(source common.Source, page *common.CdxResponse) {
	file, err := source.GetFile(page)
	if err != nil {
		errors <- err
		return
	}

	exts, _ := mime.ExtensionsByType(page.MimeType)
	if exts == nil {
		exts = []string{""}
	}

	filename := fmt.Sprintf("%v-%v-%v%v", page.Original, page.Timestamp, source.Name(), exts[0])
	escapedFilename := url.QueryEscape(filename)
	fullPath := filepath.Join(fs.outputDir, escapedFilename)

	err = common.SaveFile(file, fullPath)
	if err != nil {
		errors <- err
	}
}

func init() {
	fileCMD.Flags().StringVarP(&fileScn.outputDir, "dir", "d", "", "Path to the output directory")
	fileCMD.Flags().UintVarP(&fileScn.downloadRate, "rate", "", 5, "Download rate in seconds for each worker (thread)")
	rootCmd.AddCommand(fileCMD)
	fileCMD.MarkFlagRequired("dir")
}
