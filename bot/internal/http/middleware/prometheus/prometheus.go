package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
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
