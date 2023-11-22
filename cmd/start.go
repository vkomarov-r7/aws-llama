/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the program as a service in the background.",
	Long: `Starts the program as a service in the background.

Debug information can be found in ~/.awsllama (including logs).
`,
	Run: func(cmd *cobra.Command, args []string) {
		// Veneer on top of launchctl to start the process as a service.
		startCmd := exec.Command("launchctl", "start", LAUNCHCTL_LABEL)
		err := startCmd.Run()
		if err != nil {
			panic(err)
		}
		fmt.Println("Service started successfully.")
	},
}

func init() {
	rootCmd.AddCommand(startCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// startCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// startCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
