package main

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	log "github.com/sirupsen/logrus"
)

var (
	nexusFrom, userFrom, passFrom, repoFrom, nexusTo, userTo, passTo, repoTo, cfgFile, tmpDir string
	nexus2From, nexus2To, debug                                                               bool
)

var rootCmd = &cobra.Command{
	Use:   "nexsync",
	Short: "nexus synchronisation tool",
	Long:  "nexsync is a tool to synchronise Maven repositories and artifacts from and to Nexus2/Nexus3",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&nexusFrom, "nexusfrom", "http://localhost:8080", "Nexus endpoint to sync from")
	rootCmd.PersistentFlags().StringVar(&userFrom, "userfrom", "", "Nexus user to authenticate with in nexusFrom")
	rootCmd.PersistentFlags().StringVar(&passFrom, "passfrom", "", "Password for Nexus user in nexusFrom")
	rootCmd.PersistentFlags().StringVar(&repoFrom, "repofrom", "", "Repository to sync from in nexusFrom")
	rootCmd.PersistentFlags().BoolVar(&nexus2From, "nexus2from", false, "nexusFrom is Nexus2")

	rootCmd.PersistentFlags().StringVar(&nexusTo, "nexusto", "http://localhost:8080", "Nexus endpoint to sync to")
	rootCmd.PersistentFlags().StringVar(&userTo, "userto", "", "Nexus user to authenticate with in nexusTo")
	rootCmd.PersistentFlags().StringVar(&passTo, "passto", "", "Password for Nexus user in nexusTo")
	rootCmd.PersistentFlags().StringVar(&repoTo, "repoto", "", "Repository to sync to in nexusTo")
	rootCmd.PersistentFlags().BoolVar(&nexus2To, "nexus2to", false, "nexusTo is Nexus2")

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "Path to config file; default is ~/.nexsync")
	rootCmd.PersistentFlags().StringVar(&tmpDir, "dir", "/tmp", "Directory to store artifacts in during sync")
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "enable debug")
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".nexsync")
	}
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err == nil {
		log.Info("Using config file: ", viper.ConfigFileUsed())
	}
	enableDebug()
}

func enableDebug() {
	log.SetReportCaller(true)
	if debug {
		log.SetLevel(log.DebugLevel)
	}
}
