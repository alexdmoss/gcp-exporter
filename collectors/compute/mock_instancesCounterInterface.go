// Code generated by mockery v1.0.0

// This comment works around https://github.com/vektra/mockery/issues/155

package compute

import (
	prometheus "github.com/prometheus/client_golang/prometheus"
	mock "github.com/stretchr/testify/mock"
	compute "google.golang.org/api/compute/v1"
)

// mockInstancesCounterInterface is an autogenerated mock type for the instancesCounterInterface type
type mockInstancesCounterInterface struct {
	mock.Mock
}

// Add provides a mock function with given fields: _a0, _a1, _a2
func (_m *mockInstancesCounterInterface) Add(_a0 string, _a1 string, _a2 []*compute.Instance) {
	_m.Called(_a0, _a1, _a2)
}

// Collect provides a mock function with given fields: _a0
func (_m *mockInstancesCounterInterface) Collect(_a0 chan<- prometheus.Metric) {
	_m.Called(_a0)
}