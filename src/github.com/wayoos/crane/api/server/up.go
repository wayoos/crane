package server

import (
	"github.com/go-martini/martini"
	"github.com/wayoos/crane/api/docker"
	"github.com/wayoos/crane/api/domain"
	"github.com/wayoos/crane/config"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

func Up(params martini.Params, r *http.Request) (int, string) {

	tagName := params["name"]
	tagVersion := params["tag"]

	dockloadId, appErr := ExecuteBuild(tagName, tagVersion, r)

	if appErr != nil {
		log.Println(appErr.Error)
		return appErr.Code, appErr.Message
	}

	//	dockloadPath := config.DataPath + "/" + dockloadId

	appErr = ExecuteUp(dockloadId)
	if appErr != nil {
		log.Println(appErr.Error)
		return appErr.Code, appErr.Message
	}

	return 200, dockloadId
}

func UpOnly(params martini.Params) (int, string) {

	dockloadId := params["loadid"]

	dockloadPath := config.DataPath + "/" + dockloadId

	appErr := ExecuteUp(dockloadPath)

	if appErr != nil {
		return appErr.Code, appErr.Message
	}
	return 204, ""
}

type CraneConfig struct {
	Run struct {
		Args string
	}
}

func ExecuteUp(dockloadId string) *domain.AppError {

	//	_, appErr := BuildImage(dockloadId)
	//	if appErr != nil {
	//		return appErr
	//	}

	dockloadPath := config.DataPath + "/" + dockloadId

	// now we can run the container
	isRunning, _ := docker.IsRunning(dockloadId)
	if !isRunning {

		isExited, err := docker.IsExited(dockloadId)
		if isExited {
			docker.Start(dockloadId)
		} else {

			dockerArgs := []string{}

			// load crane.yaml
			craneConfigFile := dockloadPath + "/crane.yaml"
			if config.Exists(craneConfigFile) {
				craneConfig := CraneConfig{}
				yamlFile, err := ioutil.ReadFile(craneConfigFile)
				if err != nil {
					log.Println("error: %v", err)
				}
				err = yaml.Unmarshal(yamlFile, &craneConfig)
				if err != nil {
					log.Println("error: %v", err)
				}

				if craneConfig.Run.Args != "" {
					runArgs := strings.Split(craneConfig.Run.Args, " ")
					dockerArgs = append(dockerArgs, runArgs...)
				}

			}

			outLines, _ := docker.Run(dockloadPath, dockloadId, dockerArgs...)
			for _, line := range outLines {
				log.Println(line)
			}
		}

		if err != nil {
			log.Println(err)
		}
	}

	return nil
}
