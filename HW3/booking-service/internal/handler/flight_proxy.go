package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	flightv1 "booking-service/gen/flight/v1"
)

type FlightProxyHandler struct {
	flightClient flightv1.FlightServiceClient
	logger       *zap.Logger
}

func NewFlightProxyHandler(client flightv1.FlightServiceClient, logger *zap.Logger) *FlightProxyHandler {
	return &FlightProxyHandler{flightClient: client, logger: logger}
}

func (h *FlightProxyHandler) RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/flights", h.SearchFlights)
	r.GET("/flights/:id", h.GetFlight)
}

func (h *FlightProxyHandler) SearchFlights(c *gin.Context) {
	origin := c.Query("origin")
	destination := c.Query("destination")
	date := c.Query("date")

	if origin == "" || destination == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "origin and destination are required"})
		return
	}

	resp, err := h.flightClient.SearchFlights(c.Request.Context(), &flightv1.SearchFlightsRequest{
		Origin:      origin,
		Destination: destination,
		Date:        date,
	})
	if err != nil {
		h.logger.Error("SearchFlights proxy error", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to search flights"})
		return
	}

	flights := make([]flightJSON, 0, len(resp.Flights))
	for _, f := range resp.Flights {
		flights = append(flights, toFlightJSON(f))
	}
	c.JSON(http.StatusOK, flights)
}

func (h *FlightProxyHandler) GetFlight(c *gin.Context) {
	id := c.Param("id")

	resp, err := h.flightClient.GetFlight(c.Request.Context(), &flightv1.GetFlightRequest{
		FlightId: id,
	})
	if err != nil {
		st, _ := status.FromError(err)
		if st.Code() == codes.NotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "flight not found"})
			return
		}
		h.logger.Error("GetFlight proxy error", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get flight"})
		return
	}

	c.JSON(http.StatusOK, toFlightJSON(resp.Flight))
}

type flightJSON struct {
	ID             string  `json:"id"`
	FlightNumber   string  `json:"flight_number"`
	Airline        string  `json:"airline"`
	Origin         string  `json:"origin"`
	Destination    string  `json:"destination"`
	DepartureAt    string  `json:"departure_at"`
	ArrivalAt      string  `json:"arrival_at"`
	TotalSeats     int32   `json:"total_seats"`
	AvailableSeats int32   `json:"available_seats"`
	Price          float64 `json:"price"`
	Status         string  `json:"status"`
}

func toFlightJSON(f *flightv1.Flight) flightJSON {
	var dep, arr string
	if f.DepartureAt != nil {
		dep = f.DepartureAt.AsTime().Format("2006-01-02T15:04:05Z07:00")
	}
	if f.ArrivalAt != nil {
		arr = f.ArrivalAt.AsTime().Format("2006-01-02T15:04:05Z07:00")
	}
	return flightJSON{
		ID:             f.Id,
		FlightNumber:   f.FlightNumber,
		Airline:        f.Airline,
		Origin:         f.Origin,
		Destination:    f.Destination,
		DepartureAt:    dep,
		ArrivalAt:      arr,
		TotalSeats:     f.TotalSeats,
		AvailableSeats: f.AvailableSeats,
		Price:          f.Price,
		Status:         f.Status.String(),
	}
}
