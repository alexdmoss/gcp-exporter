package compute

import (
	"fmt"
	"net/http"
	"testing"

	"google.golang.org/api/compute/v1"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"gitlab.com/gitlab-org/ci-cd/gcp-exporter/client/services"
)

func TestInstancesCounter_Add(t *testing.T) {
	instance1 := &compute.Instance{Tags: &compute.Tags{Items: []string{"fake-tag"}}, MachineType: "n1-standard-1"}
	instance2 := &compute.Instance{Tags: &compute.Tags{Items: []string{"fake-tag"}}, MachineType: "n1-standard-1"}

	c := newInstancesCounter().(*instancesCounter)
	c.Add("project", "zone", []*compute.Instance{instance1})
	c.Add("project", "zone", []*compute.Instance{instance2})

	assert.Len(t, c.count, 1)

	p := instancesPermutation{
		Project:     "project",
		Zone:        "zone",
		Tags:        "fake-tag",
		MachineType: "n1-standard-1",
	}
	assert.Equal(t, 2, c.count[p])
}

func TestInstancesCounter_Collect(t *testing.T) {
	ch := make(chan prometheus.Metric, 50)
	defer close(ch)

	c := newInstancesCounter().(*instancesCounter)
	p := instancesPermutation{
		Project: "project",
		Zone:    "zone",
		Tags:    "",
	}
	c.count[p] = 1

	c.Collect(ch)

	assert.Len(t, ch, 1)
}

func TestInstancesCollector_GetName(t *testing.T) {
	collector := NewInstancesCollector(&Common{})
	assert.Equal(t, "instances-collector", collector.GetName())
}

func TestInstancesCollector_Init(t *testing.T) {
	collector := NewInstancesCollector(&Common{})
	err := collector.Init(http.DefaultClient)

	assert.NoError(t, err)
}

func TestInstancesCollector_Init_noClient(t *testing.T) {
	collector := NewInstancesCollector(&Common{})
	err := collector.Init(nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "error while initializing computeService:")
}

func TestInstancesCollector_GetData_withoutInitialize(t *testing.T) {
	collector := NewInstancesCollector(&Common{})
	collector.Projects = append(collector.Projects, "fake-project")
	collector.Zones = append(collector.Zones, "fake-zone")

	err := collector.GetData()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "collector not initialized")
}

func TestInstancesCollector_GetData_withoutComputeService(t *testing.T) {
	collector := NewInstancesCollector(&Common{})
	collector.Projects = append(collector.Projects, "fake-project")
	collector.Zones = append(collector.Zones, "fake-zone")
	collector.initialized = true

	err := collector.GetData()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "instances collector compute.Service is not initialized")
}

func TestInstancesCollector_GetData(t *testing.T) {
	p1 := "fake-project-1"
	p2 := "fake-project-2"
	z1 := "fake-zone-1"
	z2 := "fake-zone-2"

	collector := NewInstancesCollector(&Common{})
	collector.Projects = append(collector.Projects, []string{p1, p2}...)
	collector.Zones = append(collector.Zones, []string{z1, z2}...)

	instance1 := &compute.Instance{}
	instance2 := &compute.Instance{}
	instance3 := &compute.Instance{}

	list1 := &compute.InstanceList{Items: []*compute.Instance{instance1}}
	list2 := &compute.InstanceList{Items: []*compute.Instance{}}
	list3 := &compute.InstanceList{Items: []*compute.Instance{instance2, instance3}}
	list4 := &compute.InstanceList{Items: []*compute.Instance{}}

	service := &services.MockComputeServiceInterface{}
	service.On("ListInstances", p1, z1).Return(list1, nil).Once()
	service.On("ListInstances", p1, z2).Return(list2, nil).Once()
	service.On("ListInstances", p2, z1).Return(list3, nil).Once()
	service.On("ListInstances", p2, z2).Return(list4, nil).Once()
	collector.service = service

	ct := &mockInstancesCounterInterface{}
	ct.On("Add", p1, z1, list1.Items).Once()
	ct.On("Add", p1, z2, list2.Items).Once()
	ct.On("Add", p2, z1, list3.Items).Once()
	ct.On("Add", p2, z2, list4.Items).Once()

	newInstancesCounter = func() instancesCounterInterface {
		return ct
	}

	collector.initialized = true

	err := collector.GetData()

	require.NoError(t, err)
	service.AssertExpectations(t)
	ct.AssertExpectations(t)
}

func TestInstancesCollector_GetData_TagFiltered(t *testing.T) {
	examples := map[string]bool{
		"use-tags":      true,
		"dont-use-tags": false,
	}

	for name, useTags := range examples {
		t.Run(name, func(t *testing.T) {
			p1 := "fake-project-1"
			z1 := "fake-zone-1"

			tag1 := "tag-1"
			tag2 := "tag-2"

			collector := NewInstancesCollector(&Common{})
			collector.Projects = append(collector.Projects, p1)
			collector.Zones = append(collector.Zones, z1)

			if useTags {
				collector.MatchTags = append(collector.MatchTags, []string{tag1, tag2}...)
			}

			instance1 := &compute.Instance{Id: uint64(1), Tags: &compute.Tags{Items: []string{tag1}}}
			instance2 := &compute.Instance{Id: uint64(2), Tags: &compute.Tags{Items: []string{tag2}}}
			instance3 := &compute.Instance{Id: uint64(3), Tags: &compute.Tags{Items: []string{tag1, tag2}}}
			instance4 := &compute.Instance{Id: uint64(4), Tags: &compute.Tags{Items: []string{"tag-3"}}}
			instance5 := &compute.Instance{Id: uint64(5)}

			list1 := &compute.InstanceList{Items: []*compute.Instance{instance1, instance2, instance3, instance4, instance5}}

			service := &services.MockComputeServiceInterface{}
			service.On("ListInstances", p1, z1).Return(list1, nil).Once()
			collector.service = service

			ct := &mockInstancesCounterInterface{}
			ct.On("Add", p1, z1, mock.Anything).Run(func(args mock.Arguments) {
				list := args[2].([]*compute.Instance)
				existingInstances := make(map[uint64]*compute.Instance, 0)

				for _, instance := range list {
					existingInstances[instance.Id] = instance
				}

				assert.NotNil(t, existingInstances[instance1.Id])
				assert.NotNil(t, existingInstances[instance2.Id])
				assert.NotNil(t, existingInstances[instance3.Id])
				if useTags {
					assert.Nil(t, existingInstances[instance4.Id])
					assert.Nil(t, existingInstances[instance5.Id])
				} else {
					assert.NotNil(t, existingInstances[instance4.Id])
					assert.NotNil(t, existingInstances[instance5.Id])
				}
			}).Once()

			newInstancesCounter = func() instancesCounterInterface {
				return ct
			}

			collector.initialized = true

			err := collector.GetData()

			require.NoError(t, err)
			service.AssertExpectations(t)
			ct.AssertExpectations(t)
		})
	}
}

func TestInstancesCollector_GetData_ListInstancesError(t *testing.T) {
	collector := NewInstancesCollector(&Common{})
	collector.Projects = append(collector.Projects, "fake-project-1")
	collector.Zones = append(collector.Zones, "fake-zone-1")

	service := &services.MockComputeServiceInterface{}
	service.On("ListInstances", "fake-project-1", "fake-zone-1").Return(nil, fmt.Errorf("fake-list-instances-error")).Once()
	collector.service = service

	collector.initialized = true

	err := collector.GetData()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "error while requesting instances data: fake-list-instances-error")
	service.AssertExpectations(t)
}

func TestInstancesCollector_Describe(t *testing.T) {
	ch := make(chan<- *prometheus.Desc, 50)
	defer close(ch)

	collector := NewInstancesCollector(&Common{})
	collector.Describe(ch)

	assert.Len(t, ch, 1)
}

func TestInstancesCollector_Collect(t *testing.T) {
	ch := make(chan<- prometheus.Metric, 50)
	defer close(ch)

	ct := &mockInstancesCounterInterface{}
	ct.On("Collect", ch).Once()

	newInstancesCounter = func() instancesCounterInterface {
		return ct
	}

	collector := NewInstancesCollector(&Common{})
	collector.Collect(ch)

	ct.AssertExpectations(t)
}
