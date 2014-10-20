package main

import (
	"os"
	"fmt"
	"bytes"
	"io"
	"log"
	"strconv"
	"net/http"
	"mime/multipart"
	"path/filepath"
	"crypto/rand"
	"encoding/hex"
	"github.com/codegangsta/cli"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
//	"wayoos.com/config"
	"wayoos.com/compress"
	"wayoos.com/config"
	"io/ioutil"
)


func startServer(port int) {

	config.InitDataPath()

	fmt.Printf("dataPath=%s", config.DataPath)
	fmt.Println()


	os.Setenv("PORT", strconv.Itoa(port))
	os.Setenv("HOST", "localhost")


	os.Setenv("MARTINI_ENV", martini.Prod)

	martini.Env = martini.Prod

	m := martini.Classic()
	m.Use(render.Renderer(render.Options{
	Directory: "templates", // Specify what path to load the templates from.
	Layout: "layout", // Specify a layout template. Layouts can call {{ yield }} to render the current template.
	Charset: "UTF-8", // Sets encoding for json and html content-types.
}))


		m.Get("/", func() string {
				return "Hello world!"
			})

	m.Post("/up", func(w http.ResponseWriter, r *http.Request) {
			fmt.Printf("%v\n", "p./up" + r.Method)

			for key, val := range r.Header {
				fmt.Printf("%v\n", key)
				fmt.Printf("%v\n", val)
			}

			err := r.ParseForm();
			if err != nil {
				log.Println(err)
			}

			for key, val := range r.Form {
				fmt.Printf("%v\n", key)
				fmt.Printf("%v\n", val)
			}

			file, _, err := r.FormFile("file")

			if err != nil {
				fmt.Fprintln(w, err)
				return
			}

			defer file.Close()

			// create id and folder
			c := 6
			b := make([]byte, c)
			_, err = rand.Read(b)
			if err != nil {
				fmt.Println("error:", err)
			}
			loadId := hex.EncodeToString(b)

			loadDataPath := config.DataPath + "/" + loadId

			err = os.MkdirAll(loadDataPath, config.DataPathMode)
			if err != nil {
				fmt.Println(err)
			}

			//
			loadArchiveName := loadDataPath + "/" + "load.tar.gz"

			out, err := os.Create(loadArchiveName)
			if err != nil {
				fmt.Fprintf(w, "Failed to open the file for writing")
				return
			}
			defer out.Close()
			_, err = io.Copy(out, file)
			if err != nil {
				fmt.Fprintln(w, err)
			}

			// return loadId
			fmt.Fprintf(w, "%s", loadId)

			compress.UnTarGz(loadArchiveName, loadDataPath)
		})

		m.Run()
}


// Creates a new file upload http request with optional extra params
func newfileUploadRequest(uri string, params map[string]string, paramName, path string) (*http.Request, error) {
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

	for key, val := range params {
		_ = writer.WriteField(key, val)
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest("POST", uri, body)

	request.Header.Set("Content-Type", writer.FormDataContentType())

	return request, err
}

func main() {

//	c, err := config.LoadConfig("./conf.yaml")
//	if err != nil {
//		fmt.Println("Error loading config file:", err)
//	}

	app := cli.NewApp()
	app.Name = "Crane"
	app.Usage = "crane [command]"
	app.Version = "0.0.1"

	app.Flags = []cli.Flag {
		cli.StringFlag {
			Name: "host, H",
			Value: "http://localhost:2475",
			Usage: "Host",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:      "push",
			ShortName: "p",
			Usage: "crane push PATH",
			Description:  "push an image package or a crane package to the crane server",
			Flags: []cli.Flag {
				cli.IntFlag {
					Name: "addint",
					Value: 10,
					Usage: "test in",
				},
			},
			Action: func(c *cli.Context) {

				if (c.Args().Present()) {

					path := c.Args().First();

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
						if (err != nil) {
							log.Fatal(err)
						}
					}()

					compress.TarGz( loadPath, loadCompressedFile )

					// send the file over http
					extraParams := map[string]string {
						"title":       "My Document",
						"author":      "Matt Aimonetti",
						"description": "A document with all the Go programming language secrets",
					}
					request, err := newfileUploadRequest("http://localhost:2475/up", extraParams, "file", loadCompressedFilePath)
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
					cli.ShowCommandHelp(c, "push");
				}



			},
		},
		{
			Name:		"server",
			ShortName:	"s",
			Usage:		"start crane server",
			Flags: []cli.Flag {
				cli.IntFlag {
					Name: "p, port",
					Value: 2475,
					Usage: "port to listen on (default 2475)",
				},
			},
			Action: func(c *cli.Context) {
				startServer(c.Int("port"));
			},
		},
		{
			Name:		"exec",
			ShortName:	"e",
			Usage:		"crane exec LOADID command...",
			Flags: []cli.Flag {
				cli.IntFlag {
					Name: "p, port",
					Value: 2475,
					Usage: "port to listen on (default 2475)",
				},
			},
			Action: func(c *cli.Context) {
				fmt.Println("Execute cmd")

			},
		},

	}

	app.Action = func(c *cli.Context) {

		if len(c.Args()) <= 0 {
			cli.ShowAppHelp(c)
		} else {

			println(c.Bool("d"))

			println("boom! I say!")
		}

	}

	app.Run(os.Args)
}
