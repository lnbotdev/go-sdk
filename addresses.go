package lnbot

import (
	"context"
	"fmt"
	"net/url"
)

// AddressesService handles wallet-scoped Lightning address operations.
type AddressesService struct {
	c      *Client
	prefix string
}

// Create creates a new Lightning address for the wallet.
func (s *AddressesService) Create(ctx context.Context, params *CreateAddressParams) (*Address, error) {
	var v Address
	if err := s.c.post(ctx, s.prefix+"/addresses", params, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

// List returns all Lightning addresses for the wallet.
func (s *AddressesService) List(ctx context.Context) ([]Address, error) {
	var v []Address
	if err := s.c.get(ctx, s.prefix+"/addresses", &v); err != nil {
		return nil, err
	}
	return v, nil
}

// Delete removes a Lightning address.
func (s *AddressesService) Delete(ctx context.Context, address string) error {
	return s.c.del(ctx, fmt.Sprintf("%s/addresses/%s", s.prefix, url.PathEscape(address)))
}

// Transfer moves a Lightning address to another wallet.
func (s *AddressesService) Transfer(ctx context.Context, address string, params *TransferAddressParams) (*AddressTransfer, error) {
	var v AddressTransfer
	if err := s.c.post(ctx, fmt.Sprintf("%s/addresses/%s/transfer", s.prefix, url.PathEscape(address)), params, &v); err != nil {
		return nil, err
	}
	return &v, nil
}
