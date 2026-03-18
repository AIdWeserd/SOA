package service

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	flightv1 "booking-service/gen/flight/v1"
	"booking-service/internal/repository"
)

type CreateBookingRequest struct {
	BookingID      string
	UserID         string
	FlightID       string
	PassengerName  string
	PassengerEmail string
	SeatCount      int32
}

type BookingService struct {
	repo         *repository.Repository
	flightClient flightv1.FlightServiceClient
	logger       *zap.Logger
}

func New(repo *repository.Repository, flightClient flightv1.FlightServiceClient, logger *zap.Logger) *BookingService {
	return &BookingService{
		repo:         repo,
		flightClient: flightClient,
		logger:       logger,
	}
}

func (s *BookingService) CreateBooking(ctx context.Context, req *CreateBookingRequest) (*repository.BookingModel, error) {
	flightResp, err := s.flightClient.GetFlight(ctx, &flightv1.GetFlightRequest{
		FlightId: req.FlightID,
	})
	if err != nil {
		st, _ := status.FromError(err)
		if st.Code() == codes.NotFound {
			return nil, fmt.Errorf("flight not found: %w", ErrFlightNotFound)
		}
		return nil, fmt.Errorf("get flight: %w", err)
	}

	flight := flightResp.Flight

	// BookingID — ключ идемпотентности: при retry ReserveSeats не дублирует резервацию
	_, err = s.flightClient.ReserveSeats(ctx, &flightv1.ReserveSeatsRequest{
		FlightId:  req.FlightID,
		BookingId: req.BookingID,
		SeatCount: req.SeatCount,
	})
	if err != nil {
		st, _ := status.FromError(err)
		if st.Code() == codes.ResourceExhausted {
			return nil, fmt.Errorf("no seats available: %w", ErrNoSeatsAvailable)
		}
		return nil, fmt.Errorf("reserve seats: %w", err)
	}

	// цена фиксируется на момент бронирования (snapshot)
	totalPrice := float64(req.SeatCount) * flight.Price
	booking := &repository.BookingModel{
		UserID:         req.UserID,
		FlightID:       req.FlightID,
		PassengerName:  req.PassengerName,
		PassengerEmail: req.PassengerEmail,
		SeatCount:      req.SeatCount,
		TotalPrice:     totalPrice,
		Status:         "CONFIRMED",
	}

	created, err := s.repo.CreateBooking(ctx, booking)
	if err != nil {
		s.logger.Error("failed to persist booking", zap.String("booking_id", req.BookingID), zap.Error(err))
		// компенсирующий вызов — вернуть места если INSERT упал
		if _, relErr := s.flightClient.ReleaseReservation(ctx, &flightv1.ReleaseReservationRequest{
			BookingId: req.BookingID,
		}); relErr != nil {
			s.logger.Warn("failed to release reservation after booking DB error",
				zap.String("booking_id", req.BookingID),
				zap.Error(relErr),
			)
		}
		return nil, fmt.Errorf("create booking: %w", err)
	}
	return created, nil
}

func (s *BookingService) GetBooking(ctx context.Context, id string) (*repository.BookingModel, error) {
	b, err := s.repo.GetBookingByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrBookingNotFound
		}
		return nil, err
	}
	return b, nil
}

func (s *BookingService) ListBookings(ctx context.Context, userID string) ([]*repository.BookingModel, error) {
	if userID != "" {
		return s.repo.ListBookingsByUserID(ctx, userID)
	}
	return s.repo.ListBookings(ctx)
}

func (s *BookingService) CancelBooking(ctx context.Context, id string) error {
	b, err := s.repo.GetBookingByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrBookingNotFound
		}
		return err
	}

	if _, relErr := s.flightClient.ReleaseReservation(ctx, &flightv1.ReleaseReservationRequest{
		BookingId: id,
	}); relErr != nil {
		s.logger.Warn("release reservation on cancel",
			zap.String("booking_id", b.ID),
			zap.Error(relErr),
		)
	}

	if err := s.repo.CancelBooking(ctx, id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrBookingNotFound
		}
		return err
	}
	return nil
}

var (
	ErrFlightNotFound   = errors.New("flight not found")
	ErrNoSeatsAvailable = errors.New("no seats available")
	ErrBookingNotFound  = errors.New("booking not found")
)
