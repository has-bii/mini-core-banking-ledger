package validator

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

func UseJSONName(f reflect.StructField) string {
	name, _, _ := strings.Cut(f.Tag.Get("json"), ",")
	if name == "" || name == "-" {
		return f.Name
	}
	return name
}

func FormatErrors(err error) ErrorField {
	var ve validator.ValidationErrors
	if !errors.As(err, &ve) {
		return nil
	}
	out := make(ErrorField, len(ve))
	for _, fe := range ve {
		out[fe.Field()] = append(out[fe.Field()], messageFor(fe))
	}
	return out
}

func messageFor(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "This field is required"
	case "min":
		switch fe.Kind() {
		case reflect.String:
			return fmt.Sprintf("Must contain at least %s characters", fe.Param())
		case reflect.Slice, reflect.Map:
			return fmt.Sprintf("Must contain at least %s items", fe.Param())
		default:
			return fmt.Sprintf("Must be at least %s", fe.Param())
		}
	case "max":
		if fe.Kind() == reflect.String {
			return fmt.Sprintf("Must contain at most %s characters", fe.Param())
		}
		return fmt.Sprintf("Must be at most %s", fe.Param())
	case "email":
		return "Must be a valid email address"
	case "oneof":
		return fmt.Sprintf("Must be one of: %s", fe.Param())
	default:
		return fmt.Sprintf("Failed validation rule %q", fe.Tag())
	}
}
