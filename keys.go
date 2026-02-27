package lnbot

import (
	"context"
	"fmt"
)

// KeysService handles API key operations.
type KeysService struct{ c *Client }

// List returns all API keys for the current wallet.
func (s *KeysService) List(ctx context.Context) ([]APIKey, error) {
	var v []APIKey
	if err := s.c.get(ctx, "/v1/keys", &v); err != nil {
		return nil, err
	}
	return v, nil
}

// Rotate rotates the API key at the given slot (0 = primary, 1 = secondary).
func (s *KeysService) Rotate(ctx context.Context, slot int) (*RotatedAPIKey, error) {
	var v RotatedAPIKey
	if err := s.c.post(ctx, fmt.Sprintf("/v1/keys/%d/rotate", slot), nil, &v); err != nil {
		return nil, err
	}
	return &v, nil
}
