package main

import (
	"aws-llama/api"
)

func main() {
	engine := api.CreateGinWebserver()
	engine.Run("0.0.0.0:2600")
}
