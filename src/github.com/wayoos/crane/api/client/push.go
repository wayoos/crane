package client

import (
	"bytes"
	"fmt"
	"github.com/codegangsta/cli"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"wayoos.com/compress"
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

	for key, val := range headers {
		request.Header.Add(key, val)
	}

	return request, err
}

func PushCommand(c *cli.Context) {

	if c.Args().Present() {
		host := c.GlobalString("host")

		path := c.Args().First()

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

		tag := c.String("tag")

		// send the file over http
		headers := map[string]string{
			"load-tag": tag,
		}
		request, err := newfileUploadRequest(host+"/up", headers, "file", loadCompressedFilePath)
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
	} else {
		cli.ShowCommandHelp(c, "push")
	}

}
