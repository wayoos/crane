package server

import (
	"container/list"
	"encoding/json"
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/binding"
	"github.com/martini-contrib/render"
	"github.com/wayoos/crane/api/domain"
	"github.com/wayoos/crane/config"
	"github.com/wayoos/crane/util"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
)

func ServerCommand(c *cli.Context) {
	port := c.Int("port")

	var craneDir string = ""
	if c.IsSet("crane-dir") {
		craneDir = c.String("crane-dir")
	}

	startServer(port, craneDir)
}

func startServer(port int, craneDir string) {

	config.InitDataPath(craneDir)

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

	m.Post("/load/:loadid", Up)
	m.Post("/load/:loadid/:name", Up)
	m.Post("/load/:loadid/:name/:tag", Up)

	m.Post("/exec", binding.Bind(domain.ExecData{}), func(execData domain.ExecData, r render.Render) {
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

		outString := util.CToGoString(outByte)

		println(outString)

		execResult := domain.ExecResult{
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

		var loadRecords = make([]domain.LoadData, l.Len())

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
			var loadData domain.LoadData
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

	m.Post("/push", Build)
	m.Post("/push/:name", Build)
	m.Post("/push/:name/:tag", Build)

	m.Run()
}
