package grpcclient

import (
	"context"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type RetryConfig struct {
	MaxRetries int
	BaseMs     int
}

func noRetryCode(c codes.Code) bool {
	switch c {
	case codes.InvalidArgument,
		codes.NotFound,
		codes.ResourceExhausted,
		codes.Unauthenticated,
		codes.PermissionDenied,
		codes.AlreadyExists,
		codes.Unimplemented:
		return true
	}
	return false
}

func NewRetryInterceptor(cfg RetryConfig, logger *zap.Logger) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		var lastErr error
		backoff := time.Duration(cfg.BaseMs) * time.Millisecond

		for attempt := 0; attempt <= cfg.MaxRetries; attempt++ {
			if attempt > 0 {
				logger.Info("retrying gRPC call",
					zap.String("method", method),
					zap.Int("attempt", attempt),
					zap.Duration("backoff", backoff),
					zap.Error(lastErr),
				)
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(backoff):
				}
				backoff *= 2
			}

			err := invoker(ctx, method, req, reply, cc, opts...)
			if err == nil {
				return nil
			}
			lastErr = err

			st, ok := status.FromError(err)
			if !ok {
				return err
			}
			if noRetryCode(st.Code()) {
				return err
			}
			// only retry on UNAVAILABLE / DEADLINE_EXCEEDED
			if st.Code() != codes.Unavailable && st.Code() != codes.DeadlineExceeded {
				return err
			}
		}

		logger.Warn("gRPC call failed after retries",
			zap.String("method", method),
			zap.Int("max_retries", cfg.MaxRetries),
			zap.Error(lastErr),
		)
		return lastErr
	}
}
