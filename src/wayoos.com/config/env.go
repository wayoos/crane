package config

import (
	"os"
	"strings"
)

var RootPath string
var DataPath string

func Exists(path string) (bool) {
	_, err := os.Stat(path)
	if err == nil { return true }
	if os.IsNotExist(err) { return false }
	return false
}

func init() {
	var err error
	RootPath, err = os.Getwd()
	if err != nil {
		panic(err)
	}
}

func InitDataPath() {
	DataPath = RootPath + "/data"

	if !Exists(DataPath) {
		rootInfo,_ := os.Stat(RootPath)
		os.Mkdir(DataPath, rootInfo.Mode())
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
		rootInfo,_ := os.Stat(RootPath)
		os.MkdirAll(validPath, rootInfo.Mode())
	}
}
