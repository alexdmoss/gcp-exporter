package collectors

import (
	"context"
	"net/http"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/gitlab-org/ci-cd/gcp-exporter/collectors/collector"
)

func TestMap_Add(t *testing.T) {
	c := &collector.MockInterface{}
	c.On("GetName").Return("fake-collector").Once()
	defer c.AssertExpectations(t)

	m := &Map{}
	m.Add(c)

	assert.Len(t, m.collectors, 1)
	assert.Equal(t, c, m.collectors["fake-collector"])
}

func TestMap_AddTwice(t *testing.T) {
	c := &collector.MockInterface{}
	c.On("GetName").Return("fake-collector").Twice()
	defer c.AssertExpectations(t)

	m := &Map{}
	m.Add(c)
	m.Add(c)

	assert.Len(t, m.collectors, 1)
	assert.Equal(t, c, m.collectors["fake-collector"])
}

func TestMap_GetExisting(t *testing.T) {
	c := &collector.MockInterface{}

	m := &Map{
		collectors: map[string]collector.Interface{
			"fake-collector": c,
		},
	}

	receivedCollector := m.Get("fake-collector")

	assert.Equal(t, c, receivedCollector)
	c.AssertNotCalled(t, "GetName")
}

func TestMap_GetNonexisting(t *testing.T) {
	m := &Map{
		collectors: map[string]collector.Interface{},
	}

	receivedCollector := m.Get("fake-collector")
	assert.Nil(t, receivedCollector)
}

type fakeCollector struct {
	FakeSetting string `long:"fake-setting" description:"Fake setting"`
}

func (fc *fakeCollector) Init(*http.Client) error             { return nil }
func (fc *fakeCollector) GetName() string                     { return "fake-collector " }
func (fc *fakeCollector) GetData(ctx context.Context) error   { return nil }
func (fc *fakeCollector) Describe(ch chan<- *prometheus.Desc) {}
func (fc *fakeCollector) Collect(ch chan<- prometheus.Metric) {}

func getFakeCollector() *fakeCollector {
	return &fakeCollector{
		FakeSetting: "test",
	}
}

func TestMap_Flags(t *testing.T) {
	m := &Map{
		collectors: map[string]collector.Interface{
			"fake-collector": getFakeCollector(),
		},
	}

	flags := m.Flags()
	require.Len(t, flags, 2)

	assert.Equal(t, "fake-collector-enable", flags[0].GetName())
	assert.Contains(t, flags[0].String(), "Enables fake-collector collector")

	assert.Equal(t, "fake-setting", flags[1].GetName())
	assert.Contains(t, flags[1].String(), "Fake setting")
}

func TestMap_EnableFlagNames(t *testing.T) {
	m := &Map{
		collectors: map[string]collector.Interface{
			"fake-collector":        &collector.MockInterface{},
			"second-fake-collector": &collector.MockInterface{},
		},
	}

	flags := m.EnableFlagNames()
	require.Len(t, flags, 2)

	assert.Equal(t, "fake-collector-enable", flags["fake-collector"])
	assert.Equal(t, "second-fake-collector-enable", flags["second-fake-collector"])
}
