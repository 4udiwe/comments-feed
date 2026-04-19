package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/4udiwe/comments-feed/internal/graph/errors"
	"github.com/99designs/gqlgen/graphql"
	"github.com/sirupsen/logrus"
)

// ResolverErrorLoggingMiddleware логирует ошибки которые возникают в резолверах
func ResolverErrorLoggingMiddleware(ctx context.Context, next graphql.Resolver) (interface{}, error) {
	start := time.Now()

	// Получаем информацию о текущем поле
	fc := graphql.GetFieldContext(ctx)
	fieldName := "unknown"
	fieldType := "unknown"

	if fc != nil {
		fieldName = fc.Field.Name
		if fc.Field.Definition != nil && fc.Field.Definition.Type != nil {
			fieldType = fc.Field.Definition.Type.String()
		}
	}

	// Обрабатываем resolver
	result, err := next(ctx)
	duration := time.Since(start)

	// Если была ошибка, логируем её
	if err != nil {
		fields := logrus.Fields{
			"field_name":  fieldName,
			"field_type":  fieldType,
			"duration_ms": duration.Milliseconds(),
			"error":       err.Error(),
			"error_type":  fmt.Sprintf("%T", err),
		}

		// Если это наша GraphQL ошибка, добавляем код ошибки
		if gqlErr, ok := err.(*errors.GraphQLError); ok {
			fields["error_code"] = gqlErr.Code
			fields["http_status"] = gqlErr.StatusHttp

			// Логируем с учетом серьезности ошибки
			if gqlErr.Original != nil {
				fields["original_error"] = gqlErr.Original.Error()
			}

			if gqlErr.Code == errors.ErrCodeInternal {
				logrus.WithFields(fields).Error("Resolver error (internal)")
			} else {
				logrus.WithFields(fields).Warn("Resolver error (user)")
			}
		} else {
			logrus.WithFields(fields).Error("Resolver error (unexpected)")
		}
	} 

	return result, err
}
