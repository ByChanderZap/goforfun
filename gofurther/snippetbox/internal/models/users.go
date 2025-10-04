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

type InsertUserParams struct {
	Name     string
	Email    string
	Password string
}

func (m *UserModel) Insert(params InsertUserParams) error {
	const stmt = `
	INSERT INTO users (name, email, hashed_password, created)
	VALUES(?, ?, ?, UTC_TIMESTAMP())
	`
	_, err := m.DB.Exec(
		stmt,
		params.Name,
		params.Email,
		params.Password,
	)
	if err != nil {
		var mySQLError *mysql.MySQLError

		if errors.As(err, &mySQLError) {
			if mySQLError.Number == 1062 && strings.Contains(mySQLError.Message, "users_uc_email") {
				return ErrDuplicatedEmail
			}
		}

		return err
	}

	return nil
}

type AuthenticateUserParams struct {
	Email    string
	Password string
}

const stmtAuthenticateQuery = `
	SELECT id, hashed_password
	FROM users
	WHERE email = ?
	`

func (m *UserModel) Authenticate(params AuthenticateUserParams) (int, error) {
	var id int
	var hashedPassword []byte

	err := m.DB.QueryRow(stmtAuthenticateQuery, params.Email).Scan(
		&id,
		&hashedPassword,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrInvalidCredentials
		}
		return 0, err
	}

	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(params.Password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return 0, ErrInvalidCredentials
		}
		return 0, err
	}

	return id, nil
}

type ExistsParams struct {
	ID string
}

func (m *UserModel) Exists(params ExistsParams) (bool, error) {
	return true, nil
}
