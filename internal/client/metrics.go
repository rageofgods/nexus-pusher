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
	lastSyncTime           *prometheus.GaugeVec
	lastSrcRepoAssetsCount *prometheus.GaugeVec
	lastDstRepoAssetsCount *prometheus.GaugeVec
	lastSyncDiffCount      *prometheus.GaugeVec
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
			lastSyncTime: promauto.With(registry).NewGaugeVec(prometheus.GaugeOpts{
				Namespace: clientName,
				Subsystem: "last",
				Name:      "sync_info_seconds",
				Help:      "Represents time in unix format of last successful sync operation",
			}, []string{labelDestinationServer, labelDestinationRepo, labelId, labelErrorsCount}),
			lastSrcRepoAssetsCount: promauto.With(registry).NewGaugeVec(prometheus.GaugeOpts{
				Namespace: clientName,
				Subsystem: "last",
				Name:      "src_assets_total",
				Help:      "Represents total count of source repository assets found at last sync iteration",
			}, []string{labelSourceServer, labelSourceRepo}),
			lastDstRepoAssetsCount: promauto.With(registry).NewGaugeVec(prometheus.GaugeOpts{
				Namespace: clientName,
				Subsystem: "last",
				Name:      "dst_assets_total",
				Help:      "Represents total count of destination repository assets found at last sync iteration",
			}, []string{labelDestinationServer, labelDestinationRepo}),
			lastSyncDiffCount: promauto.With(registry).NewGaugeVec(prometheus.GaugeOpts{
				Namespace: clientName,
				Subsystem: "last",
				Name:      "sync_diff_total",
				Help:      "Represents total count of sync differences between src and dst repos",
			}, []string{labelSourceServer, labelSourceRepo, labelDestinationServer, labelDestinationRepo}),
		},
	}
}

// ClientInfo return metric for "client_info"
func (ncm nexusClientMetrics) ClientInfo() *prometheus.GaugeVec {
	return ncm.staticMetrics.clientInfo
}

func (ncm nexusClientMetrics) LastSyncTimeByLabels(server, repo, id, errorsCount string) prometheus.Gauge {
	g, err := ncm.dynamicMetrics.lastSyncTime.GetMetricWith(prometheus.Labels{
		labelDestinationServer: server,
		labelDestinationRepo:   repo,
		labelId:                id,
		labelErrorsCount:       errorsCount,
	})
	if err != nil {
		log.Errorf("LastSyncTimeByLabels: unable to set dynamic metric for destination repo %s: %v",
			repo, err)
		return nil
	}
	return g
}

func (ncm nexusClientMetrics) LastSrcAssetsCountByLabels(server, repo string) prometheus.Gauge {
	g, err := ncm.dynamicMetrics.lastSrcRepoAssetsCount.GetMetricWith(prometheus.Labels{
		labelSourceServer: server,
		labelSourceRepo:   repo,
	})
	if err != nil {
		log.Errorf("LastSrcAssetsCountByLabels: unable to set dynamic metric for destination repo %s: %v",
			repo, err)
		return nil
	}
	return g
}

func (ncm nexusClientMetrics) LastDstAssetsCountByLabels(server, repo string) prometheus.Gauge {
	g, err := ncm.dynamicMetrics.lastDstRepoAssetsCount.GetMetricWith(prometheus.Labels{
		labelDestinationServer: server,
		labelDestinationRepo:   repo,
	})
	if err != nil {
		log.Errorf("LastDstAssetsCountByLabels: unable to set dynamic metric for destination repo %s: %v",
			repo, err)
		return nil
	}
	return g
}

func (ncm nexusClientMetrics) LastSyncDiffByLabels(srcServer, srcRepo, dstServer, dstRepo string) prometheus.Gauge {
	g, err := ncm.dynamicMetrics.lastSyncDiffCount.GetMetricWith(prometheus.Labels{
		labelSourceServer:      srcServer,
		labelSourceRepo:        srcRepo,
		labelDestinationServer: dstServer,
		labelDestinationRepo:   dstRepo,
	})
	if err != nil {
		log.Errorf("LastSyncDiffByLabels: unable to set dynamic metric for destination repo %s: %v",
			dstRepo, err)
		return nil
	}
	return g
}

const (
	labelDestinationServer = "destination_server"
	labelDestinationRepo   = "destination_repo"
	labelSourceServer      = "source_server"
	labelSourceRepo        = "source_repo"
	labelId                = "id"
	labelErrorsCount       = "errors"
)
