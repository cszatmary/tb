package cmd

import (
	"github.com/TouchBistro/tb/config"
	"github.com/TouchBistro/tb/fatal"
	_ "github.com/TouchBistro/tb/release"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "tb",
	Version: "0.0.8", // TODO: Fix this hardcoded bullshit
	Short:   "tb is a CLI for running TouchBistro services on a development machine",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fatal.ExitErr(err, "Failed executing command.")
	}
}

func init() {
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	err := config.InitRC()
	if err != nil {
		fatal.ExitErr(err, "Failed to initialise .tbrc file.")
	}

	logLevel, err := log.ParseLevel(config.TBRC().LogLevel)
	if err != nil {
		fatal.ExitErr(err, "Failed to initialise logger level.")
	}

	log.SetLevel(logLevel)
	log.SetFormatter(&log.TextFormatter{
		// TODO: Remove the log level its quite ugly
		DisableTimestamp: true,
	})

	// TODO: Make this its own setting or make the format less intense.
	log.SetReportCaller(logLevel == log.DebugLevel)

	err = config.Init()
	if err != nil {
		fatal.ExitErr(err, "Failed to initialise config files.")
	}
}

func Root() *cobra.Command {
	return rootCmd
}
