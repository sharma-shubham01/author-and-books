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
	selectBooks         = "SELECT id, title, publication, published_date, author_id FROM Book"
	getAllBooks         = selectBooks + " ORDER BY id"
	getBooksByTitle     = selectBooks + " WHERE title = ? ORDER BY id"
	getBookByID         = selectBooks + " WHERE id = ?"
	findBookID          = "SELECT id FROM Book WHERE title = ? AND author_id = ?"
	insertBook          = "INSERT INTO Book (title, publication, published_date, author_id) VALUES (?, ?, ?, ?)"
	updateBook          = "UPDATE Book SET title = ?, publication = ?, published_date = ?, author_id = ? WHERE id = ?"
	deleteBookByID      = "DELETE FROM Book WHERE id = ?"
	deleteBooksByAuthor = "DELETE FROM Book WHERE author_id = ?"
)

// BookSQL implements Book for MySQL.
type BookSQL struct {
	db Querier
}

func NewBook(db Querier) *BookSQL {
	return &BookSQL{db: db}
}

type scanner interface {
	Scan(dest ...any) error
}

func scanBook(row scanner) (model.Book, error) {
	var b model.Book
	err := row.Scan(&b.ID, &b.Title, &b.Publication, &b.PublishedDate, &b.Author.ID)

	return b, err
}

func (s *BookSQL) GetAll(ctx context.Context, title string) ([]model.Book, error) {
	query, args := getAllBooks, []any(nil)
	if title != "" {
		query, args = getBooksByTitle, []any{title}
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list books: %w", err)
	}
	defer rows.Close()

	books := []model.Book{}

	for rows.Next() {
		b, err := scanBook(rows)
		if err != nil {
			return nil, fmt.Errorf("scan book: %w", err)
		}

		books = append(books, b)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("list books: %w", err)
	}

	return books, nil
}

func (s *BookSQL) GetByID(ctx context.Context, id int) (model.Book, error) {
	b, err := scanBook(s.db.QueryRowContext(ctx, getBookByID, id))
	if errors.Is(err, sql.ErrNoRows) {
		return model.Book{}, apperrors.NotFound{Entity: "book", ID: id}
	}

	if err != nil {
		return model.Book{}, fmt.Errorf("get book %d: %w", id, err)
	}

	return b, nil
}

func (s *BookSQL) FindID(ctx context.Context, title string, authorID int) (id int, found bool, err error) {
	err = s.db.QueryRowContext(ctx, findBookID, title, authorID).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, false, nil
	}

	if err != nil {
		return 0, false, fmt.Errorf("find book: %w", err)
	}

	return id, true, nil
}

func (s *BookSQL) Create(ctx context.Context, b model.Book) (int, error) {
	res, err := s.db.ExecContext(ctx, insertBook, b.Title, b.Publication, b.PublishedDate, b.Author.ID)
	if err != nil {
		return 0, fmt.Errorf("insert book: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("insert book: %w", err)
	}

	return int(id), nil
}

func (s *BookSQL) Update(ctx context.Context, b model.Book) error {
	if _, err := s.db.ExecContext(ctx, updateBook, b.Title, b.Publication, b.PublishedDate, b.Author.ID, b.ID); err != nil {
		return fmt.Errorf("update book %d: %w", b.ID, err)
	}

	return nil
}

func (s *BookSQL) Delete(ctx context.Context, id int) error {
	if _, err := s.db.ExecContext(ctx, deleteBookByID, id); err != nil {
		return fmt.Errorf("delete book %d: %w", id, err)
	}

	return nil
}

func (s *BookSQL) DeleteByAuthorID(ctx context.Context, authorID int) error {
	if _, err := s.db.ExecContext(ctx, deleteBooksByAuthor, authorID); err != nil {
		return fmt.Errorf("delete books of author %d: %w", authorID, err)
	}

	return nil
}
