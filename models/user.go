package models

import (
	"encoding/json"
	"net/http"

	"firebase.google.com/go/auth"
)

type User struct {
	Uid      string
	Username string
	Mail     string
	Password string
}

func NewUser(uid, username, mail, password string) (User, error) {
	user := User{
		Uid:      uid,
		Username: username,
		Mail:     mail,
		Password: password,
	}
	return user, nil
}

func NewUserFromRequest(r *http.Request) (User, error) {
	decoder := json.NewDecoder(r.Body)
	var user User
	err := decoder.Decode(&user)
	if err != nil {
		return User{}, err
	}
	return user, nil
}

func NewFirebaseAuthUser(r *http.Request) (*auth.UserToCreate, error) {
	user := auth.UserToCreate{}
	httpUser, err := NewUserFromRequest(r)
	if err != nil {
		return &user, err
	}
	user.DisplayName(httpUser.Username)
	user.Email(httpUser.Mail)
	user.Password(httpUser.Password)
	return &user, nil
}
