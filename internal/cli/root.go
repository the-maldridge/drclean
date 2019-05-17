package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "drclean",
	Short: "drclean cleans repos and computes versions",
	Long:  `drclean can clean your repo, and can suggest new sequential versions for you to run.`,
}

// Execute provides the entrypoint into the system.
func Execute() {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
