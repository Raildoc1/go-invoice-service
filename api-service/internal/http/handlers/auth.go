package handlers

import "net/http"

type Auth struct{}

func NewAuth() *Auth {
	return &Auth{}
}

func (h *Auth) Login(rw http.ResponseWriter, r *http.Request)    {}
func (h *Auth) Register(rw http.ResponseWriter, r *http.Request) {}
