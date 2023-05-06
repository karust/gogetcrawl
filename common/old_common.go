package common

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
