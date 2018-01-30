package services

import (
	"context"
	"fmt"
	"net"
	"net/http"
	_ "net/http/pprof" // PPROF is loading everything in its init function
	"os"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	HTTPServerShutdownTimeout = 10 * time.Second
)

type HTTPServerInterface interface {
	SetContext(context.Context)
	SetAddr(string)
	SetHandler(http.Handler)
	ListenAndServe() error
}

type HTTPServer struct {
	s *http.Server

	ctx context.Context
}

func (hs *HTTPServer) SetContext(ctx context.Context) {
	hs.ctx = ctx
}

func (hs *HTTPServer) SetAddr(addr string) {
	hs.s.Addr = addr
}

func (hs *HTTPServer) SetHandler(handler http.Handler) {
	hs.s.Handler = handler
}

func (hs *HTTPServer) ListenAndServe() error {
	finished := make(chan error, 1)

	go func() {
		finished <- hs.s.ListenAndServe()
	}()

	select {
	case <-hs.ctx.Done():
	case err := <-finished:
		logrus.WithError(err).Infoln("Listen and serve error")
		return err
	}

	logrus.Infoln("Shutting down Metrics HTTP server")
	timeoutCtx, cancelFn := context.WithTimeout(context.Background(), HTTPServerShutdownTimeout)
	defer cancelFn()

	return hs.s.Shutdown(timeoutCtx)
}

type PrometheusRegistryInterface interface {
	prometheus.Registerer
	prometheus.Gatherer
}

type MetricsService struct {
	ctx context.Context
	wg  *sync.WaitGroup

	listenAddr string
	registry   PrometheusRegistryInterface
	server     HTTPServerInterface
}

func (ms *MetricsService) StartServer() error {
	if ms.listenAddr == "" {
		logrus.Infoln("Server disabled")
		return nil
	}

	err := ms.prepareServer()
	if err != nil {
		return err
	}

	go func() {
		defer func() {
			ms.wg.Done()
		}()

		err := ms.server.ListenAndServe()
		if err != nil {
			logrus.WithError(err).Fatalln("Metrics HTTP server failure")
		}

		logrus.Infoln("Metrics HTTP server closed")
	}()

	logrus.Infof("Metrics HTTP server listening at: %s", ms.listenAddr)

	return nil
}

func (ms *MetricsService) prepareServer() error {
	_, _, err := net.SplitHostPort(ms.listenAddr)
	if err != nil {
		return fmt.Errorf("invalid metrics server address: %s", err.Error())
	}

	handler := http.NewServeMux()
	handler.Handle("/metrics", promhttp.HandlerFor(ms.registry, promhttp.HandlerOpts{}))

	if ms.server == nil {
		ms.server = &HTTPServer{
			s: &http.Server{},
		}
	}

	ms.server.SetContext(ms.ctx)
	ms.server.SetAddr(ms.listenAddr)
	ms.server.SetHandler(handler)

	return nil
}

func (ms *MetricsService) RegisterDefaultCollectors() {
	ms.initializeRegistry()

	ms.MustRegisterPrometheusCollector(prometheus.NewGoCollector())
	ms.MustRegisterPrometheusCollector(prometheus.NewProcessCollector(os.Getpid(), ""))
}

func (ms *MetricsService) MustRegisterPrometheusCollector(cs ...prometheus.Collector) {
	ms.initializeRegistry()

	ms.registry.MustRegister(cs...)
}

var newPrometheusRegistry = func() PrometheusRegistryInterface {
	return prometheus.NewRegistry()
}

func (ms *MetricsService) initializeRegistry() {
	if ms.registry == nil {
		ms.registry = newPrometheusRegistry()
	}
}

func NewMetricsService(ctx context.Context, listenAddr string, wg *sync.WaitGroup) *MetricsService {
	return &MetricsService{
		ctx:        ctx,
		listenAddr: listenAddr,
		wg:         wg,
	}
}
