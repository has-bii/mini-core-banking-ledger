package httpx

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func DecodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		if maxBytesErr, ok := errors.AsType[*http.MaxBytesError](err); ok {
			return NewHTTPError(http.StatusRequestEntityTooLarge,
				fmt.Sprintf("body must not be larger than %d bytes", maxBytesErr.Limit))
		}
		return err
	}

	body = bytes.Trim(body, " \t\r\n")
	if len(body) == 0 {
		return NewHTTPError(http.StatusBadRequest, "body must not be empty")
	}
	if body[0] != '{' {
		return NewHTTPError(http.StatusBadRequest, "body must be a JSON object")
	}

	dec := json.NewDecoder(bytes.NewReader(body))
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		var (
			syntaxErr        *json.SyntaxError
			unmarshalTypeErr *json.UnmarshalTypeError
			invalidErr       *json.InvalidUnmarshalError
		)

		switch {
		case errors.As(err, &syntaxErr):
			return NewHTTPError(http.StatusBadRequest,
				fmt.Sprintf("body contains badly-formed JSON (at character %d)", syntaxErr.Offset))

		case errors.Is(err, io.ErrUnexpectedEOF):
			return NewHTTPError(http.StatusBadRequest,
				"body contains badly-formed JSON")

		case errors.As(err, &unmarshalTypeErr):
			if unmarshalTypeErr.Field != "" {
				return NewHTTPError(http.StatusBadRequest,
					fmt.Sprintf("body contains an invalid value for field %q (expected %s)",
						unmarshalTypeErr.Field, unmarshalTypeErr.Type))
			}
			return NewHTTPError(http.StatusBadRequest,
				fmt.Sprintf("body contains an invalid JSON type (at character %d)",
					unmarshalTypeErr.Offset))

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			field := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return NewHTTPError(http.StatusBadRequest,
				fmt.Sprintf("body contains unknown field %s", field))

		case errors.As(err, &invalidErr):
			panic(err)

		default:
			return err
		}
	}

	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return NewHTTPError(http.StatusBadRequest,
			"body must contain only one JSON object")
	}

	return nil
}
