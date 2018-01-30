package compute

import (
	"fmt"
	"net/http"
	"testing"

	"google.golang.org/api/compute/v1"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/gitlab-org/ci-cd/gcp-exporter/client/services"
)

func TestRegionQuotasCounter_Add(t *testing.T) {
	quota1 := &compute.Quota{Usage: 1, Limit: 10, Metric: "quota1"}
	quota2 := &compute.Quota{Usage: 2, Limit: 20, Metric: "quota2"}

	c := newRegionQuotasCounter().(*regionQuotasCounter)
	c.Add("project", "region", []*compute.Quota{quota1})
	c.Add("project", "region", []*compute.Quota{quota2})

	assert.Len(t, c.count, 2)

	p1 := regionQuotasPermutation{Project: "project", Region: "region", Quota: "quota1"}
	assert.Equal(t, float64(1), c.count[p1][QuotaMetricTypeUsage])
	assert.Equal(t, float64(10), c.count[p1][QuotaMetricTypeLimit])

	p2 := regionQuotasPermutation{Project: "project", Region: "region", Quota: "quota2"}
	assert.Equal(t, float64(2), c.count[p2][QuotaMetricTypeUsage])
	assert.Equal(t, float64(20), c.count[p2][QuotaMetricTypeLimit])
}

func TestRegionQuotasCounter_Collect(t *testing.T) {
	ch := make(chan prometheus.Metric, 50)
	defer close(ch)

	c := newRegionQuotasCounter().(*regionQuotasCounter)
	p := regionQuotasPermutation{Project: "project", Region: "region", Quota: "quota1"}
	c.count[p] = make(map[quotaMetricType]float64)
	c.count[p][QuotaMetricTypeUsage] = 1
	c.count[p][QuotaMetricTypeLimit] = 10

	c.Collect(ch)

	assert.Len(t, ch, 2)
}

func TestRegionsCollector_GetName(t *testing.T) {
	collector := NewRegionsCollector(&Common{})
	assert.Equal(t, "regions-collector", collector.GetName())
}

func TestRegionsCollector_Init(t *testing.T) {
	collector := NewRegionsCollector(&Common{})
	err := collector.Init(http.DefaultClient)

	assert.NoError(t, err)
}

func TestRegionsCollector_Init_noClient(t *testing.T) {
	collector := NewRegionsCollector(&Common{})
	err := collector.Init(nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "error while initializing computeService:")
}

func TestRegionsCollector_GetData_withoutInitialize(t *testing.T) {
	collector := NewRegionsCollector(&Common{})
	collector.Projects = append(collector.Projects, "fake-project")
	collector.Zones = append(collector.Zones, "fake-zone")

	err := collector.GetData()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "collector not initialized")
}

func TestRegionsCollector_GetData_withoutComputeService(t *testing.T) {
	collector := NewRegionsCollector(&Common{})
	collector.Projects = append(collector.Projects, "fake-project")
	collector.Zones = append(collector.Zones, "fake-zone")
	collector.initialized = true

	err := collector.GetData()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "instances collector compute.Service is not initialized")
}

func TestRegionsCollector_GetData(t *testing.T) {
	p1 := "fake-project-1"
	p2 := "fake-project-2"
	z1 := "fake-zone-1"
	z2 := "fake-zone-2"
	r1 := "fake-zone"

	collector := NewRegionsCollector(&Common{})
	collector.Projects = append(collector.Projects, []string{p1, p2}...)
	collector.Zones = append(collector.Zones, []string{z1, z2}...)

	quota1 := &compute.Quota{}
	quota2 := &compute.Quota{}
	quota3 := &compute.Quota{}

	region1 := &compute.Region{Quotas: []*compute.Quota{quota1, quota2}}
	region2 := &compute.Region{Quotas: []*compute.Quota{quota3}}

	service := &services.MockComputeServiceInterface{}
	service.On("GetRegion", p1, r1).Return(region1, nil).Once()
	service.On("GetRegion", p2, r1).Return(region2, nil).Once()
	collector.service = service

	ct := &mockRegionQuotasCounterInterface{}
	ct.On("Add", p1, r1, region1.Quotas).Once()
	ct.On("Add", p2, r1, region2.Quotas).Once()

	newRegionQuotasCounter = func() regionQuotasCounterInterface {
		return ct
	}

	collector.initialized = true

	err := collector.GetData()

	require.NoError(t, err)
	service.AssertExpectations(t)
	ct.AssertExpectations(t)
}

func TestRegionsCollector_GetData_GetRegionError(t *testing.T) {
	collector := NewRegionsCollector(&Common{})
	collector.Projects = append(collector.Projects, "fake-project-1")
	collector.Zones = append(collector.Zones, "fake-zone-1")

	service := &services.MockComputeServiceInterface{}
	service.On("GetRegion", "fake-project-1", "fake-zone").Return(nil, fmt.Errorf("fake-get-region-error")).Once()
	collector.service = service

	collector.initialized = true

	err := collector.GetData()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "error while requesting region data: fake-get-region-error")
	service.AssertExpectations(t)
}

func TestRegionsCollector_Describe(t *testing.T) {
	ch := make(chan<- *prometheus.Desc, 50)
	defer close(ch)

	collector := NewRegionsCollector(&Common{})
	collector.Describe(ch)

	assert.Len(t, ch, 2)
}

func TestRegionsCollector_Collect(t *testing.T) {
	ch := make(chan<- prometheus.Metric, 50)
	defer close(ch)

	ct := &mockRegionQuotasCounterInterface{}
	ct.On("Collect", ch).Once()

	newRegionQuotasCounter = func() regionQuotasCounterInterface {
		return ct
	}

	collector := NewRegionsCollector(&Common{})
	collector.Collect(ch)

	ct.AssertExpectations(t)
}
