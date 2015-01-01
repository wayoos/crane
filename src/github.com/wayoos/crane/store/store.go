package store

import (
	"container/list"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"github.com/wayoos/crane/api/domain"
	"github.com/wayoos/crane/config"
	"io/ioutil"
	"os"
)

func Path(dockloadInfo domain.LoadData) string {
	return config.DataPath + "/" + dockloadInfo.ID
}

func Create() (dockloadData domain.LoadData, errApp *domain.AppError) {

	var loadId string
	var loadDataPath string

	// create id and folder
	for {
		c := 6
		b := make([]byte, c)
		_, err := rand.Read(b)
		if err != nil {
			return domain.LoadData{}, &domain.AppError{nil, "Error when create dockloadId", 500}
		}
		loadId = hex.EncodeToString(b)

		loadDataPath = config.DataPath + "/" + loadId

		if _, err := os.Stat(loadDataPath); os.IsNotExist(err) {
			// path/to/whatever does not exist
			break
		}

	}

	err := os.MkdirAll(loadDataPath, config.DataPathMode)
	if err != nil {
		return domain.LoadData{}, &domain.AppError{err, "Error creating dockload folder", 500}
	}

	loadData := domain.LoadData{
		ID: loadId,
	}

	appErr := Save(loadData)
	if appErr != nil {
		return domain.LoadData{}, appErr
	}

	return loadData, nil
}

func Save(dockloadInfo domain.LoadData) *domain.AppError {

	if dockloadInfo.ID == "" {
		return &domain.AppError{nil, "Invalid dockloadId", 400}
	}

	loadDataJson := config.DataPath + "/" + dockloadInfo.ID + ".json"

	outJson, err := os.Create(loadDataJson)
	if err != nil {
		return &domain.AppError{err, "Failed to create data file", 500}
	}
	defer outJson.Close()

	enc := json.NewEncoder(outJson)

	enc.Encode(dockloadInfo)

	return nil
}

func List() ([]domain.LoadData, *domain.AppError) {
	l := list.New()

	files, _ := ioutil.ReadDir(config.DataPath)
	for _, f := range files {
		if f.IsDir() {
			l.PushBack(f.Name())
		}
	}

	var loadRecords = make([]domain.LoadData, l.Len())

	idx := 0
	for e := l.Front(); e != nil; e = e.Next() {

		loadId := e.Value.(string)

		inJson, err := os.Open(config.DataPath + "/" + loadId + ".json")
		if err != nil {
			return []domain.LoadData{}, &domain.AppError{err, "Error opening data store", 500}
		}
		defer inJson.Close()

		decode := json.NewDecoder(inJson)
		var loadData domain.LoadData
		err = decode.Decode(&loadData)
		if err != nil {
			return []domain.LoadData{}, &domain.AppError{err, "Error opening data store", 500}
		}

		loadRecords[idx] = loadData
		idx += 1
	}

	return loadRecords, nil
}

// find docloadId by dockloadId, name and version
// tag can be in the form:
//    123456789012
//    test
//    test:1
func Find(tag string) {

}
