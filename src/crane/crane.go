package main

import (
	"bytes"
	"container/list"
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/jmcvetta/napping"
	"github.com/wayoos/crane/api/client"
	"github.com/wayoos/crane/api/domain"
	"github.com/wayoos/crane/api/server"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
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

func main() {
	app := cli.NewApp()
	app.Name = "crane"
	app.Usage = "crane [command]"
	app.Version = "0.0.1"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "host, H",
			Value: "http://localhost:2475",
			Usage: "Host",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:  "up",
			Usage: "Create and start container",
			//			Description: "Create and start container",
			Action: client.UpCommand,
		},
		{
			Name:        "rm",
			Usage:       "crane rm",
			Description: "Stop and remove container",
			Action:      client.RmCommand,
		},
		{
			Name:        "push",
			ShortName:   "p",
			Usage:       "crane push PATH",
			Description: "push an image package or a crane package to the crane server",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "tag, t",
					Value: "",
					Usage: "Load name (and optionally a tag) to be applied to the resulting",
				},
			},
			Action: func(c *cli.Context) {

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

			},
		},
		{
			Name:      "server",
			ShortName: "s",
			Usage:     "start crane server",
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "p, port",
					Value: 2475,
					Usage: "port to listen on (default 2475)",
				},
				cli.StringFlag{
					Name:  "c, crane-dir",
					Value: "current directory",
					Usage: "Path to use as the crane data directory",
				},
			},
			Action: server.ServerCommand,
		},
		{
			Name:      "exec",
			ShortName: "e",
			Usage:     "crane exec LOADID command...",
			Action: func(c *cli.Context) {
				loadId := c.Args().First()

				//fmt.Println("Execute cmd in " + loadId)

				host := c.GlobalString("host")

				var cmds []string = c.Args().Tail()

				l := list.New()

				for i := range cmds {
					//					println(cmds[i])

					val := cmds[i]
					split := strings.Split(val, " ")
					for si := range split {
						l.PushBack(split[si])
					}
				}

				cmds = make([]string, l.Len())

				idx := 0
				for e := l.Front(); e != nil; e = e.Next() {
					cmd := e.Value.(string)
					cmds[idx] = cmd
					idx++
				}

				//				for i := range cmds {
				//					println(cmds[i])
				//				}

				execData := domain.ExecData{
					LoadId: loadId,
					Cmd:    cmds,
				}

				result := domain.ExecResult{}
				resp, err := napping.Post(host+"/exec", &execData, &result, nil)
				if err != nil {
					panic(err)
				}
				if resp.Status() == 200 {
					fmt.Println(result.Out)
				}

			},
		},
		{
			Name:  "ps",
			Usage: "crane ps",
			Action: func(c *cli.Context) {

				host := c.GlobalString("host")
				result := []domain.LoadData{}
				resp, err := napping.Get(host+"/ps", nil, &result, nil)
				if err != nil {
					panic(err)
				}
				if resp.Status() == 200 {
					w := new(tabwriter.Writer)

					// Format in tab-separated columns with a tab stop of 8.
					w.Init(os.Stdout, 0, 8, 0, '\t', 0)
					fmt.Fprintln(w, "NAME\tTAG\tLOAD ID")
					for _, loadData := range result {
						fmt.Fprintln(w, loadData.Name+"\t"+loadData.Tag+"\t"+loadData.ID)
					}
					w.Flush()

				}
			},
		},
	}

	app.Action = func(c *cli.Context) {

		if len(c.Args()) <= 0 {
			cli.ShowAppHelp(c)
		} else {

			println("Invalid command")
		}

	}

	app.Run(os.Args)
}
