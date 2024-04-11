package main

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/metalmatze/transmission-exporter"
	"github.com/prometheus/client_golang/prometheus"
)

// SessionCollector exposes session metrics
type SessionCollector struct {
	client *transmission.Client

	AltSpeedDown     *prometheus.Desc
	AltSpeedUp       *prometheus.Desc
	CacheSize        *prometheus.Desc
	FreeSpace        *prometheus.GaugeVec
	TotalSpace       *prometheus.GaugeVec
	QueueDown        *prometheus.Desc
	QueueUp          *prometheus.Desc
	PeerLimitGlobal  *prometheus.Desc
	PeerLimitTorrent *prometheus.Desc
	SeedRatioLimit   *prometheus.Desc
	SpeedLimitDown   *prometheus.Desc
	SpeedLimitUp     *prometheus.Desc
	Version          *prometheus.Desc
}

// NewSessionCollector takes a transmission.Client and returns a SessionCollector
func NewSessionCollector(client *transmission.Client) *SessionCollector {
	return &SessionCollector{
		client: client,

		AltSpeedDown: prometheus.NewDesc(
			namespace+"alt_speed_down",
			"Alternative max global download speed",
			[]string{"enabled"},
			nil,
		),
		AltSpeedUp: prometheus.NewDesc(
			namespace+"alt_speed_up",
			"Alternative max global upload speed",
			[]string{"enabled"},
			nil,
		),
		CacheSize: prometheus.NewDesc(
			namespace+"cache_size_bytes",
			"Maximum size of the disk cache",
			nil,
			nil,
		),
		FreeSpace: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: strings.TrimRight(namespace, "_"),
			Name:      "free_space_bytes",
			Help:      "Free space left on device to download to",
		}, []string{"type", "path"}),
		TotalSpace: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: strings.TrimRight(namespace, "_"),
			Name:      "total_space_bytes",
			Help:      "Total space on device to download to",
		}, []string{"type", "path"}),
		QueueDown: prometheus.NewDesc(
			namespace+"queue_down",
			"Max number of torrents to download at once",
			[]string{"enabled"},
			nil,
		),
		QueueUp: prometheus.NewDesc(
			namespace+"queue_up",
			"Max number of torrents to upload at once",
			[]string{"enabled"},
			nil,
		),
		PeerLimitGlobal: prometheus.NewDesc(
			namespace+"global_peer_limit",
			"Maximum global number of peers",
			nil,
			nil,
		),
		PeerLimitTorrent: prometheus.NewDesc(
			namespace+"torrent_peer_limit",
			"Maximum number of peers for a single torrent",
			nil,
			nil,
		),
		SeedRatioLimit: prometheus.NewDesc(
			namespace+"seed_ratio_limit",
			"The default seed ratio for torrents to use",
			[]string{"enabled"},
			nil,
		),
		SpeedLimitDown: prometheus.NewDesc(
			namespace+"speed_limit_down_bytes",
			"Max global download speed",
			[]string{"enabled"},
			nil,
		),
		SpeedLimitUp: prometheus.NewDesc(
			namespace+"speed_limit_up_bytes",
			"Max global upload speed",
			[]string{"enabled"},
			nil,
		),
		Version: prometheus.NewDesc(
			namespace+"version",
			"Transmission version as label",
			[]string{"version"},
			nil,
		),
	}
}

// Describe implements the prometheus.Collector interface
func (sc *SessionCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- sc.AltSpeedDown
	ch <- sc.AltSpeedUp
	ch <- sc.CacheSize
	sc.FreeSpace.Describe(ch)
	sc.TotalSpace.Describe(ch)
	ch <- sc.QueueDown
	ch <- sc.QueueUp
	ch <- sc.PeerLimitGlobal
	ch <- sc.PeerLimitTorrent
	ch <- sc.SeedRatioLimit
	ch <- sc.SpeedLimitDown
	ch <- sc.SpeedLimitUp
	ch <- sc.Version
}

// Collect implements the prometheus.Collector interface
func (sc *SessionCollector) Collect(ch chan<- prometheus.Metric) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	session, err := sc.client.GetSession(ctx)
	if err != nil {
		log.Printf("failed to get session: %v", err)
		return
	}

	free, total, err := sc.client.FreeSpace(ctx, *session.DownloadDir)
	typ := "download-dir"
	if err != nil {
		log.Printf("failed to download dir free space: %v", err)
		sc.FreeSpace.DeleteLabelValues(typ, *session.DownloadDir)
	} else {
		sc.FreeSpace.WithLabelValues(typ, *session.DownloadDir).Set(float64(free.Byte()))
		sc.TotalSpace.WithLabelValues(typ, *session.DownloadDir).Set(float64(total.Byte()))
	}

	free, total, err = sc.client.FreeSpace(ctx, *session.IncompleteDir)
	typ = "incomplete-dir"
	if err != nil {
		log.Printf("failed to incomplete dir free space: %v", err)
		sc.FreeSpace.DeleteLabelValues(typ, *session.IncompleteDir)
	} else {
		sc.FreeSpace.WithLabelValues(typ, *session.IncompleteDir).Set(float64(free.Byte()))
		sc.TotalSpace.WithLabelValues(typ, *session.IncompleteDir).Set(float64(total.Byte()))
	}
	sc.FreeSpace.Collect(ch)
	sc.TotalSpace.Collect(ch)

	ch <- prometheus.MustNewConstMetric(
		sc.AltSpeedDown,
		prometheus.GaugeValue,
		float64(*session.AltSpeedDown),
		boolToString(*session.AltSpeedEnabled),
	)
	ch <- prometheus.MustNewConstMetric(
		sc.AltSpeedUp,
		prometheus.GaugeValue,
		float64(*session.AltSpeedUp),
		boolToString(*session.AltSpeedEnabled),
	)
	ch <- prometheus.MustNewConstMetric(
		sc.CacheSize,
		prometheus.GaugeValue,
		float64(*session.CacheSizeMB*1024*1024),
	)
	ch <- prometheus.MustNewConstMetric(
		sc.QueueDown,
		prometheus.GaugeValue,
		float64(*session.DownloadQueueSize),
		boolToString(*session.DownloadQueueEnabled),
	)
	ch <- prometheus.MustNewConstMetric(
		sc.QueueUp,
		prometheus.GaugeValue,
		float64(*session.SeedQueueSize),
		boolToString(*session.SeedQueueEnabled),
	)
	ch <- prometheus.MustNewConstMetric(
		sc.PeerLimitGlobal,
		prometheus.GaugeValue,
		float64(*session.PeerLimitGlobal),
	)
	ch <- prometheus.MustNewConstMetric(
		sc.PeerLimitTorrent,
		prometheus.GaugeValue,
		float64(*session.PeerLimitPerTorrent),
	)
	ch <- prometheus.MustNewConstMetric(
		sc.SeedRatioLimit,
		prometheus.GaugeValue,
		float64(*session.SeedRatioLimit),
		boolToString(*session.SeedRatioLimited),
	)
	ch <- prometheus.MustNewConstMetric(
		sc.SpeedLimitDown,
		prometheus.GaugeValue,
		float64(*session.SpeedLimitDown),
		boolToString(*session.SpeedLimitDownEnabled),
	)
	ch <- prometheus.MustNewConstMetric(
		sc.SpeedLimitUp,
		prometheus.GaugeValue,
		float64(*session.SpeedLimitUp),
		boolToString(*session.SpeedLimitUpEnabled),
	)
	ch <- prometheus.MustNewConstMetric(
		sc.Version,
		prometheus.GaugeValue,
		float64(1),
		*session.Version,
	)
}
