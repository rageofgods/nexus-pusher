package client

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	clientName = "client"
)

type nexusClientMetrics struct {
	serverStatus prometheus.Gauge
	serverInfo   *prometheus.GaugeVec
	clientInfo   *prometheus.GaugeVec
}

func NewMetrics(registry *prometheus.Registry) *nexusClientMetrics {
	return &nexusClientMetrics{
		serverStatus: promauto.With(registry).NewGauge(prometheus.GaugeOpts{
			Namespace: clientName,
			Name:      "server_status",
			Help:      "Nexus-pusher remote server status. 1 - Ok, 0 - Error"},
		),
		serverInfo: promauto.With(registry).NewGaugeVec(prometheus.GaugeOpts{
			Namespace: clientName,
			Name:      "server_info",
			Help:      "Represents remote nexus-pusher server version and build number",
		}, []string{"version", "build"}),
		clientInfo: promauto.With(registry).NewGaugeVec(prometheus.GaugeOpts{
			Namespace: clientName,
			Name:      "client_info",
			Help:      "Represents nexus-pusher client version and build number",
		}, []string{"version", "build"}),
	}
}

// ClientInfo return metric for "client_info"
func (ncm nexusClientMetrics) ClientInfo() *prometheus.GaugeVec {
	return ncm.clientInfo
}
