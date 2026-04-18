package middleware

import (
	"context"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/sirupsen/logrus"
)

// OperationLoggingMiddleware логирует все входящие GraphQL операции (queries и mutations)
// Логирует имя операции, переменные и время выполнения входящего запроса
func OperationLoggingMiddleware(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
	start := time.Now()
	opCtx := graphql.GetOperationContext(ctx)

	// Получить имя операции и тип
	operationName := "unknown"
	if opCtx != nil && opCtx.OperationName != "" {
		operationName = opCtx.OperationName
	}

	operationType := "query"
	if opCtx != nil && opCtx.Operation != nil {
		operationType = string(opCtx.Operation.Operation)
	}

	// Логируем входящий запрос
	logrus.WithFields(logrus.Fields{
		"operation_type": operationType,
		"operation_name": operationName,
		"variables":      opCtx.Variables,
	}).Info("GraphQL operation started")

	// Обрабатываем операцию - получаем ResponseHandler
	responseHandler := next(ctx)

	// Возвращаем новый ResponseHandler который логирует ответ
	return func(ctx context.Context) *graphql.Response {
		response := responseHandler(ctx)
		duration := time.Since(start)
		hasErrors := len(response.Errors) > 0

		fields := logrus.Fields{
			"operation_type": operationType,
			"operation_name": operationName,
			"duration_ms":    duration.Milliseconds(),
			"has_errors":     hasErrors,
		}

		if hasErrors {
			errorMessages := []string{}
			for _, err := range response.Errors {
				errorMessages = append(errorMessages, err.Error())
			}
			fields["errors"] = errorMessages
			logrus.WithFields(fields).Error("GraphQL operation completed with errors")
		} else {
			logrus.WithFields(fields).Info("GraphQL operation completed successfully")
		}

		return response
	}
}


