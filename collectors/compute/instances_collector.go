package compute

import (
	"fmt"
	"net/http"
	"strings"
	"sync"

	"google.golang.org/api/compute/v1"

	"github.com/Sirupsen/logrus"
	"github.com/prometheus/client_golang/prometheus"

	"gitlab.com/gitlab-org/ci-cd/gcp-exporter/client/services"
)

var (
	numberOfInstances = prometheus.NewDesc(
		"gcp_exporter_instances_count",
		"Current number of instances",
		[]string{"project", "zone", "tags", "machine_type"},
		nil,
	)
)

const (
	InstancesCollectorName = "instances-collector"
)

type instancesPermutation struct {
	Project     string
	Zone        string
	Tags        string
	MachineType string
}

type instancesCounterInterface interface {
	Add(string, string, []*compute.Instance)
	Collect(chan<- prometheus.Metric)
}

type instancesCounter struct {
	count map[instancesPermutation]int
	lock  sync.RWMutex
}

func (ic *instancesCounter) Add(project string, zone string, instances []*compute.Instance) {
	ic.lock.Lock()
	defer ic.lock.Unlock()

	for _, instance := range instances {
		permutation := instancesPermutation{
			Project:     project,
			Zone:        zone,
			MachineType: instance.MachineType,
		}

		if instance.Tags != nil {
			permutation.Tags = strings.Join(instance.Tags.Items, ",")
		}

		_, ok := ic.count[permutation]
		if ok {
			ic.count[permutation]++
		} else {
			ic.count[permutation] = 1
		}
	}
}

func (ic *instancesCounter) Collect(ch chan<- prometheus.Metric) {
	for permutation, count := range ic.count {
		ch <- prometheus.MustNewConstMetric(
			numberOfInstances,
			prometheus.GaugeValue,
			float64(count),
			permutation.Project,
			permutation.Zone,
			permutation.Tags,
			permutation.MachineType,
		)
	}
}

var newInstancesCounter = func() instancesCounterInterface {
	return &instancesCounter{
		count: make(map[instancesPermutation]int),
	}
}

type InstancesCollector struct {
	*Common

	MatchTags []string `long:"match-tag" description:"Count instances that are matching selected tag"`

	service   services.ComputeServiceInterface
	instances instancesCounterInterface

	initialized    bool
	initalizedLock sync.RWMutex
}

func (c *InstancesCollector) GetName() string {
	return InstancesCollectorName
}

func (c *InstancesCollector) GetData() error {
	if !c.isInitialized() {
		return fmt.Errorf("instances collector not initialized")
	}

	if c.service == nil {
		return fmt.Errorf("instances collector compute.Service is not initialized")
	}

	count := newInstancesCounter()
	for _, project := range c.GetProjects() {
		for _, zone := range c.GetZones() {
			logrus.WithFields(logrus.Fields{
				"project": project,
				"zone":    zone,
			}).Debugf("Requesting instances")

			instances, err := c.service.ListInstances(project, zone)
			if err != nil {
				return fmt.Errorf("error while requesting instances data: %v", err)
			}

			logrus.WithField("count", len(instances.Items)).Debugln("Found instances")

			selectedInstances := c.filterInstances(instances.Items)
			count.Add(project, zone, selectedInstances)
		}
	}

	c.instances = count

	return nil
}

func (c *InstancesCollector) isInitialized() bool {
	c.initalizedLock.RLock()
	defer c.initalizedLock.RUnlock()

	return c.initialized
}

func (c *InstancesCollector) filterInstances(instances []*compute.Instance) []*compute.Instance {
	if len(c.MatchTags) < 1 {
		return instances
	}

	selectedInstancesMap := make(map[uint64]*compute.Instance, 0)
	for _, matchTag := range c.MatchTags {
		logrus.WithField("matchTag", matchTag).Debugln("Filtering by tag")
		for _, instance := range instances {
			if instance.Tags == nil || len(instance.Tags.Items) < 1 {
				continue
			}

			for _, tag := range instance.Tags.Items {
				if tag == matchTag {
					selectedInstancesMap[instance.Id] = instance
					continue
				}
			}
		}
	}

	selectedInstances := make([]*compute.Instance, 0)
	for _, instance := range selectedInstancesMap {
		selectedInstances = append(selectedInstances, instance)
	}

	return selectedInstances
}

func (c *InstancesCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- numberOfInstances
}

func (c *InstancesCollector) Collect(ch chan<- prometheus.Metric) {
	c.instances.Collect(ch)
}

func (c *InstancesCollector) Init(client *http.Client) error {
	var err error

	c.service, err = services.NewComputeService(client)
	if err != nil {
		return fmt.Errorf("error while initializing computeService: %v", err)
	}

	logrus.WithFields(logrus.Fields{
		"projects":  strings.Join(c.GetProjects(), ","),
		"zones":     strings.Join(c.GetZones(), ","),
		"matchTags": strings.Join(c.MatchTags, ","),
	}).Info("Registered collector")

	c.initalizedLock.Lock()
	defer c.initalizedLock.Unlock()

	c.initialized = true

	return nil
}

func NewInstancesCollector(c *Common) *InstancesCollector {
	return &InstancesCollector{
		Common:      c,
		instances:   newInstancesCounter(),
		initialized: false,
	}
}
