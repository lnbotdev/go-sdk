package lnbot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	defaultBaseURL = "https://api.ln.bot"
	defaultTimeout = 30 * time.Second

	// Version is the SDK version sent in the User-Agent header.
	Version = "0.4.0"
)

// Option configures the Client.
type Option func(*Client)

// WithBaseURL sets a custom API base URL.
func WithBaseURL(u string) Option {
	return func(c *Client) { c.baseURL = strings.TrimRight(u, "/") }
}

// WithHTTPClient sets a custom *http.Client.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) { c.http = hc }
}

// Client is the LnBot API client.
type Client struct {
	apiKey  string
	baseURL string
	http    *http.Client

	Wallets      *WalletsService
	Keys         *KeysService
	Invoices     *InvoicesService
	Payments     *PaymentsService
	Addresses    *AddressesService
	Transactions *TransactionsService
	Webhooks     *WebhooksService
	Events       *EventsService
	Backup       *BackupService
	Restore      *RestoreService
}

// New creates a new LnBot client.
// Pass an empty string for apiKey if not needed (wallet creation, restore).
func New(apiKey string, opts ...Option) *Client {
	c := &Client{
		apiKey:  apiKey,
		baseURL: defaultBaseURL,
		http:    http.DefaultClient,
	}
	for _, o := range opts {
		o(c)
	}
	c.Wallets = &WalletsService{c: c}
	c.Keys = &KeysService{c: c}
	c.Invoices = &InvoicesService{c: c}
	c.Payments = &PaymentsService{c: c}
	c.Addresses = &AddressesService{c: c}
	c.Transactions = &TransactionsService{c: c}
	c.Webhooks = &WebhooksService{c: c}
	c.Events = &EventsService{c: c}
	c.Backup = &BackupService{c: c}
	c.Restore = &RestoreService{c: c}
	return c
}

func (c *Client) newRequest(ctx context.Context, method, path string, body any) (*http.Request, error) {
	u := c.baseURL + path

	var buf io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		buf = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, u, buf)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "lnbot-go/"+Version)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}
	return req, nil
}

func (c *Client) do(req *http.Request, v any) error {
	if _, ok := req.Context().Deadline(); !ok {
		ctx, cancel := context.WithTimeout(req.Context(), defaultTimeout)
		defer cancel()
		req = req.WithContext(ctx)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		return parseAPIError(resp.StatusCode, data)
	}

	if resp.StatusCode == 204 || v == nil {
		return nil
	}

	return json.Unmarshal(data, v)
}

func (c *Client) get(ctx context.Context, path string, v any) error {
	req, err := c.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return err
	}
	return c.do(req, v)
}

func (c *Client) post(ctx context.Context, path string, body, v any) error {
	req, err := c.newRequest(ctx, http.MethodPost, path, body)
	if err != nil {
		return err
	}
	return c.do(req, v)
}

func (c *Client) patch(ctx context.Context, path string, body, v any) error {
	req, err := c.newRequest(ctx, http.MethodPatch, path, body)
	if err != nil {
		return err
	}
	return c.do(req, v)
}

func (c *Client) del(ctx context.Context, path string) error {
	req, err := c.newRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return err
	}
	return c.do(req, nil)
}

func addListParams(path string, limit, after *int) string {
	params := url.Values{}
	if limit != nil {
		params.Set("limit", strconv.Itoa(*limit))
	}
	if after != nil {
		params.Set("after", strconv.Itoa(*after))
	}
	if len(params) == 0 {
		return path
	}
	return fmt.Sprintf("%s?%s", path, params.Encode())
}
