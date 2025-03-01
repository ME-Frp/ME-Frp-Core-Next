package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/fatedier/frp/server/metrics"
)

const (
	namespace       = "frp"
	serverSubsystem = "server"
)

var ServerMetrics metrics.ServerMetrics = newServerMetrics()

type serverMetrics struct {
	clientCount     prometheus.Gauge
	proxyCount      *prometheus.GaugeVec
	connectionCount *prometheus.GaugeVec
	trafficIn       *prometheus.CounterVec
	trafficOut      *prometheus.CounterVec
}

func (m *serverMetrics) NewClient() {
	m.clientCount.Inc()
}

func (m *serverMetrics) CloseClient() {
	m.clientCount.Dec()
}

func (m *serverMetrics) NewProxy(_ string, proxyType string) {
	m.proxyCount.WithLabelValues(proxyType).Inc()
}

func (m *serverMetrics) CloseProxy(_ string, proxyType string) {
	m.proxyCount.WithLabelValues(proxyType).Dec()
}

func (m *serverMetrics) OpenConnection(name string, proxyType string) {
	m.connectionCount.WithLabelValues(name, proxyType).Inc()
}

func (m *serverMetrics) CloseConnection(name string, proxyType string) {
	m.connectionCount.WithLabelValues(name, proxyType).Dec()
}

func (m *serverMetrics) AddTrafficIn(name string, proxyType string, trafficBytes int64) {
	m.trafficIn.WithLabelValues(name, proxyType).Add(float64(trafficBytes))
}

func (m *serverMetrics) AddTrafficOut(name string, proxyType string, trafficBytes int64) {
	m.trafficOut.WithLabelValues(name, proxyType).Add(float64(trafficBytes))
}

func newServerMetrics() *serverMetrics {
	m := &serverMetrics{
		clientCount: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: serverSubsystem,
			Name:      "client_counts",
			Help:      "当前客户端数量",
		}),
		proxyCount: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: serverSubsystem,
			Name:      "proxy_counts",
			Help:      "当前隧道数量",
		}, []string{"type"}),
		connectionCount: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: serverSubsystem,
			Name:      "connection_counts",
			Help:      "当前连接数量",
		}, []string{"name", "type"}),
		trafficIn: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: serverSubsystem,
			Name:      "traffic_in",
			Help:      "总入网流量",
		}, []string{"name", "type"}),
		trafficOut: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: serverSubsystem,
			Name:      "traffic_out",
			Help:      "总出网流量",
		}, []string{"name", "type"}),
	}
	prometheus.MustRegister(m.clientCount)
	prometheus.MustRegister(m.proxyCount)
	prometheus.MustRegister(m.connectionCount)
	prometheus.MustRegister(m.trafficIn)
	prometheus.MustRegister(m.trafficOut)
	return m
}
