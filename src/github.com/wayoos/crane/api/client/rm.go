package client

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/wayoos/crane/api/docker"
	"github.com/wayoos/crane/config"
	"log"
	"os"
	"path/filepath"
)

func RmCommand(c *cli.Context) {

	path, _ := os.Getwd()
	if c.Args().Present() {
		path = c.Args().First()
		path, _ = filepath.Abs(path)
	}
	if !config.Exists(path + "/Dockerfile") {
		fmt.Println("Dockerfile not found in " + path)
		return
	}

	// we are using the parent directory of the Dockerfile
	imageId := filepath.Base(path)

	isRunning, _ := docker.IsRunning(imageId)

	if isRunning {
		_, err := docker.Stop(imageId)
		if err != nil {
			log.Fatal(err)
		}
	}

	_, err := docker.RemoveContainer(imageId)
	if err != nil {
		log.Fatal(err)
	}

}
