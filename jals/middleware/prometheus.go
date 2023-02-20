// based on https://github.com/766b/chi-prometheus/blob/master/middleware.go
package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus"
)

// Middleware is a handler that exposes prometheus metrics for the number of requests,
// the latency and the response size, partitioned by status code, method and HTTP path.
type PrometheusMiddleware struct {
	reqs    *prometheus.CounterVec
	latency *prometheus.HistogramVec
}

// NewPrometheusMiddleware returns a new prometheus Middleware handler that groups requests by the chi routing pattern.
// EX: /users/{firstName} instead of /users/bob
func NewPrometheusMiddleware(name string, buckets ...float64) func(next http.Handler) http.Handler {
	var m PrometheusMiddleware
	m.reqs = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:        name + "_requests_total",
			Help:        "How many HTTP requests processed, partitioned by status code, method and HTTP path (with patterns).",
			ConstLabels: prometheus.Labels{"service": name},
		},
		[]string{"code", "method", "path"},
	)
	prometheus.MustRegister(m.reqs)

	if len(buckets) == 0 {
		buckets = []float64{300, 1200, 5000}
	}
	m.latency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:        name + "_request_duration_milliseconds",
		Help:        "How long it took to process the request, partitioned by status code, method and HTTP path (with patterns).",
		ConstLabels: prometheus.Labels{"service": name},
		Buckets:     buckets,
	},
		[]string{"code", "method", "path"},
	)
	prometheus.MustRegister(m.latency)
	return m.patternHandler
}

func (c PrometheusMiddleware) patternHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)

		rctx := chi.RouteContext(r.Context())
		if rctx == nil {
			// avoid nil panic
			return
		}
		routePattern := strings.Join(rctx.RoutePatterns, "")
		routePattern = strings.Replace(routePattern, "/*/", "/", -1)

		c.reqs.WithLabelValues(http.StatusText(ww.Status()), r.Method, routePattern).Inc()
		c.latency.WithLabelValues(http.StatusText(ww.Status()), r.Method, routePattern).Observe(float64(time.Since(start).Nanoseconds()) / 1000000)
	}
	return http.HandlerFunc(fn)
}
