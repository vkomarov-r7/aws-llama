package main

import (
	"aws-llama/api"
	"aws-llama/browser"
	"aws-llama/config"
	"aws-llama/log"
)

func main() {
	config.InitConfig()
	log.InitLogging()

	engine := api.CreateGinWebserver()
	go engine.Run("127.0.0.1:2600")

	// TODO: Change this to be on a timered loop.
	err := browser.Authenticate(false)
	if err != nil {
		panic(err)
	}
}
