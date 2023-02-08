package cmd

import (
	"context"
	"github.com/spf13/cobra"
	"os"
	"sync"
	"xmpp-bouncer/client"
	"xmpp-bouncer/common"
	"xmpp-bouncer/logger"
	"xmpp-bouncer/persistence"
)

var RootCmd = &cobra.Command{
	Use: "xmpp-bouncer",
	Run: runCommand,
}

func runCommand(command *cobra.Command, _ []string) {
	logger.Sugar.Infow("starting xmpp-bouncer...")

	username := common.GetString(command, "username")
	password := common.GetString(command, "password")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	connection, err := client.Connect(ctx, username, password, persistence.ReceiveMessage(persistence.NewFileWriter()))
	if err != nil {
		logger.Sugar.Fatalw("failed to establish connection", "error", err)
	}

	defer func() {
		logger.Sugar.Info("closing connection")
		if err := connection.Session.Conn().Close(); err != nil {
			logger.Sugar.Errorw("error closing connection", "error", err)
		}
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.Sugar.Infow("start serving")
		err := connection.Session.Serve(connection.Mux)
		if err != nil {
			logger.Sugar.Errorw("error handling session responses", "error", err)
		}
	}()

	client.JoinRooms(ctx, connection)

	logger.Sugar.Infow("running...")
	wg.Wait()
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		logger.Sugar.Fatalw("could not execute command. ", "error", err)
	}
}

func init() {
	flags := RootCmd.PersistentFlags()
	flags.StringP("username", "u", "", "username.")
	flags.StringP("password", "p", "", "password.")

	markPersistentFlagRequired("username")
	markPersistentFlagRequired("password")
}

func markPersistentFlagRequired(flagName string) {
	err := RootCmd.MarkPersistentFlagRequired(flagName)
	if err != nil {
		logger.Sugar.Errorw("unable to set flag to required.", "flag", flagName)
		os.Exit(1)
	}
}
