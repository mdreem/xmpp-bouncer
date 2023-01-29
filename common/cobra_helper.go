package common

import (
	"github.com/spf13/cobra"
	"xmpp-bouncer/logger"
)

func GetString(rootCmd *cobra.Command, option string) string {
	optionString, err := rootCmd.Flags().GetString(option)

	if err != nil {
		logger.Sugar.Fatalw("could not fetch %s option", "option", option, "error", err)
	}
	return optionString
}
