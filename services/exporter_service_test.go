package services

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"gitlab.com/gitlab-org/ci-cd/gcp-exporter/collectors"
)

func TestExporterService_Run(t *testing.T) {
	ctx, cancelFn := context.WithCancel(context.Background())

	p := &collectors.MockProviderInterface{}
	p.On("GetData", ctx).Twice()
	p.On("GetData", ctx).Run(func(args mock.Arguments) {
		cancelFn()
	}).Once()
	defer p.AssertExpectations(t)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	startetAt := time.Now()
	es := NewExporterService(ctx, 1, p, wg)
	err := es.Run()
	finishedAt := time.Now()

	wg.Wait()
	assert.NoError(t, err)
	assert.True(t, finishedAt.Sub(startetAt).Seconds() > 2, "Run operation with two GetData() calls and interval set to 1 should take at least 2 seconds")
}
