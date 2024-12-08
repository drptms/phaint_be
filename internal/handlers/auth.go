package auth

import (
	"context"
	"fmt"
	"phaint/internal/services"
	crypto "phaint/internal/utils"
	"phaint/models"
)

func RegisterUser(username string, mail string, password string) (string, error) {
	cryptedPassword, err := crypto.Encrypt(username)
	if err != nil {
		fmt.Println(err)
	}
	user, err := models.NewUser(username, mail, cryptedPassword)
	if err != nil {
		fmt.Println(err)
	}

	client := services.FirebaseDb().GetClient()
	ref := client.Collection("users").NewDoc()
	result, err := ref.Set(context.Background(), map[string]interface{}{
		"username": user.Username,
		"mail":     user.Mail,
		"password": user.Password,
	})
	user.Id = ref.ID
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	fmt.Printf("Result = %v\n", result)
	return user.Id, nil
}
