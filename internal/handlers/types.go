package handlers

import "github.com/steveperjesi/integra-demo/user"

type UsersResponse struct {
	Users []user.User `json:"users"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
