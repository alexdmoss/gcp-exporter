package collectors

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/urfave/cli"

	col "gitlab.com/gitlab-org/ci-cd/gcp-exporter/collectors/collector"
	"gitlab.com/gitlab-org/ci-cd/gcp-exporter/collectors/instances"
)

const (
	DefaultServiceAccountFile = "~/.google-service-account.json"
)

var (
	Collectors MapInterface = &Map{}

	timeOfLastDataRefresh = prometheus.NewDesc(
		"gcp_exporter_last_data_refresh_timestamp_seconds",
		"Time when last data refresh from GCP was done",
		[]string{},
		nil,
	)

	numberOfDataRerfeshErrors = prometheus.NewDesc(
		"gcp_exporter_data_refresh_errors_total",
		"Total number of errors raised during data refresh from GCP",
		[]string{},
		nil,
	)
)

type ProviderInterface interface {
	prometheus.Collector

	Init(*cli.Context) error
	GetData()
}

type Provider struct {
	client     *http.Client
	collectors []col.Interface

	getDataErrors        uint64
	lastGetDataTimestamp time.Time
}

func (p *Provider) Init(context *cli.Context) error {
	for collectorName, flag := range Collectors.EnableFlagNames() {
		if !context.Bool(flag) {
			continue
		}

		collector := Collectors.Get(collectorName)
		if collector == nil {
			continue
		}

		err := p.registerCollector(collectorName, collector)
		if err != nil {
			return fmt.Errorf("error while initializing collector %s: %v", collectorName, err)
		}
	}

	return nil
}

func (p *Provider) registerCollector(collectorName string, collector col.Interface) error {
	logrus.Infof("Enabling %s", collectorName)
	p.collectors = append(p.collectors, collector)
	return collector.Init(p.client)
}

func (p *Provider) GetData() {
	logrus.Infoln("Getting data from GCP")

	for _, collector := range p.collectors {
		err := collector.GetData()
		if err != nil {
			logrus.WithError(err).Errorln("Error while getting data from GCP")
			p.getDataErrors++
		}
	}

	p.lastGetDataTimestamp = time.Now()
}

func (p *Provider) Describe(ch chan<- *prometheus.Desc) {
	ch <- timeOfLastDataRefresh
	ch <- numberOfDataRerfeshErrors

	for _, collector := range p.collectors {
		collector.Describe(ch)
	}
}

func (p *Provider) Collect(ch chan<- prometheus.Metric) {
	ch <- prometheus.MustNewConstMetric(timeOfLastDataRefresh, prometheus.CounterValue, float64(p.lastGetDataTimestamp.Unix()))
	ch <- prometheus.MustNewConstMetric(numberOfDataRerfeshErrors, prometheus.CounterValue, float64(p.getDataErrors))

	for _, collector := range p.collectors {
		collector.Collect(ch)
	}
}

func NewProvider(client *http.Client) *Provider {
	provider := &Provider{
		client:               client,
		getDataErrors:        0,
		lastGetDataTimestamp: time.Unix(0, 0),
	}
	provider.collectors = make([]col.Interface, 0)

	return provider
}

func init() {
	Collectors.Add(instances.NewCollector())
}
