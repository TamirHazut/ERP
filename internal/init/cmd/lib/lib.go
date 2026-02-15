package lib

import (
	"context"
	"os"
	"time"

	"erp.localhost/infra/env"
	infra_error "erp.localhost/infra/error"
	"erp.localhost/infra/event/producer"
	"erp.localhost/infra/logging/logger"
	shared "erp.localhost/infra/model/shared"
	event "erp.localhost/init/kafka"
	"erp.localhost/init/seeder"
)

func Main() {
	// Initialize logger
	logger := logger.NewBaseLogger(shared.ModuleInit)
	defer logger.Close()

	disableInit := env.GetEnv("DISABLE_INIT", "")
	if disableInit != "" {
		logger.Info("ERP System - Init Service disabled")
		return
	}
	logger.Info("ERP System - Init Service Started")

	logger.Info("Starting Kafka configurations")
	cfg := producer.DefaultConfig()
	if err := initKafka(cfg, logger); err != nil {
		logger.Fatal("failed to init Kafka", "error", err)
		os.Exit(1)
	}
	logger.Info("Kafka configured successfully")

	// Run seeding
	logger.Info("Starting system data seeding")
	if err := initSeeder(logger); err != nil {
		logger.Fatal("failed to init seeder", "error", err)
		os.Exit(1)
	}
	logger.Info("System data seeded successfully")
	logger.Info("Init Service - Exiting")
}

func initSeeder(logger logger.Logger) *infra_error.AppError {
	s, err := seeder.NewSeeder(logger)
	if err != nil {
		return err
	}
	if err := s.SeedSystemData(); err != nil {
		return err
	}
	return nil
}

func initKafka(cfg producer.Config, logger logger.Logger) *infra_error.AppError {
	topicsInit := event.NewTopicsInitializer(cfg.Brokers, logger)

	// Create topics with timeout
	createCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := topicsInit.EnsureTopicsExist(createCtx); err != nil {
		return err
	}

	// Verify topics were created
	if err := topicsInit.VerifyTopicsExist(); err != nil {
		return err
	}
	return nil
}
