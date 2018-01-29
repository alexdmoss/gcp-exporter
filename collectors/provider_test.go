package collectors

import (
	"bytes"
	"flag"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"

	col "gitlab.com/gitlab-org/ci-cd/gcp-exporter/collectors/collector"
	"gitlab.com/gitlab-org/ci-cd/gcp-exporter/tests"
)

func TestProvider_Init(t *testing.T) {
	tests.RunOnHijackedLogrusOutput(t, func(t *testing.T, output *bytes.Buffer) {
		c1 := &col.MockInterface{}
		c1.On("Init", http.DefaultClient).Return(nil).Once()
		defer c1.AssertExpectations(t)
		c2 := &col.MockInterface{}
		c2.On("Init", http.DefaultClient).Return(nil).Once()
		defer c2.AssertExpectations(t)

		coll := &MockMapInterface{}
		coll.On("EnableFlagNames").Return(map[string]string{
			"first-fake-collector":  "first-fake-collector-enable",
			"second-fake-collector": "second-fake-collector-enable",
		}).Once()
		coll.On("Get", "first-fake-collector").Return(c1).Once()
		coll.On("Get", "second-fake-collector").Return(c2).Once()
		defer coll.AssertExpectations(t)
		Collectors = coll

		set := flag.NewFlagSet("app", flag.ContinueOnError)
		f1 := &cli.BoolFlag{Name: "first-fake-collector-enable"}
		f1.Apply(set)
		f2 := &cli.BoolFlag{Name: "second-fake-collector-enable"}
		f2.Apply(set)
		set.Parse([]string{"--first-fake-collector-enable", "--second-fake-collector-enable"})
		cliCtx := cli.NewContext(cli.NewApp(), set, nil)

		p := NewProvider(http.DefaultClient)
		p.Init(cliCtx)

		assert.Contains(t, output.String(), "Enabling first-fake-collector")
		assert.Contains(t, output.String(), "Enabling second-fake-collector")
	})
}

func TestProvider_InitCollectorFailure(t *testing.T) {
	tests.RunOnHijackedLogrusOutput(t, func(t *testing.T, output *bytes.Buffer) {
		c1 := &col.MockInterface{}
		c1.On("Init", http.DefaultClient).Return(fmt.Errorf("fake-error")).Once()
		defer c1.AssertExpectations(t)

		coll := &MockMapInterface{}
		coll.On("EnableFlagNames").Return(map[string]string{
			"first-fake-collector": "first-fake-collector-enable",
		}).Once()
		coll.On("Get", "first-fake-collector").Return(c1).Once()
		defer coll.AssertExpectations(t)
		Collectors = coll

		set := flag.NewFlagSet("app", flag.ContinueOnError)
		f1 := &cli.BoolFlag{Name: "first-fake-collector-enable"}
		f1.Apply(set)
		set.Parse([]string{"--first-fake-collector-enable"})
		cliCtx := cli.NewContext(cli.NewApp(), set, nil)

		p := NewProvider(http.DefaultClient)
		err := p.Init(cliCtx)

		assert.Error(t, err, "error while initializing collector first-fake-collector: fake-error")
		assert.Contains(t, output.String(), "Enabling first-fake-collector")
	})
}

func TestProvider_GetData(t *testing.T) {
	tests.RunOnHijackedLogrusOutput(t, func(t *testing.T, output *bytes.Buffer) {
		c1 := &col.MockInterface{}
		c1.On("Init", http.DefaultClient).Return(nil).Once()
		c1.On("GetData").Return(nil).Once()
		defer c1.AssertExpectations(t)

		p := NewProvider(http.DefaultClient)
		assert.Equal(t, int64(0), p.lastGetDataTimestamp.Unix())
		assert.Equal(t, uint64(0), p.getDataErrors)

		p.registerCollector("first-fake-collector", c1)

		p.GetData()
		assert.Contains(t, output.String(), "Getting data from GCP")
		assert.True(t, p.lastGetDataTimestamp.Unix() > time.Now().Add(-10*time.Second).Unix())
		assert.Equal(t, uint64(0), p.getDataErrors)
	})
}

func TestProvider_GetDataFailure(t *testing.T) {
	tests.RunOnHijackedLogrusOutput(t, func(t *testing.T, output *bytes.Buffer) {
		c1 := &col.MockInterface{}
		c1.On("Init", http.DefaultClient).Return(nil).Once()
		c1.On("GetData").Return(fmt.Errorf("fake-error")).Once()
		defer c1.AssertExpectations(t)

		p := NewProvider(http.DefaultClient)
		assert.Equal(t, uint64(0), p.getDataErrors)

		p.registerCollector("first-fake-collector", c1)

		p.GetData()
		assert.Contains(t, output.String(), "Error while getting data from GCP")
		assert.Contains(t, output.String(), "fake-error")
		assert.Equal(t, uint64(1), p.getDataErrors)
	})
}

func TestProvider_Describe(t *testing.T) {
	tests.RunOnHijackedLogrusOutput(t, func(t *testing.T, output *bytes.Buffer) {
		ch := make(chan<- *prometheus.Desc, 10)

		c1 := &col.MockInterface{}
		c1.On("Init", http.DefaultClient).Return(nil).Once()
		c1.On("Describe", ch).Once()
		defer c1.AssertExpectations(t)

		p := NewProvider(http.DefaultClient)
		p.registerCollector("first-fake-collector", c1)

		p.Describe(ch)
		close(ch)
	})
}

func TestProvider_Collect(t *testing.T) {
	tests.RunOnHijackedLogrusOutput(t, func(t *testing.T, output *bytes.Buffer) {
		ch := make(chan<- prometheus.Metric, 10)

		c1 := &col.MockInterface{}
		c1.On("Init", http.DefaultClient).Return(nil).Once()
		c1.On("Collect", ch).Once()
		defer c1.AssertExpectations(t)

		p := NewProvider(http.DefaultClient)
		p.registerCollector("first-fake-collector", c1)

		p.Collect(ch)
		close(ch)
	})
}
