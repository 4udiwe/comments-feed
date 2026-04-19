package errors

import (
	"fmt"

	"github.com/vektah/gqlparser/v2/gqlerror"
)

const (
	// Error codes
	ErrCodeValidation          = "VALIDATION_ERROR"
	ErrCodeNotFound            = "NOT_FOUND"
	ErrCodeUnauthorized        = "UNAUTHORIZED"
	ErrCodeForbidden           = "FORBIDDEN"
	ErrCodeConflict            = "CONFLICT"
	ErrCodeInternal            = "INTERNAL_SERVER_ERROR"
	ErrCodeBadRequest          = "BAD_REQUEST"
	ErrCodeInvalidInput        = "INVALID_INPUT"
	ErrCodeConstraintViolation = "CONSTRAINT_VIOLATION"
)

// GraphQLError представляет ошибку GraphQL с дополнительной информацией
type GraphQLError struct {
	Message    string
	Code       string
	StatusHttp int
	Original   error // исходная ошибка для логирования
}

// Error реализует интерфейс error
func (e *GraphQLError) Error() string {
	return e.Message
}

func (e *GraphQLError) ToGQLError() *gqlerror.Error {
	return &gqlerror.Error{
		Message: e.Message,
		Extensions: map[string]interface{}{
			"code":   e.Code,
			"status": e.StatusHttp,
		},
	}
}

// NewValidationError создает ошибку валидации
func NewValidationError(message string) *GraphQLError {
	return &GraphQLError{
		Message:    message,
		Code:       ErrCodeValidation,
		StatusHttp: 400,
	}
}

// NewNotFoundError создает ошибку "не найдено"
func NewNotFoundError(resource string) *GraphQLError {
	return &GraphQLError{
		Message:    fmt.Sprintf("%s not found", resource),
		Code:       ErrCodeNotFound,
		StatusHttp: 404,
	}
}

// NewUnauthorizedError создает ошибку авторизации
func NewUnauthorizedError(message string) *GraphQLError {
	return &GraphQLError{
		Message:    message,
		Code:       ErrCodeUnauthorized,
		StatusHttp: 401,
	}
}

// NewForbiddenError создает ошибку доступа
func NewForbiddenError(message string) *GraphQLError {
	return &GraphQLError{
		Message:    message,
		Code:       ErrCodeForbidden,
		StatusHttp: 403,
	}
}

// NewConflictError создает ошибку конфликта
func NewConflictError(message string) *GraphQLError {
	return &GraphQLError{
		Message:    message,
		Code:       ErrCodeConflict,
		StatusHttp: 409,
	}
}

// NewInternalError создает ошибку сервера с логированием оригинальной ошибки
func NewInternalError(message string, original error) *GraphQLError {
	return &GraphQLError{
		Message:    message,
		Code:       ErrCodeInternal,
		StatusHttp: 500,
		Original:   original,
	}
}

// NewBadRequestError создает ошибку неверного запроса
func NewBadRequestError(message string) *GraphQLError {
	return &GraphQLError{
		Message:    message,
		Code:       ErrCodeBadRequest,
		StatusHttp: 400,
	}
}

// NewInvalidInputError создает ошибку неверного ввода
func NewInvalidInputError(field string, message string) *GraphQLError {
	return &GraphQLError{
		Message:    fmt.Sprintf("invalid %s: %s", field, message),
		Code:       ErrCodeInvalidInput,
		StatusHttp: 400,
	}
}

// NewConstraintViolationError создает ошибку нарушения ограничения
func NewConstraintViolationError(message string) *GraphQLError {
	return &GraphQLError{
		Message:    message,
		Code:       ErrCodeConstraintViolation,
		StatusHttp: 409,
	}
}

// IsGraphQLError проверяет является ли ошибка GraphQLError
func IsGraphQLError(err error) bool {
	_, ok := err.(*GraphQLError)
	return ok
}
