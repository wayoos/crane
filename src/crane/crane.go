package main

import (
	"os"
	"github.com/codegangsta/cli"
	"github.com/go-martini/martini"
)

func main() {
	m := martini.Classic()
	m.Get("/", func() string {
			return "Hello world!"
		})



	m.Run()
}
