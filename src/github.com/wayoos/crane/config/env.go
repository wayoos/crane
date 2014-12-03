package config

import (
	"os"
	"strings"
)

var RootPath string
var DataPath string
var DataPathMode os.FileMode

func Exists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

func InitDataPath(craneDir string) {

	if craneDir == "" {
		var err error
		RootPath, err = os.Getwd()
		if err != nil {
			panic(err)
		}
	} else {
		RootPath = craneDir
	}

	DataPath = RootPath + "/data"

	rootInfo, err := os.Stat(RootPath)
	if err != nil {
		panic(err)
	}
	DataPathMode = rootInfo.Mode()

	if !Exists(DataPath) {
		os.Mkdir(DataPath, DataPathMode)
	}
}

func MkdirIfNotExist(path string) {
	var validPath string
	if strings.HasPrefix(path, RootPath) {
		validPath = path
	} else {
		validPath = RootPath + "/" + path
	}

	if !Exists(validPath) {
		rootInfo, _ := os.Stat(RootPath)
		os.MkdirAll(validPath, rootInfo.Mode())
	}
}
