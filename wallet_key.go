package lnbot

import "context"

// WalletKeyService handles wallet key operations.
type WalletKeyService struct {
	c      *Client
	prefix string
}

// Create creates a wallet key (max 1 per wallet).
func (s *WalletKeyService) Create(ctx context.Context) (*WalletKeyResponse, error) {
	var v WalletKeyResponse
	if err := s.c.post(ctx, s.prefix+"/key", nil, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

// Get returns wallet key info (no plaintext key).
func (s *WalletKeyService) Get(ctx context.Context) (*WalletKeyInfoResponse, error) {
	var v WalletKeyInfoResponse
	if err := s.c.get(ctx, s.prefix+"/key", &v); err != nil {
		return nil, err
	}
	return &v, nil
}

// Delete revokes the wallet key.
func (s *WalletKeyService) Delete(ctx context.Context) error {
	return s.c.del(ctx, s.prefix+"/key")
}

// Rotate rotates the wallet key. Old key is invalidated immediately.
func (s *WalletKeyService) Rotate(ctx context.Context) (*WalletKeyResponse, error) {
	var v WalletKeyResponse
	if err := s.c.post(ctx, s.prefix+"/key/rotate", nil, &v); err != nil {
		return nil, err
	}
	return &v, nil
}
