package commoncrawl

import (
	"log"
	"net/http"
	"strconv"
	"strings"
)

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
