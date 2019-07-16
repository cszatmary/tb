package git

import (
	"fmt"

	"github.com/TouchBistro/tb/config"
	"github.com/TouchBistro/tb/util"
)

func repoURL(repoName string) string {
	return fmt.Sprintf("git@github.com:TouchBistro/%s.git", repoName)
}

func Clone(repoName, destDir string) error {
	repoURL := repoURL(repoName)
	destPath := fmt.Sprintf("%s/%s", destDir, repoName)
	err := util.Exec("git", "clone", repoURL, destPath)
	return err
}

func Pull(repoPath string) error {
	err := util.Exec("git", "-C", repoPath, "pull")
	return err
}

func RepoNames(services config.ServiceMap) []string {
	var repos []string

	for name, s := range services {
		if s.IsGithubRepo {
			repos = append(repos, name)
		}
	}

	return repos
}
