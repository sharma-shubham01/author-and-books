package store

import (
	"authorAndBooks/apperrors"
	"authorAndBooks/model"
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newMock(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() {
		assert.NoError(t, mock.ExpectationsWereMet())
		db.Close()
	})

	return db, mock
}

func q(query string) string {
	return regexp.QuoteMeta(query)
}

func TestAuthorGetByID(t *testing.T) {
	db, mock := newMock(t)
	s := NewAuthor(db)
	ctx := context.Background()

	mock.ExpectQuery(q(getAuthorByID)).WithArgs(1).WillReturnRows(sqlmock.NewRows(
		[]string{"id", "first_name", "last_name", "dob", "pen_name"}).AddRow(1, "Joey", "Paul", "01-08-2001", "Joe"))
	mock.ExpectQuery(q(getAuthorByID)).WithArgs(2).WillReturnError(sql.ErrNoRows)

	got, err := s.GetByID(ctx, 1)
	assert.NoError(t, err)
	assert.Equal(t, model.Author{ID: 1, FirstName: "Joey", LastName: "Paul", Dob: "01-08-2001", PenName: "Joe"}, got)

	_, err = s.GetByID(ctx, 2)
	assert.Equal(t, apperrors.NotFound{Entity: "author", ID: 2}, err)
}

func TestAuthorFindID(t *testing.T) {
	db, mock := newMock(t)
	s := NewAuthor(db)
	a := model.Author{FirstName: "Joey", LastName: "Paul", Dob: "01-08-2001", PenName: "Joe"}

	mock.ExpectQuery(q(findAuthorID)).WithArgs("Joey", "Paul", "01-08-2001", "Joe").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(7))
	mock.ExpectQuery(q(findAuthorID)).WillReturnError(sql.ErrNoRows)

	id, found, err := s.FindID(context.Background(), a)
	assert.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, 7, id)

	_, found, err = s.FindID(context.Background(), a)
	assert.NoError(t, err)
	assert.False(t, found)
}

func TestAuthorCreate(t *testing.T) {
	db, mock := newMock(t)

	mock.ExpectExec(q(insertAuthor)).WithArgs("Joey", "Paul", "01-08-2001", "Joe").
		WillReturnResult(sqlmock.NewResult(3, 1))
	mock.ExpectExec(q(insertAuthor)).WillReturnError(errors.New("db down"))

	s := NewAuthor(db)
	a := model.Author{FirstName: "Joey", LastName: "Paul", Dob: "01-08-2001", PenName: "Joe"}

	id, err := s.Create(context.Background(), a)
	assert.NoError(t, err)
	assert.Equal(t, 3, id)

	_, err = s.Create(context.Background(), a)
	assert.EqualError(t, err, "insert author: db down")
}

func TestBookGetAll(t *testing.T) {
	db, mock := newMock(t)
	columns := []string{"id", "title", "publication", "published_date", "author_id"}

	mock.ExpectQuery(q(getAllBooks)).WillReturnRows(sqlmock.NewRows(columns).
		AddRow(1, "the wall", "Arihanth", "6/7/2017", 1).
		AddRow(2, "city", "Penguin", "11/5/2001", 2))
	mock.ExpectQuery(q(getBooksByTitle)).WithArgs("unknown").WillReturnRows(sqlmock.NewRows(columns))

	s := NewBook(db)

	got, err := s.GetAll(context.Background(), "")
	assert.NoError(t, err)
	assert.Equal(t, []model.Book{
		{ID: 1, Title: "the wall", Publication: "Arihanth", PublishedDate: "6/7/2017", Author: model.Author{ID: 1}},
		{ID: 2, Title: "city", Publication: "Penguin", PublishedDate: "11/5/2001", Author: model.Author{ID: 2}},
	}, got)

	got, err = s.GetAll(context.Background(), "unknown")
	assert.NoError(t, err)
	assert.Equal(t, []model.Book{}, got)
}

func TestBookDeleteOnlyDeletesTheBook(t *testing.T) {
	db, mock := newMock(t)

	mock.ExpectExec(q(deleteBookByID)).WithArgs(5).WillReturnResult(sqlmock.NewResult(0, 1))

	assert.NoError(t, NewBook(db).Delete(context.Background(), 5))
}

func TestTransactor(t *testing.T) {
	t.Run("commits on success", func(t *testing.T) {
		db, mock := newMock(t)

		mock.ExpectBegin()
		mock.ExpectExec(q(insertAuthor)).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := NewTransactor(db).WithTx(context.Background(), func(authors Author, _ Book) error {
			_, err := authors.Create(context.Background(), model.Author{FirstName: "Joey"})

			return err
		})
		assert.NoError(t, err)
	})

	t.Run("rolls back on error", func(t *testing.T) {
		db, mock := newMock(t)
		fnErr := errors.New("book insert failed")

		mock.ExpectBegin()
		mock.ExpectExec(q(insertAuthor)).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectRollback()

		err := NewTransactor(db).WithTx(context.Background(), func(authors Author, _ Book) error {
			if _, err := authors.Create(context.Background(), model.Author{FirstName: "Joey"}); err != nil {
				return err
			}

			return fnErr
		})
		assert.Equal(t, fnErr, err)
	})
}
