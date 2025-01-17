package cmd

import (
	"context"
	"os"

	"github.com/dusnm/minidlna-scrobble/pkg/constants"
	"github.com/dusnm/minidlna-scrobble/pkg/container"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

const (
	flagLogLevel  = "log-level"
	flagLogLevelS = "l"
)

var (
    // passed directly to the linker
	version string

	rootCmd = &cobra.Command{
		Use:   "minidlna-scrobble",
		Short: "This program watches the minidlna log file for changes and scrobbles listens to last.fm",
		Long: `Copyright (C) 2025 Dušan Mitrović <dusan@dusanmitrovic.rs>
Licensed under the terms of the GNU GPL v3 only`,
		Version: version,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			log.Logger = log.Output(zerolog.ConsoleWriter{
				Out:        os.Stderr,
				NoColor:    true,
				TimeFormat: "15:04",
			})

			level, err := cmd.Flags().GetString(flagLogLevel)
			if err != nil {
				log.Fatal().Err(err).Msg("")
			}

			logLevel, err := zerolog.ParseLevel(level)
			if err != nil {
				log.Fatal().Err(err).Msg("invalid level")
			}

			c, err := container.New(logLevel)
			if err != nil {
				log.Fatal().Err(err).Msg("")
			}

			ctx := context.WithValue(context.Background(), constants.ContextKeyContainer, c)
			cmd.SetContext(ctx)
		},
	}
)

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.
		PersistentFlags().
		StringP(
			flagLogLevel,
			flagLogLevelS,
			zerolog.ErrorLevel.String(),
			"set the logger level, can be one of: trace, debug, info, warn, error, fatal, panic",
		)
}
