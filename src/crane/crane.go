package main

import (
	"os"
	"fmt"
	"bytes"
	"io"
	"log"
	"strings"
	"strconv"
	"net/http"
	"mime/multipart"
	"path/filepath"
	"github.com/codegangsta/cli"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
//	"wayoos.com/config"
	"wayoos.com/compress"
	"wayoos.com/config"
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

			file, header, err := r.FormFile("file")

			if err != nil {
				fmt.Fprintln(w, err)
				return
			}

			defer file.Close()

			//
			loadArchiveName := config.DataPath+"/"+header.Filename

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

			// the header contains useful info, like the original file name
			fmt.Fprintf(w, "File %s uploaded successfully.", header.Filename)

			compress.UnTarGz(loadArchiveName, config.DataPath)
//			compress.UnGz(loadArchiveName)
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

	fmt.Printf("dataPath=%s", config.DataPath)
	fmt.Println()

	app := cli.NewApp()
	app.Name = "Crane"
	app.Usage = "make an explosive entrance"
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
			Usage: "crane push LOAD[:TAG] PATH",
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
					println("Push task: ", c.Args().First())

					loadName := c.Args().First();

					dirPath := c.Args().Get(1);

					_, err := os.Open( dirPath )
					if err != nil {
						log.Fatal(err)
					}

					println("Error but continue")

					// tar directory

					targetFilePath := loadName + ".tar.gz"
					inputDirPath := dirPath

					//dir, err := os.Open( dirPath )

					compress.TarGz( targetFilePath, strings.TrimRight( inputDirPath, "/" ) )

//					fileArchive, err := os.Open(targetFilePath)
//					if err != nil {
//						log.Fatal(err)
//					}

					path, _ := os.Getwd()
					path += "/"+targetFilePath
					extraParams := map[string]string {
						"title":       "My Document",
						"author":      "Matt Aimonetti",
						"description": "A document with all the Go programming language secrets",
					}
					request, err := newfileUploadRequest("http://localhost:2475/up", extraParams, "file", path)
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
						fmt.Println(resp.StatusCode)
						fmt.Println(resp.Header)
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
