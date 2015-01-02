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
	loadId, appErr := ExecuteBuild(params["name"], params["tag"], r)
	if appErr != nil {
		return appErr.Code, appErr.Message
	}
	return 200, loadId
}

func ExecuteBuild(loadName, loadTag string, r *http.Request) (loadId string, errApp *domain.AppError) {
	err := r.ParseForm()
	if err != nil {
		return "", &domain.AppError{nil, "Invalid zip file", 500}
	}

	file, _, err := r.FormFile("file")

	if err != nil {
		return "", &domain.AppError{nil, "Invalid zip file", 500}
	}

	defer file.Close()

	loadData, appErr := store.Create(loadName, loadTag)
	if appErr != nil {
		return "", appErr
	}

	loadDataPath := store.Path(loadData)
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

	imageId, appErr := BuildImage(loadData.ID)
	if appErr != nil {
		return "", appErr
	}

	loadData.ImageId = imageId

	appErr = store.Save(loadData)
	if appErr != nil {
		return "", appErr
	}

	return loadData.ID, nil
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
