package cmd

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/dusnm/minidlna-scrobble/pkg/constants"
	"github.com/dusnm/minidlna-scrobble/pkg/container"
	"github.com/spf13/cobra"
)

var scrobbleCmd = &cobra.Command{
	Use:   "scrobble",
	Short: "Start watching the minidlna log file and scrobble on changes",
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
		defer cancel()

		c := ctx.Value(constants.ContextKeyContainer).(*container.Container)
		defer c.Close()

		c.GetJobService().Work(ctx)

		logger := c.Logger.With().Str("command", "scrobble").Logger()
		watcher := c.GetWatcherService()

		logger.Info().Msg("starting watcher")

		if err := watcher.Watch(ctx); err != nil {
			logger.Fatal().Err(err).Msg("")
		}

		<-ctx.Done()
	},
}

func init() {
	rootCmd.AddCommand(scrobbleCmd)
}
