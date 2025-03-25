package app

import (
	"github.com/spf13/cobra"

	"github.com/mberwanger/admiral/server/cmd/assets"
	"github.com/mberwanger/admiral/server/config"
	"github.com/mberwanger/admiral/server/gateway"
)

type startCmd struct {
	Cmd *cobra.Command
}

func newStartCmd() *startCmd {
	root := &startCmd{}
	cmd := &cobra.Command{
		Use:               "start",
		Aliases:           []string{"s"},
		Short:             "Start the server process",
		Args:              cobra.NoArgs,
		ValidArgsFunction: cobra.NoFileCompletions,
		Run: func(cmd *cobra.Command, args []string) {
			cfg := config.Build(configFile, envVarFiles, debug)
			gateway.Run(cfg, gateway.CoreComponentFactory, assets.VirtualFS)
		},
	}

	root.Cmd = cmd
	return root
}
