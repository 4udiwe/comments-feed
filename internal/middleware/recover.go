package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/sirupsen/logrus"
)

func NewRecoverMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					logrus.Errorf("panic in HTTP: %v\n%s", rec, debug.Stack())

					w.WriteHeader(http.StatusInternalServerError)
					_, _ = w.Write([]byte("internal server error"))
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}