package server

import (
	"github.com/go-martini/martini"
	"github.com/wayoos/crane/api/docker"
	"github.com/wayoos/crane/api/domain"
	"github.com/wayoos/crane/store"
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

func ExecuteRm(tag2 string) (appErr *domain.AppError) {
	// find dockloadId

	dockloadData, appErr := store.Find(tag2)
	if appErr != nil {
		return appErr
	}

	loadDataPath := store.Path(dockloadData)

	imageId := dockloadData.ID

	isRunning, _ := docker.IsRunning(imageId)

	if isRunning {
		_, appErr := docker.Stop(imageId)
		if appErr != nil {
			return appErr
		}

	}

	hasContainer, _ := docker.HasContainer(imageId)
	if hasContainer {
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
