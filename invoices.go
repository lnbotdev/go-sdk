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

// InvoicesService handles Lightning invoice operations.
type InvoicesService struct{ c *Client }

// Create creates a new Lightning invoice.
func (s *InvoicesService) Create(ctx context.Context, params *CreateInvoiceParams) (*Invoice, error) {
	var v Invoice
	if err := s.c.post(ctx, "/v1/invoices", params, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

// List returns invoices for the current wallet.
func (s *InvoicesService) List(ctx context.Context, params *ListInvoicesParams) ([]Invoice, error) {
	path := "/v1/invoices"
	if params != nil {
		path = addListParams(path, params.Limit, params.After)
	}
	var v []Invoice
	if err := s.c.get(ctx, path, &v); err != nil {
		return nil, err
	}
	return v, nil
}

// Get returns a single invoice by number.
func (s *InvoicesService) Get(ctx context.Context, number int) (*Invoice, error) {
	var v Invoice
	if err := s.c.get(ctx, fmt.Sprintf("/v1/invoices/%d", number), &v); err != nil {
		return nil, err
	}
	return &v, nil
}

// WaitForSettlement opens an SSE stream and sends events to the returned channel.
// The channel is closed when the stream ends. Cancel the context to abort.
func (s *InvoicesService) WaitForSettlement(ctx context.Context, number int, timeout *int) (<-chan InvoiceEvent, <-chan error) {
	events := make(chan InvoiceEvent, 1)
	errs := make(chan error, 1)

	go func() {
		defer close(events)
		defer close(errs)

		path := fmt.Sprintf("%s/v1/invoices/%d/events", s.c.baseURL, number)
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
			var data Invoice
			if json.Unmarshal([]byte(raw), &data) == nil {
				events <- InvoiceEvent{Event: eventType, Data: data}
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
