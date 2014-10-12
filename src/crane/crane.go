package main

import (
	"os"
	"fmt"
	"io"
	"net/http"
	"github.com/codegangsta/cli"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	"config"
)


func startServer() {
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
			fmt.Printf("%v\n", "p./up")

			file, header, err := r.FormFile("file")
//			defer file.Close()

			if err != nil {
				fmt.Fprintln(w, err)
				return
			}

			out, err := os.Create("/tmp/file")
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
		})

		m.Run()
}


func main() {

	c, err := config.LoadConfig("./conf.yaml")
	if err != nil {
		fmt.Println("Error loading config file:", err)
	}

	println(c)

	app := cli.NewApp()
	app.Name = "Crane"
	app.Usage = "make an explosive entrance"
	app.Version = "0.0.1"

	app.Flags = []cli.Flag {
		cli.BoolFlag {
			Name:  "d, daemon",
			Usage: "Enable daemon mode",
		},
	}
	app.Action = func(c *cli.Context) {

		if c.GlobalBool("d") {
			startServer()
		} else {


		if len(c.Args()) <= 0 {
			cli.ShowAppHelp(c)
		} else {

			println(c.Bool("d"))

			println("boom! I say!")
		}
		}

	}

	app.Run(os.Args)
}
