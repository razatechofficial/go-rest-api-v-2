package main

import (
	"fmt"

	"github.com/razatechofficial/go-rest-api-v-2/config"
)

func main() {

	// Step 1: Load configuration
	// Configuration comes from YAML files and environment variables
	cfg, err := config.Load()
	if err != nil {
		// Use standard logger since our logger isn't initialized yet
		panic("Failed to load configuration: " + err.Error())
	}

	fmt.Printf("Configuration loaded successfully: %+v\n", cfg)
}
