package main

import (
	"Centralized-Data-Collector/internal/cache"
	"Centralized-Data-Collector/internal/collector"
	"Centralized-Data-Collector/internal/db"
	"Centralized-Data-Collector/pkg/logger"
	"context"
	"fmt"
	"log"
	"time"
)

func main() {
	logger.EnableDebugLog(true)
	logger.EnableWriteFileLog(true)
	logger.InitLoggers("./logs", 60)

	defer fmt.Println("main exiting")

	ctx := context.Background()
	// // Load configuration
	// cfg, err := config.LoadConfig()
	// if err != nil {
	// 	log.Fatalf("Error loading configuration: %v", err)
	// }

	// Initialize database connection
	dbUser := "postgres"
	dbPassword := "test_password2233"
	dbHost := "localhost" //coinhub_test_db
	dbPort := 5432
	dbDatabase := "coinhub_dex_test"
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
		dbHost,
		dbUser,
		dbPassword,
		dbDatabase,
		dbPort,
	)
	err := db.InitGorm(dsn)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer db.CloseGorm()
	logger.Info("Database connected.")

	// Initialize Redis connection
	redisAddr := "localhost:6379" //ccd-redis
	redisPassword := ""
	redisClient, err := cache.InitRedisClient(redisAddr, redisPassword)
	if err != nil {
		log.Fatalf("Error connecting to Redis: %v", err)
	}
	defer redisClient.Close()
	logger.Info("Redis connected.")

	// // Start data collection process

	dataCollector, err := collector.NewDataCollector(ctx, redisClient)
	if err != nil {
		logger.Fatal("Error creating data collector: %v", err)
	}

	logger.Info("Data collector created.")

	go dataCollector.Start(ctx)
	if err != nil {
		logger.Error("Error starting data collector: %v", err)
	}

	logger.Info("Starting Subscribe Dispatcher...")
	subscribeDispatcher := collector.NewSubscribeDispatcher(redisClient,
		dataCollector,
		collector.BINANCE_SUBSCRIPTION_STREAM,
		"binanceDataCollectors", "binanceDataCollector-1")
	logger.Info("Subscribe Dispatcher created.")
	for {
		err := subscribeDispatcher.Start(ctx)
		if err != nil {
			logger.Error("Error starting subscribe dispatcher: %v", err)
		}
		time.Sleep(1 * time.Second)
	}

}
