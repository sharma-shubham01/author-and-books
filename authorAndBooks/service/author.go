package service

import (
	"authorAndBooks/apperrors"
	"authorAndBooks/model"
	"authorAndBooks/store"
	"context"
)

// Author implements the business rules for authors.
type Author struct {
	authors store.Author
	tx      store.Transactor
}

func NewAuthor(authors store.Author, tx store.Transactor) *Author {
	return &Author{authors: authors, tx: tx}
}

// Create adds a new author. An author whose details all match an existing author is rejected.
func (s *Author) Create(ctx context.Context, a model.Author) (model.Author, error) {
	if err := validateNewAuthor(a); err != nil {
		return model.Author{}, err
	}

	_, found, err := s.authors.FindID(ctx, a)
	if err != nil {
		return model.Author{}, err
	}

	if found {
		return model.Author{}, apperrors.Conflict{Reason: "author already exists"}
	}

	id, err := s.authors.Create(ctx, a)
	if err != nil {
		return model.Author{}, err
	}

	a.ID = id

	return a, nil
}

// Update replaces every field of the author with the given id.
func (s *Author) Update(ctx context.Context, id int, a model.Author) (model.Author, error) {
	if err := validateID(id, "author"); err != nil {
		return model.Author{}, err
	}

	if err := validateAuthorUpdate(a); err != nil {
		return model.Author{}, err
	}

	if _, err := s.authors.GetByID(ctx, id); err != nil {
		return model.Author{}, err
	}

	existingID, found, err := s.authors.FindID(ctx, a)
	if err != nil {
		return model.Author{}, err
	}

	if found && existingID != id {
		return model.Author{}, apperrors.Conflict{Reason: "another author with the same details already exists"}
	}

	a.ID = id
	if err = s.authors.Update(ctx, a); err != nil {
		return model.Author{}, err
	}

	return a, nil
}

// Delete removes the author with the given id together with all of their books.
func (s *Author) Delete(ctx context.Context, id int) error {
	if err := validateID(id, "author"); err != nil {
		return err
	}

	if _, err := s.authors.GetByID(ctx, id); err != nil {
		return err
	}

	return s.tx.WithTx(ctx, func(authors store.Author, books store.Book) error {
		if err := books.DeleteByAuthorID(ctx, id); err != nil {
			return err
		}

		return authors.Delete(ctx, id)
	})
}
