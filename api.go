package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v4"
)

type ApiError struct {
	Error string `json:"error"`
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

func (s *APIServer) Run() error {

	router := gin.Default()

	// Ping test
	router.GET("/ping", func(c *gin.Context) {

		c.String(http.StatusOK, "pong")
	})

	router.POST("/login", s.handleLogin)

	authGroup := router.Group("/api")
	authGroup.Use(authenticate(s))
	{
		authGroup.GET("/users", s.getUsers)
		authGroup.GET("/krankenfahrten", s.getKrankenfahrten)
		authGroup.PATCH("/krankenfart", s.updateKrankenFahrt)

		authGroup.Use(isAdmin())
		{
			authGroup.DELETE("/user", s.delteUser)
			authGroup.DELETE("/krankenfahrt", s.deleteKrankenFahrt)
			authGroup.POST("/user", s.createUser)
			authGroup.POST("/krankenfahrt", s.createKrankenFahrt)
		}
	}

	log.Println("JSON Api server running on port: ", s.listenAddr)
	return http.ListenAndServe(s.listenAddr, router)
}

func isAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		u, exists := c.Get("user")

		if !exists {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		if u.(*User).Role != AdminRole {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
	}
}

func authenticate(s *APIServer) gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Println("calling JWT auth middleware")
		tokenString := c.Request.Header.Get("x-jwt-token")
		token, err := validateJWT(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ApiError{Error: "permission denied"})
			return
		}

		if !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ApiError{Error: "invalid token"})
			return
		}

		claims := token.Claims.(jwt.MapClaims)
		mail := claims["mail"].(string)
		u, err := s.db.GetUser(mail)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ApiError{Error: "invalid token"})
			return
		}

		c.Set("user", u)
	}
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

func (s *APIServer) handleLogin(c *gin.Context) {
	var req LoginRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	u, err := s.db.GetUser(req.Email)

	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	if !u.ValidPassword(req.Password) {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	token, err := createJWT(u)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	resp := LoginResponse{
		Token: token,
		Name:  u.Name,
	}

	c.JSON(http.StatusOK, resp)
}

func validateJWT(tokenString string) (*jwt.Token, error) {
	secret := os.Getenv("JWT_SECRET")

	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(secret), nil
	})
}

func (s *APIServer) getUsers(c *gin.Context) {
	u, err := s.db.GetUsers()

	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, u)
}

func (s *APIServer) getKrankenfahrten(c *gin.Context) {
	k, err := s.db.GetKrankenfahrten()

	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, k)
}

func (s *APIServer) createKrankenFahrt(c *gin.Context) {
}

func (s *APIServer) createUser(c *gin.Context) {
}

func (s *APIServer) updateKrankenFahrt(c *gin.Context) {
}

func (s *APIServer) deleteKrankenFahrt(c *gin.Context) {
}

func (s *APIServer) delteUser(c *gin.Context) {
}
