package commands

import (
	"github.com/urfave/cli"

	clihelpers "gitlab.com/ayufan/golang-cli-helpers"
)

type command interface {
	Execute(*cli.Context)
}

func PrepareCommand(name string, usage string, cmd command, additionalFlags ...cli.Flag) cli.Command {
	flags := clihelpers.GetFlagsFromStruct(cmd)

	return cli.Command{
		Name:   name,
		Usage:  usage,
		Action: cmd.Execute,
		Flags:  append(flags, additionalFlags...),
	}
}
