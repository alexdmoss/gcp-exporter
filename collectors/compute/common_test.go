package compute

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommon_GetProjects(t *testing.T) {
	p1 := "fake-project-1"
	p2 := "fake-project-2"
	c := &Common{Projects: []string{p1, p2}}

	projects := c.GetProjects()
	assert.Contains(t, projects, p1)
	assert.Contains(t, projects, p2)
}

func TestCommon_GetZones(t *testing.T) {
	z1 := "fake-zone-1"
	z2 := "fake-zone-2"
	c := &Common{Zones: []string{z1, z2}}

	zones := c.GetZones()
	assert.Contains(t, zones, z1)
	assert.Contains(t, zones, z2)
}

func TestCommon_GetRegions(t *testing.T) {
	z1 := "fake-zone-1"
	z2 := "fake-zone-2"
	c := &Common{Zones: []string{z1, z2}}

	regions := c.GetRegions()
	assert.Contains(t, regions, "fake-zone")
	assert.Len(t, regions, 1)
	assert.NotContains(t, regions, z1)
	assert.NotContains(t, regions, z2)
}
