package httpx

import (
	"bytes"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"fmt"
	"io"
	"net/http"
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

	dec := jsontext.NewDecoder(bytes.NewReader(body))
	if err := json.UnmarshalDecode(dec, dst, json.RejectUnknownMembers(true)); err != nil {
		if errors.Is(err, io.ErrUnexpectedEOF) {
			return NewHTTPError(http.StatusBadRequest, "body contains badly-formed JSON")
		}
		if syntaxErr, ok := errors.AsType[*jsontext.SyntacticError](err); ok {
			return NewHTTPError(http.StatusBadRequest,
				fmt.Sprintf("body contains badly-formed JSON (at byte offset %d)", syntaxErr.ByteOffset))
		}
		if semanticErr, ok := errors.AsType[*json.SemanticError](err); ok {
			if errors.Is(semanticErr, json.ErrUnknownName) {
				return NewHTTPError(http.StatusBadRequest,
					fmt.Sprintf("body contains unknown field %q", semanticErr.JSONPointer.LastToken()))
			}
			if semanticErr.JSONKind != jsontext.KindInvalid && semanticErr.GoType != nil {
				if semanticErr.JSONPointer != "" {
					return NewHTTPError(http.StatusBadRequest,
						fmt.Sprintf("body contains an invalid value for field %q (expected %s)",
							semanticErr.JSONPointer.LastToken(), semanticErr.GoType))
				}
				return NewHTTPError(http.StatusBadRequest,
					fmt.Sprintf("body contains an invalid JSON type (at byte offset %d)", semanticErr.ByteOffset))
			}
		}
		return err
	}

	if dec.InputOffset() != int64(len(body)) {
		return NewHTTPError(http.StatusBadRequest,
			"body must contain only one JSON object")
	}

	return nil
}
