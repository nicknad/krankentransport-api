package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")

	return json.NewEncoder(w).Encode(v)
}

type apiFunc func(http.ResponseWriter, *http.Request) error

type ApiError struct {
	Error string `json:"error"`
}

func makeHttpHandleFunc(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
	}
}

type APIServer struct {
	listenAddr string
}

func NewAPIServer(addr string) *APIServer {
	return &APIServer{
		listenAddr: addr,
	}
}

func (s *APIServer) Run() {
	router := mux.NewRouter()

	router.HandleFunc("/api/user", makeHttpHandleFunc(s.handleGetUsers))

	log.Println("JSON Api server running on port: ", s.listenAddr)
	http.ListenAndServe(s.listenAddr, router)
}

func (s *APIServer) handleUsers(w http.ResponseWriter, r *http.Request) error {

	if r.Method == "GET" {
		return s.handleGetUsers(w, r)
	}

	if r.Method == "POST" {
		return s.handleCreateUser(w, r)
	}

	if r.Method == "DELETE" {
		return s.handleDeleteUser(w, r)
	}

	return fmt.Errorf("method not allowed")
}

func (s *APIServer) handleKrankfahrten(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		return s.handleGetKrankenfahrten(w, r)
	}

	if r.Method == "POST" {
		return s.handleCreateKrankenfahrt(w, r)
	}

	if r.Method == "DELETE" {
		return s.handleDeleteKrankenfahrt(w, r)
	}

	return fmt.Errorf("method not allowed")
}

func (s *APIServer) handleGetUsers(w http.ResponseWriter, r *http.Request) error {

	return nil
}

func (s *APIServer) handleGetKrankenfahrten(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (s *APIServer) handleCreateKrankenfahrt(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (s *APIServer) handleCreateUser(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (s *APIServer) handleDeleteUser(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (s *APIServer) handleDeleteKrankenfahrt(w http.ResponseWriter, r *http.Request) error {
	return nil
}
