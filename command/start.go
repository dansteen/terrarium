package command

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// StartApp will initialize this terrarium environment by creating the workspace and starting the support applications
func StartApp(cmd *cobra.Command, args []string) {
	// grab our workspace
	workspace := viper.GetString("workspace")
	fmt.Println(workspace)

}
