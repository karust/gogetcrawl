package commoncrawl

import (
	"log"
	"net/http"
	"strconv"
	"strings"
)

func extensionByContent(content []byte) string {
	contentType := http.DetectContentType([]byte(content))
	splitted := strings.Split(contentType, "; ")[0]
	extenstions := map[string]string{"text/xml": ".xml", "text/html": ".html",
		"text/plain": ".txt"}
	if extenstions[splitted] == "" {
		log.Println("[extensionByContent] No extension for " + splitted)
		return ".none"
	}
	return extenstions[splitted]
}

func escapeURL(url string) string {
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
