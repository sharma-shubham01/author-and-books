package service

import (
	"authorAndBooks/apperrors"
	"authorAndBooks/model"
	"authorAndBooks/store"
	"context"
	"sort"
	"time"
)

// fakeAuthors is an in-memory store.Author. When err is set every method returns it.
type fakeAuthors struct {
	data   map[int]model.Author
	nextID int
	err    error
}

func newFakeAuthors(authors ...model.Author) *fakeAuthors {
	f := &fakeAuthors{data: map[int]model.Author{}}
	for _, a := range authors {
		f.data[a.ID] = a
		if a.ID > f.nextID {
			f.nextID = a.ID
		}
	}

	return f
}

func (f *fakeAuthors) GetByID(_ context.Context, id int) (model.Author, error) {
	if f.err != nil {
		return model.Author{}, f.err
	}

	a, ok := f.data[id]
	if !ok {
		return model.Author{}, apperrors.NotFound{Entity: "author", ID: id}
	}

	return a, nil
}

func (f *fakeAuthors) FindID(_ context.Context, a model.Author) (int, bool, error) {
	if f.err != nil {
		return 0, false, f.err
	}

	for id, existing := range f.data {
		existing.ID = a.ID
		if existing == a {
			return id, true, nil
		}
	}

	return 0, false, nil
}

func (f *fakeAuthors) Create(_ context.Context, a model.Author) (int, error) {
	if f.err != nil {
		return 0, f.err
	}

	f.nextID++
	a.ID = f.nextID
	f.data[a.ID] = a

	return a.ID, nil
}

func (f *fakeAuthors) Update(_ context.Context, a model.Author) error {
	if f.err != nil {
		return f.err
	}

	f.data[a.ID] = a

	return nil
}

func (f *fakeAuthors) Delete(_ context.Context, id int) error {
	if f.err != nil {
		return f.err
	}

	delete(f.data, id)

	return nil
}

// fakeBooks is an in-memory store.Book. When err is set every method returns it.
type fakeBooks struct {
	data   map[int]model.Book
	nextID int
	err    error
}

func newFakeBooks(books ...model.Book) *fakeBooks {
	f := &fakeBooks{data: map[int]model.Book{}}
	for _, b := range books {
		f.data[b.ID] = b
		if b.ID > f.nextID {
			f.nextID = b.ID
		}
	}

	return f
}

func (f *fakeBooks) GetAll(_ context.Context, title string) ([]model.Book, error) {
	if f.err != nil {
		return nil, f.err
	}

	ids := make([]int, 0, len(f.data))
	for id := range f.data {
		ids = append(ids, id)
	}

	sort.Ints(ids)

	books := []model.Book{}

	for _, id := range ids {
		if b := f.data[id]; title == "" || b.Title == title {
			books = append(books, model.Book{ID: b.ID, Title: b.Title, Publication: b.Publication,
				PublishedDate: b.PublishedDate, Author: model.Author{ID: b.Author.ID}})
		}
	}

	return books, nil
}

func (f *fakeBooks) GetByID(_ context.Context, id int) (model.Book, error) {
	if f.err != nil {
		return model.Book{}, f.err
	}

	b, ok := f.data[id]
	if !ok {
		return model.Book{}, apperrors.NotFound{Entity: "book", ID: id}
	}

	b.Author = model.Author{ID: b.Author.ID}

	return b, nil
}

func (f *fakeBooks) FindID(_ context.Context, title string, authorID int) (int, bool, error) {
	if f.err != nil {
		return 0, false, f.err
	}

	for id, b := range f.data {
		if b.Title == title && b.Author.ID == authorID {
			return id, true, nil
		}
	}

	return 0, false, nil
}

func (f *fakeBooks) Create(_ context.Context, b model.Book) (int, error) {
	if f.err != nil {
		return 0, f.err
	}

	f.nextID++
	b.ID = f.nextID
	f.data[b.ID] = b

	return b.ID, nil
}

func (f *fakeBooks) Update(_ context.Context, b model.Book) error {
	if f.err != nil {
		return f.err
	}

	f.data[b.ID] = b

	return nil
}

func (f *fakeBooks) Delete(_ context.Context, id int) error {
	if f.err != nil {
		return f.err
	}

	delete(f.data, id)

	return nil
}

func (f *fakeBooks) DeleteByAuthorID(_ context.Context, authorID int) error {
	if f.err != nil {
		return f.err
	}

	for id, b := range f.data {
		if b.Author.ID == authorID {
			delete(f.data, id)
		}
	}

	return nil
}

// fakeTx runs fn directly against the fake stores; rollback is covered by the store tests.
type fakeTx struct {
	authors *fakeAuthors
	books   *fakeBooks
}

func (f fakeTx) WithTx(_ context.Context, fn func(store.Author, store.Book) error) error {
	return fn(f.authors, f.books)
}

func fixedNow() time.Time {
	return time.Date(2026, time.November, 16, 12, 0, 0, 0, time.UTC)
}
