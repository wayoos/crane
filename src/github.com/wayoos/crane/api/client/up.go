package client

import (
	"github.com/codegangsta/cli"
	"os"
	"path/filepath"
)

func UpCommand(c *cli.Context) {

	path, _ := os.Getwd()
	if c.Args().Present() {
		path = c.Args().First()
		path, _ = filepath.Abs(path)
	}

	host := c.GlobalString("host")
	var tag string
	if c.IsSet("tag") {
		tag = c.String("tag")
	}

	BuildSend("up", host, path, tag)
}
