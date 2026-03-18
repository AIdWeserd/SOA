package service

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"

	"flight-service/internal/cache"
	"flight-service/internal/repository"
)

type FlightService struct {
	repo   *repository.Repository
	cache  *cache.FlightCache
	logger *zap.Logger
}

func New(repo *repository.Repository, c *cache.FlightCache, logger *zap.Logger) *FlightService {
	return &FlightService{repo: repo, cache: c, logger: logger}
}

func (s *FlightService) GetFlight(ctx context.Context, id string) (*repository.FlightModel, error) {
	f, err := s.cache.GetFlight(ctx, id)
	if err == nil {
		return f, nil
	}
	if !errors.Is(err, cache.ErrCacheMiss) {
		s.logger.Warn("cache get error", zap.String("id", id), zap.Error(err))
	}

	f, err = s.repo.GetFlightByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if cacheErr := s.cache.SetFlight(ctx, f); cacheErr != nil {
		s.logger.Warn("cache set error", zap.String("id", id), zap.Error(cacheErr))
	}
	return f, nil
}

func (s *FlightService) SearchFlights(ctx context.Context, origin, destination, date string) ([]*repository.FlightModel, error) {
	flights, err := s.cache.GetSearch(ctx, origin, destination, date)
	if err == nil {
		return flights, nil
	}
	if !errors.Is(err, cache.ErrCacheMiss) {
		s.logger.Warn("cache get error", zap.String("origin", origin), zap.Error(err))
	}

	flights, err = s.repo.SearchFlights(ctx, origin, destination, date)
	if err != nil {
		return nil, err
	}

	if cacheErr := s.cache.SetSearch(ctx, origin, destination, date, flights); cacheErr != nil {
		s.logger.Warn("cache set error", zap.String("origin", origin), zap.Error(cacheErr))
	}
	return flights, nil
}

func (s *FlightService) ReserveSeats(ctx context.Context, flightID, bookingID string, seatCount int32) (string, error) {
	reservationID, err := s.repo.ReserveSeatsInTx(ctx, flightID, bookingID, seatCount)
	if err != nil {
		return "", err
	}

	if delErr := s.cache.DeleteFlight(ctx, flightID); delErr != nil {
		s.logger.Warn("cache delete flight error", zap.String("id", flightID), zap.Error(delErr))
	}

	f, dbErr := s.repo.GetFlightByID(ctx, flightID)
	if dbErr == nil {
		if delErr := s.cache.DeleteSearchByFlight(ctx, f.Origin, f.Destination); delErr != nil {
			s.logger.Warn("cache delete search error", zap.Error(delErr))
		}
	}

	return reservationID, nil
}

func (s *FlightService) ReleaseReservation(ctx context.Context, bookingID string) error {
	if err := s.repo.ReleaseReservationInTx(ctx, bookingID); err != nil {
		return fmt.Errorf("release reservation: %w", err)
	}
	return nil
}
