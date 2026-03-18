package handler

import (
	"context"
	"errors"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	flightv1 "flight-service/gen/flight/v1"
	"flight-service/internal/repository"
	"flight-service/internal/service"
)

type FlightHandler struct {
	flightv1.UnimplementedFlightServiceServer
	svc    *service.FlightService
	logger *zap.Logger
}

func NewFlightHandler(svc *service.FlightService, logger *zap.Logger) *FlightHandler {
	return &FlightHandler{svc: svc, logger: logger}
}

func (h *FlightHandler) SearchFlights(ctx context.Context, req *flightv1.SearchFlightsRequest) (*flightv1.SearchFlightsResponse, error) {
	if req.Origin == "" || req.Destination == "" {
		return nil, status.Error(codes.InvalidArgument, "origin and destination are required")
	}

	flights, err := h.svc.SearchFlights(ctx, req.Origin, req.Destination, req.Date)
	if err != nil {
		h.logger.Error("SearchFlights error", zap.Error(err))
		return nil, status.Error(codes.Internal, "internal error")
	}

	resp := &flightv1.SearchFlightsResponse{
		Flights: make([]*flightv1.Flight, 0, len(flights)),
	}
	for _, f := range flights {
		resp.Flights = append(resp.Flights, toProtoFlight(f))
	}
	return resp, nil
}

func (h *FlightHandler) GetFlight(ctx context.Context, req *flightv1.GetFlightRequest) (*flightv1.GetFlightResponse, error) {
	if req.FlightId == "" {
		return nil, status.Error(codes.InvalidArgument, "flight_id is required")
	}

	f, err := h.svc.GetFlight(ctx, req.FlightId)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "flight %s not found", req.FlightId)
		}
		h.logger.Error("GetFlight error", zap.String("id", req.FlightId), zap.Error(err))
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &flightv1.GetFlightResponse{Flight: toProtoFlight(f)}, nil
}

func (h *FlightHandler) ReserveSeats(ctx context.Context, req *flightv1.ReserveSeatsRequest) (*flightv1.ReserveSeatsResponse, error) {
	if req.FlightId == "" || req.BookingId == "" {
		return nil, status.Error(codes.InvalidArgument, "flight_id and booking_id are required")
	}
	if req.SeatCount <= 0 {
		return nil, status.Error(codes.InvalidArgument, "seat_count must be positive")
	}

	reservationID, err := h.svc.ReserveSeats(ctx, req.FlightId, req.BookingId, req.SeatCount)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "flight %s not found", req.FlightId)
		}
		if errors.Is(err, repository.ErrInsufficientSeats) {
			return nil, status.Error(codes.ResourceExhausted, "insufficient seats available")
		}
		h.logger.Error("ReserveSeats error", zap.Error(err))
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &flightv1.ReserveSeatsResponse{ReservationId: reservationID}, nil
}

func (h *FlightHandler) ReleaseReservation(ctx context.Context, req *flightv1.ReleaseReservationRequest) (*flightv1.ReleaseReservationResponse, error) {
	if req.BookingId == "" {
		return nil, status.Error(codes.InvalidArgument, "booking_id is required")
	}

	if err := h.svc.ReleaseReservation(ctx, req.BookingId); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "active reservation for booking %s not found", req.BookingId)
		}
		h.logger.Error("ReleaseReservation error", zap.Error(err))
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &flightv1.ReleaseReservationResponse{Success: true}, nil
}

func toProtoFlight(f *repository.FlightModel) *flightv1.Flight {
	statusMap := map[string]flightv1.FlightStatus{
		"SCHEDULED": flightv1.FlightStatus_FLIGHT_STATUS_SCHEDULED,
		"DEPARTED":  flightv1.FlightStatus_FLIGHT_STATUS_DEPARTED,
		"CANCELLED": flightv1.FlightStatus_FLIGHT_STATUS_CANCELLED,
		"COMPLETED": flightv1.FlightStatus_FLIGHT_STATUS_COMPLETED,
	}
	s, ok := statusMap[f.Status]
	if !ok {
		s = flightv1.FlightStatus_FLIGHT_STATUS_UNSPECIFIED
	}
	return &flightv1.Flight{
		Id:             f.ID,
		FlightNumber:   f.FlightNumber,
		Airline:        f.Airline,
		Origin:         f.Origin,
		Destination:    f.Destination,
		DepartureAt:    timestamppb.New(f.DepartureAt),
		ArrivalAt:      timestamppb.New(f.ArrivalAt),
		TotalSeats:     f.TotalSeats,
		AvailableSeats: f.AvailableSeats,
		Price:          f.Price,
		Status:         s,
	}
}
