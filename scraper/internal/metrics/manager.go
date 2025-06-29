package metrics

import (
	"context"
	"runtime"
	"scraper/internal/storage/postgres"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type MetricManager struct {
	memUsage        *prometheus.GaugeVec
	dbSizeGauge     *prometheus.GaugeVec
	requestDuration *prometheus.HistogramVec
	userMessages    *prometheus.CounterVec
}

func NewMetricManager() *MetricManager {
	memUsage := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "myapp",
			Name:      "memory_usage_bytes",
			Help:      "Memory usage by type",
		},
		[]string{"type"},
	)

	dbSizeGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "myapp",
			Name:      "db_items_count",
			Help:      "Current number of items in the database",
		},
		[]string{"type"},
	)

	requestDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "myapp",
			Name:      "client_duration_seconds",
			Help:      "Duration of clients GetUpdates calls",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"type"},
	)

	userMessages := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "myapp",
			Name:      "user_messages_total",
			Help:      "Total user messages received",
		},
		[]string{"metricType"},
	)

	prometheus.MustRegister(userMessages)
	prometheus.MustRegister(dbSizeGauge)
	prometheus.MustRegister(memUsage)
	prometheus.MustRegister(requestDuration)

	return &MetricManager{memUsage: memUsage, dbSizeGauge: dbSizeGauge,
		requestDuration: requestDuration, userMessages: userMessages}
}

func (m *MetricManager) StartCollecting() {
	go func() {
		for {
			var stats runtime.MemStats

			runtime.ReadMemStats(&stats)

			m.memUsage.WithLabelValues("heap_alloc").Set(float64(stats.HeapAlloc))
			m.memUsage.WithLabelValues("heap_inuse").Set(float64(stats.HeapInuse))
			m.memUsage.WithLabelValues("stack_inuse").Set(float64(stats.StackInuse))
			m.memUsage.WithLabelValues("gc_sys").Set(float64(stats.GCSys))
			m.memUsage.WithLabelValues("sys").Set(float64(stats.Sys))

			time.Sleep(15 * time.Second)
		}
	}()
}

func (m *MetricManager) CheckDBMetric(ctx context.Context, storage postgres.Storage) {
	go func() {
		for {
			count, err := storage.UpdateMetric(ctx, "github")
			if err == nil {
				m.dbSizeGauge.WithLabelValues("Github").Set(float64(count))
			}

			count, err = storage.UpdateMetric(ctx, "stackoverflow")
			if err == nil {
				m.dbSizeGauge.WithLabelValues("StackOverFlow").Set(float64(count))
			}

			time.Sleep(5 * time.Minute)
		}
	}()
}

func (m *MetricManager) IncDBMetric(metricType string) {
	m.dbSizeGauge.WithLabelValues(metricType).Inc()
}

func (m *MetricManager) DecDBMetric(metricType string) {
	m.dbSizeGauge.WithLabelValues(metricType).Dec()
}

func (m *MetricManager) SyncDBMetricFromDB(metricType string, count int) {
	m.dbSizeGauge.WithLabelValues(metricType).Set(float64(count))
}

func (m *MetricManager) IncCounterMetric(metricType string) {
	m.userMessages.WithLabelValues(metricType).Inc()
}

func (m *MetricManager) ObserveCallDuration(metricType string, duration float64) {
	m.requestDuration.WithLabelValues(metricType).Observe(duration)
}

func (m *MetricManager) GetDBGauge(label string) prometheus.Gauge {
	return m.dbSizeGauge.WithLabelValues(label)
}

func (m *MetricManager) GetUserCounter(label string) prometheus.Counter {
	return m.userMessages.WithLabelValues(label)
}

func (m *MetricManager) GetRequestDurationCollector() prometheus.Collector {
	return m.requestDuration
}
