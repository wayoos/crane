package server

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/go-martini/martini"
	"github.com/wayoos/crane/api/docker"
	"github.com/wayoos/crane/api/domain"
	"github.com/wayoos/crane/compress"
	"github.com/wayoos/crane/config"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func Build(params martini.Params, r *http.Request) (int, string) {

	tagName := params["name"]
	tagVersion := params["tag"]

	appErr := ExecuteBuild(tagName, tagVersion, r)

	if appErr != nil {
		return appErr.Code, appErr.Message
	}
	return 204, ""
}

func ExecuteBuild(tagName, tagVersion string, r *http.Request) *domain.AppError {

	err := r.ParseForm()
	if err != nil {
		log.Println(err)
		return &domain.AppError{nil, "Invalid zip file", 500}
	}

	file, _, err := r.FormFile("file")

	if err != nil {
		log.Println(err)
		return &domain.AppError{nil, "Invalid zip file", 500}
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
			fmt.Println("error:", err)
		}
		loadId = hex.EncodeToString(b)

		loadDataPath = config.DataPath + "/" + loadId

		if _, err := os.Stat(loadDataPath); os.IsNotExist(err) {
			// path/to/whatever does not exist
			break
		}

	}
	loadDataJson := config.DataPath + "/" + loadId + ".json"

	fmt.Println("mkdir " + loadDataPath)

	err = os.MkdirAll(loadDataPath, config.DataPathMode)
	if err != nil {
		fmt.Println(err)
	}

	//
	loadArchiveName := loadDataPath + "/" + "load.zip"

	out, err := os.Create(loadArchiveName)
	if err != nil {
		return &domain.AppError{nil, "Failed to open the file for writing", 500}
	}
	defer out.Close()
	_, err = io.Copy(out, file)
	if err != nil {
		log.Println(err)
		return &domain.AppError{nil, "Open file error", 500}
	}

	//		compress.UnTarGz(loadArchiveName, loadDataPath)
	err = compress.Unzip(loadArchiveName, loadDataPath)
	if err != nil {
		log.Println(err)
		return &domain.AppError{nil, "Failed to extract file", 500}
	}

	loadData := domain.LoadData{
		ID:   loadId,
		Name: tagName,
		Tag:  tagVersion,
	}

	outJson, err := os.Create(loadDataJson)
	if err != nil {
		log.Println(err)
		return &domain.AppError{nil, "Failed to create data file", 500}
	}
	defer outJson.Close()

	enc := json.NewEncoder(outJson)

	enc.Encode(loadData)

	//		bl, _ := json.Marshal(loadData)
	//		os.Stdout.Write(bl)

	appErr := BuildImage(loadId)
	if appErr != nil {
		return appErr
	}
	// return loadId
	// TODO find a better solution as using error return structure to return correct data
	return &domain.AppError{nil, loadId, 200}
}

func BuildImage(dockloadId string) *domain.AppError {

	dockloadPath := config.DataPath + "/" + dockloadId

	if !config.Exists(dockloadPath + "/Dockerfile") {
		return &domain.AppError{nil, "Dockerfile not found in " + dockloadPath, 404}
	}

	// check if an images is present with the
	outLines, err := docker.ExecuteDocker(dockloadPath, "images")
	if err != nil {
		log.Fatal(err)
	}

	alreadyBuild := false
	for _, line := range outLines {
		if strings.HasPrefix(line, dockloadId) {
			alreadyBuild = true
		}
		//		println(line)
	}

	// if the image is not present build it

	if !alreadyBuild {
		outLines, err = docker.Build(dockloadPath, dockloadId)
		if err != nil {
			for _, line := range outLines {
				println(line)
			}

			log.Fatal(err)
		}

	}

	return nil
}
