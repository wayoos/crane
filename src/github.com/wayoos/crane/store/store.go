package store

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"github.com/wayoos/crane/api/domain"
	"github.com/wayoos/crane/config"
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

// find docloadId by dockloadId, name and version
// tag can be in the form:
//    123456789012
//    test
//    test:1
func Find(tag string) {

}
