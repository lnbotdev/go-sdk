package lnbot

import (
	"context"
	"fmt"
)

// KeysService handles API key operations.
type KeysService struct{ c *Client }

// Rotate rotates the API key at the given slot (0 = primary, 1 = secondary).
func (s *KeysService) Rotate(ctx context.Context, slot int) (*RotatedAPIKey, error) {
	var v RotatedAPIKey
	if err := s.c.post(ctx, fmt.Sprintf("/v1/keys/%d/rotate", slot), nil, &v); err != nil {
		return nil, err
	}
	return &v, nil
}
