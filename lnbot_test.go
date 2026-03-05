package lnbot

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// testServer creates an httptest server and an lnbot client pointed at it.
func testServer(t *testing.T, handler http.HandlerFunc) (*Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	c := New("uk_test", WithBaseURL(srv.URL))
	return c, srv
}

// jsonHandler returns an http.HandlerFunc that responds with JSON and captures the request.
func jsonHandler(status int, body any) (http.HandlerFunc, *capturedRequest) {
	cap := &capturedRequest{}
	return func(w http.ResponseWriter, r *http.Request) {
		cap.Method = r.Method
		cap.Path = r.URL.Path
		cap.Query = r.URL.RawQuery
		cap.Headers = r.Header
		if r.Body != nil {
			data, _ := io.ReadAll(r.Body)
			cap.Body = string(data)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		if body != nil {
			json.NewEncoder(w).Encode(body)
		}
	}, cap
}

type capturedRequest struct {
	Method  string
	Path    string
	Query   string
	Headers http.Header
	Body    string
}

func TestNew_DefaultBaseURL(t *testing.T) {
	c := New("uk_test")
	if c.baseURL != defaultBaseURL {
		t.Errorf("got %q, want %q", c.baseURL, defaultBaseURL)
	}
}

func TestNew_CustomBaseURL(t *testing.T) {
	c := New("uk_test", WithBaseURL("https://custom.api.com/"))
	if c.baseURL != "https://custom.api.com" {
		t.Errorf("got %q, want trailing slash trimmed", c.baseURL)
	}
}

func TestNew_InitializesTopLevelServices(t *testing.T) {
	c := New("uk_test")
	if c.Wallets == nil || c.Keys == nil || c.Invoices == nil ||
		c.Backup == nil || c.Restore == nil {
		t.Error("expected all top-level services to be initialized")
	}
}

func TestNew_CustomHTTPClient(t *testing.T) {
	custom := &http.Client{}
	c := New("uk_test", WithHTTPClient(custom))
	if c.http != custom {
		t.Error("expected custom HTTP client to be used")
	}
}

func TestWalletHandle_InitializesAllServices(t *testing.T) {
	c := New("uk_test")
	w := c.Wallet("wal_1")
	if w.Key == nil || w.Invoices == nil || w.Payments == nil ||
		w.Addresses == nil || w.Transactions == nil || w.Webhooks == nil ||
		w.Events == nil || w.L402 == nil {
		t.Error("expected all wallet services to be initialized")
	}
	if w.WalletID != "wal_1" {
		t.Errorf("WalletID = %q, want wal_1", w.WalletID)
	}
}

func TestRegister(t *testing.T) {
	h, cap := jsonHandler(200, map[string]any{
		"userId": "usr_1", "primaryKey": "uk_pk", "secondaryKey": "uk_sk",
		"recoveryPassphrase": "word1 word2 word3",
	})
	c, _ := testServer(t, h)

	res, err := c.Register(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if cap.Method != "POST" {
		t.Errorf("Method = %q, want POST", cap.Method)
	}
	if cap.Path != "/v1/register" {
		t.Errorf("Path = %q", cap.Path)
	}
	if res.PrimaryKey != "uk_pk" {
		t.Errorf("PrimaryKey = %q", res.PrimaryKey)
	}
	if res.RecoveryPassphrase != "word1 word2 word3" {
		t.Errorf("RecoveryPassphrase = %q", res.RecoveryPassphrase)
	}
}

func TestMe(t *testing.T) {
	h, cap := jsonHandler(200, map[string]any{"walletId": "wal_1"})
	c, _ := testServer(t, h)

	res, err := c.Me(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if cap.Method != "GET" {
		t.Errorf("Method = %q, want GET", cap.Method)
	}
	if cap.Path != "/v1/me" {
		t.Errorf("Path = %q", cap.Path)
	}
	if res.WalletID != "wal_1" {
		t.Errorf("WalletID = %q", res.WalletID)
	}
}

func TestRequest_SendsAuthorizationHeader(t *testing.T) {
	h, cap := jsonHandler(200, map[string]any{"walletId": "wal_1"})
	c, _ := testServer(t, h)

	_, _ = c.Me(context.Background())

	if got := cap.Headers.Get("Authorization"); got != "Bearer uk_test" {
		t.Errorf("Authorization = %q, want %q", got, "Bearer uk_test")
	}
}

func TestRequest_OmitsAuthWhenNoAPIKey(t *testing.T) {
	h, cap := jsonHandler(200, map[string]any{"walletId": "wal_1"})
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	c := New("", WithBaseURL(srv.URL))

	_, _ = c.Me(context.Background())

	if got := cap.Headers.Get("Authorization"); got != "" {
		t.Errorf("Authorization = %q, want empty", got)
	}
}

func TestRequest_SendsUserAgent(t *testing.T) {
	h, cap := jsonHandler(200, map[string]any{"walletId": "wal_1"})
	c, _ := testServer(t, h)

	_, _ = c.Me(context.Background())

	if got := cap.Headers.Get("User-Agent"); got != "lnbot-go/"+Version {
		t.Errorf("User-Agent = %q, want %q", got, "lnbot-go/"+Version)
	}
}

func TestRequest_SendsAcceptJSON(t *testing.T) {
	h, cap := jsonHandler(200, map[string]any{"walletId": "wal_1"})
	c, _ := testServer(t, h)

	_, _ = c.Me(context.Background())

	if got := cap.Headers.Get("Accept"); got != "application/json" {
		t.Errorf("Accept = %q, want application/json", got)
	}
}

func TestRequest_SendsContentTypeForPost(t *testing.T) {
	h, cap := jsonHandler(200, map[string]any{"number": 1, "status": "pending", "amount": 100, "bolt11": "lnbc1..."})
	c, _ := testServer(t, h)

	w := c.Wallet("wal_1")
	_, _ = w.Invoices.Create(context.Background(), &CreateInvoiceParams{Amount: 100})

	if got := cap.Headers.Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", got)
	}
}

func TestRequest_OmitsContentTypeForGet(t *testing.T) {
	h, cap := jsonHandler(200, map[string]any{"walletId": "wal_1"})
	c, _ := testServer(t, h)

	_, _ = c.Me(context.Background())

	if got := cap.Headers.Get("Content-Type"); got != "" {
		t.Errorf("Content-Type = %q, want empty", got)
	}
}

func TestRequest_SerializesBodyAsJSON(t *testing.T) {
	h, cap := jsonHandler(200, map[string]any{"number": 1, "status": "pending", "amount": 100, "bolt11": "lnbc1..."})
	c, _ := testServer(t, h)

	memo := "test memo"
	w := c.Wallet("wal_1")
	_, _ = w.Invoices.Create(context.Background(), &CreateInvoiceParams{Amount: 100, Memo: &memo})

	var body map[string]any
	json.Unmarshal([]byte(cap.Body), &body)
	if body["amount"] != float64(100) {
		t.Errorf("amount = %v, want 100", body["amount"])
	}
	if body["memo"] != "test memo" {
		t.Errorf("memo = %v, want test memo", body["memo"])
	}
}

func TestRequest_OmitsNilOptionalFields(t *testing.T) {
	h, cap := jsonHandler(200, map[string]any{"number": 1, "status": "pending", "amount": 100, "bolt11": "lnbc1..."})
	c, _ := testServer(t, h)

	w := c.Wallet("wal_1")
	_, _ = w.Invoices.Create(context.Background(), &CreateInvoiceParams{Amount: 100})

	var body map[string]any
	json.Unmarshal([]byte(cap.Body), &body)
	if _, ok := body["memo"]; ok {
		t.Error("expected memo to be omitted")
	}
	if _, ok := body["reference"]; ok {
		t.Error("expected reference to be omitted")
	}
}

func TestResponse_ParsesJSON(t *testing.T) {
	h, _ := jsonHandler(200, map[string]any{"walletId": "wal_123", "name": "My Wallet", "balance": 1000, "onHold": 50, "available": 950})
	c, _ := testServer(t, h)

	w := c.Wallet("wal_123")
	wal, err := w.Get(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if wal.WalletID != "wal_123" {
		t.Errorf("WalletID = %q, want wal_123", wal.WalletID)
	}
	if wal.Balance != 1000 {
		t.Errorf("Balance = %d, want 1000", wal.Balance)
	}
}

func TestHTTPMethods(t *testing.T) {
	tests := []struct {
		name string
		call func(*Client) error
		want string
	}{
		{"GET", func(c *Client) error { _, err := c.Wallet("wal_1").Get(context.Background()); return err }, "GET"},
		{"POST", func(c *Client) error {
			_, err := c.Wallet("wal_1").Invoices.Create(context.Background(), &CreateInvoiceParams{Amount: 100})
			return err
		}, "POST"},
		{"PATCH", func(c *Client) error {
			_, err := c.Wallet("wal_1").Update(context.Background(), &UpdateWalletParams{Name: "n"})
			return err
		}, "PATCH"},
		{"DELETE", func(c *Client) error {
			return c.Wallet("wal_1").Webhooks.Delete(context.Background(), "wh_1")
		}, "DELETE"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, cap := jsonHandler(200, map[string]any{"walletId": "wal_1", "name": "n", "balance": 0, "onHold": 0, "available": 0})
			c, _ := testServer(t, h)

			_ = tt.call(c)

			if cap.Method != tt.want {
				t.Errorf("Method = %q, want %q", cap.Method, tt.want)
			}
		})
	}
}

func TestErrorMapping(t *testing.T) {
	tests := []struct {
		status int
		check  func(error) bool
	}{
		{400, func(e error) bool { var t *BadRequestError; return errors.As(e, &t) }},
		{401, func(e error) bool { var t *UnauthorizedError; return errors.As(e, &t) }},
		{403, func(e error) bool { var t *ForbiddenError; return errors.As(e, &t) }},
		{404, func(e error) bool { var t *NotFoundError; return errors.As(e, &t) }},
		{409, func(e error) bool { var t *ConflictError; return errors.As(e, &t) }},
		{500, func(e error) bool { var t *APIError; return errors.As(e, &t) }},
	}
	for _, tt := range tests {
		t.Run(http.StatusText(tt.status), func(t *testing.T) {
			h, _ := jsonHandler(tt.status, map[string]string{"message": "test error"})
			c, _ := testServer(t, h)

			_, err := c.Me(context.Background())

			if err == nil {
				t.Fatal("expected error")
			}
			if !tt.check(err) {
				t.Errorf("status %d: got %T, want matching typed error", tt.status, err)
			}
		})
	}
}

func TestErrorMessageExtraction(t *testing.T) {
	t.Run("extracts message field", func(t *testing.T) {
		h, _ := jsonHandler(400, map[string]string{"message": "invalid amount"})
		c, _ := testServer(t, h)

		_, err := c.Me(context.Background())
		var apiErr *APIError
		if errors.As(err, &apiErr) {
			if apiErr.Message != "invalid amount" {
				t.Errorf("Message = %q, want %q", apiErr.Message, "invalid amount")
			}
		}
	})

	t.Run("extracts error field", func(t *testing.T) {
		h, _ := jsonHandler(400, map[string]string{"error": "bad input"})
		c, _ := testServer(t, h)

		_, err := c.Me(context.Background())
		var apiErr *APIError
		if errors.As(err, &apiErr) {
			if apiErr.Message != "bad input" {
				t.Errorf("Message = %q, want %q", apiErr.Message, "bad input")
			}
		}
	})

	t.Run("falls back to status text", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(500)
			w.Write([]byte("not json"))
		}))
		t.Cleanup(srv.Close)
		c := New("uk_test", WithBaseURL(srv.URL))

		_, err := c.Me(context.Background())
		var apiErr *APIError
		if errors.As(err, &apiErr) {
			if apiErr.Message != "Internal Server Error" {
				t.Errorf("Message = %q, want fallback", apiErr.Message)
			}
		}
	})
}

func TestAddListParams(t *testing.T) {
	t.Run("no params", func(t *testing.T) {
		got := addListParams("/v1/wallets/wal_1/invoices", nil, nil)
		if got != "/v1/wallets/wal_1/invoices" {
			t.Errorf("got %q", got)
		}
	})

	t.Run("limit only", func(t *testing.T) {
		limit := 10
		got := addListParams("/v1/wallets/wal_1/invoices", &limit, nil)
		if got != "/v1/wallets/wal_1/invoices?limit=10" {
			t.Errorf("got %q", got)
		}
	})

	t.Run("both params", func(t *testing.T) {
		limit, after := 10, 5
		got := addListParams("/v1/wallets/wal_1/invoices", &limit, &after)
		// url.Values encodes alphabetically
		if got != "/v1/wallets/wal_1/invoices?after=5&limit=10" {
			t.Errorf("got %q", got)
		}
	})
}

func TestPtr(t *testing.T) {
	v := Ptr(42)
	if *v != 42 {
		t.Errorf("got %d, want 42", *v)
	}
	s := Ptr("hello")
	if *s != "hello" {
		t.Errorf("got %q, want hello", *s)
	}
}
