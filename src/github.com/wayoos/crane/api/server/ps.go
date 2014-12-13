package server

import (
	"container/list"
	"encoding/json"
	"fmt"
	"github.com/martini-contrib/render"
	"github.com/wayoos/crane/api/domain"
	"github.com/wayoos/crane/config"
	"io/ioutil"
	"os"
)

func Ps(r render.Render) {

	l := list.New()

	files, _ := ioutil.ReadDir(config.DataPath)
	for _, f := range files {
		if f.IsDir() {
			fmt.Println(f.Name())
			l.PushBack(f.Name())
		}
	}

	var loadRecords = make([]domain.LoadData, l.Len())

	idx := 0
	for e := l.Front(); e != nil; e = e.Next() {

		loadId := e.Value.(string)

		inJson, err := os.Open(config.DataPath + "/" + loadId + ".json")
		if err != nil {
			fmt.Println(err)
			return
		}
		defer inJson.Close()

		decode := json.NewDecoder(inJson)
		var loadData domain.LoadData
		err = decode.Decode(&loadData)
		if err != nil {
			fmt.Println(err)
			return
		}

		loadRecords[idx] = loadData
		idx += 1
	}

	r.JSON(200, loadRecords)
}
