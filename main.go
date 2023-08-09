package main

import (
	"aws-llama/api"
	"aws-llama/browser"
)

func main() {
	engine := api.CreateGinWebserver()
	go engine.Run("0.0.0.0:2600")

	err := browser.Authenticate(false)
	if err != nil {
		panic(err)
	}
}
