package middleware

import (
	"net/http"

	"github.com/4udiwe/commnets-feed/internal/graph/loaders"
	"github.com/sirupsen/logrus"
)

// DataLoaderMiddleware инициализирует DataLoaders для каждого HTTP запроса
// и добавляет их в context, чтобы резолверы могли их использовать
//
// Это обеспечивает:
// 1. Request-scoped DataLoaders (каждый запрос имеет свой набор)
// 2. Не singleton (разные запросы не делят состояние)
// 3. Батчевание запросов внутри одного GraphQL запроса
func DataLoaderMiddleware(commentService loaders.CommentService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Создать новые DataLoaders ДЛЯ ЭТОГО ЗАПРОСА
			// Каждый HTTP запрос получит свой набор DataLoaders
			dlms := loaders.NewLoaders(commentService)

			logrus.WithFields(logrus.Fields{
				"path":   r.URL.Path,
				"method": r.Method,
			}).Debug("DataLoader middleware: created new loaders for request")

			// Добавить их в context запроса
			ctx := loaders.AttachLoaders(r.Context(), dlms)

			// Продолжить обработку с обновленным context
			// Теперь все резолверы смогут использовать эти DataLoaders
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
