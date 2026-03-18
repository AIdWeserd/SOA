package grpcclient

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type State int

const (
	StateClosed   State = iota
	StateOpen
	StateHalfOpen
)

func (s State) String() string {
	switch s {
	case StateClosed:
		return "CLOSED"
	case StateOpen:
		return "OPEN"
	case StateHalfOpen:
		return "HALF_OPEN"
	default:
		return "UNKNOWN"
	}
}

type CircuitBreakerConfig struct {
	FailureThreshold int
	TimeoutSec       int
	HalfOpenMax      int
}

type CircuitBreaker struct {
	mu               sync.Mutex
	state            State
	failures         int
	lastFailureAt    time.Time
	halfOpenAttempts int
	cfg              CircuitBreakerConfig
	logger           *zap.Logger
}

func NewCircuitBreaker(cfg CircuitBreakerConfig, logger *zap.Logger) *CircuitBreaker {
	return &CircuitBreaker{
		state:  StateClosed,
		cfg:    cfg,
		logger: logger,
	}
}

func (cb *CircuitBreaker) Interceptor() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		if err := cb.allow(); err != nil {
			return err
		}
		err := invoker(ctx, method, req, reply, cc, opts...)
		cb.record(err)
		return err
	}
}

func (cb *CircuitBreaker) allow() error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		return nil
	case StateOpen:
		if time.Since(cb.lastFailureAt) >= time.Duration(cb.cfg.TimeoutSec)*time.Second {
			cb.transition(StateHalfOpen)
			cb.halfOpenAttempts = 0
			return nil
		}
		return status.Error(codes.Unavailable, "circuit breaker is OPEN")
	case StateHalfOpen:
		if cb.halfOpenAttempts < cb.cfg.HalfOpenMax {
			cb.halfOpenAttempts++
			return nil
		}
		return status.Error(codes.Unavailable, "circuit breaker is HALF_OPEN: max probe requests reached")
	}
	return nil
}

func (cb *CircuitBreaker) record(err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err == nil {
		switch cb.state {
		case StateHalfOpen:
			cb.transition(StateClosed)
			cb.failures = 0
		case StateClosed:
			cb.failures = 0
		}
		return
	}

	st, ok := status.FromError(err)
	if !ok {
		return
	}
	// only infrastructure-level errors count toward the threshold
	switch st.Code() {
	case codes.Unavailable, codes.DeadlineExceeded, codes.Internal:
	default:
		return
	}

	switch cb.state {
	case StateClosed:
		cb.failures++
		cb.lastFailureAt = time.Now()
		if cb.failures >= cb.cfg.FailureThreshold {
			cb.transition(StateOpen)
		}
	case StateHalfOpen:
		cb.lastFailureAt = time.Now()
		cb.transition(StateOpen)
	}
}

func (cb *CircuitBreaker) transition(next State) {
	prev := cb.state
	cb.state = next
	cb.logger.Info("circuit breaker state transition",
		zap.String("from", prev.String()),
		zap.String("to", next.String()),
		zap.Int("failures", cb.failures),
	)
}

var ErrCircuitOpen = fmt.Errorf("circuit breaker open")
