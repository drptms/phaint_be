package models

type User struct {
	Id       string
	Username string
	Mail     string
	Password string
}

func NewUser(username, mail, password string) (User, error) {
	user := User{
		Username: username,
		Mail:     mail,
		Password: password,
	}
	return user, nil
}
