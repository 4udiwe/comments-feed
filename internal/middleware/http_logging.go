package middleware

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

// HTTPLoggingMiddleware логирует все входящие и исходящие HTTP запросы
// Логирует:
// - Метод и URL запроса
// - Заголовки запроса
// - Статус ответа
// - Время выполнения
// - Размер ответа
type HTTPLoggingMiddleware struct {
	logger *logrus.Logger
}

// NewHTTPLoggingMiddleware создает новый HTTP logging middleware
func NewHTTPLoggingMiddleware() *HTTPLoggingMiddleware {
	return &HTTPLoggingMiddleware{
		logger: logrus.StandardLogger(),
	}
}

// Handler возвращает HTTP middleware
func (h *HTTPLoggingMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Логируем входящий запрос
		h.logger.WithFields(logrus.Fields{
			"method":        r.Method,
			"path":          r.URL.Path,
			"query":         r.URL.RawQuery,
			"remote_addr":   r.RemoteAddr,
			"user_agent":    r.UserAgent(),
			"content_type":  r.Header.Get("Content-Type"),
			"accept":        r.Header.Get("Accept"),
		}).Info("HTTP request received")

		// Оборачиваем ResponseWriter для захвата кода статуса и размера ответа
		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
			body:           bytes.NewBuffer(nil),
		}

		// Обрабатываем запрос
		next.ServeHTTP(wrapped, r)

		// Логируем завершенный запрос
		duration := time.Since(start)
		h.logger.WithFields(logrus.Fields{
			"method":           r.Method,
			"path":             r.URL.Path,
			"status_code":      wrapped.statusCode,
			"duration_ms":      duration.Milliseconds(),
			"response_size":    wrapped.body.Len(),
			"remote_addr":      r.RemoteAddr,
		}).Info("HTTP request completed")
	})
}

// responseWriter обертка для ResponseWriter для захвата статуса и тела ответа
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	body       *bytes.Buffer
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	// Копируем тело в буфер для логирования
	// (но это может быть слишком много для больших отповедей)
	_, _ = rw.body.Write(b)
	return rw.ResponseWriter.Write(b)
}

// Flush реализует http.Flusher интерфейс
func (rw *responseWriter) Flush() {
	if flusher, ok := rw.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

// CloseNotify реализует http.CloseNotifier интерфейс
func (rw *responseWriter) CloseNotify() <-chan bool {
	if closeNotifier, ok := rw.ResponseWriter.(http.CloseNotifier); ok {
		return closeNotifier.CloseNotify()
	}
	return nil
}

// ReadFrom реализует io.ReaderFrom интерфейс
func (rw *responseWriter) ReadFrom(r io.Reader) (n int64, err error) {
	if readerFrom, ok := rw.ResponseWriter.(io.ReaderFrom); ok {
		return readerFrom.ReadFrom(r)
	}
	return io.Copy(rw.ResponseWriter, r)
}
