package client

import (
	"bytes"
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/wayoos/crane/api/domain"
	"github.com/wayoos/crane/compress"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Creates a new file upload http request with optional extra params
func newfileUploadRequest(uri string, headers map[string]string, paramName, path string) (*http.Request, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile(paramName, filepath.Base(path))
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(part, file)

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest("POST", uri, body)

	request.Header.Set("Content-Type", writer.FormDataContentType())

	if headers != nil {
		for key, val := range headers {
			request.Header.Add(key, val)
		}
	}

	return request, err
}

func BuildCommand(c *cli.Context) {

	if c.Args().Present() {
		host := c.GlobalString("host")

		path := c.Args().First()

		var tag string
		if c.IsSet("tag") {
			tag = c.String("tag")
		}

		dockloadId, appErr := BuildSend("build", host, path, tag)

		if appErr != nil {
			log.Println("Error: failed to create images")
			os.Exit(1)
		}

		fmt.Println(dockloadId)

	} else {
		cli.ShowCommandHelp(c, "build")
	}

}

func BuildSend(command, host, path, tag string) (dockloadId string, appErr *domain.AppError) {
	loadPath, err := filepath.Abs(path)
	if err != nil {
		return "", &domain.AppError{err, "Invalid path " + path, 500}
	}

	loadPathFileInfo, err := os.Stat(loadPath)
	if err != nil {
		return "", &domain.AppError{err, "Invalid path info " + loadPath, 500}
	}

	if !loadPathFileInfo.IsDir() {
		return "", &domain.AppError{nil, "Invalid path is not a directory " + loadPath, 500}
	}

	loadCompressedFile, err := ioutil.TempFile("", "crane")
	if err != nil {
		return "", &domain.AppError{err, "Create temporary file error", 500}
	}
	loadCompressedFilePath := loadCompressedFile.Name()
	defer func() {
		loadCompressedFile.Close()
		err = os.Remove(loadCompressedFilePath)
		if err != nil {
			log.Fatal(err)
		}
	}()

	compress.ZipFolder(loadPath, loadCompressedFilePath)

	urlPath := "/"
	urlPath += command

	if tag != "" {
		tagSplit := strings.Split(tag, ":")
		if len(tagSplit) > 1 {
			urlPath += "/" + tagSplit[0] + "/" + tagSplit[1]
		} else {
			urlPath += "/" + tag
		}
	}

	// send the file over http
	request, err := newfileUploadRequest(host+urlPath, nil, "file", loadCompressedFilePath)
	if err != nil {
		return "", &domain.AppError{err, "Send file error", 500}
	}

	client := &http.Client{}

	resp, err := client.Do(request)
	if err != nil {
		return "", &domain.AppError{err, "Send file error", 500}
	}

	body := &bytes.Buffer{}
	_, err = body.ReadFrom(resp.Body)
	if err != nil {
		return "", &domain.AppError{err, "Send file error", 500}
	}
	resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Println("Error response from crane daemon: " + body.String())
		log.Println("Error: failed to create and start container")
		os.Exit(1)
	}

	//TODO check status code
	//fmt.Println(resp.StatusCode)
	//fmt.Println(resp.Header)
	//	fmt.Println(body)

	return body.String(), nil
}
