package command

import (
	"os"
	"path/filepath"

	"github.com/dansteen/terrarium/consul"
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

	// get the name of this application
	appName, err := GetAppName(appPath)
	if err != nil {
		os.Exit(1)
	}
	// load up our config data
	consulService.Load(filepath.Join(appPath, "infra/data.yml"), appName, hashLabel)
}

// GetAppName will pull the name of the application from the appPath provided (it expects that appPath is a git repo)
// it modifies the names to replace underscores with hyphens
func GetAppName(appPath string) (string, error) {
	git.Repository{}
	return "", nil
}
