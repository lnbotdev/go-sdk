package lnbot

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// PaymentsService handles Lightning payment operations.
type PaymentsService struct{ c *Client }

// Create sends a new Lightning payment.
// The target can be a Lightning address (user@domain), LNURL, or BOLT11 invoice.
func (s *PaymentsService) Create(ctx context.Context, params *CreatePaymentParams) (*Payment, error) {
	var v Payment
	if err := s.c.post(ctx, "/v1/payments", params, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

// List returns payments for the current wallet.
func (s *PaymentsService) List(ctx context.Context, params *ListPaymentsParams) ([]Payment, error) {
	path := "/v1/payments"
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
	if err := s.c.get(ctx, fmt.Sprintf("/v1/payments/%d", number), &v); err != nil {
		return nil, err
	}
	return &v, nil
}

// GetByHash returns a single payment by its payment hash.
func (s *PaymentsService) GetByHash(ctx context.Context, paymentHash string) (*Payment, error) {
	var v Payment
	if err := s.c.get(ctx, fmt.Sprintf("/v1/payments/%s", paymentHash), &v); err != nil {
		return nil, err
	}
	return &v, nil
}

// Watch opens an SSE stream and sends events to the returned channel.
// The channel is closed when the stream ends. Cancel the context to abort.
func (s *PaymentsService) Watch(ctx context.Context, number int, timeout *int) (<-chan PaymentEvent, <-chan error) {
	events := make(chan PaymentEvent, 1)
	errs := make(chan error, 1)

	go func() {
		defer close(events)
		defer close(errs)

		path := fmt.Sprintf("%s/v1/payments/%d/events", s.c.baseURL, number)
		if timeout != nil {
			path = fmt.Sprintf("%s?timeout=%d", path, *timeout)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, path, nil)
		if err != nil {
			errs <- err
			return
		}
		req.Header.Set("Accept", "text/event-stream")
		req.Header.Set("User-Agent", "lnbot-go/"+Version)
		if s.c.apiKey != "" {
			req.Header.Set("Authorization", "Bearer "+s.c.apiKey)
		}

		resp, err := s.c.http.Do(req)
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

// WatchByHash opens an SSE stream for a payment identified by payment hash.
// The channel is closed when the stream ends. Cancel the context to abort.
func (s *PaymentsService) WatchByHash(ctx context.Context, paymentHash string, timeout *int) (<-chan PaymentEvent, <-chan error) {
	events := make(chan PaymentEvent, 1)
	errs := make(chan error, 1)

	go func() {
		defer close(events)
		defer close(errs)

		path := fmt.Sprintf("%s/v1/payments/%s/events", s.c.baseURL, paymentHash)
		if timeout != nil {
			path = fmt.Sprintf("%s?timeout=%d", path, *timeout)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, path, nil)
		if err != nil {
			errs <- err
			return
		}
		req.Header.Set("Accept", "text/event-stream")
		req.Header.Set("User-Agent", "lnbot-go/"+Version)
		if s.c.apiKey != "" {
			req.Header.Set("Authorization", "Bearer "+s.c.apiKey)
		}

		resp, err := s.c.http.Do(req)
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
