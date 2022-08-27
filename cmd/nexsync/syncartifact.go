package main

import (
	"github.com/marsskop/nexsync/internal/artifacts"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var artifact string

var syncArtifactCmd = &cobra.Command{
	Use:   "syncartifact",
	Short: "Synchronize artifact",
	Long:  "Synchronize artifact between repositories in Nexus2/Nexus3",
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
		diff, err := artifacts.GetDiffArtifact(repoFrom, repoTo, artifact)
		if err != nil {
			log.Fatal(err)
		}
		errs := artifacts.SyncDiffArtifact(diff, repoFrom, repoTo, artifact)
		if len(errs) > 0 {
			for _, err = range errs {
				log.Warn(err)
			}
			log.Fatal("failed to synchronise all files")
		}
	},
}

func init() {
	rootCmd.AddCommand(syncArtifactCmd)
	syncArtifactCmd.PersistentFlags().StringVar(&artifact, "artifact", "", "Artifact to sync in form <groupId>/<artifactId>")
	if err := syncArtifactCmd.MarkPersistentFlagRequired("artifact"); err != nil {
		log.Fatal(err)
	}
}