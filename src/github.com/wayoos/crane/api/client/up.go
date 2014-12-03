package client

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/jmcvetta/napping"
	"github.com/wayoos/crane/api/docker"
	"github.com/wayoos/crane/config"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func UpCommand(c *cli.Context) {
	var loadId string = ""
	if c.Args().Present() {
		loadId = c.Args().First()
	}

	host := c.GlobalString("host")
	resp, err := napping.Post(host+"/load/"+loadId, nil, nil, nil)
	if err != nil {
		log.Println("Error: failed to create and start container")
		os.Exit(1)
	}
	if resp.Status() == 204 {
		fmt.Println(loadId)
	} else {
		fmt.Println("Error response from crane daemon: " + resp.RawText())
		log.Println("Error: failed to create and start container")
		os.Exit(1)
	}

}

func UplCommand(c *cli.Context) {

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
