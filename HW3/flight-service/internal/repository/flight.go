package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type FlightModel struct {
	ID             string    `db:"id"`
	FlightNumber   string    `db:"flight_number"`
	Airline        string    `db:"airline"`
	Origin         string    `db:"origin"`
	Destination    string    `db:"destination"`
	DepartureAt    time.Time `db:"departure_at"`
	ArrivalAt      time.Time `db:"arrival_at"`
	TotalSeats     int32     `db:"total_seats"`
	AvailableSeats int32     `db:"available_seats"`
	Price          float64   `db:"price"`
	Status         string    `db:"status"`
	CreatedAt      time.Time `db:"created_at"`
}

type ReservationModel struct {
	ID        string    `db:"id"`
	FlightID  string    `db:"flight_id"`
	BookingID string    `db:"booking_id"`
	SeatCount int32     `db:"seat_count"`
	Status    string    `db:"status"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

var ErrNotFound = errors.New("not found")
var ErrInsufficientSeats = errors.New("insufficient seats")

type Repository struct {
	db     *sqlx.DB
	logger *zap.Logger
}

func New(db *sqlx.DB, logger *zap.Logger) *Repository {
	return &Repository{db: db, logger: logger}
}

func (r *Repository) GetFlightByID(ctx context.Context, id string) (*FlightModel, error) {
	var f FlightModel
	err := r.db.GetContext(ctx, &f, `SELECT * FROM flights WHERE id = $1`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("GetFlightByID: %w", err)
	}
	return &f, nil
}

func (r *Repository) SearchFlights(ctx context.Context, origin, destination, date string) ([]*FlightModel, error) {
	query := `SELECT * FROM flights WHERE origin = $1 AND destination = $2`
	args := []interface{}{origin, destination}

	if date != "" {
		query += ` AND DATE(departure_at) = $3`
		args = append(args, date)
	}
	query += ` ORDER BY departure_at`

	var flights []*FlightModel
	if err := r.db.SelectContext(ctx, &flights, query, args...); err != nil {
		return nil, fmt.Errorf("SearchFlights: %w", err)
	}
	return flights, nil
}

func (r *Repository) ReserveSeatsInTx(ctx context.Context, flightID, bookingID string, seatCount int32) (string, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// FOR UPDATE предотвращает race condition при одновременном бронировании последнего места
	var flight FlightModel
	err = tx.GetContext(ctx, &flight, `SELECT * FROM flights WHERE id = $1 FOR UPDATE`, flightID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrNotFound
		}
		return "", fmt.Errorf("lock flight: %w", err)
	}

	// идемпотентность: повторный вызов с тем же booking_id возвращает существующую резервацию
	var existing ReservationModel
	idemErr := tx.GetContext(ctx, &existing,
		`SELECT * FROM seat_reservations WHERE booking_id = $1`, bookingID)
	if idemErr == nil {
		_ = tx.Rollback()
		return existing.ID, nil
	}
	if !errors.Is(idemErr, sql.ErrNoRows) {
		return "", fmt.Errorf("idempotency check: %w", idemErr)
	}

	if flight.AvailableSeats < seatCount {
		return "", ErrInsufficientSeats
	}

	_, err = tx.ExecContext(ctx,
		`UPDATE flights SET available_seats = available_seats - $1 WHERE id = $2`,
		seatCount, flightID)
	if err != nil {
		return "", fmt.Errorf("update seats: %w", err)
	}

	var reservationID string
	err = tx.QueryRowContext(ctx,
		`INSERT INTO seat_reservations (flight_id, booking_id, seat_count, status)
		 VALUES ($1, $2, $3, 'ACTIVE') RETURNING id`,
		flightID, bookingID, seatCount,
	).Scan(&reservationID)
	if err != nil {
		return "", fmt.Errorf("insert reservation: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return "", fmt.Errorf("commit: %w", err)
	}
	return reservationID, nil
}

func (r *Repository) ReleaseReservationInTx(ctx context.Context, bookingID string) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var res ReservationModel
	err = tx.GetContext(ctx, &res,
		`SELECT * FROM seat_reservations WHERE booking_id = $1 AND status = 'ACTIVE' FOR UPDATE`,
		bookingID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return fmt.Errorf("lock reservation: %w", err)
	}

	_, err = tx.ExecContext(ctx,
		`UPDATE flights SET available_seats = available_seats + $1 WHERE id = $2`,
		res.SeatCount, res.FlightID)
	if err != nil {
		return fmt.Errorf("release seats: %w", err)
	}

	_, err = tx.ExecContext(ctx,
		`UPDATE seat_reservations SET status = 'RELEASED', updated_at = NOW() WHERE booking_id = $1`,
		bookingID)
	if err != nil {
		return fmt.Errorf("update reservation status: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}
