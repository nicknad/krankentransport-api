package main

import (
	"testing"
	"time"
)

func TestDBUser(t *testing.T) {
	u, err := NewUser("test@email.com", "TestUser", "123456", "User")

	if err != nil {
		t.Fatalf(`Err %v`, err)
	}

	db, err := NewSQLiteDatabase()

	if err != nil {
		t.Fatalf(`Err %v`, err)
	}

	err = db.CreateUser(u)

	if err != nil {
		t.Fatalf(`Err %v`, err)
	}

	u2, err := db.GetUser(u.Email)

	if err != nil {
		t.Fatalf(`Err %v`, err)
	}

	if u2.Email != u.Email {
		t.Fatalf(`Emails dont match`)
	}

	err = db.DeleteUser(u2.Id)

	if err != nil {
		t.Fatalf(`Err %v`, err)
	}
}

func TestDBKrankenfahrt(t *testing.T) {
	db, err := NewSQLiteDatabase()

	if err != nil {
		t.Fatalf(`Err %v`, err)
	}

	k, err := db.CreateKrankenfahrt("Krankenfahrt Test")

	if err != nil {
		t.Fatalf(`Err %v`, err)
	}

	userName := "Test User"
	time := time.Now()
	k.AcceptedBy = &userName
	k.AcceptedAt = &time

	err = db.UpdateKrankenfahrt(*k)

	if err != nil {
		t.Fatalf(`Err %v`, err)
	}

	k2, err := db.GetKrankenfahrt(k.Id)

	if err != nil {
		t.Fatalf(`Err %v`, err)
	}

	if *k2.AcceptedBy != userName {

		t.Fatalf("AcceptedBy Dont match")
	}

	err = db.DeleteKrankenfahrt(k2.Id)

	if err != nil {
		t.Fatalf(`Err %v`, err)
	}
}
