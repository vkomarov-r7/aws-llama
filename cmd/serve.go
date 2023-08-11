package cmd

import (
	"aws-llama/api"
	"aws-llama/browser"

	"github.com/spf13/cobra"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the main webserver instance responsible for refreshing credentials",
	Long:  `Start the main webserver instance responsible for refreshing credentials`,
	Run: func(cmd *cobra.Command, args []string) {
		go api.RunWebserver()

		browser.AuthenticationLoop()
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serveCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serveCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
