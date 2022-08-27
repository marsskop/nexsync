package main

import (
	"github.com/marsskop/nexsync/internal/artifacts"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Synchronize repositories",
	Long:  "Synchronize whole repositories from and to Nexus2/Nexus3",
	Run: func(cmd *cobra.Command, args []string) {
		versionFrom := 3
		if nexus2From {
			versionFrom = 2
		}
		versionTo := 3
		if nexus2To {
			versionTo = 2
		}
		repoFrom := &artifacts.Nexus{
			URL:        nexusFrom,
			Version:    versionFrom,
			User:       userFrom,
			Pass:       passFrom,
			Repository: repoFrom,
			TmpDir:     tmpDir,
		}
		repoTo := &artifacts.Nexus{
			URL:        nexusTo,
			Version:    versionTo,
			User:       userTo,
			Pass:       passTo,
			Repository: repoTo,
			TmpDir:     tmpDir,
		}
		log.Warn(repoFrom, repoTo)
		diff, err := artifacts.GetDiff(repoFrom, repoTo)
		if err != nil {
			log.Fatal(err)
		}

		errs := artifacts.SyncDiff(diff, repoFrom, repoTo)
		if len(errs) > 0 {
			for _, err = range errs {
				log.Warn(err)
			}
			log.Fatal("failed to synchronise all files")
		}
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
}
