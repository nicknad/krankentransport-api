package main

import (
	"flag"
	"fmt"
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
			u, err := NewUser("root", "root", "password", AdminRole)
			if err != nil {
				log.Fatal(err)
			}

			err = db.CreateUser(u)
			if err != nil {
				fmt.Println(err)
			}
		}
	}

	s := NewAPIServer(":2209", db)
	s.Run()
}
