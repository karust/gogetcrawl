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
	"Mozilla/5.0 (Windows; U; Windows NT 6.0; en-US) AppleWebKit/525.13 (KHTML, like Gecko) Chrome/0.2.149.27 Safari/525.13",
	"Mozilla/5.0 (Windows; U; Windows NT 6.0; de) AppleWebKit/525.13 (KHTML, like Gecko) Chrome/0.2.149.27 Safari/525.13",
	"Mozilla/5.0 (Windows; U; Windows NT 5.2; en-US) AppleWebKit/525.13 (KHTML, like Gecko) Chrome/0.2.149.27 Safari/525.13",
	"Mozilla/5.0 (Windows; U; Windows NT 5.1; pt-PT; rv:1.9.2.7) Gecko/20100713 Firefox/3.6.7 (.NET CLR 3.5.30729)",
	"Mozilla/5.0 (X11; U; Linux x86_64; en-US; rv:1.9.2.6) Gecko/20100628 Ubuntu/10.04 (lucid) Firefox/3.6.6 GTB7.1",
	"Mozilla/5.0 (X11; Mageia; Linux x86_64; rv:10.0.9) Gecko/20100101 Firefox/10.0.9",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10.6; rv:9.0a2) Gecko/20111101 Firefox/9.0a2",
	"Mozilla/5.0 (Windows NT 6.2; rv:9.0.1) Gecko/20100101 Firefox/9.0.1",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/42.0.2311.135 Safari/537.36 Edge/12.246",
}

// ExtensionByContent ... Returns extension of file by detecting its MIME type, `.none` returned if no MIME found
func ExtensionByContent(content []byte) string {
	contentType := http.DetectContentType([]byte(content))
	splitted := strings.Split(contentType, "; ")[0]
	extenstions := map[string]string{"text/xml": ".xml", "text/html": ".html", "image/jpeg": ".jpeg", "application/pdf": ".pdf", "text/php": ".php", "image/png": ".png",
		"text/plain": ".txt", "application/zip": ".zip", "application/javascript": ".js", "application/json": ".json", "text/css": ".css", "application/gzip": ".gzip"}
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

// IsExtensionExist ...Return `true` if there is extension in allowed or false if there is no
// returns `true` if length of `allowedExtensions` == 0
func IsExtensionExist(allowedExtensions []string, extension string) bool {
	if len(allowedExtensions) == 0 {
		return true
	}
	for _, ext := range allowedExtensions {
		if ext == extension {
			return true
		}
	}
	return false
}

/*
// MD5FileHash ... Gets MD5 hash of file
func MD5FileHash(filePath string) (string, error) {
	//Initialize variable returnMD5String now in case an error has to be returned
	var returnMD5String string

	//Open the passed argument and check for any error
	file, err := os.Open(filePath)
	if err != nil {
		return returnMD5String, err
	}

	//Tell the program to call the following function when the current function returns
	defer file.Close()

	//Open a new hash interface to write to
	hash := md5.New()

	//Copy the file in the hash interface and check for any error
	if _, err := io.Copy(hash, file); err != nil {
		return returnMD5String, err
	}

	//Get the 16 bytes hash
	hashInBytes := hash.Sum(nil)[:16]

	//Convert the bytes to a string
	returnMD5String = hex.EncodeToString(hashInBytes)

	return returnMD5String, nil
}
*/
