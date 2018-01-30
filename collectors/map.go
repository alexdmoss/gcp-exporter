package collectors

import (
	"fmt"
	"strings"
	"sync"

	"github.com/urfave/cli"

	clihelpers "gitlab.com/ayufan/golang-cli-helpers"

	col "gitlab.com/gitlab-org/ci-cd/gcp-exporter/collectors/collector"
)

type MapInterface interface {
	Add(col.Interface)
	Get(string) col.Interface
	Flags() []cli.Flag
	EnableFlagNames() map[string]string
	AddFlagsFrom(interface{})
}

type Map struct {
	collectors      map[string]col.Interface
	additionalFlags []cli.Flag

	mutex sync.RWMutex
}

func (cm *Map) Add(collector col.Interface) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	if len(cm.collectors) < 1 {
		cm.collectors = make(map[string]col.Interface)
	}

	cm.collectors[collector.GetName()] = collector
}

func (cm *Map) Get(name string) col.Interface {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	return cm.collectors[name]
}

func (cm *Map) Flags() []cli.Flag {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	flags := make([]cli.Flag, 0)
	flags = append(flags, cm.additionalFlags...)

	for collectorName, collector := range cm.collectors {
		flags = append(flags, cm.enableCollectorFlag(collectorName))
		flags = append(flags, clihelpers.GetFlagsFromStruct(collector)...)
	}

	return flags
}

func (cm *Map) enableCollectorFlag(name string) cli.BoolFlag {
	flagName := cm.enableFlagName(name)
	return cli.BoolFlag{
		Name:   flagName,
		EnvVar: strings.Replace(strings.ToUpper(flagName), "-", "_", -1),
		Usage:  fmt.Sprintf("Enables %s collector", name),
	}
}

func (cm *Map) EnableFlagNames() map[string]string {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	flagNames := make(map[string]string)
	for name := range cm.collectors {
		flagNames[name] = cm.enableFlagName(name)
	}

	return flagNames
}

func (cm *Map) enableFlagName(name string) string {
	return fmt.Sprintf("%s-enable", name)
}

func (cm *Map) AddFlagsFrom(source interface{}) {
	if len(cm.additionalFlags) == 0 {
		cm.additionalFlags = make([]cli.Flag, 0)
	}

	flags := clihelpers.GetFlagsFromStruct(source)
	cm.additionalFlags = append(cm.additionalFlags, flags...)
}
