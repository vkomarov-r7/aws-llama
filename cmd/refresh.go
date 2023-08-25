/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"aws-llama/api"
	"aws-llama/browser"
	"aws-llama/log"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
)

// refreshCmd represents the refresh command
var refreshCmd = &cobra.Command{
	Use:   "refresh",
	Short: "Make a one-time refresh of all credentials",
	Long:  `This makes a one-time refresh of all credentials`,
	Run: func(cmd *cobra.Command, args []string) {
		var r *gin.Engine
		if !api.IsWebserverRunning() {
			log.Logger.Info("Webserver is not running, starting one.")
			r = api.CreateGinWebserver()
			go api.RunWebserver(r)
		}

		browser.AttemptAuthentication()
	},
}

func init() {
	rootCmd.AddCommand(refreshCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// refreshCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// refreshCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
