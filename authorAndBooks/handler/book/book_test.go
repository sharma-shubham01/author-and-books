package book

import (
	"authorAndBooks/apperrors"
	"authorAndBooks/model"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

type fakeService struct {
	getAll  func(title string, includeAuthor bool) ([]model.Book, error)
	getByID func(id int) (model.Book, error)
	create  func(b model.Book) (model.Book, error)
	update  func(id int, b model.Book) (model.Book, error)
	delete  func(id int) error
}

func (f fakeService) GetAll(_ context.Context, title string, includeAuthor bool) ([]model.Book, error) {
	return f.getAll(title, includeAuthor)
}

func (f fakeService) GetByID(_ context.Context, id int) (model.Book, error) {
	return f.getByID(id)
}

func (f fakeService) Create(_ context.Context, b model.Book) (model.Book, error) {
	return f.create(b)
}

func (f fakeService) Update(_ context.Context, id int, b model.Book) (model.Book, error) {
	return f.update(id, b)
}

func (f fakeService) Delete(_ context.Context, id int) error {
	return f.delete(id)
}

func theWall() model.Book {
	return model.Book{ID: 1, Title: "the wall", Author: model.Author{ID: 1}, Publication: "Arihanth", PublishedDate: "6/7/2017"}
}

const theWallJSON = `{"ID":1,"title":"the wall","author":{"ID":1,"first_name":"","last_name":"","dob":"","pen_name":""},` +
	`"publication":"Arihanth","published_date":"6/7/2017"}`

func TestGetAll(t *testing.T) {
	testCases := []struct {
		desc       string
		target     string
		wantTitle  string
		wantAuthor bool
		svcErr     error
		wantStatus int
		wantBody   string
	}{
		{desc: "all books", target: "/book", wantStatus: http.StatusOK, wantBody: "[" + theWallJSON + "]"},
		{desc: "query params are passed through", target: "/book?title=the%20wall&includeAuthor=true",
			wantTitle: "the wall", wantAuthor: true, wantStatus: http.StatusOK, wantBody: "[" + theWallJSON + "]"},
		{desc: "invalid includeAuthor", target: "/book?includeAuthor=maybe", wantStatus: http.StatusBadRequest,
			wantBody: `{"error":"includeAuthor must be true or false"}`},
		{desc: "service failure", target: "/book", svcErr: errors.New("db down"),
			wantStatus: http.StatusInternalServerError, wantBody: `{"error":"internal server error"}`},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			h := New(fakeService{getAll: func(title string, includeAuthor bool) ([]model.Book, error) {
				assert.Equal(t, tc.wantTitle, title)
				assert.Equal(t, tc.wantAuthor, includeAuthor)

				return []model.Book{theWall()}, tc.svcErr
			}})
			w := httptest.NewRecorder()

			h.GetAll(w, httptest.NewRequest(http.MethodGet, tc.target, nil))

			assert.Equal(t, tc.wantStatus, w.Code)
			assert.JSONEq(t, tc.wantBody, w.Body.String())
		})
	}
}

func TestGetByID(t *testing.T) {
	testCases := []struct {
		desc       string
		id         string
		svcErr     error
		wantStatus int
		wantBody   string
	}{
		{desc: "success", id: "1", wantStatus: http.StatusOK, wantBody: theWallJSON},
		{desc: "non-numeric id", id: "abc", wantStatus: http.StatusBadRequest,
			wantBody: `{"error":"id must be a positive integer"}`},
		{desc: "not found", id: "1000", svcErr: apperrors.NotFound{Entity: "book", ID: 1000},
			wantStatus: http.StatusNotFound, wantBody: `{"error":"book with id 1000 not found"}`},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			h := New(fakeService{getByID: func(id int) (model.Book, error) {
				if tc.svcErr != nil {
					return model.Book{}, tc.svcErr
				}

				return theWall(), nil
			}})
			w := httptest.NewRecorder()
			req := mux.SetURLVars(httptest.NewRequest(http.MethodGet, "/book/"+tc.id, nil), map[string]string{"id": tc.id})

			h.GetByID(w, req)

			assert.Equal(t, tc.wantStatus, w.Code)
			assert.JSONEq(t, tc.wantBody, w.Body.String())
		})
	}
}

func TestCreate(t *testing.T) {
	testCases := []struct {
		desc       string
		body       string
		svcErr     error
		wantStatus int
		wantBody   string
	}{
		{desc: "created", body: `{"title":"the wall","author":{"ID":1},"publication":"Arihanth","published_date":"6/7/2017"}`,
			wantStatus: http.StatusCreated, wantBody: theWallJSON},
		{desc: "malformed body", body: `{"title":`, wantStatus: http.StatusBadRequest,
			wantBody: `{"error":"invalid request body"}`},
		{desc: "validation error", body: `{}`, svcErr: apperrors.Validation{Reason: "title is required"},
			wantStatus: http.StatusBadRequest, wantBody: `{"error":"title is required"}`},
		{desc: "duplicate", body: `{}`, svcErr: apperrors.Conflict{Reason: "book already exists"},
			wantStatus: http.StatusConflict, wantBody: `{"error":"book already exists"}`},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			h := New(fakeService{create: func(b model.Book) (model.Book, error) {
				if tc.svcErr != nil {
					return model.Book{}, tc.svcErr
				}

				b.ID = 1

				return b, nil
			}})
			w := httptest.NewRecorder()

			h.Create(w, httptest.NewRequest(http.MethodPost, "/book", strings.NewReader(tc.body)))

			assert.Equal(t, tc.wantStatus, w.Code)
			assert.JSONEq(t, tc.wantBody, w.Body.String())
		})
	}
}

func TestUpdate(t *testing.T) {
	testCases := []struct {
		desc       string
		id         string
		body       string
		wantStatus int
	}{
		{desc: "success", id: "1", body: `{"title":"the wall"}`, wantStatus: http.StatusOK},
		{desc: "invalid id", id: "-1", body: `{}`, wantStatus: http.StatusBadRequest},
		{desc: "malformed body", id: "1", body: `[`, wantStatus: http.StatusBadRequest},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			h := New(fakeService{update: func(id int, b model.Book) (model.Book, error) {
				assert.Equal(t, 1, id, "id comes from the path")

				b.ID = id

				return b, nil
			}})
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPut, "/book/"+tc.id, strings.NewReader(tc.body))

			h.Update(w, mux.SetURLVars(req, map[string]string{"id": tc.id}))

			assert.Equal(t, tc.wantStatus, w.Code)
		})
	}
}

func TestDelete(t *testing.T) {
	testCases := []struct {
		desc       string
		id         string
		svcErr     error
		wantStatus int
	}{
		{desc: "success", id: "1", wantStatus: http.StatusNoContent},
		{desc: "not found", id: "100", svcErr: apperrors.NotFound{Entity: "book", ID: 100}, wantStatus: http.StatusNotFound},
		{desc: "invalid id", id: "-2", wantStatus: http.StatusBadRequest},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			h := New(fakeService{delete: func(int) error { return tc.svcErr }})
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodDelete, "/book/"+tc.id, nil)

			h.Delete(w, mux.SetURLVars(req, map[string]string{"id": tc.id}))

			assert.Equal(t, tc.wantStatus, w.Code)
		})
	}
}
