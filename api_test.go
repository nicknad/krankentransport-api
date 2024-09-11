package main

import (
	jwt "github.com/golang-jwt/jwt/v4"
	"testing"
)

func TestJWT(t *testing.T) {
	u, err := NewUser("test.email@email.com", "Test", "TestPassword", AdminRole)

	if err != nil {
		t.Fatalf(`Err %v`, err)
	}

	token, err := createJWT(u)

	if err != nil {
		t.Fatalf(`Err %v`, err)
	}

	jT, err := validateJWT(token)

	if err != nil {
		t.Fatalf(`Err %v`, err)
	}

	if !jT.Valid {
		t.Fatal("Invalid token")
	}

	claims := jT.Claims.(jwt.MapClaims)

	mail := claims["mail"].(string)

	if u.Email != mail {
		t.Fatal("Not equal mails")
	}
}
