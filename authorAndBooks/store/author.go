package store

import (
	"authorAndBooks/apperrors"
	"authorAndBooks/model"
	"context"
	"database/sql"
	"errors"
	"fmt"
)

const (
	getAuthorByID    = "SELECT id, first_name, last_name, dob, pen_name FROM Author WHERE id = ?"
	findAuthorID     = "SELECT id FROM Author WHERE first_name = ? AND last_name = ? AND dob = ? AND pen_name = ?"
	insertAuthor     = "INSERT INTO Author (first_name, last_name, dob, pen_name) VALUES (?, ?, ?, ?)"
	updateAuthor     = "UPDATE Author SET first_name = ?, last_name = ?, dob = ?, pen_name = ? WHERE id = ?"
	deleteAuthorByID = "DELETE FROM Author WHERE id = ?"
)

// AuthorSQL implements Author for MySQL.
type AuthorSQL struct {
	db Querier
}

func NewAuthor(db Querier) *AuthorSQL {
	return &AuthorSQL{db: db}
}

func (s *AuthorSQL) GetByID(ctx context.Context, id int) (model.Author, error) {
	var a model.Author

	err := s.db.QueryRowContext(ctx, getAuthorByID, id).Scan(&a.ID, &a.FirstName, &a.LastName, &a.Dob, &a.PenName)
	if errors.Is(err, sql.ErrNoRows) {
		return model.Author{}, apperrors.NotFound{Entity: "author", ID: id}
	}

	if err != nil {
		return model.Author{}, fmt.Errorf("get author %d: %w", id, err)
	}

	return a, nil
}

func (s *AuthorSQL) FindID(ctx context.Context, a model.Author) (id int, found bool, err error) {
	err = s.db.QueryRowContext(ctx, findAuthorID, a.FirstName, a.LastName, a.Dob, a.PenName).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, false, nil
	}

	if err != nil {
		return 0, false, fmt.Errorf("find author: %w", err)
	}

	return id, true, nil
}

func (s *AuthorSQL) Create(ctx context.Context, a model.Author) (int, error) {
	res, err := s.db.ExecContext(ctx, insertAuthor, a.FirstName, a.LastName, a.Dob, a.PenName)
	if err != nil {
		return 0, fmt.Errorf("insert author: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("insert author: %w", err)
	}

	return int(id), nil
}

func (s *AuthorSQL) Update(ctx context.Context, a model.Author) error {
	if _, err := s.db.ExecContext(ctx, updateAuthor, a.FirstName, a.LastName, a.Dob, a.PenName, a.ID); err != nil {
		return fmt.Errorf("update author %d: %w", a.ID, err)
	}

	return nil
}

func (s *AuthorSQL) Delete(ctx context.Context, id int) error {
	if _, err := s.db.ExecContext(ctx, deleteAuthorByID, id); err != nil {
		return fmt.Errorf("delete author %d: %w", id, err)
	}

	return nil
}
