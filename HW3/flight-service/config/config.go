package config

import "github.com/ilyakaznacheev/cleanenv"

type Config struct {
	GRPCPort             string `env:"GRPC_PORT"              env-default:"50051"`
	FlightDBDSN          string `env:"FLIGHT_DB_DSN"          env-required:"true"`
	APIKey               string `env:"API_KEY"                env-required:"true"`
	RedisAddr            string `env:"REDIS_ADDR"             env-default:"redis-master:6379"`
	CacheTTLFlightSec    int    `env:"CACHE_TTL_FLIGHT_SEC"   env-default:"300"`
	CacheTTLSearchSec    int    `env:"CACHE_TTL_SEARCH_SEC"   env-default:"300"`
	RedisSentinelAddrs   string `env:"REDIS_SENTINEL_ADDRS"   env-default:""`
	RedisSentinelMaster  string `env:"REDIS_SENTINEL_MASTER"  env-default:"mymaster"`
}

func Load() (*Config, error) {
	cfg := &Config{}
	if err := cleanenv.ReadEnv(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
