package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

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
			authGroup.DELETE("/user", s.deleteUser)
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
		tokenAuth := c.Request.Header.Get("Authorization")
		if !strings.HasPrefix(tokenAuth, "Bearer") {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		var arr = strings.SplitN(tokenAuth, " ", 2)
		token, err := validateJWT(arr[1])

		if err != nil {
			fmt.Println(err.Error())
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
	var req CreateFahrtRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	k, err := s.db.CreateKrankenfahrt(req.Description)

	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	us, err := s.db.GetUsers()

	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	for _, u := range *us {
		if u.Role == AdminRole {
			continue
		}

		fmt.Println("Neue Krankenfahrt: {s} \n Email an {s}", k.Description, u.Email)
	}

	c.JSON(http.StatusOK, "")
}

func (s *APIServer) createUser(c *gin.Context) {
	var req CreateUserRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	u, err := NewUser(req.Name, req.Email, req.Password, "User")

	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	s.db.CreateUser(u)

	c.JSON(http.StatusOK, nil)
}

func (s *APIServer) updateKrankenFahrt(c *gin.Context) {
	var req ClaimKrankenfahrtRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	k, err := s.db.GetKrankenfahrt(req.Id)

	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	u, exists := c.Get("user")

	if !exists {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	t := time.Now()
	k.AcceptedAt = &t
	k.AcceptedBy = &u.(*User).Name

	err = s.db.UpdateKrankenfahrt(k)

	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}

	c.JSON(http.StatusOK, nil)
}

func (s *APIServer) deleteKrankenFahrt(c *gin.Context) {
	str := c.Query("Id")
	if str == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	i, err := strconv.Atoi(str)

	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	err = s.db.DeleteKrankenfahrt(i)

	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, "")
}

func (s *APIServer) deleteUser(c *gin.Context) {
	str := c.Query("Id")
	if str == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	i, err := strconv.Atoi(str)

	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	err = s.db.DeleteUser(i)

	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, "")

}
