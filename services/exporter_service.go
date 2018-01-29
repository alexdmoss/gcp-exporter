package services

import (
	"context"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"

	"gitlab.com/gitlab-org/ci-cd/gcp-exporter/collectors"
)

type ExporterService struct {
	ctx      context.Context
	interval time.Duration
	wg       *sync.WaitGroup

	collectorProvider collectors.ProviderInterface
}

func (es *ExporterService) Run() error {
	logrus.Infof("GCP data gathering interval: %s", es.interval)

	es.collectorProvider.GetData()
	for {
		select {
		case <-time.After(es.interval):
			es.collectorProvider.GetData()
		case <-es.ctx.Done():
			es.wg.Done()
			return nil
		}
	}
}

func NewExporterService(ctx context.Context, interval int, collectorProvider collectors.ProviderInterface, wg *sync.WaitGroup) *ExporterService {
	es := &ExporterService{
		ctx:               ctx,
		interval:          time.Duration(interval) * time.Second,
		wg:                wg,
		collectorProvider: collectorProvider,
	}

	return es
}
