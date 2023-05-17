package cmd

import (
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/karust/gogetcrawl/common"
	"github.com/spf13/cobra"
)

type fileScenario struct {
	finishedWorkers uint
	outputDir       string
	downloadRate    float32
}

var fileScn = fileScenario{}

var fileCMD = &cobra.Command{
	Use:     "file",
	Aliases: []string{"download"},
	Short:   "Download files located in web arhives for desired domains",
	Args:    cobra.MatchAll(cobra.MinimumNArgs(1), cobra.OnlyValidArgs),
	Run:     fileScn.spawnWorkers,
}

func (fs *fileScenario) worker(configs <-chan common.RequestConfig) {
	for {
		select {
		case config, ok := <-configs:
			if ok {
				var wg sync.WaitGroup
				for _, s := range sources {

					wg.Add(1)
					go func(s common.Source) {
						defer wg.Done()
						s.FetchPages(config, results, errors)
					}(s)

					//wg.Add(1)
					go func() {
						//defer wg.Done()
						common.SaveFiles(results, fs.outputDir, errors, fs.downloadRate)
					}()
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
	fp, _ := filepath.Abs(fs.outputDir)
	err := os.MkdirAll(fp, os.ModePerm)
	if err != nil {
		log.Fatalf("Cannot get access to '%v' dir: %v", fileScn.outputDir, err)
	} else {
		log.Printf("Setting '%v' as output directorty", fp)
	}

	configs := getRequestConfigs(args)
	initSources()

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

func init() {
	fileCMD.Flags().StringVarP(&fileScn.outputDir, "dir", "d", "", "Path to the output directory")
	fileCMD.Flags().Float32VarP(&fileScn.downloadRate, "rate", "", 1.0, "Download rate in seconds for each worker (thread). Ex: 5, 1.5")
	rootCmd.AddCommand(fileCMD)
	fileCMD.MarkFlagRequired("dir")
}
