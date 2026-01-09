package httpapi

import (
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	readTimeout       = 15 * time.Second
	readHeaderTimeout = 5 * time.Second
	writeTimeout      = 15 * time.Second
	idleTimeout       = 60 * time.Second
)

var registerMetricsOnce sync.Once

func NewServer(listenAddr string) *http.Server {
	return NewServerWithRegisterer(listenAddr, prometheus.DefaultRegisterer)
}

func NewServerWithRegisterer(listenAddr string, reg prometheus.Registerer) *http.Server {
	registerMetricsOnce.Do(func() {
		RegisterMetrics(reg)
	})

	mux := http.NewServeMux()

	mux.Handle("/healthz", withPromMetrics("/healthz", http.HandlerFunc(handleHealthz)))
	mux.Handle("/metrics", promhttp.Handler())

	return &http.Server{
		Addr:              listenAddr,
		Handler:           mux,
		ReadTimeout:       readTimeout,
		ReadHeaderTimeout: readHeaderTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}
}
