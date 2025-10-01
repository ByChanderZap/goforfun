package models

import (
	"database/sql"
	"errors"
	"time"
)

type Snippet struct {
	ID      int
	Title   string
	Content string
	Created time.Time
	Expires time.Time
}

type SnippetModel struct {
	DB *sql.DB
}

const stmt = `
	INSERT INTO snippets (title, content, created, expires)
	VALUES(?, ?, UTC_TIMESTAMP(), DATE_ADD(UTC_TIMESTAMP(), INTERVAL ? DAY))
	`

type InsertSnippetParams struct {
	Title   string
	Content string
	Expires int
}

func (m *SnippetModel) Insert(params InsertSnippetParams) (int, error) {
	result, err := m.DB.Exec(
		stmt,
		params.Title,
		params.Content,
		params.Expires,
	)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

const stmtGet = `
	SELECT id, title, content, created, expires FROM snippets
	WHERE expires > UTC_TIMESTAMP() AND id = ?
	`

func (m *SnippetModel) Get(id int) (Snippet, error) {
	var s Snippet

	err := m.DB.QueryRow(stmtGet, id).Scan(
		&s.ID,
		&s.Title,
		&s.Content,
		&s.Created,
		&s.Expires,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Snippet{}, ErrNoRecord
		}
		return Snippet{}, err
	}

	return s, nil
}

const stmtGetLastTen = `
SELECT id, title, content, created, expires FROM snippets
	WHERE expires > UTC_TIMESTAMP() ORDER BY id DESC LIMIT 10
	`

func (m *SnippetModel) Latest() ([]Snippet, error) {
	var snippets []Snippet

	rows, err := m.DB.Query(stmtGetLastTen)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var s Snippet
		err := rows.Scan(
			&s.ID,
			&s.Title,
			&s.Content,
			&s.Created,
			&s.Expires,
		)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, ErrNoRecord
			}
			return nil, err
		}
		snippets = append(snippets, s)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return snippets, nil
}
