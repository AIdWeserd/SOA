package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"booking-service/internal/repository"
	"booking-service/internal/service"
)

type BookingHandler struct {
	svc    *service.BookingService
	logger *zap.Logger
}

func NewBookingHandler(svc *service.BookingService, logger *zap.Logger) *BookingHandler {
	return &BookingHandler{svc: svc, logger: logger}
}

func (h *BookingHandler) RegisterRoutes(r *gin.RouterGroup) {
	r.POST("/bookings", h.CreateBooking)
	r.GET("/bookings", h.ListBookings)
	r.GET("/bookings/:id", h.GetBooking)
	r.POST("/bookings/:id/cancel", h.CancelBooking)
}

type createBookingRequest struct {
	UserID         string `json:"user_id"         binding:"required"`
	FlightID       string `json:"flight_id"       binding:"required"`
	PassengerName  string `json:"passenger_name"  binding:"required"`
	PassengerEmail string `json:"passenger_email" binding:"required"`
	SeatCount      int32  `json:"seat_count"      binding:"required,min=1"`
}

func (h *BookingHandler) CreateBooking(c *gin.Context) {
	var req createBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// booking_id генерируется здесь — он же служит ключом идемпотентности для ReserveSeats
	bookingID := uuid.New().String()

	created, err := h.svc.CreateBooking(c.Request.Context(), &service.CreateBookingRequest{
		BookingID:      bookingID,
		UserID:         req.UserID,
		FlightID:       req.FlightID,
		PassengerName:  req.PassengerName,
		PassengerEmail: req.PassengerEmail,
		SeatCount:      req.SeatCount,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrFlightNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "flight not found"})
		case errors.Is(err, service.ErrNoSeatsAvailable):
			c.JSON(http.StatusConflict, gin.H{"error": "no seats available"})
		default:
			h.logger.Error("CreateBooking", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		}
		return
	}

	c.JSON(http.StatusCreated, toBookingResponse(created))
}

func (h *BookingHandler) GetBooking(c *gin.Context) {
	id := c.Param("id")
	b, err := h.svc.GetBooking(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrBookingNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "booking not found"})
			return
		}
		h.logger.Error("GetBooking", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, toBookingResponse(b))
}

func (h *BookingHandler) ListBookings(c *gin.Context) {
	userID := c.Query("user_id")
	bookings, err := h.svc.ListBookings(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("ListBookings", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	resp := make([]bookingResponse, 0, len(bookings))
	for _, b := range bookings {
		resp = append(resp, toBookingResponse(b))
	}
	c.JSON(http.StatusOK, resp)
}

func (h *BookingHandler) CancelBooking(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.CancelBooking(c.Request.Context(), id); err != nil {
		if errors.Is(err, service.ErrBookingNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "booking not found"})
			return
		}
		h.logger.Error("CancelBooking", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "booking cancelled"})
}

type bookingResponse struct {
	ID             string  `json:"id"`
	UserID         string  `json:"user_id"`
	FlightID       string  `json:"flight_id"`
	PassengerName  string  `json:"passenger_name"`
	PassengerEmail string  `json:"passenger_email"`
	SeatCount      int32   `json:"seat_count"`
	TotalPrice     float64 `json:"total_price"`
	Status         string  `json:"status"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
}

func toBookingResponse(b *repository.BookingModel) bookingResponse {
	return bookingResponse{
		ID:             b.ID,
		UserID:         b.UserID,
		FlightID:       b.FlightID,
		PassengerName:  b.PassengerName,
		PassengerEmail: b.PassengerEmail,
		SeatCount:      b.SeatCount,
		TotalPrice:     b.TotalPrice,
		Status:         b.Status,
		CreatedAt:      b.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:      b.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
