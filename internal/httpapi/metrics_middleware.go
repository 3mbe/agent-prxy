package httpapi

import (
	"net/http"
	"strconv"
	"time"
)

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (sw *statusWriter) WriteHeader(code int) {
	sw.status = code
	sw.ResponseWriter.WriteHeader(code)
}

func (sw *statusWriter) Write(p []byte) (int, error) {
	if sw.status == 0 {
		sw.status = http.StatusOK
	}
	return sw.ResponseWriter.Write(p)
}

func (sw *statusWriter) Status() int {
	if sw.status == 0 {
		return http.StatusOK
	}
	return sw.status
}

// withPromMetrics instruments an HTTP handler with Prometheus metrics.
// The route label must be a fixed value (e.g. "/healthz"), not a raw path.
func withPromMetrics(route string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sw := &statusWriter{ResponseWriter: w}
		start := time.Now()

		httpInFlight.WithLabelValues(route).Inc()
		defer httpInFlight.WithLabelValues(route).Dec()

		next.ServeHTTP(sw, r)

		method := r.Method
		code := strconv.Itoa(sw.Status())

		httpRequestsTotal.WithLabelValues(method, route, code).Inc()
		httpRequestDuration.WithLabelValues(method, route).
			Observe(time.Since(start).Seconds())
	})
}
