package main

import (
	"bytes"
	"container/list"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/go-martini/martini"
	"github.com/jmcvetta/napping"
	"github.com/martini-contrib/binding"
	"github.com/martini-contrib/render"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	//	"wayoos.com/config"
	"io/ioutil"
	"text/tabwriter"
	"wayoos.com/compress"
	"wayoos.com/config"
)

//type load_record struct {
//	LoadId string
//}

type LoadData struct {
	ID   string
	Name string
	Tag  string
}

type ExecData struct {
	LoadId string
	Cmd    []string
}

type ExecResult struct {
	ExitCode int
	Out      string
}

func CToGoString(c []byte) string {
	n := -1
	for i, b := range c {
		if b == 0 {
			break
		}
		n = i
	}
	return string(c[:n+1])
}

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
		//	Directory: "templates", // Specify what path to load the templates from.
		//	Layout: "layout", // Specify a layout template. Layouts can call {{ yield }} to render the current template.
		Charset: "UTF-8", // Sets encoding for json and html content-types.
	}))

	m.Post("/exec", binding.Bind(ExecData{}), func(execData ExecData, r render.Render) {
		fmt.Printf("LoadId: %s\n",
			execData.LoadId)
		cmds := execData.Cmd

		program := cmds[0]

		for i := range cmds {
			println(cmds[i])
		}

		var programArgs = make([]string, len(cmds)-1)

		for i := 0; i < len(programArgs); i++ {
			println(cmds[i+1])
			programArgs[i] = cmds[i+1]
		}

		fmt.Printf("Program: %s\n", program)

		cmd := exec.Command(program, programArgs...)
		cmd.Dir = config.DataPath + "/" + execData.LoadId
		cmd.Stdin = os.Stdin
		//		cmd.Stdout = os.Stdout
		//		cmd.Stderr = os.Stderr
		//		err := cmd.Run()
		outByte, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("%v\n", err)
		}

		outString := CToGoString(outByte)

		println(outString)

		execResult := ExecResult{
			ExitCode: 0,
			Out:      outString,
		}

		r.JSON(200, execResult)
	})

	m.Get("/ps", func(r render.Render) {

		l := list.New()

		files, _ := ioutil.ReadDir(config.DataPath)
		for _, f := range files {
			if f.IsDir() {
				fmt.Println(f.Name())
				l.PushBack(f.Name())
			}
		}

		var loadRecords = make([]LoadData, l.Len())

		idx := 0
		for e := l.Front(); e != nil; e = e.Next() {

			loadId := e.Value.(string)

			fmt.Println("LoadId: " + loadId)

			inJson, err := os.Open(config.DataPath + "/" + loadId + ".json")
			if err != nil {
				fmt.Println(err)
				return
			}
			defer inJson.Close()

			decode := json.NewDecoder(inJson)
			var loadData LoadData
			err = decode.Decode(&loadData)
			if err != nil {
				fmt.Println(err)
				return
			}

			fmt.Println(loadData.Name)

			loadRecords[idx] = loadData
			idx += 1
		}

		r.JSON(200, loadRecords)
	})

	m.Post("/up", func(w http.ResponseWriter, r *http.Request) {

		nameTag := r.Header.Get("Load-tag")

		err := r.ParseForm()
		if err != nil {
			log.Println(err)
		}

		file, _, err := r.FormFile("file")

		if err != nil {
			fmt.Fprintln(w, err)
			return
		}

		defer file.Close()

		var loadId string = ""
		var loadDataPath string = ""
		// create id and folder
		for {
			c := 6
			b := make([]byte, c)
			_, err = rand.Read(b)
			if err != nil {
				fmt.Println("error:", err)
			}
			loadId = hex.EncodeToString(b)

			loadDataPath = config.DataPath + "/" + loadId

			if _, err := os.Stat(loadDataPath); os.IsNotExist(err) {
				// path/to/whatever does not exist
				break
			}

		}
		loadDataJson := config.DataPath + "/" + loadId + ".json"

		fmt.Println("mkdir " + loadDataPath)

		err = os.MkdirAll(loadDataPath, config.DataPathMode)
		if err != nil {
			fmt.Println(err)
		}

		//
		loadArchiveName := loadDataPath + "/" + "load.zip"

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

		//		compress.UnTarGz(loadArchiveName, loadDataPath)
		err = compress.Unzip(loadArchiveName, loadDataPath)
		if err != nil {
			fmt.Println(err)
			fmt.Fprintf(w, "Failed to extract file")
			return
		}

		split := strings.Split(nameTag, ":")
		name := split[0]
		tag := ""
		if len(split) > 1 {
			tag = split[1]
		}

		loadData := LoadData{
			ID:   loadId,
			Name: name,
			Tag:  tag,
		}

		outJson, err := os.Create(loadDataJson)
		if err != nil {
			fmt.Fprintf(w, "Failed to open the file for writing")
			return
		}
		defer outJson.Close()

		enc := json.NewEncoder(outJson)

		enc.Encode(loadData)

		//		bl, _ := json.Marshal(loadData)
		//		os.Stdout.Write(bl)

		// return loadId
		fmt.Fprintf(w, "%s", loadId)

	})

	m.Run()
}

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

	//	c, err := config.LoadConfig("./conf.yaml")
	//	if err != nil {
	//		fmt.Println("Error loading config file:", err)
	//	}

	app := cli.NewApp()
	app.Name = "Crane"
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
			},
			Action: func(c *cli.Context) {
				startServer(c.Int("port"))
			},
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

				execData := ExecData{
					LoadId: loadId,
					Cmd:    cmds,
				}

				result := ExecResult{}
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
				result := []LoadData{}
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
