package main

import (
	"fmt"
	"log"
	"os"

	"github.com/fsnotify/fsnotify"
	"github.com/skushnerchuk/simda/internal/config"
	"github.com/skushnerchuk/simda/internal/daemon"
	"github.com/skushnerchuk/simda/internal/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	dmnConfig  *config.DaemonConfig
	configFile string
)

const DaemonVersion = "0.0.1"

var rootCmd = &cobra.Command{
	Use:     "simda",
	Short:   "System Information Monitoring DAemon",
	Version: DaemonVersion,
	Run: func(_ *cobra.Command, _ []string) {
		if !utils.IsRoot() {
			fatal("This program must be run as root.\n")
		}

		daemonApp, err := daemon.NewDaemon(dmnConfig)
		if err != nil {
			log.Fatalln(err)
		}

		if err := daemonApp.Run(); err != nil {
			fatal(err.Error())
		}
	},
}

func fatal(msg string, args ...any) {
	fmt.Printf(msg, args...)
	os.Exit(1)
}

func loadConfig() {
	err := config.Load(configFile, &dmnConfig)
	if err != nil {
		log.Printf("load config error: %v\n", err.Error())
		log.Println("use default values")
	}
	err = dmnConfig.Validate()
	if err != nil {
		log.Printf("parse config error: %v\n", err.Error())
		log.Println("use previous values")
	}
}

func initConfig() {
	loadConfig()
	viper.WatchConfig()
	viper.OnConfigChange(func(_ fsnotify.Event) {
		loadConfig()
	})
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.Flags().StringVarP(
		&configFile,
		"config",
		"c",
		"/etc/simda/config.yml",
		"Path to configuration file",
	)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fatal("%s\n", err.Error())
	}
}
