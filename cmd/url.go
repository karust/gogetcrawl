package cmd

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"

	"github.com/karust/gogetcrawl/common"
	"github.com/spf13/cobra"
)

type urlScenario struct {
	outputFile      string
	finishedWorkers uint
}

var urlScn = urlScenario{}

var urlCMD = &cobra.Command{
	Use:     "url",
	Aliases: []string{"collect"},
	Short:   "Collect URLs from web archives for desired domain",
	Args:    cobra.MatchAll(cobra.MinimumNArgs(1), cobra.OnlyValidArgs),
	Run:     urlScn.spawnWorkers,
}

func (us *urlScenario) worker(configs chan common.RequestConfig) {
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
				}
				wg.Wait()
			} else {
				us.finishedWorkers += 1
				return
			}
		}
	}
}

func (us *urlScenario) spawnWorkers(cmd *cobra.Command, args []string) {
	output, err := us.getOutputTarget()
	if err != nil {
		log.Fatalf("Error obtaining output: %v", err)
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
				us.worker(configs)
			}()
		}
	}()

	close(configs)

	// Read Results and errors
	for us.finishedWorkers != maxWorkers {
		select {
		case res, ok := <-results:
			if ok {
				fmt.Fprintf(output, "%v", us.formatResultOutput(res))
			}
		case err, ok := <-errors:
			if ok {
				log.Println(err)
			}
		}
	}

	wg.Wait()
	close(results)
	close(errors)
}

func (us *urlScenario) getOutputTarget() (io.Writer, error) {
	if us.outputFile != "" {
		file, err := os.OpenFile(us.outputFile, os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			return nil, err
		}
		return file, nil
	}

	return os.Stdout, nil
}

func (us *urlScenario) formatResultOutput(results []*common.CdxResponse) string {
	output := ""
	for _, r := range results {
		output += r.Original + "\n"
	}
	return output
}

func init() {
	urlCMD.Flags().StringVarP(&urlScn.outputFile, "output", "o", "", "Path to the output file")
	rootCmd.AddCommand(urlCMD)
}
