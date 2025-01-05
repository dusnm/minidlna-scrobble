package cmd

import (
	"fmt"
	"net/url"

	"github.com/dusnm/minidlna-scrobble/pkg/constants"
	"github.com/dusnm/minidlna-scrobble/pkg/container"
	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authenticate with last.fm",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		c := ctx.Value(constants.ContextKeyContainer).(*container.Container)
		logger := c.Logger.With().Str("command", "auth").Logger()
		authService := c.GetAuthService()
		sessionCacheService := c.GetSessionCacheService()
		token, err := authService.GetToken(ctx)
		if err != nil {
			logger.Fatal().Err(err).Msg("")
		}

		userAPI, _ := url.Parse(constants.UserAPIBaseURL)
		userAPI.Path += "/auth/"

		query := userAPI.Query()
		query.Add("api_key", c.Cfg.Credentials.APIKey)
		query.Add("token", token)

		userAPI.RawQuery = query.Encode()

		fmt.Printf(
			"Authenticate with last.fm by following the provided link. Afterwards, press RETURN to continue.\n\n%s\n",
			userAPI.String(),
		)

		// Block until the user authenticates the session
		fmt.Scanln()

		sessionKey, err := authService.GetSessionKey(ctx, token)
		if err != nil {
			logger.Fatal().Err(err).Msg("")
		}

		if err = sessionCacheService.Save(sessionKey); err != nil {
			logger.Fatal().Err(err).Msg("")
		}

		fmt.Println("Authentication details saved.")
	},
}

func init() {
	rootCmd.AddCommand(authCmd)
}
