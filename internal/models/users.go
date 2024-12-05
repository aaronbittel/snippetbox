package models

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID             int
	Name           string
	Email          string
	HashedPassword []byte
	Created        time.Time
}

type UserModel struct {
	DB *sql.DB
}

func (m *UserModel) Insert(name, email, password string) error {
	stmt := `INSERT INTO users (name, email, hashed_password, created)
	VALUES (?, ?, ?, NOW())`

	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}

	_, err = m.DB.Exec(stmt, name, email, hash)
	if err != nil {

		var mySQLError *mysql.MySQLError
		if errors.As(err, &mySQLError) {
			if mySQLError.Number == 1062 &&
				strings.Contains(mySQLError.Message, "users_uc_email") {
				return ErrDuplicateEmail
			}
		}
		return err
	}

	return nil
}

func (m *UserModel) Authenticate(email, password string) (int, error) {
	stmt := `SELECT id, hashed_password FROM users WHERE email = ?`

	var (
		id             int
		hashedPassword []byte
	)

	row := m.DB.QueryRow(stmt, email)
	if err := row.Scan(&id, &hashedPassword); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrInvalidCredentials
		} else {
			return 0, err
		}
	}

	err := bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return 0, ErrInvalidCredentials
		} else {
			return 0, err
		}
	}

	return id, nil
}

func (m *UserModel) Exists(id int) (bool, error) {
	stmt := `SELECT id FROM users WHERE id = ?`

	row := m.DB.QueryRow(stmt, id)

	if err := row.Scan(); err != nil {
		return false, err
	}

	return true, nil
}
