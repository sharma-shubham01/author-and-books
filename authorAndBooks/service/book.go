package service

import (
	"authorAndBooks/apperrors"
	"authorAndBooks/model"
	"authorAndBooks/store"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	// publishedDateLayout is d/m/yyyy; day and month may or may not be zero-padded.
	publishedDateLayout = "2/1/2006"
	minPublicationYear  = 1880
)

// Book implements the business rules for books.
type Book struct {
	books   store.Book
	authors store.Author
	tx      store.Transactor
	now     func() time.Time
}

func NewBook(books store.Book, authors store.Author, tx store.Transactor) *Book {
	return &Book{books: books, authors: authors, tx: tx, now: time.Now}
}

// GetAll returns all books, filtered by title when title is not empty. Each book carries only
// its author's ID unless includeAuthor is set.
func (s *Book) GetAll(ctx context.Context, title string, includeAuthor bool) ([]model.Book, error) {
	books, err := s.books.GetAll(ctx, title)
	if err != nil || !includeAuthor {
		return books, err
	}

	authors := make(map[int]model.Author)

	for i := range books {
		authorID := books[i].Author.ID

		a, ok := authors[authorID]
		if !ok {
			if a, err = s.authors.GetByID(ctx, authorID); err != nil {
				return nil, err
			}

			authors[authorID] = a
		}

		books[i].Author = a
	}

	return books, nil
}

// GetByID returns the book with the given id, including its author.
func (s *Book) GetByID(ctx context.Context, id int) (model.Book, error) {
	if err := validateID(id, "book"); err != nil {
		return model.Book{}, err
	}

	b, err := s.books.GetByID(ctx, id)
	if err != nil {
		return model.Book{}, err
	}

	if b.Author, err = s.authors.GetByID(ctx, b.Author.ID); err != nil {
		return model.Book{}, err
	}

	return b, nil
}

// Create adds a new book. The author and the book are written in one transaction, so a book is
// never created without its author:
//   - if b.Author.ID is set, that author must already exist;
//   - otherwise an existing author with exactly the same details is reused, or a new author is created.
func (s *Book) Create(ctx context.Context, b model.Book) (model.Book, error) {
	if err := s.validate(b); err != nil {
		return model.Book{}, err
	}

	err := s.tx.WithTx(ctx, func(authors store.Author, books store.Book) error {
		a, err := resolveAuthor(ctx, authors, b.Author)
		if err != nil {
			return err
		}

		b.Author = a

		if err = ensureUniqueTitle(ctx, books, b.Title, a.ID, 0); err != nil {
			return err
		}

		b.ID, err = books.Create(ctx, b)

		return err
	})
	if err != nil {
		return model.Book{}, err
	}

	return b, nil
}

// Update replaces the book with the given id. The referenced author must already exist.
func (s *Book) Update(ctx context.Context, id int, b model.Book) (model.Book, error) {
	if err := validateID(id, "book"); err != nil {
		return model.Book{}, err
	}

	if err := s.validate(b); err != nil {
		return model.Book{}, err
	}

	if b.Author.ID <= 0 {
		return model.Book{}, apperrors.Validation{Reason: "author.ID is required"}
	}

	if _, err := s.books.GetByID(ctx, id); err != nil {
		return model.Book{}, err
	}

	a, err := existingAuthor(ctx, s.authors, b.Author.ID)
	if err != nil {
		return model.Book{}, err
	}

	if err = ensureUniqueTitle(ctx, s.books, b.Title, a.ID, id); err != nil {
		return model.Book{}, err
	}

	b.ID, b.Author = id, a
	if err = s.books.Update(ctx, b); err != nil {
		return model.Book{}, err
	}

	return b, nil
}

// Delete removes the book with the given id.
func (s *Book) Delete(ctx context.Context, id int) error {
	if err := validateID(id, "book"); err != nil {
		return err
	}

	if _, err := s.books.GetByID(ctx, id); err != nil {
		return err
	}

	return s.books.Delete(ctx, id)
}

func (s *Book) validate(b model.Book) error {
	var problems []string

	if strings.TrimSpace(b.Title) == "" {
		problems = append(problems, "title is required")
	}

	if !isValidPublication(b.Publication) {
		problems = append(problems, "publication must be one of Penguin, Scholastic, Arihanth")
	}

	if reason := s.publishedDateProblem(b.PublishedDate); reason != "" {
		problems = append(problems, reason)
	}

	if len(problems) > 0 {
		return apperrors.Validation{Reason: strings.Join(problems, "; ")}
	}

	return nil
}

func (s *Book) publishedDateProblem(value string) string {
	date, err := time.Parse(publishedDateLayout, value)

	switch {
	case err != nil:
		return "published_date must be a valid date in d/m/yyyy format"
	case date.Year() < minPublicationYear:
		return fmt.Sprintf("published_date must not be before %d", minPublicationYear)
	case date.After(s.now()):
		return "published_date must not be in the future"
	default:
		return ""
	}
}

func isValidPublication(publication string) bool {
	switch publication {
	case "Penguin", "Scholastic", "Arihanth":
		return true
	default:
		return false
	}
}

// resolveAuthor returns the author a new book is linked to, creating the author when needed.
func resolveAuthor(ctx context.Context, authors store.Author, a model.Author) (model.Author, error) {
	if a.ID != 0 {
		return existingAuthor(ctx, authors, a.ID)
	}

	if err := validateNewAuthor(a); err != nil {
		return model.Author{}, apperrors.Validation{Reason: "author: " + err.Error()}
	}

	id, found, err := authors.FindID(ctx, a)
	if err != nil {
		return model.Author{}, err
	}

	if !found {
		if id, err = authors.Create(ctx, a); err != nil {
			return model.Author{}, err
		}
	}

	a.ID = id

	return a, nil
}

// existingAuthor loads a referenced author, reporting a missing one as invalid input.
func existingAuthor(ctx context.Context, authors store.Author, id int) (model.Author, error) {
	a, err := authors.GetByID(ctx, id)

	var notFound apperrors.NotFound
	if errors.As(err, &notFound) {
		return model.Author{}, apperrors.Validation{Reason: fmt.Sprintf("author with id %d does not exist", id)}
	}

	return a, err
}

// ensureUniqueTitle rejects a title the author already has on a book other than excludeID.
func ensureUniqueTitle(ctx context.Context, books store.Book, title string, authorID, excludeID int) error {
	existingID, found, err := books.FindID(ctx, title, authorID)
	if err != nil {
		return err
	}

	if found && existingID != excludeID {
		return apperrors.Conflict{Reason: fmt.Sprintf("book %q by author %d already exists", title, authorID)}
	}

	return nil
}
