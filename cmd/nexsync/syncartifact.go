package main

import (
	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

var syncArtifactCmd = &cobra.Command{
	Use:   "syncartifact",
	Short: "Synchronize artifact",
	Long:  "Synchronize artifact between repositories in Nexus2/Nexus3",
	Run: func(cmd *cobra.Command, args []string) {
		log.Debug("sync artifact")
	},
}

func init() {
	rootCmd.AddCommand(syncArtifactCmd)
}
