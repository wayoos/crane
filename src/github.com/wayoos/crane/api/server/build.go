package server

import (
	"github.com/go-martini/martini"
	"github.com/wayoos/crane/api/docker"
	"github.com/wayoos/crane/api/domain"
	"github.com/wayoos/crane/compress"
	"github.com/wayoos/crane/config"
	"github.com/wayoos/crane/store"
	"io"
	"net/http"
	"os"
	"strings"
)

func Build(params martini.Params, r *http.Request) (int, string) {

	tagName := params["name"]
	tagVersion := params["tag"]

	dockloadId, appErr := ExecuteBuild(tagName, tagVersion, r)

	if appErr != nil {
		return appErr.Code, appErr.Message
	}
	return 200, dockloadId
}

func ExecuteBuild(tagName, tagVersion string, r *http.Request) (dockloadId string, errApp *domain.AppError) {

	err := r.ParseForm()
	if err != nil {
		return "", &domain.AppError{nil, "Invalid zip file", 500}
	}

	file, _, err := r.FormFile("file")

	if err != nil {
		return "", &domain.AppError{nil, "Invalid zip file", 500}
	}

	defer file.Close()

	dockloadInfo, appErr := store.Create()
	if appErr != nil {
		return "", appErr
	}

	dockloadInfo.Name = tagName
	dockloadInfo.Tag = tagVersion

	appErr = store.Save(dockloadInfo)
	if appErr != nil {
		return "", appErr
	}

	loadDataPath := store.Path(dockloadInfo)

	//
	loadArchiveName := loadDataPath + "/" + "load.zip"

	out, err := os.Create(loadArchiveName)
	if err != nil {
		return "", &domain.AppError{nil, "Failed to open the file for writing", 500}
	}
	defer out.Close()
	_, err = io.Copy(out, file)
	if err != nil {
		return "", &domain.AppError{err, "Open file error", 500}
	}

	//		compress.UnTarGz(loadArchiveName, loadDataPath)
	err = compress.Unzip(loadArchiveName, loadDataPath)
	if err != nil {
		return "", &domain.AppError{err, "Failed to extract file", 500}
	}

	imageId, appErr := BuildImage(dockloadInfo.ID)
	if appErr != nil {
		return "", appErr
	}

	dockloadInfo.ImageId = imageId

	appErr = store.Save(dockloadInfo)
	if appErr != nil {
		return "", appErr
	}

	return dockloadInfo.ID, nil
}

func BuildImage(dockloadId string) (imageId string, appErr *domain.AppError) {

	dockloadPath := config.DataPath + "/" + dockloadId

	if !config.Exists(dockloadPath + "/Dockerfile") {
		return "", &domain.AppError{nil, "Dockerfile not found in " + dockloadPath, 404}
	}

	isImageBuild, err := docker.IsImageBuild(dockloadId)
	if err != nil {
		return "", err
	}

	// if the image is not present build it

	if !isImageBuild {
		outLines, err := docker.Build(dockloadPath, dockloadId)
		if err != nil {
			return "", err
		}

		// find image id
		return strings.Split(outLines[len(outLines)-1], " ")[2], nil
	}

	return "", nil
}
