package domain

type ErrorCode string

const (
	CodeIncedentExists ErrorCode = "INCEDENT_EXISTS"
	CodeInvalidRequest ErrorCode = "INVALID_REQUEST"
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

func ErrIncedentExists() error {
	return &AppError{Code: CodeIncedentExists, Message: "incedent_title already exists"}
}

func ErrInvalidRequest(msg string) error {
	return &AppError{Code: CodeInvalidRequest, Message: msg}
}
