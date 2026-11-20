package book

import (
	"authorAndBooks/apperrors"
	"authorAndBooks/handler/rest"
	"authorAndBooks/model"
	"context"
	"net/http"
	"strconv"
)

// Service is the book business logic the handler depends on.
type Service interface {
	GetAll(ctx context.Context, title string, includeAuthor bool) ([]model.Book, error)
	GetByID(ctx context.Context, id int) (model.Book, error)
	Create(ctx context.Context, b model.Book) (model.Book, error)
	Update(ctx context.Context, id int, b model.Book) (model.Book, error)
	Delete(ctx context.Context, id int) error
}

// Handler serves the /book endpoints.
type Handler struct {
	svc Service
}

func New(svc Service) *Handler {
	return &Handler{svc: svc}
}

// GetAll handles GET /book with the optional query parameters title and includeAuthor.
func (h *Handler) GetAll(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	includeAuthor := false

	if value := query.Get("includeAuthor"); value != "" {
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			rest.WriteError(w, apperrors.Validation{Reason: "includeAuthor must be true or false"})
			return
		}

		includeAuthor = parsed
	}

	books, err := h.svc.GetAll(r.Context(), query.Get("title"), includeAuthor)
	if err != nil {
		rest.WriteError(w, err)
		return
	}

	rest.WriteJSON(w, http.StatusOK, books)
}

// GetByID handles GET /book/{id}.
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := rest.PathID(r)
	if err != nil {
		rest.WriteError(w, err)
		return
	}

	b, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		rest.WriteError(w, err)
		return
	}

	rest.WriteJSON(w, http.StatusOK, b)
}

// Create handles POST /book.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var b model.Book
	if err := rest.DecodeJSON(w, r, &b); err != nil {
		rest.WriteError(w, err)
		return
	}

	created, err := h.svc.Create(r.Context(), b)
	if err != nil {
		rest.WriteError(w, err)
		return
	}

	rest.WriteJSON(w, http.StatusCreated, created)
}

// Update handles PUT /book/{id}.
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := rest.PathID(r)
	if err != nil {
		rest.WriteError(w, err)
		return
	}

	var b model.Book
	if err = rest.DecodeJSON(w, r, &b); err != nil {
		rest.WriteError(w, err)
		return
	}

	updated, err := h.svc.Update(r.Context(), id, b)
	if err != nil {
		rest.WriteError(w, err)
		return
	}

	rest.WriteJSON(w, http.StatusOK, updated)
}

// Delete handles DELETE /book/{id}.
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := rest.PathID(r)
	if err != nil {
		rest.WriteError(w, err)
		return
	}

	if err = h.svc.Delete(r.Context(), id); err != nil {
		rest.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
