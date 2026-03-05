package lnbot

import "context"

// L402Service handles wallet-scoped L402 paywall authentication operations.
type L402Service struct {
	c      *Client
	prefix string
}

// CreateChallenge creates an L402 challenge (invoice + macaroon) for paywall authentication.
func (s *L402Service) CreateChallenge(ctx context.Context, params *CreateL402ChallengeParams) (*L402Challenge, error) {
	var v L402Challenge
	if err := s.c.post(ctx, s.prefix+"/l402/challenges", params, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

// Verify verifies an L402 authorization token. Stateless — checks signature, preimage, and caveats.
func (s *L402Service) Verify(ctx context.Context, params *VerifyL402Params) (*VerifyL402Response, error) {
	var v VerifyL402Response
	if err := s.c.post(ctx, s.prefix+"/l402/verify", params, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

// Pay pays an L402 challenge and returns a ready-to-use Authorization header.
func (s *L402Service) Pay(ctx context.Context, params *PayL402Params) (*L402PayResponse, error) {
	var v L402PayResponse
	if err := s.c.post(ctx, s.prefix+"/l402/pay", params, &v); err != nil {
		return nil, err
	}
	return &v, nil
}
