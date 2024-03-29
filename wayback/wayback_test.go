package wayback

import (
	"testing"
	"time"

	common "github.com/karust/gogetcrawl/common"
)

// Example request: https://web.archive.org/cdx/search/cdx?url=kamaloff.ru/*&output=json&limit=100&collapse=urlkey
const RESPONSE = `[["urlkey","timestamp","original","mimetype","statuscode","digest","length"],
["ru,kamaloff)/", "20130522121421", "http://kamaloff.ru/", "text/html", "200", "FXOQP7LM7FWUC7S5MTDHZS2WMKNLCW2E", "2558"],
["ru,kamaloff)/favicon.ico", "20180104074528", "http://kamaloff.ru/favicon.ico", "text/html", "301", "WHT3EXKF6XVIKYVG67BXESI75TWESKWU", "463"],
["ru,kamaloff)/login?next=/", "20180104100356", "https://kamaloff.ru/login/?next=/", "text/html", "200", "7AK62UMEB5LCDYN5JSLXZ7ZZG5Z7XNCQ", "1998"],
["ru,kamaloff)/robots.txt", "20130801111119", "http://kamaloff.ru/robots.txt", "text/html", "404", "6443ZIMC2V4HX7MZ4YUEY2VI3OEI36HM", "351"]]`

// Test interface
var wb common.Source = &Wayback{}

func init() {
	wb, _ = New(15, 2)
}

func TestParseResponse(t *testing.T) {
	want := "http://kamaloff.ru/"

	parsedResp, err := wb.ParseResponse([]byte(RESPONSE))
	if err != nil {
		t.Fatalf("%v", err)
	}

	got := parsedResp[0].Original
	if got != want {
		t.Fatalf("parsed result doesn't contain wanted value: Want=%v, Got=%v", want, got)
	}
}

func TestGetNumPages(t *testing.T) {
	want := 1

	parsedResp, err := wb.GetNumPages("kamaloff.ru")
	if err != nil {
		t.Fatalf("%v", err)
	}

	got := parsedResp
	if got != want {
		t.Fatalf("Parsed result doesn't contain wanted value: Want=%v, Got=%v", want, got)
	}
}

func TestGetPages(t *testing.T) {
	config := common.RequestConfig{
		URL:     "*.kamaloff.ru/*",
		Filters: []string{"statuscode:200"},
		Limit:   10,
	}
	results, err := wb.GetPages(config)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if len(results) == 0 {
		t.Fatalf("No pages fetched from Web Archive")
	}

	if len(results) != 10 {
		t.Fatalf("Incorrect number of pages returned: %v, want=10", len(results))
	}
}

func TestFetchPages(t *testing.T) {
	config1 := common.RequestConfig{
		URL:        "tutorialspoint.com/*",
		Filters:    []string{"statuscode:200", "mimetype:text/html"},
		Limit:      6,
		SinglePage: true,
	}

	config2 := common.RequestConfig{
		URL:        "example.com/*",
		Filters:    []string{"statuscode:200", "mimetype:text/html"},
		Limit:      6,
		SinglePage: true,
	}

	resultsChan := make(chan []*common.CdxResponse)
	errorsChan := make(chan error)

	go func() {
		wb.FetchPages(config1, resultsChan, errorsChan)
	}()

	go func() {
		wb.FetchPages(config2, resultsChan, errorsChan)
	}()

	var results []*common.CdxResponse
	timeout := false

	for {
		select {
		case err := <-errorsChan:
			t.Fatalf("FetchPages goroutine failed %v", err)
		case res, ok := <-resultsChan:
			if len(res) > 0 && res[0].StatusCode != "200" {
				t.Fatalf("Incorrect response")
			}
			if ok {
				results = append(results, res...)
			}
		case <-time.After(time.Second * 10):
			timeout = true
			break
		}

		if timeout {
			t.Log("Timeout passed")
			break
		}
	}

	t.Logf("Got %v results", len(results))

	if len(results) != 12 {
		t.Fatalf("Got less that 12 results")
	}
}

func TestGetFile(t *testing.T) {
	config := common.RequestConfig{
		URL:     "kamaloff.ru/*",
		Filters: []string{"statuscode:200", "mimetype:text/html"},
		Limit:   5,
	}
	results, err := wb.GetPages(config)
	if err != nil {
		t.Fatalf("%v", err)
	}

	file, err := wb.GetFile(results[0])
	if err != nil {
		t.Fatalf("Cannot get file: %v", err)
	}

	t.Logf("Obtained file length: %v", len(file))
	if len(file) != 11011 {
		t.Fatalf("Got incorrect length file")
	}
}
