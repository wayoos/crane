package server

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/go-martini/martini"
	"github.com/wayoos/crane/api/domain"
	"github.com/wayoos/crane/compress"
	"github.com/wayoos/crane/config"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

func Build(params martini.Params, w http.ResponseWriter, r *http.Request) {

	tagName := params["name"]
	tagVersion := params["version"]

	fmt.Println("Tab name: " + tagName)
	fmt.Println("Tab version: " + tagVersion)

	nameTag := r.Header.Get("Load-tag")

	err := r.ParseForm()
	if err != nil {
		log.Println(err)
	}

	file, _, err := r.FormFile("file")

	if err != nil {
		fmt.Fprintln(w, err)
		return
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
		fmt.Fprintf(w, "Failed to open the file for writing")
		return
	}
	defer out.Close()
	_, err = io.Copy(out, file)
	if err != nil {
		fmt.Fprintln(w, err)
	}

	//		compress.UnTarGz(loadArchiveName, loadDataPath)
	err = compress.Unzip(loadArchiveName, loadDataPath)
	if err != nil {
		fmt.Println(err)
		fmt.Fprintf(w, "Failed to extract file")
		return
	}

	split := strings.Split(nameTag, ":")
	name := split[0]
	tag := ""
	if len(split) > 1 {
		tag = split[1]
	}

	loadData := domain.LoadData{
		ID:   loadId,
		Name: name,
		Tag:  tag,
	}

	outJson, err := os.Create(loadDataJson)
	if err != nil {
		fmt.Fprintf(w, "Failed to open the file for writing")
		return
	}
	defer outJson.Close()

	enc := json.NewEncoder(outJson)

	enc.Encode(loadData)

	//		bl, _ := json.Marshal(loadData)
	//		os.Stdout.Write(bl)

	// return loadId
	fmt.Fprintf(w, "%s", loadId)

}
