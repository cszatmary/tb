package cmd

import (
	"fmt"
	"os"

	"github.com/TouchBistro/tb/config"
	"github.com/TouchBistro/tb/deps"
	"github.com/TouchBistro/tb/docker"
	"github.com/TouchBistro/tb/git"
	"github.com/TouchBistro/tb/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type nukeOptions struct {
	shouldNukeContainers bool
	shouldNukeImages     bool
	shouldNukeVolumes    bool
	shouldNukeNetworks   bool
	shouldNukeRepos      bool
	shouldNukeConfig     bool
	shouldNukeAll        bool
}

var nukeOpts nukeOptions

var nukeCmd = &cobra.Command{
	Use:   "nuke",
	Short: "Removes all docker images, containers, volumes and networks",
	PreRun: func(cmd *cobra.Command, args []string) {
		if !nukeOpts.shouldNukeContainers &&
			!nukeOpts.shouldNukeImages &&
			!nukeOpts.shouldNukeVolumes &&
			!nukeOpts.shouldNukeNetworks &&
			!nukeOpts.shouldNukeRepos &&
			!nukeOpts.shouldNukeConfig &&
			!nukeOpts.shouldNukeAll {
			log.Fatalln("Error: Must specify what to nuke")
		}

		err := deps.Resolve(deps.Docker)
		if err != nil {
			log.WithFields(log.Fields{"error": err.Error()}).Fatal("Failed to resolve dependencies.")
		}

		for _, repo := range git.RepoNames(config.Services()) {
			path := fmt.Sprintf("%s/%s", config.TBRootPath(), repo)

			if !util.FileOrDirExists(path) {
				log.Fatalf("Repo %s is missing. Please ensure all repos exist before running nuke.\n", repo)
			}
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		err := docker.StopContainersAndServices()
		if err != nil {
			log.WithFields(log.Fields{"error": err.Error()}).Fatal("Failed stopping docker containers and services.")
		}

		if nukeOpts.shouldNukeContainers || nukeOpts.shouldNukeAll {
			log.Infoln("Removing containers...")
			err = docker.RmContainers()
			if err != nil {
				log.WithFields(log.Fields{"error": err.Error()}).Fatal("Failed removing docker containers")
			}
			log.Infoln("...done")
		}

		if nukeOpts.shouldNukeImages || nukeOpts.shouldNukeAll {
			log.Infoln("Removing images...")
			err = docker.RmImages()
			if err != nil {
				log.WithFields(log.Fields{"error": err.Error()}).Fatal("Failed removing docker images")
			}
			log.Infoln("...done")
		}

		if nukeOpts.shouldNukeNetworks || nukeOpts.shouldNukeAll {
			log.Infoln("Removing networks...")
			err = docker.RmNetworks()
			if err != nil {
				log.WithFields(log.Fields{"error": err.Error()}).Fatal("Failed removing docker networks")
			}
			log.Infoln("...done")
		}

		if nukeOpts.shouldNukeVolumes || nukeOpts.shouldNukeAll {
			log.Infoln("Removing volumes...")
			err = docker.RmVolumes()
			if err != nil {
				log.WithFields(log.Fields{"error": err.Error()}).Fatal("Failed removing docker volumes")
			}
			log.Infoln("...done")
		}

		if nukeOpts.shouldNukeRepos || nukeOpts.shouldNukeAll {
			log.Infoln("Removing repos...")
			for _, repo := range git.RepoNames(config.Services()) {
				log.Debugf("Removing repo %s...", repo)
				repoPath := fmt.Sprintf("%s/%s", config.TBRootPath(), repo)
				err = os.RemoveAll(repoPath)
				if err != nil {
					log.WithFields(log.Fields{"error": err.Error(), "repo": repo}).Fatal("Failed removing git repo")
				}
			}
			log.Infoln("...done")
		}

		if nukeOpts.shouldNukeConfig || nukeOpts.shouldNukeAll {
			log.Infoln("Removing config files...")
			config.RmFiles()
			log.Infoln("...done")
		}
	},
}

func init() {
	rootCmd.AddCommand(nukeCmd)
	nukeCmd.Flags().BoolVar(&nukeOpts.shouldNukeContainers, "containers", false, "nuke all containers")
	nukeCmd.Flags().BoolVar(&nukeOpts.shouldNukeImages, "images", false, "nuke all images")
	nukeCmd.Flags().BoolVar(&nukeOpts.shouldNukeVolumes, "volumes", false, "nuke all volumes")
	nukeCmd.Flags().BoolVar(&nukeOpts.shouldNukeNetworks, "networks", false, "nuke all networks")
	nukeCmd.Flags().BoolVar(&nukeOpts.shouldNukeRepos, "repos", false, "nuke all repos")
	nukeCmd.Flags().BoolVar(&nukeOpts.shouldNukeConfig, "config", false, "nuke all config files")
	nukeCmd.Flags().BoolVar(&nukeOpts.shouldNukeAll, "all", false, "nuke everything")
}