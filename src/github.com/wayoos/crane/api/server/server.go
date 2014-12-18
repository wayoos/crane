package server

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	"github.com/wayoos/crane/config"
	"log"
	"os"
)

func ServerCommand(c *cli.Context) {
	//	port := c.Int("port")

	host := c.String("addr")

	var craneDir string = ""
	if c.IsSet("crane-dir") {
		craneDir = c.String("crane-dir")
	}

	startServer(host, craneDir)
}

func startServer(host string, craneDir string) {

	config.InitDataPath(craneDir)

	log.Println("Start crane server")
	fmt.Printf("dataPath=%s", config.DataPath)
	fmt.Println()

	//	os.Setenv("PORT", strconv.Itoa(port))
	//	os.Setenv("HOST", "localhost")

	os.Setenv("MARTINI_ENV", martini.Prod)

	martini.Env = martini.Prod

	m := martini.Classic()

	m.Use(render.Renderer(render.Options{
		//	Directory: "templates", // Specify what path to load the templates from.
		//	Layout: "layout", // Specify a layout template. Layouts can call {{ yield }} to render the current template.
		Charset: "UTF-8", // Sets encoding for json and html content-types.
	}))

	m.Post("/build", Build)
	m.Post("/build/:name", Build)
	m.Post("/build/:name/:tag", Build)

	m.Get("/ps", Ps)

	m.Delete("/dockload/:id", Rm)

	m.Post("/up", Up)
	m.Post("/up/:name", Up)
	m.Post("/up/:name/:tag", Up)

	//m.Run()
	m.RunOnAddr(host)
}
