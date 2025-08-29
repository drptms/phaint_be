package models

import (
	"encoding/json"
	"net/http"
)

type Project struct {
	Uid string
	Pid string
}

func GetProjectFromRequest(r *http.Request) (Project, error) {
	decoder := json.NewDecoder(r.Body)
	var project Project
	err := decoder.Decode(&project)
	if err != nil {
		return Project{}, err
	}
	return project, nil
}

func GetUidFromRequest(r *http.Request) (string, error) {
	type UidStruct struct {
		Uid string
	}

	decoder := json.NewDecoder(r.Body)
	var uid UidStruct
	err := decoder.Decode(&uid)
	if err != nil {
		return "", err
	}
	return uid.Uid, nil
}
