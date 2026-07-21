package main

import (
	"net/http"
	"time"

	"github.com/gofrs/uuid"
	"go.uber.org/zap"
)

type responseWriterWithCode struct {
	http.ResponseWriter // embedding
	statusCode          int
}

func newResponseWriterWithCode(w http.ResponseWriter) *responseWriterWithCode {
	return &responseWriterWithCode{ResponseWriter: w, statusCode: http.StatusOK}
}

func (rw *responseWriterWithCode) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func Log(l *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			requestID := uuid.Must(uuid.NewV4()).String()

			l.Info("request started",
				zap.String("request_id", requestID),
				zap.String("path", r.URL.Path),
				zap.String("method", r.Method),
			)
			defer func() {
				if rec := recover(); rec != nil {
					l.Panic("request panicked",
						zap.String("request_id", requestID),
						zap.String("path", r.URL.Path),
						zap.String("method", r.Method),
						zap.Any("panic", rec),
						zap.Duration("duration", time.Since(start)),
					)
				}
			}()

			rw := newResponseWriterWithCode(w)
			next.ServeHTTP(rw, r)

			l.Info("request finished",
				zap.String("request_id", requestID),
				zap.String("path", r.URL.Path),
				zap.String("method", r.Method),
				zap.Int("status_code", rw.statusCode),
				zap.Duration("duration", time.Since(start)),
			)
		})
	}
}
