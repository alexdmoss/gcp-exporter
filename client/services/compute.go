package services

import (
	"context"
	"fmt"
	"net/http"

	"google.golang.org/api/compute/v1"
)

type ComputeServiceInterface interface {
	ListInstances(project string, zone string, perPage int64) ([]*compute.Instance, error)
	GetRegion(project string, region string) (*compute.Region, error)
}

type ComputeService struct {
	service *compute.Service
}

func (cs *ComputeService) ListInstances(project string, zone string, perPage int64) ([]*compute.Instance, error) {
	err := cs.failIfInitialized()
	if err != nil {
		return nil, err
	}

	instances := make([]*compute.Instance, 0)

	ilc := cs.service.Instances.List(project, zone)
	ilc.MaxResults(perPage)
	err = ilc.Pages(context.TODO(), func(page *compute.InstanceList) error {
		instances = append(instances, page.Items...)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return instances, nil
}

func (cs *ComputeService) GetRegion(project string, region string) (*compute.Region, error) {
	err := cs.failIfInitialized()
	if err != nil {
		return nil, err
	}

	rgc := cs.service.Regions.Get(project, region)

	return rgc.Do()
}

func (cs *ComputeService) failIfInitialized() error {
	if cs.service != nil {
		return nil
	}

	return fmt.Errorf("service not initialized")
}

func NewComputeService(client *http.Client) (*ComputeService, error) {
	service, err := compute.New(client)
	if err != nil {
		return nil, err
	}

	cs := &ComputeService{
		service: service,
	}

	return cs, nil
}
