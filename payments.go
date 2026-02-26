package lnbot

import (
	"context"
	"fmt"
)

// PaymentsService handles Lightning payment operations.
type PaymentsService struct{ c *Client }

// Create sends a new Lightning payment.
func (s *PaymentsService) Create(ctx context.Context, params *CreatePaymentParams) (*Payment, error) {
	var v Payment
	if err := s.c.post(ctx, "/v1/payments", params, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

// List returns payments for the current wallet.
func (s *PaymentsService) List(ctx context.Context, params *ListPaymentsParams) ([]Payment, error) {
	path := "/v1/payments"
	if params != nil {
		path = addListParams(path, params.Limit, params.After)
	}
	var v []Payment
	if err := s.c.get(ctx, path, &v); err != nil {
		return nil, err
	}
	return v, nil
}

// Get returns a single payment by number.
func (s *PaymentsService) Get(ctx context.Context, number int) (*Payment, error) {
	var v Payment
	if err := s.c.get(ctx, fmt.Sprintf("/v1/payments/%d", number), &v); err != nil {
		return nil, err
	}
	return &v, nil
}
