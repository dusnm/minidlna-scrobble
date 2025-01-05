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
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
		defer cancel()

		c := ctx.Value(constants.ContextKeyContainer).(*container.Container)
		defer c.Close()

		watcher := c.GetWatcherService()
		if err := watcher.Watch(ctx); err != nil {
			return err
		}

		<-ctx.Done()

		return nil
	},
}

func init() {
	rootCmd.AddCommand(scrobbleCmd)
}
