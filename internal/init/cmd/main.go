package cmd

import (
	"os"

	"erp.localhost/internal/infra/logging/logger"
	shared "erp.localhost/internal/infra/model/shared"
	"erp.localhost/internal/init/seeder"
)

func Main() {
	// Initialize logger
	logger := logger.NewBaseLogger(shared.ModuleInit)
	logger.Info("ERP System - Init Service Started")

	// Run seeding
	logger.Info("Starting system data seeding")
	s, err := seeder.NewSeeder(logger)
	if err != nil {
		logger.Fatal("failed to init seeder", "error", err)
		os.Exit(1)
	}
	if err := s.SeedSystemData(); err != nil {
		logger.Error("Seeding failed", "error", err)
		os.Exit(1)
	}

	logger.Info("System data seeded successfully")
	logger.Info("Init Service - Exiting")
}
