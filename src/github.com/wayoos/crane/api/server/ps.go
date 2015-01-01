package server

import (
	"github.com/martini-contrib/render"
	"github.com/wayoos/crane/store"
)

func Ps(r render.Render) {
	loadRecords, appErr := store.List()
	if appErr != nil {
		r.Error(appErr.Code)
	}
	r.JSON(200, loadRecords)
}
