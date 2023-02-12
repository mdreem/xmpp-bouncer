package cmd

import (
	"context"
	"fmt"
	"sync"
	"xmpp-bouncer/client"
	"xmpp-bouncer/common"
	"xmpp-bouncer/logger"
	"xmpp-bouncer/persistence"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var RootCmd = &cobra.Command{
	Use: "xmpp-bouncer",
	Run: runCommand,
}

func getEnvVar(name string) string {
	result := viper.Get(name)
	if result == nil {
		logger.Sugar.Fatalw("failed to get env variable", "name", name)
	}

	return result.(string)
}

func runCommand(command *cobra.Command, _ []string) {
	logger.Sugar.Infow("starting xmpp-bouncer...")

	viper.SetEnvPrefix("xmpp")
	viper.AutomaticEnv()

	username := getEnvVar("USERNAME")
	password := getEnvVar("PASSWORD")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	connectionString := getConnectionString(command)
	dbWriter := persistence.NewDBWriter(connectionString)

	err := dbWriter.Migrate("migrations")
	if err != nil {
		logger.Sugar.Fatalw("unable to migrate database", "error", err)
	}

	connection, err := client.Connect(ctx, username, password, persistence.ReceiveMessage(dbWriter))
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

func getConnectionString(command *cobra.Command) string {
	hostname := common.GetString(command, "hostname")
	port := common.GetString(command, "port")
	database := common.GetString(command, "database")

	dbUsername := getEnvVar("DB_USERNAME")
	dbPassword := getEnvVar("DB_PASSWORD")
	connectionString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", dbUsername, dbPassword, hostname, port, database)
	return connectionString
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		logger.Sugar.Fatalw("could not execute command. ", "error", err)
	}
}

func init() {
	flags := RootCmd.PersistentFlags()
	flags.StringP("hostname", "H", "", "hostname.")
	flags.StringP("port", "p", "3306", "port.")
	flags.StringP("database", "d", "xmpp-bouncer", "database name.")

	markPersistentFlagRequired("hostname")
}

func markPersistentFlagRequired(flagName string) {
	err := RootCmd.MarkPersistentFlagRequired(flagName)
	if err != nil {
		logger.Sugar.Fatalw("unable to set flag to required.", "flag", flagName)
	}
}
