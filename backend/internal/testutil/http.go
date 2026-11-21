package testutil

// ProblemDetails decodes the common fields of an HTTP problem response in tests.
type ProblemDetails struct {
	Type     string `json:"type"`
	Title    string `json:"title"`
	Detail   string `json:"detail"`
	Instance string `json:"instance"`
	Status   int    `json:"status"`
}
