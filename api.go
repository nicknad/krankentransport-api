package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	jwt "github.com/golang-jwt/jwt/v4"
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
	db         DataBase
}

func NewAPIServer(addr string, db DataBase) *APIServer {
	return &APIServer{
		listenAddr: addr,
		db:         db,
	}
}

func (s *APIServer) Run() {
	router := mux.NewRouter()

	router.HandleFunc("/login", makeHttpHandleFunc(s.handleLogin))
	router.HandleFunc("/api/user", withJWTAuth(makeHttpHandleFunc(s.handleGetUsers), s.db))
	router.HandleFunc("/api/krankenfahrt", withJWTAuth(makeHttpHandleFunc(s.handleKrankfahrten), s.db))

	log.Println("JSON Api server running on port: ", s.listenAddr)
	http.ListenAndServe(s.listenAddr, router)
}

func createJWT(user *User) (string, error) {
	claims := &jwt.MapClaims{
		"expiresAt": 15000,
		"mail":      user.Email,
	}

	secret := os.Getenv("JWT_SECRET")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(secret))
}

func (s *APIServer) handleLogin(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "POST" {
		return fmt.Errorf("method not allowed %s", r.Method)
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return err
	}

	u, err := s.db.GetUser(req.Email)

	if err != nil {
		return err
	}

	if !u.ValidPassword(req.Password) {
		return fmt.Errorf("not authenticated")
	}

	token, err := createJWT(u)
	if err != nil {
		return err
	}

	resp := LoginResponse{
		Token: token,
		Name:  u.Name,
	}

	return WriteJSON(w, http.StatusOK, resp)
}

func permissionDenied(w http.ResponseWriter) {
	WriteJSON(w, http.StatusForbidden, ApiError{Error: "permission denied"})
}

func withJWTAuth(handlerFunc http.HandlerFunc, db DataBase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("calling JWT auth middleware")

		tokenString := r.Header.Get("x-jwt-token")
		token, err := validateJWT(tokenString)
		if err != nil {
			permissionDenied(w)
			return
		}
		if !token.Valid {
			permissionDenied(w)
			return
		}

		claims := token.Claims.(jwt.MapClaims)
		mail := claims["mail"].(string)
		_, err = db.GetUser(mail)

		if err != nil {
			WriteJSON(w, http.StatusForbidden, ApiError{Error: "invalid token"})
			return
		}

		handlerFunc(w, r)
	}
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

func validateJWT(tokenString string) (*jwt.Token, error) {
	secret := os.Getenv("JWT_SECRET")

	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return []byte(secret), nil
	})
}

func (s *APIServer) handleGetUsers(w http.ResponseWriter, r *http.Request) error {
	u, err := s.db.GetUsers()

	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, u)
}

func (s *APIServer) handleGetKrankenfahrten(w http.ResponseWriter, r *http.Request) error {
	k, err := s.db.GetKrankenfahrten()

	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, k)
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
