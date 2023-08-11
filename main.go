package main

import (
	"aws-llama/cmd"
	"aws-llama/config"
	"aws-llama/log"
)

func main() {
	config.InitConfig()
	log.InitLogging()
	cmd.Execute()
}
