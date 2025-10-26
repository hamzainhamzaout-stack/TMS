package validator

import (
	"bytes"
	"encoding/json"
	"slices"
	"strings"
)

type ValidationErrors struct {
	FieldErrors map[string][]string
}

func NewValidationErrors() *ValidationErrors {
	return &ValidationErrors{
		FieldErrors: make(map[string][]string),
	}
}

func (v *ValidationErrors) Merge(input *ValidationErrors) {
	for key, value := range input.FieldErrors {
		v.FieldErrors[key] = slices.Concat(v.FieldErrors[key], value)
	}
}

func (v *ValidationErrors) Add(key string, err error) {
	v.FieldErrors[key] = append(v.FieldErrors[key], err.Error())
}

func (e ValidationErrors) Error() string {
	buff := bytes.NewBufferString("")

	for field, errs := range e.FieldErrors {
		buff.WriteString("Field:")
		buff.WriteString(field)
		buff.WriteString(" Errors:")
		for i := range errs {
			buff.WriteString(errs[i])
			if i != len(errs)-1 {
				buff.WriteString(", ")
			}
		}
		buff.WriteString("\n")
	}

	return strings.TrimSpace(buff.String())
}

func (e ValidationErrors) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.FieldErrors)
}

func (e ValidationErrors) HasErrors() bool {
	return len(e.FieldErrors) > 0
}
