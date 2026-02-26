package lnbot

import "context"

// RestoreService handles wallet restore operations.
type RestoreService struct{ c *Client }

// Recovery restores a wallet using a recovery passphrase. Does not require authentication.
func (s *RestoreService) Recovery(ctx context.Context, params *RecoveryRestoreParams) (*RestoredWallet, error) {
	var v RestoredWallet
	if err := s.c.post(ctx, "/v1/restore/recovery", params, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

// PasskeyBegin starts the passkey authentication flow for wallet restore.
func (s *RestoreService) PasskeyBegin(ctx context.Context) (*PasskeyAuthenticationChallenge, error) {
	var v PasskeyAuthenticationChallenge
	if err := s.c.post(ctx, "/v1/restore/passkey/begin", nil, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

// PasskeyComplete finishes the passkey authentication flow and restores the wallet.
func (s *RestoreService) PasskeyComplete(ctx context.Context, params *PasskeyAssertionParams) (*RestoredWallet, error) {
	var v RestoredWallet
	if err := s.c.post(ctx, "/v1/restore/passkey/complete", params, &v); err != nil {
		return nil, err
	}
	return &v, nil
}
