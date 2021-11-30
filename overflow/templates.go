package overflow

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"os"
)

func splitByWidthMake(str string, size int) []string {
	strLength := len(str)
	splitedLength := int(math.Ceil(float64(strLength) / float64(size)))
	splited := make([]string, splitedLength)
	var start, stop int
	for i := 0; i < splitedLength; i += 1 {
		start = i * size
		stop = start + size
		if stop > strLength {
			stop = strLength
		}
		splited[i] = str[start:stop]
	}
	return splited
}

func fileAsImageData(path string) (string, error) {
	f, _ := os.Open(path)

	defer f.Close()

	// Read entire JPG into byte slice.
	reader := bufio.NewReader(f)
	content, err := ioutil.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("could not read imageFile %s, %w", path, err)
	}

	return contentAsImageDataUrl(content), nil
}

func contentAsImageDataUrl(content []byte) string {
	contentType := http.DetectContentType(content)

	// Encode as base64.
	encoded := base64.StdEncoding.EncodeToString(content)

	return "data:" + contentType + ";base64, " + encoded
}

func fileAsBase64(path string) (string, error) {
	f, _ := os.Open(path)

	defer f.Close()

	// Read entire JPG into byte slice.
	reader := bufio.NewReader(f)
	content, err := ioutil.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("could not read file %s, %w", path, err)
	}

	// Encode as base64.
	encoded := base64.StdEncoding.EncodeToString(content)

	return encoded, nil
}

//UploadFile reads a file, base64 encodes it and chunk upload to /storage/upload
func (f *GoWithTheFlow) UploadFile(filename string, accountName string) error {
	content, err := fileAsBase64(filename)
	if err != nil {
		return err
	}

	return f.UploadString(content, accountName)
}

func getUrl(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}

//DownloadAndUploadFile reads a file, base64 encodes it and chunk upload to /storage/upload
func (f *GoWithTheFlow) DownloadAndUploadFile(url string, accountName string) error {
	body, err := getUrl(url)
	if err != nil {
		return err
	}

	encoded := base64.StdEncoding.EncodeToString(body)
	return f.UploadString(encoded, accountName)
}

//DownloadImageAndUploadAsDataUrl download an image and upload as data url
func (f *GoWithTheFlow) DownloadImageAndUploadAsDataUrl(url, accountName string) error {
	body, err := getUrl(url)
	if err != nil {
		return err
	}
	content := contentAsImageDataUrl(body)

	return f.UploadString(content, accountName)
}

//UploadImageAsDataUrl will upload a image file from the filesystem into /storage/upload of the given account
func (f *GoWithTheFlow) UploadImageAsDataUrl(filename string, accountName string) error {
	content, err := fileAsImageData(filename)
	if err != nil {
		return err
	}

	return f.UploadString(content, accountName)
}

//UploadString will upload the given string data in 1mb chunkts to /storage/upload of the given account
func (f *GoWithTheFlow) UploadString(content string, accountName string) error {
	//unload previous content if any.
	if _, err := f.Transaction(`
	transaction {
		prepare(signer: AuthAccount) {
			let path = /storage/upload
			let existing = signer.load<String>(from: path) ?? ""
			log(existing)
		}
	}
	  `).SignProposeAndPayAs(accountName).RunE(); err != nil {
		return err
	}

	parts := splitByWidthMake(content, 1_000_000)
	for _, part := range parts {
		if _, err := f.Transaction(`
		transaction(part: String) {
			prepare(signer: AuthAccount) {
				let path = /storage/upload
				let existing = signer.load<String>(from: path) ?? ""
				signer.save(existing.concat(part), to: path)
				log(signer.address.toString())
				log(part)
			}
		}
			`).SignProposeAndPayAs(accountName).StringArgument(part).RunE(); err != nil {
			return err
		}
	}

	return nil
}
