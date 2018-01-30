package version

import (
	"fmt"
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var AppVersion AppVersionInfo

var (
	NAME     = "gcp-exporter"
	VERSION  = "dev"
	REVISION = "HEAD"
	BRANCH   = "HEAD"
	BUILT    = "now"
)

type AppVersionInfo struct {
	Name         string
	Version      string
	Revision     string
	Branch       string
	GOVersion    string
	BuiltAt      time.Time
	OS           string
	Architecture string
}

func (v *AppVersionInfo) UserAgent() string {
	return fmt.Sprintf("%s %s (%s; %s; %s/%s)", v.Name, v.Version, v.Branch, v.GOVersion, v.OS, v.Architecture)
}

func (v *AppVersionInfo) Line() string {
	return fmt.Sprintf("%s %s (%s)", v.Name, v.Version, v.Revision)
}

func (v *AppVersionInfo) ShortLine() string {
	return fmt.Sprintf("%s (%s)", v.Version, v.Revision)
}

func (v *AppVersionInfo) Extended() string {
	version := fmt.Sprintf("Version:      %s\n", v.Version)
	version += fmt.Sprintf("Git revision: %s\n", v.Revision)
	version += fmt.Sprintf("Git branch:   %s\n", v.Branch)
	version += fmt.Sprintf("GO version:   %s\n", v.GOVersion)
	version += fmt.Sprintf("Built:        %s\n", v.BuiltAt.Format(time.RFC3339))
	version += fmt.Sprintf("OS/Arch:      %s/%s\n", v.OS, v.Architecture)

	return version
}

func (v *AppVersionInfo) VersionCollector() *prometheus.GaugeVec {
	labels := map[string]string{
		"name":         v.Name,
		"version":      v.Version,
		"revision":     v.Revision,
		"branch":       v.Branch,
		"go_version":   v.GOVersion,
		"built_at":     v.BuiltAt.String(),
		"os":           v.OS,
		"architecture": v.Architecture,
	}

	labelNames := make([]string, 0, len(labels))
	for n := range labels {
		labelNames = append(labelNames, n)
	}

	buildInfo := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "hanging_droplets_cleaner_version_info",
			Help: "A metric with a constant '1' value labeled by different build stats fields.",
		},
		labelNames,
	)
	buildInfo.With(labels).Set(1)

	return buildInfo
}

func NewAppVersionInfo(name, version, revision, branch, builtAt string) AppVersionInfo {
	built := time.Now()
	if builtAt != "now" {
		built, _ = time.Parse(time.RFC3339, builtAt)
	}

	return AppVersionInfo{
		Name:         name,
		Version:      version,
		Revision:     revision,
		Branch:       branch,
		GOVersion:    runtime.Version(),
		BuiltAt:      built,
		OS:           runtime.GOOS,
		Architecture: runtime.GOARCH,
	}
}

func init() {
	AppVersion = NewAppVersionInfo(NAME, VERSION, REVISION, BRANCH, BUILT)
}
