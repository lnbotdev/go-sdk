package lnbot

import "context"

// TransactionsService handles wallet transaction operations.
type TransactionsService struct{ c *Client }

// List returns transactions for the current wallet.
func (s *TransactionsService) List(ctx context.Context, params *ListTransactionsParams) ([]Transaction, error) {
	path := "/v1/transactions"
	if params != nil {
		path = addListParams(path, params.Limit, params.After)
	}
	var v []Transaction
	if err := s.c.get(ctx, path, &v); err != nil {
		return nil, err
	}
	return v, nil
}
