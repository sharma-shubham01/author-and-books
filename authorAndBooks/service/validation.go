package service

import (
	"authorAndBooks/apperrors"
	"authorAndBooks/model"
	"fmt"
	"strings"
)

type field struct {
	name  string
	value string
}

// requireFields returns a validation error naming every field whose value is blank.
func requireFields(fields ...field) error {
	var missing []string

	for _, f := range fields {
		if strings.TrimSpace(f.value) == "" {
			missing = append(missing, f.name)
		}
	}

	if len(missing) > 0 {
		return apperrors.Validation{Reason: "missing required fields: " + strings.Join(missing, ", ")}
	}

	return nil
}

func validateID(id int, entity string) error {
	if id <= 0 {
		return apperrors.Validation{Reason: fmt.Sprintf("invalid %s id %d", entity, id)}
	}

	return nil
}

// validateNewAuthor checks the fields needed to create an author.
func validateNewAuthor(a model.Author) error {
	return requireFields(field{"first_name", a.FirstName}, field{"dob", a.Dob})
}

// validateAuthorUpdate checks the fields needed to replace an author; every field is required.
func validateAuthorUpdate(a model.Author) error {
	return requireFields(
		field{"first_name", a.FirstName},
		field{"last_name", a.LastName},
		field{"dob", a.Dob},
		field{"pen_name", a.PenName},
	)
}
