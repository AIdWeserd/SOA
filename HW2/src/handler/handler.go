package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/go-faster/jx"
	"github.com/google/uuid"
	ogenerrors "github.com/ogen-go/ogen/ogenerrors"

	api "hw2/.build/openapi"
	"hw2/src/apierr"
	"hw2/src/service"
)

type Handler struct {
	svc *service.Service
}

func New(svc *service.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) CreateProduct(ctx context.Context, req *api.ProductCreate) (*api.ProductResponse, error) {
	resp, err := h.svc.CreateProduct(ctx, req)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (h *Handler) GetProductById(ctx context.Context, params api.GetProductByIdParams) (*api.ProductResponse, error) {
	id, err := uuid.Parse(params.ID)
	if err != nil {
		return nil, apierr.ValidationError([]apierr.ValidationDetail{
			{Field: "id", Message: "Невалидный UUID"},
		})
	}

	resp, err := h.svc.GetProductByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (h *Handler) ListProducts(ctx context.Context, params api.ListProductsParams) (*api.ProductPage, error) {
	resp, err := h.svc.ListProducts(ctx, params)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (h *Handler) UpdateProduct(ctx context.Context, req *api.ProductUpdate, params api.UpdateProductParams) (*api.ProductResponse, error) {
	id, err := uuid.Parse(params.ID)
	if err != nil {
		return nil, apierr.ValidationError([]apierr.ValidationDetail{
			{Field: "id", Message: "Невалидный UUID"},
		})
	}

	resp, err := h.svc.UpdateProduct(ctx, id, req)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (h *Handler) DeleteProduct(ctx context.Context, params api.DeleteProductParams) error {
	id, err := uuid.Parse(params.ID)
	if err != nil {
		return apierr.ValidationError([]apierr.ValidationDetail{
			{Field: "id", Message: "Невалидный UUID"},
		})
	}

	return h.svc.DeleteProduct(ctx, id)
}

func (h *Handler) NewError(ctx context.Context, err error) *api.ErrorStatusCode {
	log.Printf("NewError type: %T, value: %v", err, err)

	var ogenDecodeErr *ogenerrors.DecodeRequestError
	if errors.As(err, &ogenDecodeErr) {
		return toStatusCode(apierr.ValidationError([]apierr.ValidationDetail{
			{Field: "request", Message: ogenDecodeErr.Error()},
		}))
	}

	var ogenParamErr *ogenerrors.DecodeParamError
	if errors.As(err, &ogenParamErr) {
		return toStatusCode(apierr.ValidationError([]apierr.ValidationDetail{
			{Field: ogenParamErr.Name, Message: ogenParamErr.Err.Error()},
		}))
	}

	var e *apierr.Error
	if errors.As(err, &e) {
		return toStatusCode(e)
	}

	return toStatusCode(apierr.InternalError())
}

func (h *Handler) ErrorHandler() ogenerrors.ErrorHandler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
		resp := h.NewError(ctx, err)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.StatusCode)

		data, _ := json.Marshal(resp.Response)
		_, _ = w.Write(data)
	}
}

func toStatusCode(e *apierr.Error) *api.ErrorStatusCode {
	resp := api.ErrorStatusCode{
		StatusCode: e.Status,
		Response: api.Error{
			ErrorCode: e.Code,
			Message:   e.Message,
		},
	}

	if e.Details != nil {
		resp.Response.Details = toAPIDetails(e.Details)
	}

	return &resp
}

func toAPIDetails(details any) api.OptNilErrorDetails {
	d, ok := details.([]apierr.ValidationDetail)
	if !ok {
		return api.OptNilErrorDetails{}
	}

	m := make(api.ErrorDetails, len(d))
	for _, v := range d {
		b, _ := json.Marshal(v.Message)
		m[v.Field] = jx.Raw(b)
	}

	return api.NewOptNilErrorDetails(m)
}

