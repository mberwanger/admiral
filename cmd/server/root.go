package server

import (
	"errors"
	"strings"

	goversion "github.com/caarlos0/go-version"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	configFile  string
	envVarFiles envFiles
	debug       bool
)

type rootCmd struct {
	cmd  *cobra.Command
	log  *zap.Logger
	exit func(int)
}

type envFiles []string

func (f *envFiles) String() string {
	return strings.Join(*f, ",")
}

func (f *envFiles) Set(value string) error {
	*f = append(*f, value)
	return nil
}

func (f *envFiles) Type() string {
	return "envFiles"
}

// Execute initializes and runs the root command.
func Execute(versionInfo goversion.Info, exitFunc func(int), args []string) {
	cmd := newRootCmd(versionInfo, exitFunc)
	if err := cmd.Execute(args); err != nil {
		exitFunc(1)
	}
}

func newRootCmd(versionInfo goversion.Info, exit func(int)) *rootCmd {
	// Set up logging with a production configuration.
	logger, err := zap.NewProduction()
	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}
	defer func() { _ = logger.Sync() }()

	root := &rootCmd{
		log:  logger,
		exit: exit,
	}

	cmd := &cobra.Command{
		Use:               "admiral-server",
		Short:             "Admiral - Platform Orchestrator that helps developers build, deploy, and manage their applications",
		Version:           versionInfo.String(),
		SilenceUsage:      true,
		SilenceErrors:     true,
		Args:              cobra.NoArgs,
		ValidArgsFunction: cobra.NoFileCompletions,
		PreRun: func(cmd *cobra.Command, args []string) {
			logger.Info("using configuration", zap.String("file", configFile))
		},
	}
	cmd.SetVersionTemplate("{{.Version}}")
	cmd.CompletionOptions.DisableDefaultCmd = true

	cmd.PersistentFlags().StringVarP(&configFile, "config", "c", "config.yaml", "Load configuration from file")
	_ = cmd.MarkFlagFilename("config", "yaml", "yml")
	cmd.PersistentFlags().BoolVar(&debug, "debug", false, "print the final configuration file to stdout")
	cmd.PersistentFlags().Var(&envVarFiles, "env", "path to additional .env files to load")

	cmd.AddCommand(
		newMigrateCmd().Cmd,
		newStartCmd().Cmd,
	)
	root.cmd = cmd
	return root
}

func (cmd *rootCmd) Execute(args []string) error {
	cmd.cmd.SetArgs(args)

	if err := cmd.cmd.Execute(); err != nil {
		code := 1
		msg := "command failed"

		eerr := &exitError{}
		if errors.As(err, &eerr) {
			code = eerr.code
			if eerr.details != "" {
				msg = eerr.details
			}
		}
		cmd.log.Error(msg, zap.Error(err))
		cmd.exit(code)

		return err
	}

	return nil
}
