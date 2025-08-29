package auth

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"phaint/internal/services"
	"phaint/models"

	"google.golang.org/api/iterator"
)

type ProjectHandler struct{}

func (p *ProjectHandler) getProjects(w http.ResponseWriter, r *http.Request) {
	client := services.FirebaseDb().GetClient()
	ctx := context.Background()

	uid := ""
	if r.ContentLength > 0 {
		newUid, err := models.GetUidFromRequest(r)
		if err != nil {
			log.Println(err)
		}
		uid = newUid
	}

	iter := client.Collection("projects").Documents(ctx)
	defer iter.Stop()

	var arr []map[string]interface{}
	for {
		doc, err := iter.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			log.Println("Err during collection iteration")
		}
		if uid == "" || doc.Data()["UID"] == uid {
			arr = append(arr, doc.Data())
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(arr)
}

func (p *ProjectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodGet && r.URL.Path == "/projects/get":
		p.getProjects(w, r)
		return
	case r.Method == http.MethodPost && r.URL.Path == "projects/add":
		log.Println("Add ph")
		return
	}
}
