package server

import (
	"github.com/go-martini/martini"
	"github.com/wayoos/crane/api/docker"
	"github.com/wayoos/crane/api/domain"
	"github.com/wayoos/crane/config"
	"log"
	"os"
)

func Rm(params martini.Params) (int, string) {
	id := params["id"]
	appErr := ExecuteRm(id)
	if appErr != nil {
		log.Println(appErr.Error)
		return appErr.Code, appErr.Message
	}
	return 204, ""
}

func ExecuteRm(tag string) (appErr *domain.AppError) {
	// find dockloadId
	loadDataPath := config.DataPath + "/" + tag

	imageId := tag

	isRunning, _ := docker.IsRunning(imageId)

	if isRunning {
		_, appErr := docker.Stop(imageId)
		if appErr != nil {
			return appErr
		}

	}

	isExited, _ := docker.IsExited(imageId)
	if isExited {
		_, appErr = docker.RemoveContainer(imageId)
		if appErr != nil {
			return appErr
		}
	}

	isImageBuild, appErr := docker.IsImageBuild(imageId)
	if appErr != nil {
		return appErr
	}

	if isImageBuild {
		_, appErr := docker.RemoveImage(imageId)
		if appErr != nil {
			return appErr
		}
	}

	// remove dockload storage
	err := os.RemoveAll(loadDataPath)
	if err != nil {
		return &domain.AppError{err, "Remove dockload error.", 500}
	}

	err = os.Remove(loadDataPath + ".json")
	if err != nil {
		return &domain.AppError{err, "Remove dockload json error.", 500}
	}

	return nil
}
