package errors

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

// ErrorPresenter преобразует ошибки в GraphQL формат
// Логирование ошибок происходит в ResolverErrorLoggingMiddleware, поэтому здесь мы только преобразуем
func ErrorPresenter(ctx context.Context, err error) *gqlerror.Error {
	// Если это уже GraphQL ошибка, просто преобразуем (логирование уже произошло в middleware)
	if gqlErr, ok := err.(*GraphQLError); ok {
		return gqlErr.ToGQLError()
	}

	// Если это уже gqlerror.Error (обёрнута в middleware), просто вернем её
	// Это значит что error уже залогирована в ResolverErrorLoggingMiddleware
	if gqlErr, ok := err.(*gqlerror.Error); ok {
		return gqlErr
	}

	// Если это неожиданная ошибка которая не была обёрнута в GraphQLError,
	// логируем её как неожиданную и преобразуем в Generic internal error
	logger := logrus.WithContext(ctx)
	logger.WithFields(logrus.Fields{
		"error_message": err.Error(),
		"error_type":    fmt.Sprintf("%T", err),
	}).Error("Unexpected error in resolver")

	return &gqlerror.Error{
		Message: "Internal server error",
		Extensions: map[string]interface{}{
			"code":   ErrCodeInternal,
			"status": 500,
		},
	}
}

// ValidateInput валидирует входные данные GraphQL мутаций
// Проверяет пустые строки, длины и другие стандартные ограничения
type InputValidator struct{}

// NewInputValidator создает новый валидатор
func NewInputValidator() *InputValidator {
	return &InputValidator{}
}

// ValidateString валидирует строковое поле
// Проверяет:
// - не пусто
// - не состоит только из пробелов
// - длина в пределах лимита
func (v *InputValidator) ValidateString(value string, fieldName string, minLen int, maxLen int) *GraphQLError {
	// Проверка на пустую строку
	if value == "" {
		return NewValidationError(fmt.Sprintf("%s cannot be empty", fieldName))
	}

	// Проверка на строку только из пробелов
	if len(value) > 0 && isWhitespaceOnly(value) {
		return NewValidationError(fmt.Sprintf("%s cannot be empty or contain only whitespace", fieldName))
	}

	// Проверка минимальной длины
	if minLen > 0 && len(value) < minLen {
		return NewValidationError(fmt.Sprintf("%s must be at least %d characters", fieldName, minLen))
	}

	// Проверка максимальной длины
	if maxLen > 0 && len(value) > maxLen {
		return NewValidationError(fmt.Sprintf("%s must not exceed %d characters", fieldName, maxLen))
	}

	return nil
}

// ValidateID валидирует ID поле
func (v *InputValidator) ValidateID(id string, fieldName string) *GraphQLError {
	if id == "" {
		return NewValidationError(fmt.Sprintf("%s cannot be empty", fieldName))
	}

	if !isValidID(id) {
		return NewValidationError(fmt.Sprintf("%s has invalid format", fieldName))
	}

	return nil
}

// ValidateInt валидирует целое число
func (v *InputValidator) ValidateInt(value int32, fieldName string, min int32, max int32) *GraphQLError {
	if min > 0 && value < min {
		return NewValidationError(fmt.Sprintf("%s must be at least %d", fieldName, min))
	}

	if max > 0 && value > max {
		return NewValidationError(fmt.Sprintf("%s must not exceed %d", fieldName, max))
	}

	return nil
}

// ValidateOptionalString валидирует опциональное строковое поле
func (v *InputValidator) ValidateOptionalString(value *string, fieldName string, minLen int, maxLen int) *GraphQLError {
	// Если значение не установлено, это OK для опционального поля
	if value == nil {
		return nil
	}

	// Если значение пустое, это тоже OK для опционального поля
	if *value == "" {
		return nil
	}

	// Если значение установлено, обычная валидация
	return v.ValidateString(*value, fieldName, minLen, maxLen)
}

// isWhitespaceOnly проверяет состоит ли строка только из пробелов
func isWhitespaceOnly(s string) bool {
	for _, c := range s {
		if c != ' ' && c != '\t' && c != '\n' && c != '\r' {
			return false
		}
	}
	return true
}

// isValidID простая проверка формата ID
// UUID v4 формат или простая строка из букв и цифр
func isValidID(id string) bool {
	if len(id) == 0 {
		return false
	}

	// Проверяем что ID не состоит только из пробелов
	if isWhitespaceOnly(id) {
		return false
	}

	return true
}
