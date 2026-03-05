package lnbot

import (
	"context"
	"fmt"
	"net/url"
)

// WebhooksService handles wallet-scoped webhook operations.
type WebhooksService struct {
	c      *Client
	prefix string
}

// Create registers a new webhook endpoint.
func (s *WebhooksService) Create(ctx context.Context, params *CreateWebhookParams) (*WebhookWithSecret, error) {
	var v WebhookWithSecret
	if err := s.c.post(ctx, s.prefix+"/webhooks", params, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

// List returns all registered webhooks for the wallet.
func (s *WebhooksService) List(ctx context.Context) ([]Webhook, error) {
	var v []Webhook
	if err := s.c.get(ctx, s.prefix+"/webhooks", &v); err != nil {
		return nil, err
	}
	return v, nil
}

// Delete removes a webhook endpoint.
func (s *WebhooksService) Delete(ctx context.Context, id string) error {
	return s.c.del(ctx, fmt.Sprintf("%s/webhooks/%s", s.prefix, url.PathEscape(id)))
}
