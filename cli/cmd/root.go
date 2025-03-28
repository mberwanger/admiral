package cmd

import (
	"errors"
	"fmt"
	utillog "github.com/mberwanger/admiral/cli/util/log"
	"github.com/mberwanger/admiral/cli/util/text"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/mberwanger/admiral/server/version"
)

func Execute(version version.Info, exit func(int), args []string) {
	newRootCmd(version, exit).Execute(args)
}

type MultiError struct {
	Errors []error
}

// Error implements the error interface for MultiError.
func (me *MultiError) Error() string {
	var errorStrings []string
	for _, err := range me.Errors {
		errorStrings = append(errorStrings, err.Error())
	}
	return strings.Join(errorStrings, ", ")
}

func (cmd *rootCmd) Execute(args []string) {
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
		log.WithError(err).Error(msg)
		cmd.exit(code)
	}
}

type rootCmd struct {
	cmd  *cobra.Command
	exit func(int)
}

func newRootCmd(version version.Info, exit func(int)) *rootCmd {
	var logFormat, logLevel string

	root := &rootCmd{
		exit: exit,
	}

	cmd := &cobra.Command{
		Use:               "admiral",
		Short:             "Admiral - Platform Orchestrator",
		Version:           version.String(),
		SilenceUsage:      true,
		SilenceErrors:     true,
		Args:              cobra.NoArgs,
		ValidArgsFunction: cobra.NoFileCompletions,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if err := SetLogFormat(logFormat); err != nil {
				return wrapError(err, "failed to set log format")
			}

			if err := SetLogLevel(logLevel); err != nil {
				return wrapError(err, "failed to set log level")
			}

			return nil
		},
	}
	cmd.SetVersionTemplate("{{.Version}}")

	// general options
	cmd.PersistentFlags().BoolP("help", "h", false, "help for admiral cli")
	cmd.PersistentFlags().StringVar(&logFormat, "logformat", "text", "Set the logging format. One of: text|json")
	cmd.PersistentFlags().StringVar(&logLevel, "loglevel", "info", "Set the logging level. One of: debug|info|warn|error")

	cmd.AddCommand(
		newManCmd().cmd,
	)

	root.cmd = cmd
	return root
}

func SetLogFormat(logFormat string) (err error) {
	switch strings.ToLower(logFormat) {
	case utillog.JsonFormat:
		err = os.Setenv(utillog.EnvLogFormat, utillog.JsonFormat)
	case utillog.TextFormat, "":
		err = os.Setenv(utillog.EnvLogFormat, utillog.TextFormat)
	default:
		err = fmt.Errorf("unknown log format '%s'", logFormat)
	}
	log.SetFormatter(utillog.CreateFormatter(logFormat))
	return err
}

func SetLogLevel(logLevel string) (err error) {
	level, err := log.ParseLevel(text.FirstNonEmpty(logLevel, log.InfoLevel.String()))
	if err != nil {
		return err
	}
	err = os.Setenv(utillog.EnvLogLevel, level.String())
	log.SetLevel(level)
	return err
}
