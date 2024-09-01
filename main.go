package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

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
			password := os.Getenv("ADMINPASSWORD")
			u, err := NewUser("root", "root", password, AdminRole)
			if err != nil {
				log.Fatal(err)
			}

			err = db.CreateUser(u)
			if err != nil {
				fmt.Println(err)
			}
		}
	}

	addr := os.Getenv("ADDR")
	s := NewAPIServer(addr, db)
	log.Fatal(s.Run())

}
