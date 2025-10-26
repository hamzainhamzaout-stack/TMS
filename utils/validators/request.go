package validator

import (
	"fmt"
	"reflect"

	"github.com/Robotech-Org/driver-management/utils/gens"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

// Validate performs validation and returns errors in a structured format.
func Validate(s any) map[string]string {
	err := validate.Struct(s)
	if err != nil {
		errors := make(map[string]string)
		if fieldErrors, ok := err.(validator.ValidationErrors); ok {
			for _, fe := range fieldErrors {
				fieldName := getJSONFieldName(s, fe.StructField())
				errors[fieldName] = msgForTag(fe, fieldName)
			}
		}
		return errors
	}
	return nil
}

// msgForTag maps validation errors to human-readable messages.
func msgForTag(fe validator.FieldError, fieldName string) string {
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("The field '%s' is required.", gens.ToVerboseName(fieldName))
	case "email":
		return fmt.Sprintf("The field '%s' must be a valid email address.", fieldName)
	case "min":
		return fmt.Sprintf("The field '%s' must be at least %s characters long.", fieldName, fe.Param())
	case "max":
		return fmt.Sprintf("The field '%s' must be at most %s characters long.", fieldName, fe.Param())
	default:
		return fe.Error()
	}
}

// getJSONFieldName retrieves the JSON field name from struct tags.
func getJSONFieldName(s any, structFieldName string) string {
	t := reflect.TypeOf(s)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if field, found := t.FieldByName(structFieldName); found {
		if jsonTag := field.Tag.Get("json"); jsonTag != "" && jsonTag != "-" {
			return jsonTag
		}
	}

	return structFieldName
}
