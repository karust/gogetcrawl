package commoncrawl

import (
	"fmt"
	"testing"
)

func TestGetIndexIDs(t *testing.T) {
	index_ids, err := GetIndexIDs()
	if err != nil {
		t.Fatal(err)
	}

	if len(index_ids) == 0 {
		t.Fatal("No index IDs found")
	}
}

func TestGetPages(t *testing.T) {
	pages, err := GetPages("CC-MAIN-2023-14", "example.com/*", "", 100, 30, true)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(pages)
}
