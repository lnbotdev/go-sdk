package lnbot

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// sseHandler returns an HTTP handler that writes raw SSE text and closes.
func sseHandler(t *testing.T, body string) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(200)
		fmt.Fprint(w, body)
	}
}

// sseServer creates a test server and client for SSE tests.
func sseServer(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return New("key_test", WithBaseURL(srv.URL))
}

// collectInvoiceEvents drains the events channel into a slice.
func collectInvoiceEvents(events <-chan InvoiceEvent, errs <-chan error) ([]InvoiceEvent, error) {
	var result []InvoiceEvent
	for ev := range events {
		result = append(result, ev)
	}
	if err, ok := <-errs; ok && err != nil {
		return result, err
	}
	return result, nil
}

// collectPaymentEvents drains the events channel into a slice.
func collectPaymentEvents(events <-chan PaymentEvent, errs <-chan error) ([]PaymentEvent, error) {
	var result []PaymentEvent
	for ev := range events {
		result = append(result, ev)
	}
	if err, ok := <-errs; ok && err != nil {
		return result, err
	}
	return result, nil
}

// collectWalletEvents drains the events channel into a slice.
func collectWalletEvents(events <-chan WalletEvent, errs <-chan error) ([]WalletEvent, error) {
	var result []WalletEvent
	for ev := range events {
		result = append(result, ev)
	}
	if err, ok := <-errs; ok && err != nil {
		return result, err
	}
	return result, nil
}

// ---------------------------------------------------------------------------
// Invoice Watch
// ---------------------------------------------------------------------------

func TestInvoiceWatch_YieldsEvents(t *testing.T) {
	sse := "event: settled\ndata: {\"number\":1,\"status\":\"settled\",\"amount\":100,\"bolt11\":\"lnbc1...\"}\n\n"
	c := sseServer(t, sseHandler(t, sse))

	evs, err := collectInvoiceEvents(c.Invoices.Watch(context.Background(), 1, nil))
	if err != nil {
		t.Fatal(err)
	}
	if len(evs) != 1 {
		t.Fatalf("len = %d, want 1", len(evs))
	}
	if evs[0].Event != "settled" {
		t.Errorf("Event = %q", evs[0].Event)
	}
	if evs[0].Data.Number != 1 {
		t.Errorf("Number = %d", evs[0].Data.Number)
	}
	if evs[0].Data.Amount != 100 {
		t.Errorf("Amount = %d", evs[0].Data.Amount)
	}
}

func TestInvoiceWatch_MultipleEvents(t *testing.T) {
	sse := "event: pending\ndata: {\"number\":1,\"status\":\"pending\",\"amount\":50,\"bolt11\":\"lnbc1...\"}\n\n" +
		"event: settled\ndata: {\"number\":1,\"status\":\"settled\",\"amount\":50,\"bolt11\":\"lnbc1...\"}\n\n"
	c := sseServer(t, sseHandler(t, sse))

	evs, err := collectInvoiceEvents(c.Invoices.Watch(context.Background(), 1, nil))
	if err != nil {
		t.Fatal(err)
	}
	if len(evs) != 2 {
		t.Fatalf("len = %d, want 2", len(evs))
	}
	if evs[0].Event != "pending" {
		t.Errorf("evs[0].Event = %q", evs[0].Event)
	}
	if evs[1].Event != "settled" {
		t.Errorf("evs[1].Event = %q", evs[1].Event)
	}
}

func TestInvoiceWatch_SkipsCommentLines(t *testing.T) {
	sse := ": keepalive\n\n" +
		"event: settled\ndata: {\"number\":1,\"status\":\"settled\",\"amount\":100,\"bolt11\":\"lnbc1...\"}\n\n"
	c := sseServer(t, sseHandler(t, sse))

	evs, err := collectInvoiceEvents(c.Invoices.Watch(context.Background(), 1, nil))
	if err != nil {
		t.Fatal(err)
	}
	if len(evs) != 1 {
		t.Fatalf("len = %d, want 1", len(evs))
	}
}

func TestInvoiceWatch_EmptyStream(t *testing.T) {
	c := sseServer(t, sseHandler(t, ""))

	evs, err := collectInvoiceEvents(c.Invoices.Watch(context.Background(), 1, nil))
	if err != nil {
		t.Fatal(err)
	}
	if len(evs) != 0 {
		t.Errorf("len = %d, want 0", len(evs))
	}
}

func TestInvoiceWatch_BuildsCorrectPath(t *testing.T) {
	var gotPath, gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(200)
	}))
	t.Cleanup(srv.Close)
	c := New("key_test", WithBaseURL(srv.URL))

	evs, errs := c.Invoices.Watch(context.Background(), 42, Ptr(120))
	for range evs {
	}
	<-errs

	if gotPath != "/v1/invoices/42/events" {
		t.Errorf("Path = %q", gotPath)
	}
	if gotQuery != "timeout=120" {
		t.Errorf("Query = %q", gotQuery)
	}
}

func TestInvoiceWatch_OmitsTimeoutWhenNil(t *testing.T) {
	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(200)
	}))
	t.Cleanup(srv.Close)
	c := New("key_test", WithBaseURL(srv.URL))

	evs, errs := c.Invoices.Watch(context.Background(), 1, nil)
	for range evs {
	}
	<-errs

	if gotQuery != "" {
		t.Errorf("Query = %q, want empty", gotQuery)
	}
}

func TestInvoiceWatch_SendsSseHeaders(t *testing.T) {
	var gotAccept, gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAccept = r.Header.Get("Accept")
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(200)
	}))
	t.Cleanup(srv.Close)
	c := New("key_test", WithBaseURL(srv.URL))

	evs, errs := c.Invoices.Watch(context.Background(), 1, nil)
	for range evs {
	}
	<-errs

	if gotAccept != "text/event-stream" {
		t.Errorf("Accept = %q", gotAccept)
	}
	if gotAuth != "Bearer key_test" {
		t.Errorf("Authorization = %q", gotAuth)
	}
}

func TestInvoiceWatch_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(401)
		w.Write([]byte(`{"message":"unauthorized"}`))
	}))
	t.Cleanup(srv.Close)
	c := New("key_test", WithBaseURL(srv.URL))

	evs, errs := c.Invoices.Watch(context.Background(), 1, nil)
	for range evs {
	}

	err := <-errs
	var apiErr *UnauthorizedError
	if !errors.As(err, &apiErr) {
		t.Errorf("got %T, want *UnauthorizedError", err)
	}
}

// ---------------------------------------------------------------------------
// Invoice WatchByHash
// ---------------------------------------------------------------------------

func TestInvoiceWatchByHash_BuildsCorrectPath(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(200)
	}))
	t.Cleanup(srv.Close)
	c := New("key_test", WithBaseURL(srv.URL))

	evs, errs := c.Invoices.WatchByHash(context.Background(), "abc123", nil)
	for range evs {
	}
	<-errs

	if gotPath != "/v1/invoices/abc123/events" {
		t.Errorf("Path = %q", gotPath)
	}
}

func TestInvoiceWatchByHash_WithTimeout(t *testing.T) {
	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(200)
	}))
	t.Cleanup(srv.Close)
	c := New("key_test", WithBaseURL(srv.URL))

	evs, errs := c.Invoices.WatchByHash(context.Background(), "abc123", Ptr(60))
	for range evs {
	}
	<-errs

	if gotQuery != "timeout=60" {
		t.Errorf("Query = %q", gotQuery)
	}
}

// ---------------------------------------------------------------------------
// Payment Watch
// ---------------------------------------------------------------------------

func TestPaymentWatch_YieldsEvents(t *testing.T) {
	sse := "event: settled\ndata: {\"number\":1,\"status\":\"settled\",\"amount\":50,\"maxFee\":10,\"serviceFee\":0,\"address\":\"user@ln.bot\"}\n\n"
	c := sseServer(t, sseHandler(t, sse))

	evs, err := collectPaymentEvents(c.Payments.Watch(context.Background(), 1, nil))
	if err != nil {
		t.Fatal(err)
	}
	if len(evs) != 1 {
		t.Fatalf("len = %d, want 1", len(evs))
	}
	if evs[0].Event != "settled" {
		t.Errorf("Event = %q", evs[0].Event)
	}
	if evs[0].Data.Amount != 50 {
		t.Errorf("Amount = %d", evs[0].Data.Amount)
	}
}

func TestPaymentWatch_BuildsCorrectPath(t *testing.T) {
	var gotPath, gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(200)
	}))
	t.Cleanup(srv.Close)
	c := New("key_test", WithBaseURL(srv.URL))

	evs, errs := c.Payments.Watch(context.Background(), 7, Ptr(60))
	for range evs {
	}
	<-errs

	if gotPath != "/v1/payments/7/events" {
		t.Errorf("Path = %q", gotPath)
	}
	if gotQuery != "timeout=60" {
		t.Errorf("Query = %q", gotQuery)
	}
}

// ---------------------------------------------------------------------------
// Payment WatchByHash
// ---------------------------------------------------------------------------

func TestPaymentWatchByHash_BuildsCorrectPath(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(200)
	}))
	t.Cleanup(srv.Close)
	c := New("key_test", WithBaseURL(srv.URL))

	evs, errs := c.Payments.WatchByHash(context.Background(), "hash123", nil)
	for range evs {
	}
	<-errs

	if gotPath != "/v1/payments/hash123/events" {
		t.Errorf("Path = %q", gotPath)
	}
}

// ---------------------------------------------------------------------------
// Events Stream
// ---------------------------------------------------------------------------

func TestEventsStream_YieldsEvents(t *testing.T) {
	sse := "data: {\"event\":\"invoice.settled\",\"createdAt\":\"2024-01-01T00:00:00Z\",\"data\":{\"number\":1}}\n"
	c := sseServer(t, sseHandler(t, sse))

	evs, err := collectWalletEvents(c.Events.Stream(context.Background()))
	if err != nil {
		t.Fatal(err)
	}
	if len(evs) != 1 {
		t.Fatalf("len = %d, want 1", len(evs))
	}
	if evs[0].Event != "invoice.settled" {
		t.Errorf("Event = %q", evs[0].Event)
	}

	var data map[string]any
	json.Unmarshal(evs[0].Data, &data)
	if data["number"] != float64(1) {
		t.Errorf("data.number = %v", data["number"])
	}
}

func TestEventsStream_MultipleEvents(t *testing.T) {
	sse := "data: {\"event\":\"invoice.settled\",\"createdAt\":\"2024-01-01T00:00:00Z\",\"data\":{\"number\":1}}\n" +
		"data: {\"event\":\"payment.settled\",\"createdAt\":\"2024-01-01T00:00:00Z\",\"data\":{\"number\":2}}\n"
	c := sseServer(t, sseHandler(t, sse))

	evs, err := collectWalletEvents(c.Events.Stream(context.Background()))
	if err != nil {
		t.Fatal(err)
	}
	if len(evs) != 2 {
		t.Fatalf("len = %d, want 2", len(evs))
	}
	if evs[0].Event != "invoice.settled" {
		t.Errorf("evs[0].Event = %q", evs[0].Event)
	}
	if evs[1].Event != "payment.settled" {
		t.Errorf("evs[1].Event = %q", evs[1].Event)
	}
}

func TestEventsStream_SkipsNonDataLines(t *testing.T) {
	sse := ": keepalive\n" +
		"event: ignored\n" +
		"data: {\"event\":\"payment.settled\",\"createdAt\":\"2024-01-01T00:00:00Z\",\"data\":{\"number\":1}}\n"
	c := sseServer(t, sseHandler(t, sse))

	evs, err := collectWalletEvents(c.Events.Stream(context.Background()))
	if err != nil {
		t.Fatal(err)
	}
	if len(evs) != 1 {
		t.Fatalf("len = %d, want 1", len(evs))
	}
	if evs[0].Event != "payment.settled" {
		t.Errorf("Event = %q", evs[0].Event)
	}
}

func TestEventsStream_BuildsCorrectPath(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(200)
	}))
	t.Cleanup(srv.Close)
	c := New("key_test", WithBaseURL(srv.URL))

	evs, errs := c.Events.Stream(context.Background())
	for range evs {
	}
	<-errs

	if !strings.HasSuffix(gotPath, "/v1/events") {
		t.Errorf("Path = %q", gotPath)
	}
}

func TestEventsStream_EmptyStream(t *testing.T) {
	c := sseServer(t, sseHandler(t, ""))

	evs, err := collectWalletEvents(c.Events.Stream(context.Background()))
	if err != nil {
		t.Fatal(err)
	}
	if len(evs) != 0 {
		t.Errorf("len = %d, want 0", len(evs))
	}
}

func TestEventsStream_SendsSseHeaders(t *testing.T) {
	var gotAccept string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAccept = r.Header.Get("Accept")
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(200)
	}))
	t.Cleanup(srv.Close)
	c := New("key_test", WithBaseURL(srv.URL))

	evs, errs := c.Events.Stream(context.Background())
	for range evs {
	}
	<-errs

	if gotAccept != "text/event-stream" {
		t.Errorf("Accept = %q", gotAccept)
	}
}

func TestEventsStream_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(403)
		w.Write([]byte(`{"message":"forbidden"}`))
	}))
	t.Cleanup(srv.Close)
	c := New("key_test", WithBaseURL(srv.URL))

	evs, errs := c.Events.Stream(context.Background())
	for range evs {
	}

	err := <-errs
	var apiErr *ForbiddenError
	if !errors.As(err, &apiErr) {
		t.Errorf("got %T, want *ForbiddenError", err)
	}
}
