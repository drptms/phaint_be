package models

import (
	"encoding/json"
	"net/http"
	"phaint/internal/utils"
)

type Invitation struct {
	CreatorUid   string `firestore:"UID" json:"UID"`
	Link         string `firestore:"link" json:"link"`
	ProjectID    string `firestore:"PID" json:"PID"`
}

func GetInvitationFromRequest(r *http.Request) (Invitation, error) {

	decoder := json.NewDecoder(r.Body)
	var invitation Invitation
	err := decoder.Decode(&invitation)
	if err != nil {
		return Invitation{}, err
	}
	return Invitation{
		Link:         utils.GenerateRandomString(32),
		CreatorUid:   invitation.CreatorUid,
		ProjectID:    invitation.ProjectID,
	}, nil
}