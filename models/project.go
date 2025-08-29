package models

import (
	"encoding/json"
	"net/http"
)

type Project struct {
	Uid string
	Pid string
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
