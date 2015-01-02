package server

import (
	"github.com/codegangsta/cli"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	"github.com/wayoos/crane/config"
	"log"
	"os"
)

func ServerCommand(c *cli.Context) {
	host := c.String("addr")

	var craneDir string = ""
	if c.IsSet("crane-dir") {
		craneDir = c.String("crane-dir")
	}

	startServer(host, craneDir)
}

func startServer(host string, craneDir string) {

	config.InitDataPath(craneDir)

	log.Println("Start crane server (dataPath=" + config.DataPath + ")")

	os.Setenv("MARTINI_ENV", martini.Prod)

	martini.Env = martini.Prod

	m := Classic()

	//	m.Martini.logger = log.New(os.Stdout, "[crane] ", 0)

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

func Classic() *martini.ClassicMartini {
	r := martini.NewRouter()
	m := martini.New()
	//	m.Use(martini.Logger())
	m.Use(martini.Recovery())
	//	m.Use(martini.Static("public"))
	m.MapTo(r, (*martini.Routes)(nil))
	m.Action(r.Handle)
	return &martini.ClassicMartini{m, r}
}
