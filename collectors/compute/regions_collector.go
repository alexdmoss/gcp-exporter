package compute

import (
	"fmt"
	"net/http"
	"strings"
	"sync"

	"google.golang.org/api/compute/v1"

	"github.com/Sirupsen/logrus"
	"github.com/prometheus/client_golang/prometheus"

	"gitlab.com/gitlab-org/ci-cd/gcp-exporter/client/services"
)

const (
	RegionsCollectorName = "regions-collector"
)

var (
	regionQuotaUsages = prometheus.NewDesc(
		"gcp_exporter_region_quota_usage",
		"Current usage for regions quotas",
		[]string{"project", "region", "quota"},
		nil,
	)

	regionQuotaLimits = prometheus.NewDesc(
		"gcp_exporter_region_quota_limit",
		"Current limit for regions quotas",
		[]string{"project", "region", "quota"},
		nil,
	)
)

type quotaMetricType int

const (
	QuotaMetricTypeUsage quotaMetricType = iota
	QuotaMetricTypeLimit
)

type regionQuotasPermutation struct {
	Project string
	Region  string
	Quota   string
}

type regionQuotasCounterInterface interface {
	Add(string, string, []*compute.Quota)
	Collect(chan<- prometheus.Metric)
}

type regionQuotasCounter struct {
	count map[regionQuotasPermutation]map[quotaMetricType]float64
	lock  sync.RWMutex
}

func (ic *regionQuotasCounter) Add(project string, region string, quotas []*compute.Quota) {
	ic.lock.Lock()
	defer ic.lock.Unlock()

	for _, quota := range quotas {
		permutation := regionQuotasPermutation{
			Project: project,
			Region:  region,
			Quota:   quota.Metric,
		}

		_, ok := ic.count[permutation]
		if !ok {
			ic.count[permutation] = make(map[quotaMetricType]float64)
		}
		ic.count[permutation][QuotaMetricTypeUsage] = quota.Usage
		ic.count[permutation][QuotaMetricTypeLimit] = quota.Limit
	}
}

func (ic *regionQuotasCounter) Collect(ch chan<- prometheus.Metric) {
	for permutation, count := range ic.count {
		ch <- prometheus.MustNewConstMetric(
			regionQuotaUsages,
			prometheus.GaugeValue,
			count[QuotaMetricTypeUsage],
			permutation.Project,
			permutation.Region,
			permutation.Quota,
		)

		ch <- prometheus.MustNewConstMetric(
			regionQuotaLimits,
			prometheus.GaugeValue,
			count[QuotaMetricTypeLimit],
			permutation.Project,
			permutation.Region,
			permutation.Quota,
		)
	}
}

var newRegionQuotasCounter = func() regionQuotasCounterInterface {
	return &regionQuotasCounter{
		count: make(map[regionQuotasPermutation]map[quotaMetricType]float64),
	}
}

type RegionsCollector struct {
	*Common

	service services.ComputeServiceInterface

	regionQuotas regionQuotasCounterInterface

	initialized    bool
	initalizedLock sync.RWMutex
}

func (c *RegionsCollector) GetName() string {
	return RegionsCollectorName
}

func (c *RegionsCollector) GetData() error {
	if !c.isInitialized() {
		return fmt.Errorf("instances collector not initialized")
	}

	if c.service == nil {
		return fmt.Errorf("instances collector compute.Service is not initialized")
	}

	regionQuotas := newRegionQuotasCounter()
	for _, project := range c.GetProjects() {
		for _, region := range c.GetRegions() {
			logrus.WithFields(logrus.Fields{
				"project": project,
				"region":  region,
			}).Debugf("Requesting region")

			reg, err := c.service.GetRegion(project, region)
			if err != nil {
				return fmt.Errorf("error while requesting region data: %v", err)
			}

			regionQuotas.Add(project, region, reg.Quotas)
		}
	}

	c.regionQuotas = regionQuotas

	return nil
}

func (c *RegionsCollector) isInitialized() bool {
	c.initalizedLock.RLock()
	defer c.initalizedLock.RUnlock()

	return c.initialized
}

func (c *RegionsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- regionQuotaUsages
	ch <- regionQuotaLimits
}

func (c *RegionsCollector) Collect(ch chan<- prometheus.Metric) {
	c.regionQuotas.Collect(ch)
}

func (c *RegionsCollector) Init(client *http.Client) error {
	var err error

	c.service, err = services.NewComputeService(client)
	if err != nil {
		return fmt.Errorf("error while initializing computeService: %v", err)
	}

	logrus.WithFields(logrus.Fields{
		"projects": strings.Join(c.GetProjects(), ","),
		"regions":  strings.Join(c.GetRegions(), ","),
	}).Info("Registered collector")

	c.initalizedLock.Lock()
	defer c.initalizedLock.Unlock()

	c.initialized = true

	return nil
}

func NewRegionsCollector(c *Common) *RegionsCollector {
	return &RegionsCollector{
		Common:       c,
		regionQuotas: newRegionQuotasCounter(),
		initialized:  false,
	}
}
