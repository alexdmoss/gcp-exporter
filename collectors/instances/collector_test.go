package instances

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

func TestCount_Add(t *testing.T) {
	instance1 := &compute.Instance{}
	instance2 := &compute.Instance{}

	c := newCounter().(*counter)
	c.Add("project", "zone", []*compute.Instance{instance1})
	c.Add("project", "zone", []*compute.Instance{instance2})

	assert.Len(t, c.count, 1)

	p := permutation{
		Project: "project",
		Zone:    "zone",
		Tags:    "",
	}
	assert.Equal(t, 2, c.count[p])
}

func TestCount_Collect(t *testing.T) {
	ch := make(chan prometheus.Metric, 50)
	defer close(ch)

	c := newCounter().(*counter)
	p := permutation{
		Project: "project",
		Zone:    "zone",
		Tags:    "",
	}
	c.count[p] = 1

	c.Collect(ch)

	assert.Len(t, ch, 1)
}

func TestCollector_Init(t *testing.T) {
	collector := NewCollector()
	err := collector.Init(http.DefaultClient)

	assert.NoError(t, err)
}

func TestCollector_Init_noClient(t *testing.T) {
	collector := NewCollector()
	err := collector.Init(nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "error while initializing computeService:")
}

func TestCollector_GetData_withoutInitialize(t *testing.T) {
	collector := NewCollector()
	collector.Projects = append(collector.Projects, "fake-project")
	collector.Zones = append(collector.Zones, "fake-project")

	err := collector.GetData()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "collector not initialized")
}

func TestCollector_GetData(t *testing.T) {
	p1 := "fake-project-1"
	p2 := "fake-project-2"
	z1 := "fake-zone-1"
	z2 := "fake-zone-2"

	collector := NewCollector()
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

	ct := &mockCounterInterface{}
	ct.On("Add", p1, z1, list1.Items).Once()
	ct.On("Add", p1, z2, list2.Items).Once()
	ct.On("Add", p2, z1, list3.Items).Once()
	ct.On("Add", p2, z2, list4.Items).Once()

	newCounter = func() counterInterface {
		return ct
	}

	collector.initialized = true

	err := collector.GetData()

	require.NoError(t, err)
	service.AssertExpectations(t)
	ct.AssertExpectations(t)
}

func TestCollector_GetData_TagFiltered(t *testing.T) {
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

			collector := NewCollector()
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

			ct := &mockCounterInterface{}
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

			newCounter = func() counterInterface {
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

func TestCollector_GetData_ListInstancesError(t *testing.T) {
	collector := NewCollector()
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
