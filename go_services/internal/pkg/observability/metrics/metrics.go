package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Counter 计数器接口
type Counter interface {
	Inc()
	Add(float64)
}

// Gauge 仪表盘接口
type Gauge interface {
	Set(float64)
	Inc()
	Dec()
	Add(float64)
	Sub(float64)
}

// Histogram 直方图接口
type Histogram interface {
	Observe(float64)
}

// NewCounter 创建一个新的计数器
func NewCounter(name, help string, constLabels map[string]string) Counter {
	return promauto.NewCounter(prometheus.CounterOpts{
		Name:        name,
		Help:        help,
		ConstLabels: constLabels,
	})
}

// NewCounterVec 创建一个新的计数器向量
func NewCounterVec(name, help string, labelNames []string) *prometheus.CounterVec {
	return promauto.NewCounterVec(prometheus.CounterOpts{
		Name: name,
		Help: help,
	}, labelNames)
}

// NewGauge 创建一个新的仪表盘
func NewGauge(name, help string, constLabels map[string]string) Gauge {
	return promauto.NewGauge(prometheus.GaugeOpts{
		Name:        name,
		Help:        help,
		ConstLabels: constLabels,
	})
}

// NewGaugeVec 创建一个新的仪表盘向量
func NewGaugeVec(name, help string, labelNames []string) *prometheus.GaugeVec {
	return promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: name,
		Help: help,
	}, labelNames)
}

// NewHistogram 创建一个新的直方图
func NewHistogram(name, help string, buckets []float64, constLabels map[string]string) Histogram {
	return promauto.NewHistogram(prometheus.HistogramOpts{
		Name:        name,
		Help:        help,
		Buckets:     buckets,
		ConstLabels: constLabels,
	})
}

// NewHistogramVec 创建一个新的直方图向量
func NewHistogramVec(name, help string, buckets []float64, labelNames []string) *prometheus.HistogramVec {
	return promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    name,
		Help:    help,
		Buckets: buckets,
	}, labelNames)
}
