package service

import (
	"authorAndBooks/apperrors"
	"authorAndBooks/model"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

var errDB = errors.New("db down")

func joey() model.Author {
	return model.Author{ID: 1, FirstName: "Joey", LastName: "Paul", Dob: "01-08-2001", PenName: "Joe"}
}

func newAuthorService(authors *fakeAuthors, books *fakeBooks) *Author {
	return NewAuthor(authors, fakeTx{authors: authors, books: books})
}

func TestAuthorCreate(t *testing.T) {
	testCases := []struct {
		desc    string
		input   model.Author
		dbErr   error
		want    model.Author
		wantErr error
	}{
		{desc: "success", input: model.Author{FirstName: "Rohan", LastName: "Gupta", Dob: "01/07/2001", PenName: "GCC"},
			want: model.Author{ID: 2, FirstName: "Rohan", LastName: "Gupta", Dob: "01/07/2001", PenName: "GCC"}},
		{desc: "only first name and dob are required", input: model.Author{FirstName: "Rohan", Dob: "01/07/2001"},
			want: model.Author{ID: 2, FirstName: "Rohan", Dob: "01/07/2001"}},
		{desc: "missing first name and dob", input: model.Author{LastName: "Gupta"},
			wantErr: apperrors.Validation{Reason: "missing required fields: first_name, dob"}},
		{desc: "duplicate author", input: model.Author{FirstName: "Joey", LastName: "Paul", Dob: "01-08-2001", PenName: "Joe"},
			wantErr: apperrors.Conflict{Reason: "author already exists"}},
		{desc: "store failure", input: model.Author{FirstName: "Rohan", Dob: "01/07/2001"}, dbErr: errDB, wantErr: errDB},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			authors := newFakeAuthors(joey())
			authors.err = tc.dbErr

			got, err := newAuthorService(authors, newFakeBooks()).Create(context.Background(), tc.input)

			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestAuthorUpdate(t *testing.T) {
	rohan := model.Author{ID: 2, FirstName: "Rohan", LastName: "Gupta", Dob: "01/07/2001", PenName: "GCC"}
	update := model.Author{FirstName: "Mohan", LastName: "Chandra", Dob: "02/07/2001", PenName: "MC"}

	testCases := []struct {
		desc    string
		id      int
		input   model.Author
		want    model.Author
		wantErr error
	}{
		{desc: "success", id: 2, input: update,
			want: model.Author{ID: 2, FirstName: "Mohan", LastName: "Chandra", Dob: "02/07/2001", PenName: "MC"}},
		{desc: "unchanged details are not a conflict with itself", id: 2, input: rohan, want: rohan},
		{desc: "invalid id", id: 0, input: update, wantErr: apperrors.Validation{Reason: "invalid author id 0"}},
		{desc: "all fields are required", id: 2, input: model.Author{FirstName: "Mohan", Dob: "02/07/2001"},
			wantErr: apperrors.Validation{Reason: "missing required fields: last_name, pen_name"}},
		{desc: "author not found", id: 100, input: update, wantErr: apperrors.NotFound{Entity: "author", ID: 100}},
		{desc: "details match another author", id: 2, input: joey(),
			wantErr: apperrors.Conflict{Reason: "another author with the same details already exists"}},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			authors := newFakeAuthors(joey(), rohan)

			got, err := newAuthorService(authors, newFakeBooks()).Update(context.Background(), tc.id, tc.input)

			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.want, got)

			if tc.wantErr == nil {
				assert.Equal(t, tc.want, authors.data[tc.id])
			}
		})
	}
}

func TestAuthorDelete(t *testing.T) {
	testCases := []struct {
		desc    string
		id      int
		wantErr error
	}{
		{desc: "success", id: 1},
		{desc: "invalid id", id: -3, wantErr: apperrors.Validation{Reason: "invalid author id -3"}},
		{desc: "author not found", id: 100, wantErr: apperrors.NotFound{Entity: "author", ID: 100}},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			authors := newFakeAuthors(joey(), model.Author{ID: 2, FirstName: "Rohan", Dob: "01/07/2001"})
			books := newFakeBooks(
				model.Book{ID: 1, Title: "the wall", Author: model.Author{ID: 1}},
				model.Book{ID: 2, Title: "city", Author: model.Author{ID: 2}},
			)

			err := newAuthorService(authors, books).Delete(context.Background(), tc.id)

			assert.Equal(t, tc.wantErr, err)

			if tc.wantErr == nil {
				assert.NotContains(t, authors.data, 1)
				assert.Equal(t, []int{2}, keys(books.data), "only the deleted author's books are removed")
			}
		})
	}
}

func keys(books map[int]model.Book) []int {
	ids := []int{}
	for id := range books {
		ids = append(ids, id)
	}

	return ids
}
