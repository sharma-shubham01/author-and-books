package author

import (
	"authorAndBooks/apperrors"
	"authorAndBooks/model"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

type fakeService struct {
	create func(a model.Author) (model.Author, error)
	update func(id int, a model.Author) (model.Author, error)
	delete func(id int) error
}

func (f fakeService) Create(_ context.Context, a model.Author) (model.Author, error) {
	return f.create(a)
}

func (f fakeService) Update(_ context.Context, id int, a model.Author) (model.Author, error) {
	return f.update(id, a)
}

func (f fakeService) Delete(_ context.Context, id int) error {
	return f.delete(id)
}

const joeyJSON = `{"ID":1,"first_name":"Joey","last_name":"Paul","dob":"01-08-2001","pen_name":"Joe"}`

func TestCreate(t *testing.T) {
	testCases := []struct {
		desc       string
		body       string
		svcErr     error
		wantStatus int
		wantBody   string
	}{
		{desc: "created", body: `{"first_name":"Joey","last_name":"Paul","dob":"01-08-2001","pen_name":"Joe"}`,
			wantStatus: http.StatusCreated, wantBody: joeyJSON},
		{desc: "malformed body", body: `not json`, wantStatus: http.StatusBadRequest,
			wantBody: `{"error":"invalid request body"}`},
		{desc: "missing fields", body: `{}`, svcErr: apperrors.Validation{Reason: "missing required fields: first_name, dob"},
			wantStatus: http.StatusBadRequest, wantBody: `{"error":"missing required fields: first_name, dob"}`},
		{desc: "duplicate", body: `{}`, svcErr: apperrors.Conflict{Reason: "author already exists"},
			wantStatus: http.StatusConflict, wantBody: `{"error":"author already exists"}`},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			h := New(fakeService{create: func(a model.Author) (model.Author, error) {
				if tc.svcErr != nil {
					return model.Author{}, tc.svcErr
				}

				a.ID = 1

				return a, nil
			}})
			w := httptest.NewRecorder()

			h.Create(w, httptest.NewRequest(http.MethodPost, "/author", strings.NewReader(tc.body)))

			assert.Equal(t, tc.wantStatus, w.Code)
			assert.JSONEq(t, tc.wantBody, w.Body.String())
		})
	}
}

func TestUpdate(t *testing.T) {
	testCases := []struct {
		desc       string
		id         string
		svcErr     error
		wantStatus int
	}{
		{desc: "success", id: "1", wantStatus: http.StatusOK},
		{desc: "invalid id", id: "abc", wantStatus: http.StatusBadRequest},
		{desc: "not found", id: "1000", svcErr: apperrors.NotFound{Entity: "author", ID: 1000}, wantStatus: http.StatusNotFound},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			h := New(fakeService{update: func(id int, a model.Author) (model.Author, error) {
				a.ID = id

				return a, tc.svcErr
			}})
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPut, "/author/"+tc.id, strings.NewReader(joeyJSON))

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
		{desc: "author does not exist", id: "100", svcErr: apperrors.NotFound{Entity: "author", ID: 100},
			wantStatus: http.StatusNotFound},
		{desc: "invalid id", id: "-3", wantStatus: http.StatusBadRequest},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			h := New(fakeService{delete: func(int) error { return tc.svcErr }})
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodDelete, "/author/"+tc.id, nil)

			h.Delete(w, mux.SetURLVars(req, map[string]string{"id": tc.id}))

			assert.Equal(t, tc.wantStatus, w.Code)
		})
	}
}
