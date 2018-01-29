package instances

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
		[]string{"project", "zone", "tags"},
		nil,
	)
)

const (
	CollectorName = "instances-collector"
)

type permutation struct {
	Project string
	Zone    string
	Tags    string
}

type counterInterface interface {
	Add(project string, zone string, instances []*compute.Instance)
	Collect(ch chan<- prometheus.Metric)
}

type counter struct {
	count map[permutation]int
	lock  sync.RWMutex
}

func (ic *counter) Add(project string, zone string, instances []*compute.Instance) {
	ic.lock.Lock()
	defer ic.lock.Unlock()

	for _, instance := range instances {
		permutation := permutation{
			Project: project,
			Zone:    zone,
		}

		if instance.Tags != nil {
			permutation.Tags = strings.Join(instance.Tags.Items, ",")
		}

		if _, ok := ic.count[permutation]; ok {
			ic.count[permutation]++
		} else {
			ic.count[permutation] = 1
		}
	}
}

func (ic *counter) Collect(ch chan<- prometheus.Metric) {
	for permutation, count := range ic.count {
		ch <- prometheus.MustNewConstMetric(
			numberOfInstances,
			prometheus.GaugeValue,
			float64(count),
			permutation.Project,
			permutation.Zone,
			permutation.Tags,
		)
	}
}

var newCounter = func() counterInterface {
	return &counter{
		count: make(map[permutation]int),
	}
}

type Collector struct {
	Projects  []string `long:"project" description:"Count instances that belong to selected project"`
	Zones     []string `long:"zone" description:"Count instances that belong to selected zone"`
	MatchTags []string `long:"match-tag" description:"Count instances that are matching selected tag"`

	service   services.ComputeServiceInterface
	instances counterInterface

	initialized    bool
	initalizedLock sync.RWMutex
}

func (ic *Collector) GetName() string {
	return CollectorName
}

func (ic *Collector) GetData() error {
	if !ic.isInitialized() {
		return fmt.Errorf("instances collector not initialized")
	}

	if ic.service == nil {
		return fmt.Errorf("instances collector compute.Service is not initialized")
	}

	count := newCounter()

	for _, project := range ic.Projects {
		for _, zone := range ic.Zones {
			logrus.WithFields(logrus.Fields{
				"project": project,
				"zone":    zone,
			}).Debugf("Requesting instances")

			instances, err := ic.service.ListInstances(project, zone)
			if err != nil {
				return fmt.Errorf("error while requesting instances data: %v", err)
			}

			logrus.WithField("count", len(instances.Items)).Debugln("Found instances")

			selectedInstances := ic.filterInstances(instances.Items)
			count.Add(project, zone, selectedInstances)
		}
	}

	ic.instances = count

	return nil
}

func (ic *Collector) isInitialized() bool {
	ic.initalizedLock.RLock()
	defer ic.initalizedLock.RUnlock()

	return ic.initialized
}

func (ic *Collector) filterInstances(instances []*compute.Instance) []*compute.Instance {
	if len(ic.MatchTags) < 1 {
		return instances
	}

	selectedInstancesMap := make(map[uint64]*compute.Instance, 0)
	for _, matchTag := range ic.MatchTags {
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

func (ic *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- numberOfInstances
}

func (ic *Collector) Collect(ch chan<- prometheus.Metric) {
	ic.instances.Collect(ch)
}

func (ic *Collector) Init(client *http.Client) error {
	var err error

	ic.service, err = services.NewComputeService(client)
	if err != nil {
		return fmt.Errorf("error while initializing computeService: %v", err)
	}

	logrus.WithFields(logrus.Fields{
		"projects":  strings.Join(ic.Projects, ","),
		"zones":     strings.Join(ic.Zones, ","),
		"matchTags": strings.Join(ic.MatchTags, ","),
	}).Info("Registered collector")

	ic.initalizedLock.Lock()
	defer ic.initalizedLock.Unlock()

	ic.initialized = true

	return nil
}

func NewCollector() *Collector {
	return &Collector{
		instances:   newCounter(),
		initialized: false,
	}
}
