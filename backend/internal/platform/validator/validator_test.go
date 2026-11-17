package validator_test

import (
	pValidator "backend/internal/platform/validator"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/go-playground/validator/v10"
)

func Test_UseJSONName(t *testing.T) {
	t.Parallel()

	type DTO struct {
		Email  string `json:"email,omitempty"`
		Name   string
		Hidden string `json:"-"`
	}
	typ := reflect.TypeFor[DTO]()
	for i, want := range []string{"email", "Name", "Hidden"} {
		field := typ.Field(i)
		t.Run(field.Name, func(t *testing.T) {
			t.Parallel()
			if got := pValidator.UseJSONName(field); got != want {
				t.Errorf("UseJSONName() = %q, want %q", got, want)
			}
		})
	}
}

func Test_FormatErrors(t *testing.T) {
	t.Parallel()

	validate := validator.New()
	validate.RegisterTagNameFunc(pValidator.UseJSONName)

	type DTO struct {
		Required string            `json:"required" validate:"required"`
		MinText  string            `json:"min_text" validate:"min=3"`
		MinSlice []string          `json:"min_slice" validate:"min=2"`
		MinMap   map[string]string `json:"min_map" validate:"min=2"`
		MinValue int               `json:"min_value" validate:"min=10"`
		MaxText  string            `json:"max_text" validate:"max=3"`
		MaxValue int               `json:"max_value" validate:"max=10"`
		Email    string            `json:"email" validate:"email"`
		Currency string            `json:"currency" validate:"oneof=USD IDR"`
		Code     string            `json:"code" validate:"len=3"`
	}
	invalidErr := validate.Struct(DTO{
		MaxText:  "long",
		MaxValue: 11,
		Email:    "invalid",
		Currency: "EUR",
	})
	invalidFields := pValidator.ErrorField{
		"required":  {"This field is required"},
		"min_text":  {"Must contain at least 3 characters"},
		"min_slice": {"Must contain at least 2 items"},
		"min_map":   {"Must contain at least 2 items"},
		"min_value": {"Must be at least 10"},
		"max_text":  {"Must contain at most 3 characters"},
		"max_value": {"Must be at most 10"},
		"email":     {"Must be a valid email address"},
		"currency":  {"Must be one of: USD IDR"},
		"code":      {`Failed validation rule "len"`},
	}

	type Child struct {
		Name string `json:"name" validate:"required"`
	}
	groupedErr := validate.Struct(struct {
		First  Child
		Second Child
	}{})

	tests := []struct {
		name string
		err  error
		want pValidator.ErrorField
	}{
		{
			name: "nil error",
		},
		{
			name: "non-validation error",
			err:  errors.New("unrelated error"),
		},
		{
			name: "invalid validation input",
			err:  validate.Struct(42),
		},
		{
			name: "valid input",
			err:  validate.Struct(Child{Name: "Alice"}),
		},
		{
			name: "validation messages with JSON field names",
			err:  invalidErr,
			want: invalidFields,
		},
		{
			name: "wrapped validation errors",
			err:  fmt.Errorf("validation failed: %w", invalidErr),
			want: invalidFields,
		},
		{
			name: "multiple errors for the same field",
			err:  groupedErr,
			want: pValidator.ErrorField{
				"name": {"This field is required", "This field is required"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := pValidator.FormatErrors(tt.err); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FormatErrors() = %#v, want %#v", got, tt.want)
			}
		})
	}
}
