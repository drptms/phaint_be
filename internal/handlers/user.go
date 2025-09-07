package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"phaint/internal/services"
	"phaint/internal/utils"
	"phaint/models"

	"firebase.google.com/go/auth"
	"google.golang.org/api/iterator"
)

type UserHandler struct{}

// RegisterUser Create a new user on Firebase Authentication system from an user passed in the http request body
func (h *UserHandler) RegisterUser(w http.ResponseWriter, r *http.Request, userReq models.User) {
	app := services.FirebaseDb().GetApp()

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)
	}
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	var user *auth.UserToCreate
	user, err = models.NewFirebaseAuthUser(r)

	ctx := context.Background()
	client := services.FirebaseDb().GetClient()

	userId := userReq.Uid
	if len(userId) == 0 {
		userId = generateUserID()
	}

	_, _, err = client.Collection("users").Add(ctx, map[string]interface{}{
		"UID":      userId,
		"mail":     userReq.Mail,
		"username": userReq.Username,
	})

	if err != nil {
		fmt.Println(err)
	}
	authx, err := app.Auth(ctx)
	if err != nil {
		fmt.Println(err)
	}
	_, err = authx.CreateUser(ctx, user)
	if err != nil {
		log.Fatalf("Error during user craetion : %v\n", err)
	}

	_, err = utils.SignInWithPassword(userReq.Mail, userReq.Password)
	if err != nil {
		fmt.Println(err)
	}

	writeResponse(w, map[string]string{"userId": userId, "username": userReq.Username})
}

// LoginUser Authenticate a user passed in the http request body
func (h *UserHandler) LoginUser(w http.ResponseWriter, r *http.Request, user models.User) {
	_, err := utils.SignInWithPassword(user.Mail, user.Password)
	if err != nil {
		fmt.Println(err)
	}

	ctx := context.Background()
	client := services.FirebaseDb().GetClient()
	iter := client.Collection("users").Documents(ctx)
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
		if doc.Data()["mail"] == user.Mail {
			arr = append(arr, doc.Data())
		}
	}

	writeResponse(w, map[string]string{"userId": arr[0]["UID"].(string), "username": arr[0]["username"].(string)})
}

func writeResponse(w http.ResponseWriter, toWrite map[string]string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(toWrite)
}

func (h *UserHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)
	}
	var user models.User
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	user, err = models.NewUserFromRequest(r)
	if err != nil {
		fmt.Println(err)
	}

	switch {
	case r.Method == http.MethodPost && len(user.Username) > 0:
		h.RegisterUser(w, r, user)
		return
	case r.Method == http.MethodPost && len(user.Username) == 0:
		h.LoginUser(w, r, user)
		return
	}
}
