package lnbot

import "context"

// WalletsService handles account-level wallet operations.
type WalletsService struct{ c *Client }

// Create creates a new wallet. Requires user key authentication.
func (s *WalletsService) Create(ctx context.Context) (*CreateWalletResponse, error) {
	var v CreateWalletResponse
	if err := s.c.post(ctx, "/v1/wallets", nil, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

// List returns all wallets for the authenticated user.
func (s *WalletsService) List(ctx context.Context) ([]WalletListItem, error) {
	var v []WalletListItem
	if err := s.c.get(ctx, "/v1/wallets", &v); err != nil {
		return nil, err
	}
	return v, nil
}
