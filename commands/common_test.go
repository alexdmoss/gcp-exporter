package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli"
)

type FakeCommand struct {
	FlagOne string `long:"flag-one" usage:"Flag one"`
}

func (fc *FakeCommand) Execute(*cli.Context) {}

func TestPrepareCommand(t *testing.T) {
	flags := []cli.Flag{
		cli.StringFlag{
			Name:  "string-flag",
			Usage: "String flag",
		},
	}
	cmd := PrepareCommand("fake-command", "Fake command", &FakeCommand{}, flags...)

	t.Log(cmd)

	assert.Equal(t, "fake-command", cmd.Name)
	assert.Equal(t, "Fake command", cmd.Usage)

	require.Len(t, cmd.Flags, 2)

	flag1 := cmd.Flags[0]
	assert.Equal(t, "flag-one", flag1.GetName())

	flag2 := cmd.Flags[1]
	assert.Equal(t, "string-flag", flag2.GetName())
}
