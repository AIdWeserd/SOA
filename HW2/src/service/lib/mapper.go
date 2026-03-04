package lib

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	api "hw2/.build/openapi"
	db "hw2/.build/sqlc"
)

func ProductStatusFromAPI(s api.ProductStatus) db.ProductStatus {
	switch s {
	case api.ProductStatusACTIVE:
		return db.ProductStatusACTIVE
	case api.ProductStatusINACTIVE:
		return db.ProductStatusINACTIVE
	case api.ProductStatusARCHIVED:
		return db.ProductStatusARCHIVED
	default:
		return db.ProductStatusACTIVE
	}
}

func ProductStatusToAPI(s db.ProductStatus) api.ProductStatus {
	switch s {
	case db.ProductStatusACTIVE:
		return api.ProductStatusACTIVE
	case db.ProductStatusINACTIVE:
		return api.ProductStatusINACTIVE
	case db.ProductStatusARCHIVED:
		return api.ProductStatusARCHIVED
	default:
		return api.ProductStatusACTIVE
	}
}

func OptStringToPtr(s api.OptNilString) *string {
	if s.IsNull() || !s.IsSet() {
		return nil
	}
	v := s.Value
	return &v
}

func PtrToOptNilString(s *string) api.OptNilString {
	if s == nil {
		return api.OptNilString{}
	}
	return api.NewOptNilString(*s)
}

func DBProductToAPI(p db.Product) api.ProductResponse {
	return api.ProductResponse{
		ID:          p.ID,
		Name:        p.Name,
		Description: PtrToOptNilString(p.Description),
		Price:       p.Price,
		Stock:       int(p.Stock),
		Category:    p.Category,
		Status:      ProductStatusToAPI(p.Status),
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

func CreateParamsFromAPI(req *api.ProductCreate) (name string, description *string, price decimal.Decimal, stock int32, category string, status db.ProductStatus) {
	name = req.Name
	description = OptStringToPtr(req.Description)
	price = req.Price
	stock = int32(req.Stock)
	category = req.Category
	status = ProductStatusFromAPI(req.Status)
	return
}

func UpdateParamsFromAPI(id uuid.UUID, req *api.ProductUpdate) (uuid.UUID, string, *string, decimal.Decimal, int32, string, db.ProductStatus) {
	description := OptStringToPtr(req.Description)
	price := req.Price.Value
	stock := int32(req.Stock.Value)
	status := ProductStatusFromAPI(req.Status.Value)
	return id, req.Name.Value, description, price, stock, req.Category.Value, status
}

func ParseUUID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}

func TimeNow() time.Time {
	return time.Now()
}
