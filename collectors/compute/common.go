package compute

import (
	"strings"
)

type Common struct {
	Projects []string `long:"project" description:"Count instances that belong to selected project"`
	Zones    []string `long:"zone" description:"Count instances that belong to selected zone"`
}

func (c *Common) GetProjects() []string {
	return c.Projects
}

func (c *Common) GetZones() []string {
	return c.Zones
}

func (c *Common) GetRegions() []string {
	regionsMap := make(map[string]bool, 0)
	for _, zone := range c.Zones {
		zoneParts := strings.Split(zone, "-")
		region := strings.Join(zoneParts[0:len(zoneParts)-1], "-")

		regionsMap[region] = true
	}

	regions := make([]string, 0)
	for region := range regionsMap {
		regions = append(regions, region)
	}

	return regions
}
