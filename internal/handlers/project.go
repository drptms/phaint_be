package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"phaint/internal/services"
	"phaint/models"

	"cloud.google.com/go/firestore"
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

func (p *ProjectHandler) addProject(w http.ResponseWriter, r *http.Request) {
	client := services.FirebaseDb().GetClient()
	ctx := context.Background()

	project, err := models.GetProjectFromRequest(r)
	if err != nil {
		log.Println(err)
	}

	_, _, err = client.Collection("projects").Add(ctx, map[string]interface{}{
		"UID":          project.Uid,
		"PID":          project.Pid,
		"ProjectName":  project.ProjectName,
		"CreationDate": project.CreationDate,
	})
	if err != nil {
		log.Println(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(project.Pid)
}

func (p *ProjectHandler) getProjectByName(name string) (*firestore.DocumentRef, error) {
    ctx := context.Background()
    client := services.FirebaseDb().GetClient()

	// Query to find the document where ProjectName field matches name
    query := client.Collection("projects").Where("ProjectName", "==", name).Limit(1)
    docs, err := query.Documents(ctx).GetAll()
    if err != nil {
        log.Println("Error querying document by ProjectName:", err)
        return nil, err
    }
    if len(docs) == 0 {
        log.Println("No project found with ProjectName:", name)
        return nil, fmt.Errorf("no project found with ProjectName: %s", name)
    }

    return docs[0].Ref, nil
}

func (p *ProjectHandler) updateProjectCanvasesData(hub *Hub) error {
	ctx := context.Background()

    docRef, err := p.getProjectByName(hub.projectID)
    if err != nil {
        return err
    }

    // Get all canvases from the CanvasService
    canvases := hub.workBoard.GetAllCanvases()

    // Update "CanvasesData" field with current canvases
    _, err = docRef.Update(ctx, []firestore.Update{
        {
            Path:  "CanvasesData",
            Value: canvases,
        },
    })

    if err != nil {
        log.Println("Failed to update CanvasesData:", err)
        return err
    }

    return nil
}

func (p *ProjectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodGet:
		p.getProjects(w, r)
		return
	case r.Method == http.MethodPost:
		p.addProject(w, r)
		return
	}
}
