package models

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
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

func (m *UserModel) Authenticate(params AuthenticateUserParams) (int, error) {
	return 0, nil
}

type ExistsParams struct {
	ID string
}

func (m *UserModel) Exists(params ExistsParams) (bool, error) {
	return true, nil
}
