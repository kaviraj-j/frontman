package main

import (
	"frontman/internal/config"
	"frontman/internal/engine"
	"frontman/internal/server"
	"frontman/internal/stats"
	"log"
	"os"
)

func main() {
	// open (or create) a minimal log file at project root
	logFile, err := os.OpenFile("frontman.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		log.Fatalf("failed to open log file: %v", err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)
	log.SetFlags(log.LstdFlags)

	configPath := getConfigPath()
	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatal(err)
	}

	apiStats := stats.NewApiStats(cfg)

	eng := engine.NewEngine(cfg, apiStats)
	srv := server.NewServer(eng, cfg, apiStats)
	srv.Run()
}

func getConfigPath() string {
	configPath, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	// check if user has provided path
	args := os.Args
	if len(args) > 1 {
		configPath = args[1]
	}
	return configPath
}
