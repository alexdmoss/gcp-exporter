package collector

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
)

type Interface interface {
	prometheus.Collector

	Init(*http.Client) error
	GetName() string
	GetData() error
}
