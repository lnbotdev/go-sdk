package lnbot

import "context"

// WalletsService handles wallet operations.
type WalletsService struct{ c *Client }

// Create creates a new wallet. Does not require authentication.
func (s *WalletsService) Create(ctx context.Context, params *CreateWalletParams) (*WalletCredentials, error) {
	var v WalletCredentials
	if err := s.c.post(ctx, "/v1/wallets", params, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

// Current returns the wallet associated with the current API key.
func (s *WalletsService) Current(ctx context.Context) (*Wallet, error) {
	var v Wallet
	if err := s.c.get(ctx, "/v1/wallets/current", &v); err != nil {
		return nil, err
	}
	return &v, nil
}

// Update modifies the current wallet's settings.
func (s *WalletsService) Update(ctx context.Context, params *UpdateWalletParams) (*Wallet, error) {
	var v Wallet
	if err := s.c.patch(ctx, "/v1/wallets/current", params, &v); err != nil {
		return nil, err
	}
	return &v, nil
}
