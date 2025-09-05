package auth

import (
	"net/http"
)

type WorkBoardHandler struct{}


func (w *WorkBoardHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	switch {
	case request.Method == http.MethodGet && request.URL.Path == "/workboard":
		return
	case request.Method == http.MethodPost && request.URL.Path == "/workboard":
		return
	default:
        response.WriteHeader(http.StatusMethodNotAllowed)
    }
}