// Package store contains the MySQL persistence layer. It performs no business validation;
// that is the responsibility of the service layer.
package store

import (
	"authorAndBooks/model"
	"context"
	"database/sql"
	"fmt"
)

// Author persists authors.
type Author interface {
	GetByID(ctx context.Context, id int) (model.Author, error)
	// FindID returns the ID of the author whose details all match a, if any.
	FindID(ctx context.Context, a model.Author) (id int, found bool, err error)
	Create(ctx context.Context, a model.Author) (int, error)
	Update(ctx context.Context, a model.Author) error
	Delete(ctx context.Context, id int) error
}

// Book persists books. Only the author ID of a book is stored and returned.
type Book interface {
	// GetAll returns all books, filtered by title when title is not empty.
	GetAll(ctx context.Context, title string) ([]model.Book, error)
	GetByID(ctx context.Context, id int) (model.Book, error)
	// FindID returns the ID of the book with the given title and author, if any.
	FindID(ctx context.Context, title string, authorID int) (id int, found bool, err error)
	Create(ctx context.Context, b model.Book) (int, error)
	Update(ctx context.Context, b model.Book) error
	Delete(ctx context.Context, id int) error
	DeleteByAuthorID(ctx context.Context, authorID int) error
}

// Transactor runs a function with stores bound to a single transaction. The transaction is
// committed when fn returns nil and rolled back otherwise.
type Transactor interface {
	WithTx(ctx context.Context, fn func(authors Author, books Book) error) error
}

// Querier is satisfied by both *sql.DB and *sql.Tx.
type Querier interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// SQLTransactor implements Transactor on top of *sql.DB.
type SQLTransactor struct {
	db *sql.DB
}

func NewTransactor(db *sql.DB) *SQLTransactor {
	return &SQLTransactor{db: db}
}

func (t *SQLTransactor) WithTx(ctx context.Context, fn func(authors Author, books Book) error) error {
	tx, err := t.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	// Rollback is a no-op once the transaction has been committed.
	defer func() { _ = tx.Rollback() }()

	if err = fn(NewAuthor(tx), NewBook(tx)); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
