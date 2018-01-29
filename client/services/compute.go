package services

import (
	"fmt"
	"net/http"

	"google.golang.org/api/compute/v1"
)

type ComputeServiceInterface interface {
	ListInstances(project string, zone string) (*compute.InstanceList, error)
}

type ComputeService struct {
	service *compute.Service
}

func (cs *ComputeService) ListInstances(project string, zone string) (*compute.InstanceList, error) {
	if cs.service != nil {
		return nil, fmt.Errorf("service not initialized")
	}

	ilc := cs.service.Instances.List(project, zone)
	return ilc.Do()
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
