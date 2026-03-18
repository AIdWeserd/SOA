package main

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	flightv1 "flight-service/gen/flight/v1"
	"flight-service/config"
	"flight-service/internal/cache"
	"flight-service/internal/handler"
	"flight-service/internal/middleware"
	"flight-service/internal/repository"
	"flight-service/internal/service"
)

func main() {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync() //nolint:errcheck

	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("failed to load config", zap.Error(err))
	}

	db, err := connectDB(cfg.FlightDBDSN, logger)
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	if err := runMigrations(cfg.FlightDBDSN, logger); err != nil {
		logger.Fatal("failed to run migrations", zap.Error(err))
	}

	redisClient := newRedisClient(cfg)
	logger.Info("redis client initialized",
		zap.String("addr", cfg.RedisAddr),
		zap.String("sentinel_addrs", cfg.RedisSentinelAddrs),
	)

	repo := repository.New(db, logger)
	flightCache := cache.New(redisClient, cfg.CacheTTLFlightSec, cfg.CacheTTLSearchSec, logger)
	svc := service.New(repo, flightCache, logger)
	flightHandler := handler.NewFlightHandler(svc, logger)

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(middleware.UnaryServerAuthInterceptor(cfg.APIKey, logger)),
	)
	flightv1.RegisterFlightServiceServer(grpcServer, flightHandler)

	addr := fmt.Sprintf(":%s", cfg.GRPCPort)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Fatal("failed to listen", zap.String("addr", addr), zap.Error(err))
	}

	logger.Info("flight-service gRPC server starting", zap.String("addr", addr))
	if err := grpcServer.Serve(lis); err != nil {
		logger.Fatal("gRPC server failed", zap.Error(err))
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

// newRedisClient — standalone или Sentinel, в зависимости от конфига
func newRedisClient(cfg *config.Config) redis.UniversalClient {
	if cfg.RedisSentinelAddrs != "" {
		addrs := strings.Split(cfg.RedisSentinelAddrs, ",")
		return redis.NewUniversalClient(&redis.UniversalOptions{
			MasterName: cfg.RedisSentinelMaster,
			Addrs:      addrs,
		})
	}
	return redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs: []string{cfg.RedisAddr},
	})
}
