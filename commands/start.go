package commands

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/Sirupsen/logrus"
	"github.com/urfave/cli"

	google_client "gitlab.com/gitlab-org/ci-cd/gcp-exporter/client"
	"gitlab.com/gitlab-org/ci-cd/gcp-exporter/collectors"
	"gitlab.com/gitlab-org/ci-cd/gcp-exporter/services"

	"gitlab.com/gitlab-org/ci-cd/gcp-exporter/version"
)

const (
	DefaultInterval = 60
)

type StartExporterServiceCommand struct {
	ListenAddr         string `long:"listen" env:"GCP_EXPORTER_LISTEN" description:"Metrics and debug server listen address"`
	Interval           int    `long:"interval" env:"GCP_EXPORTER_INTERVAL" description:"Number of seconds between requesting data from GCP"`
	ServiceAccountFile string `long:"service-account-file" env:"GCP_EXPORTER_SERVICE_ACCOUNT_FILE" description:"Path to GCP Service Account JSON file"`

	ctx      context.Context
	client   *http.Client
	provider collectors.ProviderInterface

	wg *sync.WaitGroup
}

func (sc *StartExporterServiceCommand) Execute(cliCtx *cli.Context) {
	logrus.Infoln("Starting metrics service")

	methods := []func(context *cli.Context) error{
		sc.registerSignalHandler,
		sc.prepareClient,
		sc.prepareProvider,
		sc.startMetricsServer,
		sc.startExporterService,
	}

	for _, method := range methods {
		err := method(cliCtx)
		if err != nil {
			logrus.WithError(err).Fatalln("Exporter service failed")
		}
	}
}

func (sc *StartExporterServiceCommand) registerSignalHandler(cliCtx *cli.Context) error {
	logrus.Debugln("Registering signal handler")

	sc.wg = &sync.WaitGroup{}
	ctx, cancelFn := context.WithCancel(context.Background())
	sc.ctx = ctx

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-signals
		logrus.WithField("signal", sig).Warnln("Received signal. Exiting...")

		cancelFn()

		logrus.Infoln("Terminated. Goodbye!")
		sc.wg.Wait()
	}()

	return nil
}

func (sc *StartExporterServiceCommand) prepareClient(cliCtx *cli.Context) error {
	var err error
	serviceAccountFilePath := cliCtx.String("service-account-file")

	sc.client, err = google_client.New(serviceAccountFilePath)
	if err != nil {
		return fmt.Errorf("failed to create HTTP client: %v", err)
	}

	return nil
}

func (sc *StartExporterServiceCommand) prepareProvider(cliCtx *cli.Context) error {
	sc.provider = collectors.NewProvider(sc.client)
	return sc.provider.Init(cliCtx)
}

func (sc *StartExporterServiceCommand) startMetricsServer(cliCtx *cli.Context) error {
	listenAddr := cliCtx.String("listen")

	sc.wg.Add(1)

	ms := services.NewMetricsService(sc.ctx, listenAddr, sc.wg)
	ms.MustRegisterPrometheusCollector(sc.provider)
	ms.MustRegisterPrometheusCollector(version.AppVersion.VersionCollector())
	err := ms.StartServer()
	if err != nil {
		return fmt.Errorf("failed to start metrics HTTP server: %v", err)
	}

	return nil
}

func (sc *StartExporterServiceCommand) startExporterService(cliCtx *cli.Context) error {
	interval := cliCtx.Int("interval")

	sc.wg.Add(1)
	es := services.NewExporterService(sc.ctx, interval, sc.provider, sc.wg)

	err := es.Run()
	if err != nil {
		return fmt.Errorf("failure during exporter service execution: %v", err)
	}

	return nil
}

func NewStartCommand() cli.Command {
	cmd := &StartExporterServiceCommand{
		Interval:           DefaultInterval,
		ServiceAccountFile: collectors.DefaultServiceAccountFile,
	}

	return PrepareCommand("start", "Start exporter service", cmd, collectors.Collectors.Flags()...)
}
