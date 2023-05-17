package commoncrawl

import (
	"testing"
	"time"

	common "github.com/karust/gogetcrawl/common"
)

// ! Currently impossinble to run tests all at once, due to the index server timeouts

// Use pre-saved response, hard to get live one
// Example request: http://index.commoncrawl.org/CC-MAIN-2023-14-index?url=tutorialspoint.com/*&output=json&limit=6&filter=statuscode:200&filter=mimetype:application/pdf
const RESPONSE = `{"urlkey": "com,tutorialspoint)/accounting_basics/accounting_basics_tutorial.pdf", "timestamp": "20230320100841", "url": "http://www.tutorialspoint.com/accounting_basics/accounting_basics_tutorial.pdf", "mime": "application/pdf", "mime-detected": "application/pdf", "status": "200", "digest": "2JQ2AQ3HQZIMXHB5CJGSADUGOHYBIRJJ", "length": "787172", "offset": "102849414", "filename": "crawl-data/CC-MAIN-2023-14/segments/1679296943471.24/warc/CC-MAIN-20230320083513-20230320113513-00267.warc.gz"}
{"urlkey": "com,tutorialspoint)/add_and_subtract_whole_numbers/pdf/subtracting_of_two_2digit_numbers_with_borrowing_worksheet10_1.pdf", "timestamp": "20230326185123", "url": "https://www.tutorialspoint.com/add_and_subtract_whole_numbers/pdf/subtracting_of_two_2digit_numbers_with_borrowing_worksheet10_1.pdf", "mime": "application/pdf", "mime-detected": "application/pdf", "status": "200", "digest": "T4OQARBGDQ2Z3ZMJ57MWZTUIBCFR65QG", "length": "120114", "offset": "1156945883", "filename": "crawl-data/CC-MAIN-2023-14/segments/1679296946445.46/warc/CC-MAIN-20230326173112-20230326203112-00412.warc.gz"}
{"urlkey": "com,tutorialspoint)/add_and_subtract_whole_numbers/pdf/subtracting_of_two_2digit_numbers_with_borrowing_worksheet10_2.pdf", "timestamp": "20230322123716", "url": "https://www.tutorialspoint.com/add_and_subtract_whole_numbers/pdf/subtracting_of_two_2digit_numbers_with_borrowing_worksheet10_2.pdf", "mime": "application/pdf", "mime-detected": "application/pdf", "status": "200", "digest": "EJJMOG5QPWIV7YXADIFOPML45UTJKYWW", "length": "118702", "offset": "1159004265", "filename": "crawl-data/CC-MAIN-2023-14/segments/1679296943809.76/warc/CC-MAIN-20230322114226-20230322144226-00733.warc.gz"}
{"urlkey": "com,tutorialspoint)/add_and_subtract_whole_numbers/pdf/subtracting_of_two_2digit_numbers_with_borrowing_worksheet10_3.pdf", "timestamp": "20230324124641", "url": "https://www.tutorialspoint.com/add_and_subtract_whole_numbers/pdf/subtracting_of_two_2digit_numbers_with_borrowing_worksheet10_3.pdf", "mime": "application/pdf", "mime-detected": "application/pdf", "status": "200", "digest": "AOTDOZIAULAYGY3AOMD7662BJBEPYKWJ", "length": "210009", "offset": "1172608792", "filename": "crawl-data/CC-MAIN-2023-14/segments/1679296945282.33/warc/CC-MAIN-20230324113500-20230324143500-00254.warc.gz"}
{"urlkey": "com,tutorialspoint)/adding_and_subtracting_decimals/pdf/addition_with_money_worksheet8_1.pdf", "timestamp": "20230330141211", "url": "https://www.tutorialspoint.com/adding_and_subtracting_decimals/pdf/addition_with_money_worksheet8_1.pdf", "mime": "application/pdf", "mime-detected": "application/pdf", "status": "200", "digest": "MOODQKFMHRVSZK4UOZO3E6H2MGHTK2VW", "length": "226484", "offset": "1136155166", "filename": "crawl-data/CC-MAIN-2023-14/segments/1679296949331.26/warc/CC-MAIN-20230330132508-20230330162508-00514.warc.gz"}
{"urlkey": "com,tutorialspoint)/adding_and_subtracting_decimals/pdf/addition_with_money_worksheet8_2.pdf", "timestamp": "20230330112743", "url": "https://www.tutorialspoint.com/adding_and_subtracting_decimals/pdf/addition_with_money_worksheet8_2.pdf", "mime": "application/pdf", "mime-detected": "application/pdf", "status": "200", "digest": "ZYCDOJ2JTPPWFTCNYEIXCWKEJQXTA7UD", "length": "226957", "offset": "1167440233", "filename": "crawl-data/CC-MAIN-2023-14/segments/1679296949181.44/warc/CC-MAIN-20230330101355-20230330131355-00035.warc.gz"}
`

// Test interface
var cctest common.Source = &CommonCrawl{}
var cc *CommonCrawl

func init() {
	cc, _ = New(15, 2)
}

func TestGetIndexes(t *testing.T) {
	index_ids, err := cc.GetIndexes()
	if err != nil {
		t.Fatal(err)
	}

	if len(index_ids) == 0 {
		t.Fatal("No index IDs found")
	}
}

// Commented due to excessive request to test
// func TestGetNumPagesIndex(t *testing.T) {
// 	want := 989

// 	got, err := GetNumPagesIndex("*.wikipedia.org/", "CC-MAIN-2015-11")
// 	if err != nil {
// 		t.Fatalf("%v", err)
// 	}
// 	if got != want {
// 		t.Fatalf("Parsed result doesn't contain wanted value: Want=%v, Got=%v", want, got)
// 	}
// }

func TestGetNumPages(t *testing.T) {
	got, err := cc.GetNumPages("*.wikipedia.org/")
	if err != nil {
		t.Fatalf("%v", err)
	}
	if got == 0 {
		t.Fatalf("Parsed result doesn't contain wanted value: Got=%v", got)
	}
}

func TestParseResponse(t *testing.T) {
	want := "http://www.tutorialspoint.com/accounting_basics/accounting_basics_tutorial.pdf"
	parsedResp, err := cc.ParseResponse([]byte(RESPONSE))
	if err != nil {
		t.Fatalf("%v", err)
	}

	got := parsedResp[0].Original
	if got != want {
		t.Fatalf("parsed result doesn't contain wanted value: Want=%v, Got=%v", want, got)
	}
}

// Commented due to excessive request to test
// func TestGetPagesIndex(t *testing.T) {
// 	config := common.RequestConfig{
// 		URL: "wikipedia.org/",
// 		//Filters:    []string{"statuscode:200", "mimetype:text/html"},
// 		Limit: 10,
// 		//Collapse:   true,
// 		Timeout:    60,
// 		MaxRetries: 2,
// 	}
// 	results, err := GetPagesIndex(config, "CC-MAIN-2015-11")
// 	if err != nil {
// 		t.Fatalf("%v", err)
// 	}
// 	fmt.Println(results)

// 	if len(results) == 0 {
// 		t.Fatalf("No pages fetched from Web Archive")
// 	}
// }

func TestGetPages(t *testing.T) {
	config := common.RequestConfig{
		URL:     "wikipedia.org/",
		Filters: []string{"statuscode:200", "mimetype:text/html"},
		Limit:   10,
	}
	results, err := cc.GetPages(config)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if len(results) == 0 {
		t.Fatalf("No pages fetched from Web Archive")
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
		cc.FetchPages(config1, resultsChan, errorsChan)
	}()

	go func() {
		cc.FetchPages(config2, resultsChan, errorsChan)
	}()

	var results []*common.CdxResponse
	timeout := false

	for {
		select {
		case err := <-errorsChan:
			t.Fatalf("FetchPages goroutine failed %v", err)
		case res, ok := <-resultsChan:
			if ok {
				results = append(results, res...)
			}
		case <-time.After(time.Second * 100):
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
	pages, err := cc.ParseResponse([]byte(RESPONSE))
	if err != nil {
		t.Fatalf("Cannot parse response: %v", err)
	}

	file, err := cc.GetFile(pages[4])
	if err != nil {
		t.Fatalf("Cannot get file: %v", err)
	}

	if len(file) == 0 {
		t.Fatalf("Obtained file length: %v", len(file))
	}
	t.Logf("Obtained file length: %v", len(file))
}
