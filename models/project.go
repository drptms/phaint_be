package models

import (
	"encoding/json"
	"net/http"
	"phaint/internal/utils"
)

type Project struct {
	Uid          string
	Pid          string
	ProjectName  string
	CreationDate string
}

func GetProjectFromRequest(r *http.Request) (Project, error) {
	type NoPidProject struct {
		Uid          string
		ProjectName  string
		CreationDate string
	}

	decoder := json.NewDecoder(r.Body)
	var project NoPidProject
	err := decoder.Decode(&project)
	if err != nil {
		return Project{}, err
	}
	return Project{
		Pid:          utils.GenerateRandomString(32),
		Uid:          project.Uid,
		ProjectName:  project.ProjectName,
		CreationDate: project.CreationDate,
	}, nil
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
