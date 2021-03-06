package main

import (
	"fmt"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/urfave/cli"

	"gitlab.com/gitlab-org/ci-cd/gcp-exporter/commands"
	"gitlab.com/gitlab-org/ci-cd/gcp-exporter/version"
)

var mainFlags = []cli.Flag{
	cli.BoolFlag{
		Name:   "debug",
		EnvVar: "DEBUG",
		Usage:  "Set debug log-level",
	},
	cli.BoolFlag{
		Name:   "no-color",
		EnvVar: "NO_COLOR",
		Usage:  "Disable output coloring",
	},
}

func setupLogging(app *cli.App) {
	appBefore := app.Before
	app.Before = func(c *cli.Context) error {
		logrus.SetOutput(os.Stderr)

		formatter := new(logrus.TextFormatter)
		if c.Bool("no-color") {
			formatter.DisableColors = true
		}

		logrus.SetFormatter(formatter)

		if c.Bool("debug") {
			logrus.SetLevel(logrus.DebugLevel)
		}

		if appBefore != nil {
			return appBefore(c)
		}
		return nil
	}
}

func logStartup(app *cli.App) {
	appBefore := app.Before
	app.Before = func(c *cli.Context) error {
		logrus.Infof("Starting %s", version.AppVersion.Line())

		if appBefore != nil {
			return appBefore(c)
		}
		return nil
	}
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			// log panics forces exit
			if _, ok := r.(*logrus.Entry); ok {
				os.Exit(1)
			}
			panic(r)
		}
	}()

	app := cli.NewApp()
	app.Name = "GCP exporter"
	app.Usage = "Exports metrics about Google Cloud Platform"
	app.Version = version.AppVersion.ShortLine()
	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Println(version.AppVersion.Extended())
	}
	app.Authors = []cli.Author{
		{
			Name:  "GitLab Inc.",
			Email: "support@gitlab.com",
		},
	}
	app.Flags = mainFlags
	app.CommandNotFound = func(context *cli.Context, command string) {
		logrus.Fatalln("Command", command, "not found.")
	}

	setupLogging(app)
	logStartup(app)

	app.Commands = []cli.Command{
		commands.NewStartCommand(),
		commands.NewGetTokenCommand(),
	}

	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}
