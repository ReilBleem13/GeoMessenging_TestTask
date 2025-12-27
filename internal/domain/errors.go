package domain

type ErrorCode string

const (
	CodeAlreadyExists     ErrorCode = "ALREADY_EXISTS"
	CodeInvalidRequest    ErrorCode = "INVALID_REQUEST"
	CodeInvalidValidation ErrorCode = "INVALID_VALIDATION"
	CodeNotFound          ErrorCode = "NOT_FOUND"
	CodeUnauthorized      ErrorCode = "UNAUTHORIZED"
)

type AppError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
}

func (e *AppError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return string(e.Code)
}

func ErrAlreadyExists(msg string) error {
	return &AppError{Code: CodeAlreadyExists, Message: msg}
}

func ErrInvalidRequest(msg string) error {
	return &AppError{Code: CodeInvalidRequest, Message: msg}
}

func ErrInvalidValidation(msg string) error {
	return &AppError{Code: CodeInvalidValidation, Message: msg}
}

func ErrNotFound(msg string) error {
	return &AppError{Code: CodeNotFound, Message: msg}
}

func ErrUnauthorized(msg string) error {
	return &AppError{Code: CodeUnauthorized, Message: msg}
}
