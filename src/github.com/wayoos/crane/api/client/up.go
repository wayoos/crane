package client

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/wayoos/crane/api/docker"
	"log"
	"os"
	"path/filepath"
	"strings"
	"wayoos.com/config"
)

func UpCommand(c *cli.Context) {
	//	println("Up command")

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

	// check if an images is present with the
	outLines, err := docker.ExecuteDocker(path, "images")
	if err != nil {
		log.Fatal(err)
	}

	alreadyBuild := false
	for _, line := range outLines {
		if strings.HasPrefix(line, imageId) {
			alreadyBuild = true
		}
		//		println(line)
	}

	// if the image is not present build it

	if !alreadyBuild {
		outLines, err = docker.Build(path, imageId)
		if err != nil {
			for _, line := range outLines {
				println(line)
			}

			log.Fatal(err)
		}

	}

	// now we can run the container
	isRunning, err := docker.IsRunning(imageId)
	if !isRunning {

		isExited, err := docker.IsExited(imageId)
		if isExited {
			docker.Start(imageId)
		} else {
			outLines, err = docker.Run(path, imageId)
			for _, line := range outLines {
				println(line)
			}
		}

		if err != nil {

			log.Fatal(err)
		}
	}

}
