package command

import (
	"path/filepath"

	"github.com/dansteen/terrarium/consul"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// StartApp will initialize this terrarium environment by creating the workspace and starting the support applications
func StartApp(cmd *cobra.Command, args []string) {
	// grab our workspace
	//workspace := viper.GetString("workspace")

	// load up our config data
	consul.Load(filepath.Join(viper.GetString("appPath"), "infra/data.yml"), viper.GetString("hashLabel"))

}
