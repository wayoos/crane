package server

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	"github.com/wayoos/crane/config"
	"os"
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

	m.Post("/build", Build)
	m.Post("/build/:name", Build)
	m.Post("/build/:name/:tag", Build)

	m.Get("/ps", Ps)

	m.Post("/up", Up)
	m.Post("/up/:name", Up)
	m.Post("/up/:name/:tag", Up)

	m.Run()
}
