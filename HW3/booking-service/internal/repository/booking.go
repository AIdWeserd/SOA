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

type BookingModel struct {
	ID             string    `db:"id"`
	UserID         string    `db:"user_id"`
	FlightID       string    `db:"flight_id"`
	PassengerName  string    `db:"passenger_name"`
	PassengerEmail string    `db:"passenger_email"`
	SeatCount      int32     `db:"seat_count"`
	TotalPrice     float64   `db:"total_price"`
	Status         string    `db:"status"`
	CreatedAt      time.Time `db:"created_at"`
	UpdatedAt      time.Time `db:"updated_at"`
}

var ErrNotFound = errors.New("not found")

type Repository struct {
	db     *sqlx.DB
	logger *zap.Logger
}

func New(db *sqlx.DB, logger *zap.Logger) *Repository {
	return &Repository{db: db, logger: logger}
}

func (r *Repository) CreateBooking(ctx context.Context, b *BookingModel) (*BookingModel, error) {
	query := `
		INSERT INTO bookings (user_id, flight_id, passenger_name, passenger_email, seat_count, total_price, status)
		VALUES (:user_id, :flight_id, :passenger_name, :passenger_email, :seat_count, :total_price, :status)
		RETURNING id, created_at, updated_at`

	rows, err := r.db.NamedQueryContext(ctx, query, b)
	if err != nil {
		return nil, fmt.Errorf("insert booking: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.Scan(&b.ID, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan returning: %w", err)
		}
	}
	return b, nil
}

func (r *Repository) GetBookingByID(ctx context.Context, id string) (*BookingModel, error) {
	var b BookingModel
	err := r.db.GetContext(ctx, &b, `SELECT * FROM bookings WHERE id = $1`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get booking: %w", err)
	}
	return &b, nil
}

func (r *Repository) ListBookingsByUserID(ctx context.Context, userID string) ([]*BookingModel, error) {
	var bookings []*BookingModel
	err := r.db.SelectContext(ctx, &bookings,
		`SELECT * FROM bookings WHERE user_id = $1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("list bookings: %w", err)
	}
	return bookings, nil
}

func (r *Repository) ListBookings(ctx context.Context) ([]*BookingModel, error) {
	var bookings []*BookingModel
	err := r.db.SelectContext(ctx, &bookings, `SELECT * FROM bookings ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("list all bookings: %w", err)
	}
	return bookings, nil
}

func (r *Repository) CancelBooking(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE bookings SET status = 'CANCELLED', updated_at = NOW() WHERE id = $1 AND status = 'CONFIRMED'`, id)
	if err != nil {
		return fmt.Errorf("cancel booking: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
