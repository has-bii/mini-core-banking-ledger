package httpx

import "fmt"

type MalformedRequestBody struct {
	Status int
	Msg    string
}

func (mr *MalformedRequestBody) Error() string { return mr.Msg }

func BadRequest(status int, format string, args ...any) error {
	return &MalformedRequestBody{Status: status, Msg: fmt.Sprintf(format, args...)}
}
