package lib

import (
	"github.com/shopspring/decimal"

	api "hw2/.build/openapi"
	"hw2/src/apierr"
)

func ValidateProductCreate(req *api.ProductCreate) *apierr.Error {
	var details []apierr.ValidationDetail

	if len(req.Name) < 1 {
		details = append(details, apierr.ValidationDetail{
			Field:   "name",
			Message: "Не может быть пустым",
		})
	} else if len(req.Name) > 255 {
		details = append(details, apierr.ValidationDetail{
			Field:   "name",
			Message: "Максимальная длина 255 символов",
		})
	}

	if req.Description.IsSet() && !req.Description.IsNull() {
		if len(req.Description.Value) > 4000 {
			details = append(details, apierr.ValidationDetail{
				Field:   "description",
				Message: "Максимальная длина 4000 символов",
			})
		}
	}

	price := req.Price
	if price.LessThanOrEqual(decimal.Zero) {
		details = append(details, apierr.ValidationDetail{
			Field:   "price",
			Message: "Цена должна быть больше 0",
		})
	}

	if req.Stock < 0 {
		details = append(details, apierr.ValidationDetail{
			Field:   "stock",
			Message: "Остаток не может быть отрицательным",
		})
	}

	if len(req.Category) < 1 {
		details = append(details, apierr.ValidationDetail{
			Field:   "category",
			Message: "Не может быть пустой",
		})
	} else if len(req.Category) > 100 {
		details = append(details, apierr.ValidationDetail{
			Field:   "category",
			Message: "Максимальная длина 100 символов",
		})
	}

	if len(details) > 0 {
		return apierr.ValidationError(details)
	}
	return nil
}

func ValidateProductUpdate(req *api.ProductUpdate) *apierr.Error {
	var details []apierr.ValidationDetail

	if req.Name.IsSet() {
		if len(req.Name.Value) < 1 {
			details = append(details, apierr.ValidationDetail{
				Field:   "name",
				Message: "Не может быть пустым",
			})
		} else if len(req.Name.Value) > 255 {
			details = append(details, apierr.ValidationDetail{
				Field:   "name",
				Message: "Максимальная длина 255 символов",
			})
		}
	}

	if req.Description.IsSet() && !req.Description.IsNull() {
		if len(req.Description.Value) > 4000 {
			details = append(details, apierr.ValidationDetail{
				Field:   "description",
				Message: "Максимальная длина 4000 символов",
			})
		}
	}

	if req.Price.IsSet() {
		price := req.Price.Value
		if price.LessThanOrEqual(decimal.Zero) {
			details = append(details, apierr.ValidationDetail{
				Field:   "price",
				Message: "Цена должна быть больше 0",
			})
		}
	}

	if req.Stock.IsSet() && req.Stock.Value < 0 {
		details = append(details, apierr.ValidationDetail{
			Field:   "stock",
			Message: "Остаток не может быть отрицательным",
		})
	}

	if req.Category.IsSet() {
		if len(req.Category.Value) < 1 {
			details = append(details, apierr.ValidationDetail{
				Field:   "category",
				Message: "Не может быть пустой",
			})
		} else if len(req.Category.Value) > 100 {
			details = append(details, apierr.ValidationDetail{
				Field:   "category",
				Message: "Максимальная длина 100 символов",
			})
		}
	}

	if len(details) > 0 {
		return apierr.ValidationError(details)
	}
	return nil
}
