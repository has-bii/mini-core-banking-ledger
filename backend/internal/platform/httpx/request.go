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
			return invalidBodyError(http.StatusRequestEntityTooLarge,
				fmt.Errorf("body exceeds %d bytes", maxBytesErr.Limit))
		}
		return err
	}

	const jsonWhitespace = " \t\r\n"
	trimmed := bytes.TrimLeft(body, jsonWhitespace)
	if len(trimmed) == 0 {
		return invalidBodyError(http.StatusBadRequest, errors.New("empty body"))
	}
	if trimmed[0] != '{' {
		return invalidBodyError(http.StatusBadRequest,
			fmt.Errorf("body is not a JSON object (at byte offset %d)", len(body)-len(trimmed)))
	}

	dec := jsontext.NewDecoder(bytes.NewReader(body))
	if err := json.UnmarshalDecode(dec, dst, json.RejectUnknownMembers(true)); err != nil {
		return classifyJSONError(err)
	}

	if trailing := bytes.TrimLeft(body[dec.InputOffset():], jsonWhitespace); len(trailing) != 0 {
		return invalidBodyError(http.StatusBadRequest,
			fmt.Errorf("trailing data after JSON object (at byte offset %d)", len(body)-len(trailing)))
	}

	return nil
}

func invalidBodyError(status int, cause error) error {
	return &HTTPError{Status: status, Message: "Invalid request body", cause: cause}
}

func classifyJSONError(err error) error {
	// Decoder errors can contain request data; retain only fixed reasons and offsets.
	if syntaxErr, ok := errors.AsType[*jsontext.SyntacticError](err); ok {
		return invalidBodyError(http.StatusBadRequest,
			fmt.Errorf("malformed JSON (at byte offset %d)", syntaxErr.ByteOffset))
	}
	if errors.Is(err, io.ErrUnexpectedEOF) {
		return invalidBodyError(http.StatusBadRequest, errors.New("malformed JSON: unexpected EOF"))
	}
	semanticErr, ok := errors.AsType[*json.SemanticError](err)
	if !ok {
		return err
	}

	var reason string
	switch {
	case errors.Is(semanticErr, json.ErrUnknownName):
		reason = "unknown field"
	case semanticErr.JSONKind != jsontext.KindInvalid && semanticErr.GoType != nil:
		reason = "type mismatch"
	default:
		return err
	}
	return invalidBodyError(http.StatusBadRequest,
		fmt.Errorf("%s (at byte offset %d)", reason, semanticErr.ByteOffset))
}
