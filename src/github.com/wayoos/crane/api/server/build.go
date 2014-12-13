package server

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"github.com/go-martini/martini"
	"github.com/wayoos/crane/api/docker"
	"github.com/wayoos/crane/api/domain"
	"github.com/wayoos/crane/compress"
	"github.com/wayoos/crane/config"
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

	var loadId string = ""
	var loadDataPath string = ""
	// create id and folder
	for {
		c := 6
		b := make([]byte, c)
		_, err = rand.Read(b)
		if err != nil {
			return "", &domain.AppError{nil, "Error when create dockloadId", 500}
		}
		loadId = hex.EncodeToString(b)

		loadDataPath = config.DataPath + "/" + loadId

		if _, err := os.Stat(loadDataPath); os.IsNotExist(err) {
			// path/to/whatever does not exist
			break
		}

	}
	loadDataJson := config.DataPath + "/" + loadId + ".json"

	err = os.MkdirAll(loadDataPath, config.DataPathMode)
	if err != nil {
		return "", &domain.AppError{err, "Error creating dockload folder", 500}
	}

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

	imageId, appErr := BuildImage(loadId)
	if appErr != nil {
		return "", appErr
	}

	loadData := domain.LoadData{
		ID:      loadId,
		Name:    tagName,
		Tag:     tagVersion,
		ImageId: imageId,
	}

	outJson, err := os.Create(loadDataJson)
	if err != nil {
		return "", &domain.AppError{err, "Failed to create data file", 500}
	}
	defer outJson.Close()

	enc := json.NewEncoder(outJson)

	enc.Encode(loadData)

	return loadId, nil
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
