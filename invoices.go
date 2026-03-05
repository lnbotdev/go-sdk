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

// InvoicesService handles wallet-scoped Lightning invoice operations.
type InvoicesService struct {
	c      *Client
	prefix string
}

// Create creates a new Lightning invoice.
func (s *InvoicesService) Create(ctx context.Context, params *CreateInvoiceParams) (*Invoice, error) {
	var v Invoice
	if err := s.c.post(ctx, s.prefix+"/invoices", params, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

// List returns invoices for the wallet.
func (s *InvoicesService) List(ctx context.Context, params *ListInvoicesParams) ([]Invoice, error) {
	path := s.prefix + "/invoices"
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
	if err := s.c.get(ctx, fmt.Sprintf("%s/invoices/%d", s.prefix, number), &v); err != nil {
		return nil, err
	}
	return &v, nil
}

// GetByHash returns a single invoice by its payment hash.
func (s *InvoicesService) GetByHash(ctx context.Context, paymentHash string) (*Invoice, error) {
	var v Invoice
	if err := s.c.get(ctx, fmt.Sprintf("%s/invoices/%s", s.prefix, paymentHash), &v); err != nil {
		return nil, err
	}
	return &v, nil
}

// Watch opens an SSE stream and sends events to the returned channel.
// The channel is closed when the stream ends. Cancel the context to abort.
func (s *InvoicesService) Watch(ctx context.Context, number int, timeout *int) (<-chan InvoiceEvent, <-chan error) {
	path := fmt.Sprintf("%s%s/invoices/%d/events", s.c.baseURL, s.prefix, number)
	if timeout != nil {
		path = fmt.Sprintf("%s?timeout=%d", path, *timeout)
	}
	return readSSEInvoice(ctx, s.c, path)
}

// WatchByHash opens an SSE stream for an invoice identified by payment hash.
func (s *InvoicesService) WatchByHash(ctx context.Context, paymentHash string, timeout *int) (<-chan InvoiceEvent, <-chan error) {
	path := fmt.Sprintf("%s%s/invoices/%s/events", s.c.baseURL, s.prefix, paymentHash)
	if timeout != nil {
		path = fmt.Sprintf("%s?timeout=%d", path, *timeout)
	}
	return readSSEInvoice(ctx, s.c, path)
}

// PublicInvoicesService handles public (no-auth) invoice creation.
type PublicInvoicesService struct{ c *Client }

// CreateForWallet creates an invoice for a specific wallet by its ID (wal_xxx).
// No authentication required. Rate limited by IP.
func (s *PublicInvoicesService) CreateForWallet(ctx context.Context, params *CreateInvoiceForWalletParams) (*AddressInvoice, error) {
	var v AddressInvoice
	if err := s.c.post(ctx, "/v1/invoices/for-wallet", params, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

// CreateForAddress creates an invoice for the wallet owning the given Lightning address.
// No authentication required. Rate limited by IP.
func (s *PublicInvoicesService) CreateForAddress(ctx context.Context, params *CreateInvoiceForAddressParams) (*AddressInvoice, error) {
	var v AddressInvoice
	if err := s.c.post(ctx, "/v1/invoices/for-address", params, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

// ── SSE helpers ─────────────────────────────────────────

func readSSEInvoice(ctx context.Context, c *Client, url string) (<-chan InvoiceEvent, <-chan error) {
	events := make(chan InvoiceEvent, 1)
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
