package config

import "github.com/ilyakaznacheev/cleanenv"

type Config struct {
	HTTPPort             string `env:"HTTP_PORT"               env-default:"8080"`
	BookingDBDSN         string `env:"BOOKING_DB_DSN"          env-required:"true"`
	FlightServiceAddr    string `env:"FLIGHT_SERVICE_ADDR"     env-default:"flight-service:50051"`
	FlightServiceAPIKey  string `env:"FLIGHT_SERVICE_API_KEY"  env-required:"true"`
	GRPCMaxRetries       int    `env:"GRPC_MAX_RETRIES"        env-default:"3"`
	GRPCRetryBaseMs      int    `env:"GRPC_RETRY_BASE_MS"      env-default:"100"`
	CBFailureThreshold   int    `env:"CB_FAILURE_THRESHOLD"    env-default:"5"`
	CBTimeoutSec         int    `env:"CB_TIMEOUT_SEC"          env-default:"30"`
	CBHalfOpenMax        int    `env:"CB_HALF_OPEN_MAX"        env-default:"1"`
}

func Load() (*Config, error) {
	cfg := &Config{}
	if err := cleanenv.ReadEnv(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
