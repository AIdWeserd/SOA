package apierr

import "fmt"

type Error struct {
	Code    string
	Message string
	Status  int
	Details any
}

func (e *Error) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

type ValidationDetail struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func ProductNotFound() *Error {
	return &Error{
		Code:    "PRODUCT_NOT_FOUND",
		Message: "Товар не найден",
		Status:  404,
	}
}

func ProductInactive() *Error {
	return &Error{
		Code:    "PRODUCT_INACTIVE",
		Message: "Товар неактивен",
		Status:  409,
	}
}

func OrderNotFound() *Error {
	return &Error{
		Code:    "ORDER_NOT_FOUND",
		Message: "Заказ не найден",
		Status:  404,
	}
}

func OrderLimitExceeded() *Error {
	return &Error{
		Code:    "ORDER_LIMIT_EXCEEDED",
		Message: "Превышен лимит частоты создания/обновления заказа",
		Status:  429,
	}
}

func OrderHasActive() *Error {
	return &Error{
		Code:    "ORDER_HAS_ACTIVE",
		Message: "У пользователя уже есть активный заказ",
		Status:  409,
	}
}

func InvalidStateTransition(from, to string) *Error {
	return &Error{
		Code:    "INVALID_STATE_TRANSITION",
		Message: fmt.Sprintf("Недопустимый переход состояния: %s → %s", from, to),
		Status:  409,
	}
}

func InsufficientStock(productID string) *Error {
	return &Error{
		Code:    "INSUFFICIENT_STOCK",
		Message: fmt.Sprintf("Недостаточно товара на складе: %s", productID),
		Status:  409,
	}
}

func PromoCodeInvalid() *Error {
	return &Error{
		Code:    "PROMO_CODE_INVALID",
		Message: "Промокод не найден, истёк, исчерпан или неактивен",
		Status:  422,
	}
}

func PromoCodeMinAmount(min string) *Error {
	return &Error{
		Code:    "PROMO_CODE_MIN_AMOUNT",
		Message: fmt.Sprintf("Сумма заказа ниже минимальной для промокода (%s)", min),
		Status:  422,
	}
}

func OrderOwnershipViolation() *Error {
	return &Error{
		Code:    "ORDER_OWNERSHIP_VIOLATION",
		Message: "Заказ принадлежит другому пользователю",
		Status:  403,
	}
}

func ValidationError(details []ValidationDetail) *Error {
	return &Error{
		Code:    "VALIDATION_ERROR",
		Message: "Ошибка валидации входных данных",
		Status:  400,
		Details: details,
	}
}

func InternalError() *Error {
	return &Error{
		Code:    "INTERNAL_ERROR",
		Message: "Внутренняя ошибка сервера",
		Status:  500,
	}
}
