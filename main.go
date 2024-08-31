package main

import (
	"flag"
	"log"
)

func main() {
	seed := flag.Bool("seed", false, "seed the db")
	flag.Parse()

	db, err := NewSQLiteDatabase()

	if err != nil {
		log.Fatal(err)
	}

	if *seed {
		db.CreateDB()
		_, err := db.GetUser("root")
		if err != nil {
			db.CreateUser("root", "root", "password", AdminRole)
		}
	}

	s := NewAPIServer(":2209")
	s.Run()
}
