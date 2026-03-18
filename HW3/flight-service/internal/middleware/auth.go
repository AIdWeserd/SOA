package middleware

import (
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const apiKeyHeader = "x-api-key"

func UnaryServerAuthInterceptor(expectedKey string, logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			logger.Warn("missing metadata", zap.String("method", info.FullMethod))
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		values := md.Get(apiKeyHeader)
		if len(values) == 0 || values[0] == "" {
			logger.Warn("missing api key", zap.String("method", info.FullMethod))
			return nil, status.Error(codes.Unauthenticated, "missing api key")
		}

		if values[0] != expectedKey {
			logger.Warn("invalid api key", zap.String("method", info.FullMethod))
			return nil, status.Error(codes.Unauthenticated, "invalid api key")
		}

		return handler(ctx, req)
	}
}
