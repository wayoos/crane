package client

import (
	"bytes"
	"fmt"
	"github.com/codegangsta/cli"
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

		BuildSend("build", host, path, tag)
	} else {
		cli.ShowCommandHelp(c, "build")
	}

}

func BuildSend(command, host, path, tag string) {
	loadPath, err := filepath.Abs(path)
	if err != nil {
		log.Fatal(err)
	}

	loadPathFileInfo, err := os.Stat(loadPath)
	if err != nil {
		log.Fatal(err)
	}

	if !loadPathFileInfo.IsDir() {
		log.Fatal("Path is not a directory")
	}

	loadCompressedFile, err := ioutil.TempFile("", "crane")
	if err != nil {
		log.Fatal(err)
	}
	loadCompressedFilePath := loadCompressedFile.Name()
	defer func() {
		loadCompressedFile.Close()
		err = os.Remove(loadCompressedFilePath)
		if err != nil {
			log.Fatal(err)
		}
	}()

	//					compress.TarGz(loadPath, loadCompressedFile)

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
		log.Fatal(err)
	}
	client := &http.Client{}

	resp, err := client.Do(request)
	if err != nil {
		log.Fatal(err)
	} else {
		body := &bytes.Buffer{}
		_, err := body.ReadFrom(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		resp.Body.Close()
		//TODO check status code
		//fmt.Println(resp.StatusCode)
		//fmt.Println(resp.Header)
		fmt.Println(body)
	}
}