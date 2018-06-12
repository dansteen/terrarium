package command

import (
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/dansteen/terrarium/consul"
	"github.com/dansteen/terrarium/vault"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	git "gopkg.in/src-d/go-git.v4"
)

// StartApp will initialize this terrarium environment by creating the workspace and starting the support applications
func StartApp(cmd *cobra.Command, args []string) {
	// grab our workspace
	workspace := viper.GetString("workspace")
	appPath := viper.GetString("appPath")
	hashLabel := viper.GetString("hashLabel")

	// first get our consul service information
	consulService, err := consul.GetService(workspace)
	if err != nil {
		os.Exit(1)
	}

	// then get our vault service information
	vaultService, err := vault.GetService(workspace)
	if err != nil {
		os.Exit(1)
	}

	// get the name of this application
	appName, err := GetAppName(appPath)
	if err != nil {
		os.Exit(1)
	}
	// load up our config data
	err = consulService.Load(filepath.Join(appPath, "infra/data.yml"), appName, hashLabel)
	if err != nil {
		os.Exit(1)
	}

	// load up our vault instance
	err = vaultService.Load(filepath.Join(appPath, "infra/secrets.yml"))
	if err != nil {
		os.Exit(1)
	}
}

// GetAppName will pull the name of the application from the appPath provided (it expects that appPath is a git repo)
// it modifies the names to replace underscores with hyphens
func GetAppName(appPath string) (string, error) {
	r, err := git.PlainOpen(appPath)
	if err != nil {
		log.Error().Err(err).Msgf("Could not get app name from repository at %s:", appPath)
		return "", err
	}
	origin, err := r.Remote("origin")
	if err != nil {
		log.Error().Err(err).Msgf("Could not get app name from repository at %s:", appPath)
		return "", err
	}
	// if there are no URLs set we fail
	if len(origin.Config().URLs) < 1 {
		log.Error().Msgf("Could not get app name form repository at %s: does not have a remote origin url.", appPath)
		return "", err
	}
	// we grab the first url and strip out the repository name.
	appName := path.Base(origin.Config().URLs[0])
	// then we remove anything following a period
	appName = strings.SplitN(appName, ".", 2)[0]
	// then we convert underscores to hyphens
	appName = strings.Replace(appName, "_", "-", -1)

	return appName, nil
}
