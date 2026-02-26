package lnbot

import (
	"context"
	"fmt"
	"net/url"
)

// AddressesService handles Lightning address operations.
type AddressesService struct{ c *Client }

// Create creates a new Lightning address for the current wallet.
func (s *AddressesService) Create(ctx context.Context, params *CreateAddressParams) (*Address, error) {
	var v Address
	if err := s.c.post(ctx, "/v1/addresses", params, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

// List returns all Lightning addresses for the current wallet.
func (s *AddressesService) List(ctx context.Context) ([]Address, error) {
	var v []Address
	if err := s.c.get(ctx, "/v1/addresses", &v); err != nil {
		return nil, err
	}
	return v, nil
}

// Delete removes a Lightning address.
func (s *AddressesService) Delete(ctx context.Context, address string) error {
	return s.c.del(ctx, fmt.Sprintf("/v1/addresses/%s", url.PathEscape(address)))
}

// Transfer moves a Lightning address to another wallet.
func (s *AddressesService) Transfer(ctx context.Context, address string, params *TransferAddressParams) (*AddressTransfer, error) {
	var v AddressTransfer
	if err := s.c.post(ctx, fmt.Sprintf("/v1/addresses/%s/transfer", url.PathEscape(address)), params, &v); err != nil {
		return nil, err
	}
	return &v, nil
}
