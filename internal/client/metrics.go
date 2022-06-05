package client

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	log "github.com/sirupsen/logrus"
)

const (
	clientName = "client"
)

type nexusClientMetrics struct {
	staticMetrics  *staticMetrics
	dynamicMetrics *syncConfigMetrics
}

type staticMetrics struct {
	serverStatus prometheus.Gauge
	serverInfo   *prometheus.GaugeVec
	clientInfo   *prometheus.GaugeVec
}

type syncConfigMetrics struct {
	lastErrorsCount *prometheus.GaugeVec
}

func NewMetrics(registry *prometheus.Registry) *nexusClientMetrics {
	return &nexusClientMetrics{
		&staticMetrics{
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
		},
		&syncConfigMetrics{
			lastErrorsCount: promauto.With(registry).NewGaugeVec(prometheus.GaugeOpts{
				Namespace: clientName,
				Subsystem: "last",
				Name:      "sync_errors_total",
				Help:      "Represents errors count for sync config",
			}, []string{labelDestinationServer, labelDestinationRepo, labelId}),
		},
	}
}

// ClientInfo return metric for "client_info"
func (ncm nexusClientMetrics) ClientInfo() *prometheus.GaugeVec {
	return ncm.staticMetrics.clientInfo
}

func (ncm nexusClientMetrics) SyncErrorsCountByLabels(server string, repo string, id string) prometheus.Gauge {
	g, err := ncm.dynamicMetrics.lastErrorsCount.GetMetricWith(prometheus.Labels{
		labelDestinationServer: server,
		labelDestinationRepo:   repo,
		labelId:                id,
	})
	if err != nil {
		log.Errorf("unable to set dynamic metric for destination repo %s: %v", repo, err)
		return nil
	}
	return g
}

const (
	labelDestinationServer = "destination_server"
	labelDestinationRepo   = "destination_repo"
	labelId                = "id"
)
