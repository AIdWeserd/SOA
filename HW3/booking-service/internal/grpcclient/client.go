package grpcclient

import (
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	flightv1 "booking-service/gen/flight/v1"
	"booking-service/internal/middleware"
)

// chain order: CB → Retry → Auth
func NewFlightServiceClient(
	addr string,
	apiKey string,
	retryCfg RetryConfig,
	cbCfg CircuitBreakerConfig,
	logger *zap.Logger,
) (flightv1.FlightServiceClient, *grpc.ClientConn, error) {
	cb := NewCircuitBreaker(cbCfg, logger)
	retryInterceptor := NewRetryInterceptor(retryCfg, logger)
	authInterceptor := middleware.UnaryClientAuthInterceptor(apiKey)

	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(
			cb.Interceptor(),
			retryInterceptor,
			authInterceptor,
		),
	)
	if err != nil {
		return nil, nil, err
	}

	client := flightv1.NewFlightServiceClient(conn)
	return client, conn, nil
}
