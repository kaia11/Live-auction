package http

import (
	"bufio"
	"net"
	"net/http"
	"time"

	"auction-live/backend/internal/logger"
	"auction-live/backend/internal/monitoring"
)

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (r *statusRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *statusRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := r.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, http.ErrNotSupported
	}
	return hijacker.Hijack()
}

func WithRequestLogging(next http.Handler, metrics *monitoring.Metrics) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		recorder := &statusRecorder{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(recorder, r)
		durationMS := time.Since(start).Milliseconds()
		if metrics != nil {
			metrics.RecordHTTPRequest(durationMS)
		}

		logger.Info(
			"http request method=%s path=%s status=%d duration_ms=%d remote=%s",
			r.Method,
			r.URL.Path,
			recorder.statusCode,
			durationMS,
			r.RemoteAddr,
		)
	})
}
