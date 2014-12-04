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

func ExecuteUp(dockloadId string) *domain.AppError {

	appErr := BuildImage(dockloadId)
	if appErr != nil {
		return appErr
	}

	dockloadPath := config.DataPath + "/" + dockloadId

	// now we can run the container
	isRunning, _ := docker.IsRunning(dockloadId)
	if !isRunning {

		isExited, err := docker.IsExited(dockloadId)
		if isExited {
			docker.Start(dockloadId)
		} else {
			outLines, _ := docker.Run(dockloadPath, dockloadId)
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
