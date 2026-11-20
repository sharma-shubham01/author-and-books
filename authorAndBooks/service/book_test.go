package service

import (
	"authorAndBooks/apperrors"
	"authorAndBooks/model"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func theWall() model.Book {
	return model.Book{ID: 1, Title: "the wall", Author: model.Author{ID: 1}, Publication: "Arihanth", PublishedDate: "6/7/2017"}
}

func newBookService(authors *fakeAuthors, books *fakeBooks) *Book {
	s := NewBook(books, authors, fakeTx{authors: authors, books: books})
	s.now = fixedNow

	return s
}

func TestBookGetAll(t *testing.T) {
	city := model.Book{ID: 2, Title: "city", Author: model.Author{ID: 1}, Publication: "Penguin", PublishedDate: "11/5/2001"}
	cityWithAuthor := city
	cityWithAuthor.Author = joey()
	theWallWithAuthor := theWall()
	theWallWithAuthor.Author = joey()

	testCases := []struct {
		desc          string
		title         string
		includeAuthor bool
		want          []model.Book
	}{
		{desc: "all books", want: []model.Book{theWall(), city}},
		{desc: "filter by title", title: "city", want: []model.Book{city}},
		{desc: "no book with title", title: "unknown", want: []model.Book{}},
		{desc: "include author", includeAuthor: true, want: []model.Book{theWallWithAuthor, cityWithAuthor}},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			s := newBookService(newFakeAuthors(joey()), newFakeBooks(theWall(), city))

			got, err := s.GetAll(context.Background(), tc.title, tc.includeAuthor)

			assert.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestBookGetByID(t *testing.T) {
	want := theWall()
	want.Author = joey()

	testCases := []struct {
		desc    string
		id      int
		want    model.Book
		wantErr error
	}{
		{desc: "success includes author", id: 1, want: want},
		{desc: "invalid id", id: -1, wantErr: apperrors.Validation{Reason: "invalid book id -1"}},
		{desc: "not found", id: 1000, wantErr: apperrors.NotFound{Entity: "book", ID: 1000}},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			s := newBookService(newFakeAuthors(joey()), newFakeBooks(theWall()))

			got, err := s.GetByID(context.Background(), tc.id)

			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

//nolint:funlen // table of cases
func TestBookCreate(t *testing.T) {
	newBook := func(title string, a model.Author) model.Book {
		return model.Book{Title: title, Author: a, Publication: "Penguin", PublishedDate: "11/5/2001"}
	}
	rohan := model.Author{FirstName: "Rohan", LastName: "Gupta", Dob: "01/07/2001", PenName: "GCC"}
	rohanCreated := rohan
	rohanCreated.ID = 2

	testCases := []struct {
		desc        string
		input       model.Book
		want        model.Book
		wantErr     error
		wantAuthors int
	}{
		{desc: "existing author referenced by id", input: newBook("city", model.Author{ID: 1}),
			want:        model.Book{ID: 2, Title: "city", Author: joey(), Publication: "Penguin", PublishedDate: "11/5/2001"},
			wantAuthors: 1},
		{desc: "existing author matched by details is reused", input: newBook("city", model.Author{FirstName: "Joey",
			LastName: "Paul", Dob: "01-08-2001", PenName: "Joe"}),
			want:        model.Book{ID: 2, Title: "city", Author: joey(), Publication: "Penguin", PublishedDate: "11/5/2001"},
			wantAuthors: 1},
		{desc: "new author is created before the book", input: newBook("city", rohan),
			want:        model.Book{ID: 2, Title: "city", Author: rohanCreated, Publication: "Penguin", PublishedDate: "11/5/2001"},
			wantAuthors: 2},
		{desc: "unknown author id", input: newBook("city", model.Author{ID: 100}),
			wantErr: apperrors.Validation{Reason: "author with id 100 does not exist"}, wantAuthors: 1},
		{desc: "author details missing", input: newBook("city", model.Author{}),
			wantErr: apperrors.Validation{Reason: "author: missing required fields: first_name, dob"}, wantAuthors: 1},
		{desc: "duplicate title for author", input: newBook("the wall", model.Author{ID: 1}),
			wantErr: apperrors.Conflict{Reason: `book "the wall" by author 1 already exists`}, wantAuthors: 1},
		{desc: "invalid fields", input: model.Book{Author: model.Author{ID: 1}, Publication: "the sun", PublishedDate: "2001"},
			wantErr: apperrors.Validation{Reason: "title is required; publication must be one of Penguin, Scholastic, " +
				"Arihanth; published_date must be a valid date in d/m/yyyy format"}, wantAuthors: 1},
		{desc: "published before 1880", input: model.Book{Title: "the river", Author: model.Author{ID: 1},
			Publication: "Penguin", PublishedDate: "3/5/1879"},
			wantErr: apperrors.Validation{Reason: "published_date must not be before 1880"}, wantAuthors: 1},
		{desc: "published in the future", input: model.Book{Title: "the river", Author: model.Author{ID: 1},
			Publication: "Scholastic", PublishedDate: "17/11/2026"},
			wantErr: apperrors.Validation{Reason: "published_date must not be in the future"}, wantAuthors: 1},
		{desc: "impossible date", input: model.Book{Title: "the river", Author: model.Author{ID: 1},
			Publication: "Scholastic", PublishedDate: "31/2/2001"},
			wantErr: apperrors.Validation{Reason: "published_date must be a valid date in d/m/yyyy format"}, wantAuthors: 1},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			authors, books := newFakeAuthors(joey()), newFakeBooks(theWall())

			got, err := newBookService(authors, books).Create(context.Background(), tc.input)

			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.want, got)
			assert.Len(t, authors.data, tc.wantAuthors)

			if tc.wantErr != nil {
				assert.Len(t, books.data, 1, "no book is created on failure")
			}
		})
	}
}

func TestBookUpdate(t *testing.T) {
	update := model.Book{Title: "the wall 2", Author: model.Author{ID: 1}, Publication: "Scholastic", PublishedDate: "4/5/2018"}
	city := model.Book{ID: 2, Title: "city", Author: model.Author{ID: 1}, Publication: "Penguin", PublishedDate: "11/5/2001"}

	testCases := []struct {
		desc    string
		id      int
		input   model.Book
		want    model.Book
		wantErr error
	}{
		{desc: "success", id: 1, input: update, want: model.Book{ID: 1, Title: "the wall 2", Author: joey(),
			Publication: "Scholastic", PublishedDate: "4/5/2018"}},
		{desc: "invalid id", id: 0, input: update, wantErr: apperrors.Validation{Reason: "invalid book id 0"}},
		{desc: "missing details", id: 1, input: model.Book{}, wantErr: apperrors.Validation{Reason: "title is required; " +
			"publication must be one of Penguin, Scholastic, Arihanth; published_date must be a valid date in d/m/yyyy format"}},
		{desc: "missing author id", id: 1, input: model.Book{Title: "x", Publication: "Penguin", PublishedDate: "1/1/2000"},
			wantErr: apperrors.Validation{Reason: "author.ID is required"}},
		{desc: "book not found", id: 100, input: update, wantErr: apperrors.NotFound{Entity: "book", ID: 100}},
		{desc: "author not found", id: 1, input: model.Book{Title: "x", Author: model.Author{ID: 9}, Publication: "Penguin",
			PublishedDate: "1/1/2000"}, wantErr: apperrors.Validation{Reason: "author with id 9 does not exist"}},
		{desc: "title clashes with another book of the author", id: 1, input: model.Book{Title: "city",
			Author: model.Author{ID: 1}, Publication: "Penguin", PublishedDate: "1/1/2000"},
			wantErr: apperrors.Conflict{Reason: `book "city" by author 1 already exists`}},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			books := newFakeBooks(theWall(), city)

			got, err := newBookService(newFakeAuthors(joey()), books).Update(context.Background(), tc.id, tc.input)

			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.want, got)

			if tc.wantErr == nil {
				assert.Equal(t, tc.want, books.data[tc.id])
			}
		})
	}
}

func TestBookDelete(t *testing.T) {
	testCases := []struct {
		desc    string
		id      int
		wantErr error
	}{
		{desc: "success", id: 1},
		{desc: "invalid id", id: -2, wantErr: apperrors.Validation{Reason: "invalid book id -2"}},
		{desc: "not found", id: 100, wantErr: apperrors.NotFound{Entity: "book", ID: 100}},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			sameAuthor := model.Book{ID: 2, Title: "city", Author: model.Author{ID: 1}}
			books := newFakeBooks(theWall(), sameAuthor)

			err := newBookService(newFakeAuthors(joey()), books).Delete(context.Background(), tc.id)

			assert.Equal(t, tc.wantErr, err)

			if tc.wantErr == nil {
				assert.Equal(t, []int{2}, keys(books.data), "other books of the same author are kept")
			}
		})
	}
}
