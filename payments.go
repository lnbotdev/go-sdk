package lnbot

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// PaymentsService handles wallet-scoped Lightning payment operations.
type PaymentsService struct {
	c      *Client
	prefix string
}

// Create sends a new Lightning payment.
func (s *PaymentsService) Create(ctx context.Context, params *CreatePaymentParams) (*Payment, error) {
	var v Payment
	if err := s.c.post(ctx, s.prefix+"/payments", params, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

// List returns payments for the wallet.
func (s *PaymentsService) List(ctx context.Context, params *ListPaymentsParams) ([]Payment, error) {
	path := s.prefix + "/payments"
	if params != nil {
		path = addListParams(path, params.Limit, params.After)
	}
	var v []Payment
	if err := s.c.get(ctx, path, &v); err != nil {
		return nil, err
	}
	return v, nil
}

// Get returns a single payment by number.
func (s *PaymentsService) Get(ctx context.Context, number int) (*Payment, error) {
	var v Payment
	if err := s.c.get(ctx, fmt.Sprintf("%s/payments/%d", s.prefix, number), &v); err != nil {
		return nil, err
	}
	return &v, nil
}

// GetByHash returns a single payment by its payment hash.
func (s *PaymentsService) GetByHash(ctx context.Context, paymentHash string) (*Payment, error) {
	var v Payment
	if err := s.c.get(ctx, fmt.Sprintf("%s/payments/%s", s.prefix, paymentHash), &v); err != nil {
		return nil, err
	}
	return &v, nil
}

// Resolve inspects a payment target (Lightning address, LNURL, or BOLT11) before sending.
func (s *PaymentsService) Resolve(ctx context.Context, target string) (*ResolveTargetResponse, error) {
	var v ResolveTargetResponse
	path := fmt.Sprintf("%s/payments/resolve?target=%s", s.prefix, url.QueryEscape(target))
	if err := s.c.get(ctx, path, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

// Watch opens an SSE stream and sends events to the returned channel.
func (s *PaymentsService) Watch(ctx context.Context, number int, timeout *int) (<-chan PaymentEvent, <-chan error) {
	path := fmt.Sprintf("%s%s/payments/%d/events", s.c.baseURL, s.prefix, number)
	if timeout != nil {
		path = fmt.Sprintf("%s?timeout=%d", path, *timeout)
	}
	return readSSEPayment(ctx, s.c, path)
}

// WatchByHash opens an SSE stream for a payment identified by payment hash.
func (s *PaymentsService) WatchByHash(ctx context.Context, paymentHash string, timeout *int) (<-chan PaymentEvent, <-chan error) {
	path := fmt.Sprintf("%s%s/payments/%s/events", s.c.baseURL, s.prefix, paymentHash)
	if timeout != nil {
		path = fmt.Sprintf("%s?timeout=%d", path, *timeout)
	}
	return readSSEPayment(ctx, s.c, path)
}

// ── SSE helpers ─────────────────────────────────────────

func readSSEPayment(ctx context.Context, c *Client, url string) (<-chan PaymentEvent, <-chan error) {
	events := make(chan PaymentEvent, 1)
	errs := make(chan error, 1)

	go func() {
		defer close(events)
		defer close(errs)

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			errs <- err
			return
		}
		req.Header.Set("Accept", "text/event-stream")
		req.Header.Set("User-Agent", "lnbot-go/"+Version)
		if c.apiKey != "" {
			req.Header.Set("Authorization", "Bearer "+c.apiKey)
		}

		resp, err := c.http.Do(req)
		if err != nil {
			errs <- err
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 400 {
			body, _ := io.ReadAll(resp.Body)
			errs <- parseAPIError(resp.StatusCode, body)
			return
		}

		scanner := bufio.NewScanner(resp.Body)
		var (
			eventType string
			dataLines []string
		)
		dispatch := func() {
			if eventType == "" || len(dataLines) == 0 {
				return
			}
			raw := strings.Join(dataLines, "\n")
			var data Payment
			if json.Unmarshal([]byte(raw), &data) == nil {
				events <- PaymentEvent{Event: eventType, Data: data}
			}
		}

		for scanner.Scan() {
			line := scanner.Text()
			switch {
			case line == "":
				dispatch()
				eventType = ""
				dataLines = nil
			case strings.HasPrefix(line, "event:"):
				eventType = strings.TrimSpace(line[6:])
			case strings.HasPrefix(line, "data:"):
				dataLines = append(dataLines, strings.TrimSpace(line[5:]))
			}
		}
		dispatch()

		if err := scanner.Err(); err != nil {
			errs <- err
		}
	}()

	return events, errs
}
