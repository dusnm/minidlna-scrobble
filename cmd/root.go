package cmd

import (
	"context"
	"log"
	"os"

	"github.com/dusnm/minidlna-scrobble/pkg/constants"
	"github.com/dusnm/minidlna-scrobble/pkg/container"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "minidlna-scrobble",
	Short: "This program watches the minidlna log file for changes and scrobbles listens to last.fm",
	Long: `Copyright (C) 2025 Dušan Mitrović <dusan@dusanmitrovic.rs>
Licensed under the terms of the GNU GPL v3 only`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	c, err := container.New()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.WithValue(context.Background(), constants.ContextKeyContainer, c)

	rootCmd.SetContext(ctx)
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}
