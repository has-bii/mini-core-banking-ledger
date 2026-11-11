package httpx

const (
	CodeInternalError       = "INTERNAL_ERROR"
	CodeNotFound            = "NOT_FOUND"
	CodeInvalidJSON         = "INVALID_JSON"
	CodeValidationError     = "VALIDATION_ERROR"
	CodeUnauthenticated     = "UNAUTHENTICATED" // HTTP 401: identity is missing or invalid.
	CodeForbidden           = "FORBIDDEN"       // HTTP 403: authenticated but lacks permission.
	CodeInvalidCredentials  = "INVALID_CREDENTIALS"
	CodeEmailTaken          = "EMAIL_TAKEN"
	CodeTooManyRequests     = "TOO_MANY_REQUESTS"
	CodeUnprocessableEntity = "UNPROCESSABLE_ENTITY"
)
