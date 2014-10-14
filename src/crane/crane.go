package main

import (
	"os"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"net/http"
	"archive/tar"
	"compress/gzip"
	"github.com/codegangsta/cli"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	"config"
)


func startServer(port int) {

	os.Setenv("PORT", strconv.Itoa(port))

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

func handleError( _e error ) {
	if _e != nil {
		log.Fatal( _e )
	}
}

func TarGzWrite( _path string, tw *tar.Writer, fi os.FileInfo ) {
	fr, err := os.Open( _path )
	handleError( err )
	defer fr.Close()

	h := new( tar.Header )
	h.Name = _path
	//fi.Name()
	h.Size = fi.Size()
	h.Mode = int64( fi.Mode() )
	h.ModTime = fi.ModTime()

	err = tw.WriteHeader( h )
	handleError( err )

	_, err = io.Copy( tw, fr )
	handleError( err )
}

func IterDirectory( dirPath string, tw *tar.Writer ) {
	dir, err := os.Open( dirPath )
	handleError( err )
	defer dir.Close()
	fis, err := dir.Readdir( 0 )
	handleError( err )
	for _, fi := range fis {
		curPath := dirPath + "/" + fi.Name()
		if fi.IsDir() {
			//TarGzWrite( curPath, tw, fi )
			IterDirectory( curPath, tw )
		} else {
			fmt.Printf( "adding... %s\n", curPath )
			TarGzWrite( curPath, tw, fi )
		}
	}
}

func TarGz( outFilePath string, inPath string ) {
	// file write
	fw, err := os.Create( outFilePath )
	handleError( err )
	defer fw.Close()

	// gzip write
	gw := gzip.NewWriter( fw )
	defer gw.Close()

	// tar write
	tw := tar.NewWriter( gw )
	defer tw.Close()

	IterDirectory( inPath, tw )

	fmt.Println( "tar.gz ok" )
}

func main() {

	c, err := config.LoadConfig("./conf.yaml")
	if err != nil {
//		fmt.Println("Error loading config file:", err)
	}

	println(c)

	app := cli.NewApp()
	app.Name = "Crane"
	app.Usage = "make an explosive entrance"
	app.Version = "0.0.1"

	app.Flags = []cli.Flag {
		cli.IntFlag {
			Name: "int",
			Value: 10,
			Usage: "test in",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:      "push",
			ShortName: "p",
			Usage:     "push an image package or a crane package to the crane server",
			Flags: []cli.Flag {
				cli.IntFlag {
					Name: "addint",
					Value: 10,
					Usage: "test in",
				},
			},
			Action: func(c *cli.Context) {
				println("Push task: ", c.Args().First())

				// tar directory

				targetFilePath := "test.tar.gz"
				inputDirPath := "haproxy/"
				TarGz( targetFilePath, strings.TrimRight( inputDirPath, "/" ) )

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
