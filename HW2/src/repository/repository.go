package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	db "hw2/.build/sqlc"
)

type Repository struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{
		pool:    pool,
		queries: db.New(pool),
	}
}

type CreateProductParams struct {
	Name        string
	Description *string
	Price       decimal.Decimal
	Stock       int32
	Category    string
	Status      db.ProductStatus
}

type UpdateProductParams struct {
	ID          uuid.UUID
	Name        string
	Description *string
	Price       decimal.Decimal
	Stock       int32
	Category    string
	Status      db.ProductStatus
}

type ListProductsParams struct {
	Status   *db.ProductStatus
	Category *string
	Limit    int32
	Offset   int32
}

var ErrNotFound = errors.New("product not found")

func (r *Repository) CreateProduct(ctx context.Context, p CreateProductParams) (db.Product, error) {
	product, err := r.queries.CreateProduct(ctx, db.CreateProductParams{
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
		Stock:       p.Stock,
		Category:    p.Category,
		Status:      p.Status,
	})
	if err != nil {
		return db.Product{}, fmt.Errorf("repository.CreateProduct: %w", err)
	}
	return product, nil
}

func (r *Repository) GetProductByID(ctx context.Context, id uuid.UUID) (db.Product, error) {
	product, err := r.queries.GetProductByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Product{}, ErrNotFound
		}
		return db.Product{}, fmt.Errorf("repository.GetProductByID: %w", err)
	}
	return product, nil
}

func (r *Repository) ListProducts(ctx context.Context, p ListProductsParams) ([]db.Product, error) {
	productStatus := db.NullProductStatus{
		Valid: false,
	}
	if p.Status != nil {
		productStatus.ProductStatus = *p.Status
		productStatus.Valid = true
	}

	products, err := r.queries.GetProducts(ctx, db.GetProductsParams{
		Status:   productStatus,
		Category: p.Category,
		Limit:    p.Limit,
		Offset:   p.Offset,
	})
	if err != nil {
		return nil, fmt.Errorf("repository.ListProducts: %w", err)
	}
	return products, nil
}

func (r *Repository) CountProducts(ctx context.Context, status *db.ProductStatus, category *string) (int64, error) {
	productStatus := db.NullProductStatus{
		Valid: false,
	}
	if status != nil {
		productStatus.ProductStatus = *status
		productStatus.Valid = true
	}

	count, err := r.queries.CountProducts(ctx, db.CountProductsParams{
		Status:   productStatus,
		Category: category,
	})
	if err != nil {
		return 0, fmt.Errorf("repository.CountProducts: %w", err)
	}
	return count, nil
}

func (r *Repository) UpdateProduct(ctx context.Context, p UpdateProductParams) (db.Product, error) {
	product, err := r.queries.UpdateProduct(ctx, db.UpdateProductParams{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
		Stock:       p.Stock,
		Category:    p.Category,
		Status:      p.Status,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Product{}, ErrNotFound
		}
		return db.Product{}, fmt.Errorf("repository.UpdateProduct: %w", err)
	}
	return product, nil
}

func (r *Repository) ArchiveProduct(ctx context.Context, id uuid.UUID) (db.Product, error) {
	product, err := r.queries.ArchiveProduct(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Product{}, ErrNotFound
		}
		return db.Product{}, fmt.Errorf("repository.ArchiveProduct: %w", err)
	}
	return product, nil
}
