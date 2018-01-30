package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComputeService_ListInstances_notInitialized(t *testing.T) {
	c := &ComputeService{}
	instancesList, err := c.ListInstances("fake-project", "fake-zone")

	assert.Nil(t, instancesList)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "service not initialized")
}

func TestComputeService_GetRegion_notInitialized(t *testing.T) {
	c := &ComputeService{}
	reg, err := c.GetRegion("fake-project", "fake-region")

	assert.Nil(t, reg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "service not initialized")
}
