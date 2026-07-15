package main

import (
	"flag"
	"log"

	"pqr.info/mev/pqrld/internal/config"
	"pqr.info/mev/pqrld/internal/executor"
	"pqr.info/mev/pqrld/internal/grpc"
)

func main() {
	configPath := flag.String("config", "/etc/sos/runlevels.toml", "path to runlevels toml configuration")
	grpcPort := flag.Int("port", 11112, "gRPC server control port")
	flag.Parse()

	log.Println("Initializing PQRL.d bootloader engine...")

	// 1. Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Printf("Warning: failed to load config from %s (%v). Using default stub configurations.", *configPath, err)
		cfg = &config.Config{
			Bootloader: config.BootloaderConfig{
				Version:             "1.0.0",
				Mode:                "sovereign",
				HeartbeatIntervalMs: 5000,
			},
			Runlevels: make(map[string]config.RunlevelConfig),
		}
	}
	cfg.SetDefaults()

	// 2. Instantiate FSM boot engine
	engine := executor.NewExecutor(cfg)

	// 3. Start boot sequence in background
	go func() {
		if err := engine.Run(); err != nil {
			log.Fatalf("Fatal: boot sequence failed: %v", err)
		}
	}()

	// 4. Start gRPC remote control server
	if err := grpc.StartGRPCServer(*grpcPort, engine); err != nil {
		log.Fatalf("Fatal: gRPC server failed to start: %v", err)
	}
}
