package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"phaint/internal/services"
	"phaint/models"

	"cloud.google.com/go/firestore"
)

type InvitationHandler struct{}

type InvitationAcceptBody struct {
	UID        string `json:"UID"`
	InviteLink string `json:"inviteLink"`
}

func (i *InvitationHandler) getInvitations(invitation InvitationAcceptBody) (string, error) {
	client := services.FirebaseDb().GetClient()
	ctx := context.Background()

	query := client.Collection("invitations").Where("Link", "==", invitation.InviteLink).Limit(1)
	docs, err := query.Documents(ctx).GetAll()
	if err != nil {
		log.Println("Error querying document by ProjectName:", err)
		return "", err
	}
	if len(docs) == 0 {
		log.Println("No project found with ProjectName:", invitation.InviteLink)
		return "", fmt.Errorf("no project found with ProjectName: %s", invitation.InviteLink)
	}

	// Get the document snapshot
	docSnap, err := docs[0].Ref.Get(context.Background())
	if err != nil {
		return "", err
	}

	// Assuming your Firestore doc has a field "CanvasesData" which is a slice or map of canvases
	var rawData map[string]interface{}
	if err := docSnap.DataTo(&rawData); err != nil {
		return "", err
	}

	projectID, ok := rawData["ProjectID"].(string)
	if !ok {
		return "", fmt.Errorf("ProjectID field not found")
	}

	return projectID, nil
}

func (i *InvitationHandler) createInvitation(w http.ResponseWriter, r *http.Request) {
	client := services.FirebaseDb().GetClient()
	ctx := context.Background()

	invitation, err := models.GetInvitationFromRequest(r)
	if err != nil {
		log.Println(err)
	}

	_, _, err = client.Collection("invitations").Add(ctx, map[string]interface{}{
		"CreatorUID": invitation.CreatorUid,
		"Link":       invitation.Link,
		"ProjectID":  invitation.ProjectID,
		"Used":       false,
	})
	if err != nil {
		log.Println(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"data": invitation.Link,
	})
}

func (i *InvitationHandler) acceptInvitation(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	decoder := json.NewDecoder(r.Body)
	var invitation InvitationAcceptBody
	err := decoder.Decode(&invitation)

	if err != nil {
		log.Println(err)
		return
	}
	projectID, err := i.getInvitations(invitation)
	if err != nil {
		log.Println(err)
		return
	}

	docRef, err := GetProjectById(projectID)
	if err != nil {
		log.Println(err)
		return
	}

	_, err = docRef.Update(ctx, []firestore.Update{
		{
			Path:  "Collaborators",
			Value: firestore.ArrayUnion(invitation.UID),
		},
	})
}

func (i *InvitationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodPost && r.URL.Path == "/invitations/accept":
		i.acceptInvitation(w, r)
	case r.Method == http.MethodPost:
		i.createInvitation(w, r)
	}
}
