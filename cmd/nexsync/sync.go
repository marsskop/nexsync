package main

import (
	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Synchronize repositories",
	Long:  "Synchronize whole repositories from and to Nexus2/Nexus3",
	Run: func(cmd *cobra.Command, args []string) {
		log.Debug("sync")
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
}
