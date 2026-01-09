package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	dto "github.com/prometheus/client_model/go"
)

func TestStatusWriter_StatusTracking(t *testing.T) {
	t.Parallel()

	t.Run("defaults to 200 when nothing written", func(t *testing.T) {
		t.Parallel()

		sw := &statusWriter{ResponseWriter: httptest.NewRecorder()}
		if got, want := sw.Status(), http.StatusOK; got != want {
			t.Fatalf("Status()=%d want=%d", got, want)
		}
	})

	t.Run("WriteHeader sets status", func(t *testing.T) {
		t.Parallel()

		sw := &statusWriter{ResponseWriter: httptest.NewRecorder()}
		sw.WriteHeader(http.StatusTeapot)
		if got, want := sw.Status(), http.StatusTeapot; got != want {
			t.Fatalf("Status()=%d want=%d", got, want)
		}
	})

	t.Run("Write sets status to 200 if not set", func(t *testing.T) {
		t.Parallel()

		sw := &statusWriter{ResponseWriter: httptest.NewRecorder()}
		_, _ = sw.Write([]byte("x"))
		if got, want := sw.Status(), http.StatusOK; got != want {
			t.Fatalf("Status()=%d want=%d", got, want)
		}
	})

	t.Run("WriteHeader wins over Write", func(t *testing.T) {
		t.Parallel()

		sw := &statusWriter{ResponseWriter: httptest.NewRecorder()}
		sw.WriteHeader(http.StatusCreated)
		_, _ = sw.Write([]byte("x"))
		if got, want := sw.Status(), http.StatusCreated; got != want {
			t.Fatalf("Status()=%d want=%d", got, want)
		}
	})
}

func TestWithPromMetrics_RecordsLabelsAndCounts(t *testing.T) {
	// Not parallel: this test swaps package-level metric vars.

	const (
		route        = "/healthz"
		method       = http.MethodGet
		code         = "200"
		metricReqDur = "requests_duration_seconds_test"
	)

	reg := prometheus.NewRegistry()

	inFlight := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{Name: "in_flight_requests_test"},
		[]string{"route"},
	)
	reqTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "requests_total_test"},
		[]string{"method", "route", "code"},
	)
	reqDur := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{Name: metricReqDur},
		[]string{"method", "route"},
	)

	reg.MustRegister(inFlight, reqTotal, reqDur)

	oldInFlight, oldReqTotal, oldReqDur := httpInFlight, httpRequestsTotal, httpRequestDuration
	httpInFlight, httpRequestsTotal, httpRequestDuration = inFlight, reqTotal, reqDur
	t.Cleanup(func() {
		httpInFlight, httpRequestsTotal, httpRequestDuration = oldInFlight, oldReqTotal, oldReqDur
	})

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok")) // implicit 200
	})

	h := withPromMetrics(route, next)

	req := httptest.NewRequest(method, "http://localhost/healthz", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if got := testutil.ToFloat64(reqTotal.WithLabelValues(method, route, code)); got != 1 {
		t.Fatalf("requests_total=%v want=1", got)
	}

	if got := histogramSampleCount(t, reg, metricReqDur, method, route); got != 1 {
		t.Fatalf("request_duration observations=%d want=1", got)
	}

	if got := testutil.ToFloat64(inFlight.WithLabelValues(route)); got != 0 {
		t.Fatalf("in_flight=%v want=0", got)
	}
}

func histogramSampleCount(t *testing.T, reg prometheus.Gatherer, metricName, method, route string) uint64 {
	t.Helper()

	families, err := reg.Gather()
	if err != nil {
		t.Fatalf("gather: %v", err)
	}

	for _, fam := range families {
		if fam.GetName() != metricName {
			continue
		}
		for _, m := range fam.GetMetric() {
			if hasLabelValue(m.GetLabel(), "method", method) &&
				hasLabelValue(m.GetLabel(), "route", route) {

				h := m.GetHistogram()
				if h == nil {
					t.Fatalf("expected histogram metric")
				}
				return h.GetSampleCount()
			}
		}
	}

	t.Fatalf("no histogram series matched metric=%q method=%q route=%q", metricName, method, route)
	return 0
}

func hasLabelValue(labels []*dto.LabelPair, key, value string) bool {
	for _, lp := range labels {
		if lp.GetName() == key && lp.GetValue() == value {
			return true
		}
	}
	return false
}
