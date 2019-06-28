package commoncrawl

import (
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Safari/537.36",
	"Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/604.1.38 (KHTML, like Gecko) Version/11.0 Safari/604.1.38",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:56.0) Gecko/20100101 Firefox/56.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13) AppleWebKit/604.1.38 (KHTML, like Gecko) Version/11.0 Safari/604.1.38",
}

// ExtensionByContent ... Returns extension of file by detecting its MIME type, `.none` returned if no MIME found
func ExtensionByContent(content []byte) string {
	contentType := http.DetectContentType([]byte(content))
	splitted := strings.Split(contentType, "; ")[0]
	extenstions := map[string]string{"text/xml": ".xml", "text/html": ".html", "image/jpeg": ".jpeg", "application/pdf": ".pdf", "text/plain": ".txt"}
	if extenstions[splitted] == "" {
		log.Println("[extensionByContent] No extension for " + splitted)
		return ".none"
	}
	return extenstions[splitted]
}

// EscapeURL ... Makes possible to create files with URL name by replacing system chars with %<char code>
func EscapeURL(url string) string {
	badChars := []string{"/", "\\", ":", "?"}
	escapedURL := ""

	for _, urlChar := range url {
		for i, badChar := range badChars {
			if string(urlChar) == badChar {
				conv := strconv.Itoa(int(urlChar))
				escapedURL += "%" + conv
				break
			} else if i == len(badChars)-1 {
				escapedURL += string(urlChar)
			}
		}
	}
	return escapedURL
}

func randomOption(options []string) string {
	rand.Seed(time.Now().Unix())
	randNum := rand.Int() % len(options)
	return options[randNum]
}
