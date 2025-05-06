package middleware

import (
	"net/http"
	"strconv"
	"time"
	"weather-api/pkg/metrics"

	"github.com/gorilla/mux"
)

// MetricsMiddleware создает middleware для сбора метрик HTTP запросов
func MetricsMiddleware(metrics *metrics.Metrics) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Создаем ResponseWriter, который может записывать статус-код
			ww := NewResponseWriter(w)

			// Вызываем следующий обработчик
			next.ServeHTTP(ww, r)

			// Рассчитываем длительность запроса
			duration := time.Since(start).Seconds()

			// Получаем информацию о запросе
			route := mux.CurrentRoute(r)
			path, _ := route.GetPathTemplate()
			method := r.Method
			status := strconv.Itoa(ww.Status())

			// Обновляем метрики
			metrics.HttpRequestsTotal.WithLabelValues(method, path, status).Inc()
			metrics.HttpRequestDuration.WithLabelValues(method, path).Observe(duration)
		})
	}
}

// ResponseWriter оборачивает http.ResponseWriter для отслеживания статус-кода
type ResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

// NewResponseWriter создает новый ResponseWriter
func NewResponseWriter(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{w, http.StatusOK}
}

// WriteHeader реализует интерфейс http.ResponseWriter
func (rw *ResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Status возвращает статус-код ответа
func (rw *ResponseWriter) Status() int {
	return rw.statusCode
}
