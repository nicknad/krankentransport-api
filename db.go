package main

import (
	"database/sql"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type DataBase interface {
	GetKrankenfahrten() (*[]Krankenfahrt, error)
	GetKrankenfahrt(id int) (*Krankenfahrt, error)
	UpdateKrankenfahrt(Krankenfahrt) error
	DeleteKrankenfahrt(id int) error
	CreateKrankenfahrt(desc string) (*Krankenfahrt, error)
	GetUsers() (*[]User, error)
	GetUser(email string) (*User, error)
	DeleteUser(id int) error
	CreateUser(*User) error
}

type SQLiteDatebase struct {
	db *sql.DB
}

func NewSQLiteDatabase() (*SQLiteDatebase, error) {

	path := os.Getenv("DATABASE")

	db, err := sql.Open("sqlite3", path)

	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &SQLiteDatebase{
		db: db,
	}, nil
}

func (s *SQLiteDatebase) GetKrankenfahrten() (*[]Krankenfahrt, error) {
	results, err := s.db.Query("SELECT id, description, createdAt, acceptedBy, acceptedAt, finished FROM krankenfahrten;")
	if err != nil {
		return nil, err
	}
	defer results.Close()

	var fahrten []Krankenfahrt
	for results.Next() {
		var fahrt Krankenfahrt
		var createdAt int64
		var acceptedAt sql.NullInt64 // Handles NULL values
		var acceptedBy sql.NullString

		if err := results.Scan(&fahrt.Id, &fahrt.Description, &createdAt, &acceptedBy, &acceptedAt, &fahrt.Finished); err != nil {
			return nil, err
		}

		fahrt.CreatedAt = time.Unix(createdAt, 0)
		if acceptedAt.Valid {
			acceptedAtTime := time.Unix(acceptedAt.Int64, 0)
			fahrt.AcceptedAt = &acceptedAtTime
		} else {
			fahrt.AcceptedAt = nil
		}

		if acceptedBy.Valid {
			fahrt.AcceptedBy = &acceptedBy.String
		} else {
			fahrt.AcceptedBy = nil
		}

		fahrten = append(fahrten, fahrt)
	}

	return &fahrten, nil
}

func (s *SQLiteDatebase) GetKrankenfahrt(id int) (*Krankenfahrt, error) {

	stmt, err := s.db.Prepare("SELECT id, description, createdAt, acceptedBy, acceptedAt, finished FROM krankenfahrten WHERE id = ?;")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var fahrt Krankenfahrt
	var createdAt int64
	var acceptedAt sql.NullInt64 // Handles NULL values
	var acceptedBy sql.NullString

	if err := stmt.QueryRow(id).Scan(&fahrt.Id, &fahrt.Description, &createdAt, &acceptedBy, &acceptedAt, &fahrt.Finished); err != nil {
		return nil, err
	}

	fahrt.CreatedAt = time.Unix(createdAt, 0)
	if acceptedAt.Valid {
		acceptedAtTime := time.Unix(acceptedAt.Int64, 0)
		fahrt.AcceptedAt = &acceptedAtTime
	} else {
		fahrt.AcceptedAt = nil
	}

	if acceptedBy.Valid {
		fahrt.AcceptedBy = &acceptedBy.String
	} else {
		fahrt.AcceptedBy = nil
	}

	return &fahrt, nil
}

func (s *SQLiteDatebase) DeleteKrankenfahrt(id int) error {
	stmt, err := s.db.Prepare("DELETE FROM krankenfahrten WHERE id = ?;")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(id)
	if err != nil {
		return err
	}

	return nil
}

func (s *SQLiteDatebase) CreateKrankenfahrt(desc string) (*Krankenfahrt, error) {
	stmt, err := s.db.Prepare("INSERT INTO krankenfahrten (description, createdAt, finished) VALUES (?, ?, ?) RETURNING id, description, finished;")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	unixTime := time.Now().Unix()
	var createdFahrt Krankenfahrt
	err = stmt.QueryRow(desc, unixTime, false).Scan(&createdFahrt.Id, &createdFahrt.Description, &createdFahrt.Finished)
	if err != nil {
		return nil, err
	}

	return &createdFahrt, nil
}

func (s *SQLiteDatebase) GetUsers() (*[]User, error) {
	results, err := s.db.Query("SELECT id, email, name, role FROM users;")
	if err != nil {
		return nil, err
	}
	defer results.Close()

	var users []User
	for results.Next() {
		var user User

		if err := results.Scan(&user.Id, &user.Email, &user.Name, &user.Role); err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	return &users, nil
}

func (s *SQLiteDatebase) GetUser(email string) (*User, error) {
	stmt, err := s.db.Prepare("SELECT email, name, passwordhash, role FROM users WHERE email = ?;")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var foundUser User
	err = stmt.QueryRow(email).Scan(&foundUser.Email, &foundUser.Name, &foundUser.PasswordHash, &foundUser.Role)
	if err != nil {
		return nil, err
	}

	return &foundUser, nil
}

func (s *SQLiteDatebase) CreateUser(u *User) error {
	stmt, err := s.db.Prepare("INSERT INTO users (email, name, passwordhash, role) VALUES (?, ?, ?, ?) RETURNING id, email, name, role;")
	if err != nil {
		return err
	}
	defer stmt.Close()

	var createdUser User
	err = stmt.QueryRow(u.Email, u.Name, u.PasswordHash, u.Role).Scan(&createdUser.Id, &createdUser.Email, &createdUser.Name, &createdUser.Role)
	if err != nil {
		return err
	}

	return nil
}

func (s *SQLiteDatebase) DeleteUser(id int) error {
	stmt, err := s.db.Prepare("DELETE FROM users WHERE id = ?;")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(id)
	if err != nil {
		return err
	}

	return nil
}

func (s *SQLiteDatebase) UpdateKrankenfahrt(k Krankenfahrt) error {
	stmt, err := s.db.Prepare("Update krankenfahrten SET description = ?,acceptedAt = ?, acceptedBy = ?, finished = ? WHERE id = ?;")

	if err != nil {
		return err
	}

	defer stmt.Close()

	timeInt := k.AcceptedAt.Unix()
	_, err = stmt.Exec(k.Description, timeInt, k.AcceptedBy, k.Finished, k.Id)

	if err != nil {
		return err
	}

	return nil
}

func (s *SQLiteDatebase) CreateDB() {

	stmt, err := s.db.Prepare("CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY, email TEXT, name TEXT, passwordhash TEXT, role TEXT);")

	if err != nil {
		panic(err)
	}
	stmt.Exec()

	// add a unique constraint to the email column if it does not exists
	stmt, err = s.db.Prepare("CREATE UNIQUE INDEX IF NOT EXISTS user_unique ON users (email);")
	if err != nil {
		panic(err)
	}
	stmt.Exec()

	stmt, err = s.db.Prepare("CREATE TABLE IF NOT EXISTS krankenfahrten (id INTEGER PRIMARY KEY, description TEXT, createdAt INTEGER DEFAULT (strftime('%s', 'now')), acceptedBy TEXT, acceptedAt INTEGER, finished BOOLEAN DEFAULT 0);")

	if err != nil {
		panic(err)
	}
	stmt.Exec()
}
