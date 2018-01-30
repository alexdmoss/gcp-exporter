package services

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/prometheus/client_golang/prometheus"
	"gitlab.com/gitlab-org/ci-cd/gcp-exporter/tests"
)

func TestMetricsService_StartServer_Disabled(t *testing.T) {
	tests.RunOnHijackedLogrusOutput(t, func(t *testing.T, output *bytes.Buffer) {
		ms := &MetricsService{}
		err := ms.StartServer()

		assert.NoError(t, err)
		assert.Contains(t, output.String(), "Server disabled")
	})
}

func TestMetricsService_StartServer_InvalidPort(t *testing.T) {
	examples := []struct {
		listenAddr    string
		expectedError string
	}{
		{listenAddr: "1.2.3.4", expectedError: "missing port in address"},
		{listenAddr: "1.2.3.4::", expectedError: "too many colons in address"},
		{listenAddr: "::1234", expectedError: "too many colons in address"},
	}

	for id, example := range examples {
		t.Run(strconv.Itoa(id), func(t *testing.T) {
			s := &MockHTTPServerInterface{}

			ms := &MetricsService{
				listenAddr: example.listenAddr,
				server:     s,
			}

			err := ms.StartServer()
			require.Error(t, err)
			assert.Contains(t, err.Error(), "invalid metrics server address")
			assert.Contains(t, err.Error(), example.expectedError)

			s.AssertNotCalled(t, "SetContext")
			s.AssertNotCalled(t, "SetAddr")
			s.AssertNotCalled(t, "SetHandler")
			s.AssertNotCalled(t, "ListenAndServe")
		})
	}
}

func TestMetricsService_StartServer_ValidPort(t *testing.T) {
	examples := []string{
		"127.0.0.1:1234",
		":1235",
	}

	for id, example := range examples {
		t.Run(strconv.Itoa(id), func(t *testing.T) {
			tests.RunOnHijackedLogrusOutput(t, func(t *testing.T, output *bytes.Buffer) {
				ctx, cancelFn := context.WithCancel(context.Background())

				s := &MockHTTPServerInterface{}
				s.On("SetContext", ctx).Once()
				s.On("SetAddr", example).Once()
				s.On("SetHandler", mock.Anything).Once()
				s.On("ListenAndServe", mock.Anything).Return(nil).Once()
				defer s.AssertExpectations(t)

				wg := &sync.WaitGroup{}
				wg.Add(1)

				ms := NewMetricsService(ctx, example, wg)
				ms.server = s

				err := ms.StartServer()
				assert.NoError(t, err)

				cancelFn()

				wg.Wait()

				assert.Contains(t, output.String(), "Metrics HTTP server listening at")
				assert.Contains(t, output.String(), example)
				assert.Contains(t, output.String(), "Metrics HTTP server closed")
			})
		})
	}
}

func tesetListenAndServeFailure(t *testing.T) {
	s := &MockHTTPServerInterface{}
	s.On("SetContext", mock.Anything).Once()
	s.On("SetAddr", mock.Anything).Once()
	s.On("SetHandler", mock.Anything).Once()
	s.On("ListenAndServe", mock.Anything).Return(fmt.Errorf("HTTPServer error")).Once()
	defer s.AssertExpectations(t)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	ms := &MetricsService{
		listenAddr: ":1234",
		server:     s,
		wg:         wg,
	}

	ms.StartServer()

	wg.Wait()
}

func TestMetricsService_StartServer_ListenAndServeFailure(t *testing.T) {
	if os.Getenv("DO_TEST_FAIL_ON_LISTEN_AND_SERVE") == "1" {
		tesetListenAndServeFailure(t)

		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestMetricsService_StartServer_ListenAndServeFailure")
	cmd.Env = append(os.Environ(), "DO_TEST_FAIL_ON_LISTEN_AND_SERVE=1")

	combinedOut, err := cmd.CombinedOutput()

	output := string(combinedOut)
	assert.Contains(t, output, "Metrics HTTP server listening at: :1234")
	assert.NotContains(t, output, "Metrics HTTP server closed")
	assert.Contains(t, output, "Metrics HTTP server failure")
	assert.Contains(t, output, "HTTPServer error")

	e, ok := err.(*exec.ExitError)
	require.True(t, ok, "Process should be finished with non-zero exit code")
	assert.False(t, e.Success(), "Process should be finished with non-zero exit code")
}

type fakeCollector struct{}

func (fc *fakeCollector) Describe(chan<- *prometheus.Desc) {}
func (fc *fakeCollector) Collect(chan<- prometheus.Metric) {}

func TestMetricsService_MustRegisterPrometheusCollector(t *testing.T) {
	collector1 := &fakeCollector{}
	collector2 := &fakeCollector{}
	collector3 := &fakeCollector{}

	r := &MockPrometheusRegistryInterface{}
	r.On("MustRegister", collector1, collector2).Once()
	r.On("MustRegister", collector3).Once()

	ms := &MetricsService{
		registry: r,
	}
	ms.MustRegisterPrometheusCollector(collector1, collector2)
	ms.MustRegisterPrometheusCollector(collector3)

	r.AssertExpectations(t)
}

func TestMetricsService_MustRegisterPrometheusCollector_withUninitializedRegistry(t *testing.T) {
	collector1 := &fakeCollector{}

	r := &MockPrometheusRegistryInterface{}
	r.On("MustRegister", collector1).Once()

	newPrometheusRegistry = func() PrometheusRegistryInterface {
		return r
	}

	ms := &MetricsService{}

	ms.MustRegisterPrometheusCollector(collector1)
	r.AssertExpectations(t)
}

func TestMetricsService_RegisterDefaultCollectors(t *testing.T) {
	r := &MockPrometheusRegistryInterface{}
	r.On("MustRegister", mock.Anything).Twice()

	ms := &MetricsService{
		registry: r,
	}

	ms.RegisterDefaultCollectors()

	r.AssertExpectations(t)
}
