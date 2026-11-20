package author

import (
	"authorAndBooks/handler/rest"
	"authorAndBooks/model"
	"context"
	"net/http"
)

// Service is the author business logic the handler depends on.
type Service interface {
	Create(ctx context.Context, a model.Author) (model.Author, error)
	Update(ctx context.Context, id int, a model.Author) (model.Author, error)
	Delete(ctx context.Context, id int) error
}

// Handler serves the /author endpoints.
type Handler struct {
	svc Service
}

func New(svc Service) *Handler {
	return &Handler{svc: svc}
}

// Create handles POST /author.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var a model.Author
	if err := rest.DecodeJSON(w, r, &a); err != nil {
		rest.WriteError(w, err)
		return
	}

	created, err := h.svc.Create(r.Context(), a)
	if err != nil {
		rest.WriteError(w, err)
		return
	}

	rest.WriteJSON(w, http.StatusCreated, created)
}

// Update handles PUT /author/{id}.
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := rest.PathID(r)
	if err != nil {
		rest.WriteError(w, err)
		return
	}

	var a model.Author
	if err = rest.DecodeJSON(w, r, &a); err != nil {
		rest.WriteError(w, err)
		return
	}

	updated, err := h.svc.Update(r.Context(), id, a)
	if err != nil {
		rest.WriteError(w, err)
		return
	}

	rest.WriteJSON(w, http.StatusOK, updated)
}

// Delete handles DELETE /author/{id}.
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
