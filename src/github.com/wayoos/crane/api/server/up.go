package server

import (
	"github.com/go-martini/martini"
	"github.com/wayoos/crane/api/docker"
	"github.com/wayoos/crane/api/domain"
	"github.com/wayoos/crane/config"
	"log"
	"path/filepath"
	"strings"
)

func Up(params martini.Params) (int, string) {

	dockloadId := params["loadid"]

	dockloadPath := config.DataPath + "/" + dockloadId

	appErr := ExecuteUp(dockloadPath)

	if appErr != nil {
		return appErr.Code, appErr.Message
	}
	return 204, ""
}

func ExecuteUp(path string) *domain.AppError {
	if !config.Exists(path + "/Dockerfile") {
		return &domain.AppError{nil, "Dockerfile not found in " + path, 404}
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
			log.Println(err)
		}
	}

	return nil
}
