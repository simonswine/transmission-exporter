package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	transmission "github.com/metalmatze/transmission-exporter"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sync/errgroup"
)

type portStatusCollector struct {
	client *transmission.Client

	portStatus *prometheus.GaugeVec
	interval   time.Duration
	lastUpdate time.Time
}

func NewPortStatusCollector(client *transmission.Client, interval time.Duration) prometheus.Collector {
	return &portStatusCollector{
		client:   client,
		interval: interval,
		portStatus: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: strings.TrimRight(namespace, "_"),
			Name:      "port_open",
			Help:      "Free space left on device to download to",
		}, []string{"port"}),
	}
}

func (c *portStatusCollector) Describe(ch chan<- *prometheus.Desc) {
	c.portStatus.Describe(ch)
}

func (c *portStatusCollector) Collect(ch chan<- prometheus.Metric) {
	defer c.portStatus.Collect(ch)
	if c.lastUpdate.Before(time.Now().Add(-c.interval)) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		value := 0.0
		portNumber := -1

		g, ctx := errgroup.WithContext(ctx)
		g.Go(func() error {
			portOpen, err := c.client.TestPort(ctx)
			if err != nil {
				return err
			}
			if portOpen {
				value = 1.0
			}
			return nil
		})
		g.Go(func() error {
			stats, err := c.client.GetSession(ctx)
			if err != nil {
				return err
			}
			portNumber = int(*stats.PeerPort)
			return nil
		})

		if err := g.Wait(); err != nil {
			log.Printf("Failed to get port status: %v", err)
			return
		}

		c.portStatus.WithLabelValues(fmt.Sprintf("%d", portNumber)).Set(value)
		c.lastUpdate = time.Now()
	}

}
