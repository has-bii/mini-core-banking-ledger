package request

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
)

func HttpRequestJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	// Limit body size to 1 MB
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		var (
			syntaxErr        *json.SyntaxError
			unmarshalTypeErr *json.UnmarshalTypeError
			maxBytesErr      *http.MaxBytesError
			invalidErr       *json.InvalidUnmarshalError
		)

		switch {
		case errors.As(err, &syntaxErr):
			return BadRequest(http.StatusBadRequest,
				"body contains badly-formed JSON (at character %d)", syntaxErr.Offset)

		case errors.Is(err, io.ErrUnexpectedEOF):
			return BadRequest(http.StatusBadRequest,
				"body contains badly-formed JSON")

		case errors.As(err, &unmarshalTypeErr):
			if unmarshalTypeErr.Field != "" {
				return BadRequest(http.StatusBadRequest,
					"body contains an invalid value for field %q (expected %s)",
					unmarshalTypeErr.Field, unmarshalTypeErr.Type)
			}
			return BadRequest(http.StatusBadRequest,
				"body contains an invalid JSON type (at character %d)",
				unmarshalTypeErr.Offset)

		case errors.Is(err, io.EOF):
			return BadRequest(http.StatusBadRequest, "body must not be empty")

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			field := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return BadRequest(http.StatusBadRequest,
				"body contains unknown field %s", field)

		case errors.As(err, &maxBytesErr):
			return BadRequest(http.StatusRequestEntityTooLarge,
				"body must not be larger than %d bytes", maxBytesErr.Limit)

		case errors.As(err, &invalidErr):
			panic(err)

		default:
			return err
		}
	}

	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return BadRequest(http.StatusBadRequest,
			"body must contain only one JSON object")
	}

	return nil
}
