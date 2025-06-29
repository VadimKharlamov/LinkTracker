package prometheus

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus"

	"net/http"
	"strconv"
	"time"
)

var (
	httpRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "myapp",
			Name:      "http_requests_total",
			Help:      "Всего HTTP-запросов",
		},
		[]string{"handler", "method", "code"},
	)
)

var errorsTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "myapp",
		Name:      "errors_total",
		Help:      "Количество ошибок",
	},
	[]string{"handler", "method", "code"},
)

var httpDuration = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Namespace: "myapp",
		Name:      "http_request_duration_seconds",
		Help:      "Время обработки HTTP-запросов",
		Buckets:   prometheus.DefBuckets,
	},
	[]string{"handler", "method"},
)

func init() {
	prometheus.MustRegister(httpRequests)
	prometheus.MustRegister(httpDuration)
	prometheus.MustRegister(errorsTotal)
}

func PrometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rr := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		next.ServeHTTP(rr, r)

		duration := time.Since(start).Seconds()
		route := chi.RouteContext(r.Context()).RoutePattern()

		httpRequests.WithLabelValues(
			route,
			r.Method,
			strconv.Itoa(rr.Status()),
		).Inc()
		httpDuration.WithLabelValues(route, r.Method).Observe(duration)

		if rr.Status() >= 400 {
			errorsTotal.WithLabelValues(route, r.Method, strconv.Itoa(rr.Status())).Inc()
		}
	})
}
