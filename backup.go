package lnbot

import "context"

// BackupService handles wallet backup operations.
type BackupService struct{ c *Client }

// Recovery generates a recovery passphrase for the current wallet.
func (s *BackupService) Recovery(ctx context.Context) (*RecoveryPassphrase, error) {
	var v RecoveryPassphrase
	if err := s.c.post(ctx, "/v1/backup/recovery", nil, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

// PasskeyBegin starts the passkey registration flow for wallet backup.
func (s *BackupService) PasskeyBegin(ctx context.Context) (*PasskeyRegistrationChallenge, error) {
	var v PasskeyRegistrationChallenge
	if err := s.c.post(ctx, "/v1/backup/passkey/begin", nil, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

// PasskeyComplete finishes the passkey registration flow by verifying the attestation.
func (s *BackupService) PasskeyComplete(ctx context.Context, params *PasskeyAttestationParams) error {
	return s.c.post(ctx, "/v1/backup/passkey/complete", params, nil)
}
