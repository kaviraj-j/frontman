package main

import (
	"fmt"
	"frontman/internal/config"
	"log"
	"os"
)

func main() {
	configPath, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	// check if user has provided path
	args := os.Args
	fmt.Println(args)
	if len(args) > 1 {
		configPath = args[1]
	}

	config, err := config.Load(configPath)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(config)
}
