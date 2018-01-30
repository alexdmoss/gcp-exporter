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
