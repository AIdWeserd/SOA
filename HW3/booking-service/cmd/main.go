package main

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go.uber.org/zap"

	"booking-service/config"
	"booking-service/internal/grpcclient"
	"booking-service/internal/handler"
	"booking-service/internal/repository"
	"booking-service/internal/service"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync() //nolint:errcheck

	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("failed to load config", zap.Error(err))
	}

	db, err := connectDB(cfg.BookingDBDSN, logger)
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	if err := runMigrations(cfg.BookingDBDSN, logger); err != nil {
		logger.Fatal("failed to run migrations", zap.Error(err))
	}

	retryCfg := grpcclient.RetryConfig{
		MaxRetries: cfg.GRPCMaxRetries,
		BaseMs:     cfg.GRPCRetryBaseMs,
	}
	cbCfg := grpcclient.CircuitBreakerConfig{
		FailureThreshold: cfg.CBFailureThreshold,
		TimeoutSec:       cfg.CBTimeoutSec,
		HalfOpenMax:      cfg.CBHalfOpenMax,
	}
	flightClient, conn, err := grpcclient.NewFlightServiceClient(
		cfg.FlightServiceAddr,
		cfg.FlightServiceAPIKey,
		retryCfg,
		cbCfg,
		logger,
	)
	if err != nil {
		logger.Fatal("failed to create flight service gRPC client", zap.Error(err))
	}
	defer conn.Close()

	logger.Info("connected to flight-service", zap.String("addr", cfg.FlightServiceAddr))

	bookingRepo := repository.New(db, logger)
	bookingSvc := service.New(bookingRepo, flightClient, logger)

	router := gin.New()
	router.Use(gin.Recovery())

	api := router.Group("/api/v1")
	handler.NewBookingHandler(bookingSvc, logger).RegisterRoutes(api)
	handler.NewFlightProxyHandler(flightClient, logger).RegisterRoutes(api)

	addr := fmt.Sprintf(":%s", cfg.HTTPPort)
	logger.Info("booking-service HTTP server starting", zap.String("addr", addr))
	if err := router.Run(addr); err != nil {
		logger.Fatal("HTTP server failed", zap.Error(err))
	}
}

func connectDB(dsn string, logger *zap.Logger) (*sqlx.DB, error) {
	var db *sqlx.DB
	var err error
	for i := 0; i < 10; i++ {
		db, err = sqlx.Connect("postgres", dsn)
		if err == nil {
			return db, nil
		}
		logger.Warn("waiting for database", zap.Int("attempt", i+1), zap.Error(err))
		time.Sleep(3 * time.Second)
	}
	return nil, fmt.Errorf("could not connect to database: %w", err)
}

func runMigrations(dsn string, logger *zap.Logger) error {
	m, err := migrate.New("file:///app/migrations", dsn)
	if err != nil {
		return fmt.Errorf("migrate.New: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migrate.Up: %w", err)
	}
	logger.Info("migrations applied successfully")
	return nil
}
