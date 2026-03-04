package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	api "hw2/.build/openapi"
	db "hw2/.build/sqlc"
	"hw2/src/apierr"
	"hw2/src/repository"
	"hw2/src/service/lib"
)

type Service struct {
	repo *repository.Repository
}

func New(repo *repository.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateProduct(ctx context.Context, req *api.ProductCreate) (api.ProductResponse, error) {
	if err := lib.ValidateProductCreate(req); err != nil {
		return api.ProductResponse{}, err
	}
	name, description, price, stock, category, status := lib.CreateParamsFromAPI(req)

	if price.LessThanOrEqual(decimal.Zero) {
		return api.ProductResponse{}, apierr.ValidationError([]apierr.ValidationDetail{
			{Field: "price", Message: "Цена должна быть больше 0"},
		})
	}
	

	product, err := s.repo.CreateProduct(ctx, repository.CreateProductParams{
		Name:        name,
		Description: description,
		Price:       price,
		Stock:       stock,
		Category:    category,
		Status:      status,
	})
	if err != nil {
		return api.ProductResponse{}, fmt.Errorf("service.CreateProduct: %w", err)
	}

	return lib.DBProductToAPI(product), nil
}

func (s *Service) GetProductByID(ctx context.Context, id uuid.UUID) (api.ProductResponse, error) {
	product, err := s.repo.GetProductByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return api.ProductResponse{}, apierr.ProductNotFound()
		}
		return api.ProductResponse{}, fmt.Errorf("service.GetProductByID: %w", err)
	}

	return lib.DBProductToAPI(product), nil
}

func (s *Service) ListProducts(ctx context.Context, params api.ListProductsParams) (api.ProductPage, error) {
	page := int32(params.Page.Or(0))
	size := int32(params.Size.Or(20))
	offset := page * size

	var statusPtr *db.ProductStatus
	if params.Status.IsSet() {
		v := lib.ProductStatusFromAPI(params.Status.Value)
		statusPtr = &v
	}

	var categoryPtr *string
	if params.Category.IsSet() {
		v := params.Category.Value
		categoryPtr = &v
	}

	products, err := s.repo.ListProducts(ctx, repository.ListProductsParams{
		Status:   statusPtr,
		Category: categoryPtr,
		Limit:    size,
		Offset:   offset,
	})
	if err != nil {
		return api.ProductPage{}, fmt.Errorf("service.ListProducts: %w", err)
	}

	total, err := s.repo.CountProducts(ctx, statusPtr, categoryPtr)
	if err != nil {
		return api.ProductPage{}, fmt.Errorf("service.ListProducts count: %w", err)
	}

	content := make([]api.ProductResponse, 0, len(products))
	for _, p := range products {
		content = append(content, lib.DBProductToAPI(p))
	}

	return api.ProductPage{
		Content:       content,
		TotalElements: int(total),
		Page:          int(page),
		Size:          int(size),
	}, nil
}

func (s *Service) UpdateProduct(ctx context.Context, id uuid.UUID, req *api.ProductUpdate) (api.ProductResponse, error) {
	if err := lib.ValidateProductUpdate(req); err != nil {
		return api.ProductResponse{}, err
	}

	existing, err := s.repo.GetProductByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return api.ProductResponse{}, apierr.ProductNotFound()
		}
		return api.ProductResponse{}, fmt.Errorf("service.UpdateProduct: %w", err)
	}

	name := existing.Name
	if req.Name.IsSet() {
		name = req.Name.Value
	}

	description := existing.Description
	if req.Description.IsSet() {
		description = lib.OptStringToPtr(req.Description)
	}

	price := existing.Price
	if req.Price.IsSet() {
		price = req.Price.Value
		if price.LessThanOrEqual(decimal.Zero) {
			return api.ProductResponse{}, apierr.ValidationError([]apierr.ValidationDetail{
				{Field: "price", Message: "Цена должна быть больше 0"},
		})
    }
}

	stock := existing.Stock
	if req.Stock.IsSet() {
		stock = int32(req.Stock.Value)
	}

	category := existing.Category
	if req.Category.IsSet() {
		category = req.Category.Value
	}

	status := existing.Status
	if req.Status.IsSet() {
		status = lib.ProductStatusFromAPI(req.Status.Value)
	}

	product, err := s.repo.UpdateProduct(ctx, repository.UpdateProductParams{
		ID:          id,
		Name:        name,
		Description: description,
		Price:       price,
		Stock:       stock,
		Category:    category,
		Status:      status,
	})
	if err != nil {
		return api.ProductResponse{}, fmt.Errorf("service.UpdateProduct: %w", err)
	}

	return lib.DBProductToAPI(product), nil
}

func (s *Service) DeleteProduct(ctx context.Context, id uuid.UUID) error {
	_, err := s.repo.ArchiveProduct(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return apierr.ProductNotFound()
		}
		return fmt.Errorf("service.DeleteProduct: %w", err)
	}

	return nil
}
