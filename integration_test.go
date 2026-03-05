//go:build integration

package lnbot

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

// Integration tests — run against the live API with real sats.
//
// Required environment variables:
//   LNBOT_USER_KEY=uk_...   # user key that owns the prefunded wallet
//   LNBOT_WALLET_ID=wal_... # prefunded wallet ID
//
// Run:
//   go test -tags=integration -v -count=1 -timeout=120s

var (
	w1Client   *Client
	w1WalletID string
	w1Wallet   *WalletHandle

	w2Client   *Client
	w2WalletID string
	w2Wallet   *WalletHandle
	w2WalletKey string

	w1BalanceBefore int64
)

func TestMain(m *testing.M) {
	userKey := os.Getenv("LNBOT_USER_KEY")
	walletID := os.Getenv("LNBOT_WALLET_ID")
	if userKey == "" || walletID == "" {
		fmt.Println("SKIP: LNBOT_USER_KEY and LNBOT_WALLET_ID required")
		os.Exit(0)
	}

	ctx := context.Background()
	w1Client = New(userKey)
	w1WalletID = walletID
	w1Wallet = w1Client.Wallet(walletID)

	// Record w1 balance
	wal, err := w1Wallet.Get(ctx)
	if err != nil {
		fmt.Printf("FATAL: cannot get w1 balance: %v\n", err)
		os.Exit(1)
	}
	w1BalanceBefore = wal.Balance

	// Create w2 account + wallet
	anon := New("")
	account, err := anon.Register(ctx)
	if err != nil {
		fmt.Printf("FATAL: cannot register w2 account: %v\n", err)
		os.Exit(1)
	}
	w2Client = New(account.PrimaryKey)
	w2Resp, err := w2Client.Wallets.Create(ctx)
	if err != nil {
		fmt.Printf("FATAL: cannot create w2 wallet: %v\n", err)
		os.Exit(1)
	}
	w2WalletID = w2Resp.WalletID
	w2Wallet = w2Client.Wallet(w2WalletID)

	// Create wallet key for w2 (needed for SSE)
	wk, err := w2Wallet.Key.Create(ctx)
	if err != nil {
		fmt.Printf("FATAL: cannot create w2 wallet key: %v\n", err)
		os.Exit(1)
	}
	w2WalletKey = wk.Key

	os.Exit(m.Run())
}

// ── Account ─────────────────────────────────────────────

func TestInteg_Register(t *testing.T) {
	anon := New("")
	res, err := anon.Register(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(res.PrimaryKey, "uk_") {
		t.Errorf("PrimaryKey = %q, want uk_ prefix", res.PrimaryKey)
	}
	if !strings.HasPrefix(res.SecondaryKey, "uk_") {
		t.Errorf("SecondaryKey = %q, want uk_ prefix", res.SecondaryKey)
	}
	words := strings.Fields(res.RecoveryPassphrase)
	if len(words) != 12 {
		t.Errorf("RecoveryPassphrase has %d words, want 12", len(words))
	}
}

func TestInteg_Me_UserKey(t *testing.T) {
	res, err := w1Client.Me(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if res == nil {
		t.Error("Me() returned nil")
	}
}

func TestInteg_Me_WalletKey(t *testing.T) {
	wkClient := New(w2WalletKey)
	res, err := wkClient.Me(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if res.WalletID != w2WalletID {
		t.Errorf("WalletID = %q, want %q", res.WalletID, w2WalletID)
	}
}

func TestInteg_Me_InvalidKey(t *testing.T) {
	bad := New("uk_invalid")
	_, err := bad.Me(context.Background())
	var apiErr *UnauthorizedError
	if !errors.As(err, &apiErr) {
		t.Errorf("expected UnauthorizedError, got %T", err)
	}
}

// ── Wallets ─────────────────────────────────────────────

func TestInteg_Wallets_Create(t *testing.T) {
	// w2 wallet was created in TestMain
	if w2WalletID == "" {
		t.Fatal("w2WalletID is empty")
	}
	if !strings.HasPrefix(w2WalletID, "wal_") {
		t.Errorf("WalletID = %q, want wal_ prefix", w2WalletID)
	}
}

func TestInteg_Wallets_List(t *testing.T) {
	wallets, err := w2Client.Wallets.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(wallets) < 1 {
		t.Error("expected at least 1 wallet")
	}
	found := false
	for _, w := range wallets {
		if w.WalletID == w2WalletID {
			found = true
		}
	}
	if !found {
		t.Errorf("w2 wallet %q not found in list", w2WalletID)
	}
}

func TestInteg_Wallet_Get(t *testing.T) {
	wal, err := w1Wallet.Get(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if wal.WalletID != w1WalletID {
		t.Errorf("WalletID = %q, want %q", wal.WalletID, w1WalletID)
	}
}

func TestInteg_Wallet_Update(t *testing.T) {
	name := fmt.Sprintf("go-test-%d", time.Now().Unix())
	wal, err := w2Wallet.Update(context.Background(), &UpdateWalletParams{Name: name})
	if err != nil {
		t.Fatal(err)
	}
	if wal.Name != name {
		t.Errorf("Name = %q, want %q", wal.Name, name)
	}
}

func TestInteg_Wallet_Get_Nonexistent(t *testing.T) {
	_, err := w1Client.Wallet("wal_nonexistent").Get(context.Background())
	var notFound *NotFoundError
	if !errors.As(err, &notFound) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

// ── Wallet Keys ─────────────────────────────────────────

func TestInteg_WalletKey_DuplicateRejected(t *testing.T) {
	// w2 already has a key from TestMain
	_, err := w2Wallet.Key.Create(context.Background())
	var conflict *ConflictError
	if !errors.As(err, &conflict) {
		t.Errorf("expected ConflictError, got %T: %v", err, err)
	}
}

func TestInteg_WalletKey_Get(t *testing.T) {
	info, err := w2Wallet.Key.Get(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if info.Hint == "" {
		t.Error("Hint is empty")
	}
}

func TestInteg_WalletKey_Rotate(t *testing.T) {
	res, err := w2Wallet.Key.Rotate(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(res.Key, "wk_") {
		t.Errorf("Key = %q, want wk_ prefix", res.Key)
	}
	w2WalletKey = res.Key
}

func TestInteg_WalletKey_CurrentAlias(t *testing.T) {
	wkClient := New(w2WalletKey)
	wal, err := wkClient.Wallet("current").Get(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if wal.WalletID != w2WalletID {
		t.Errorf("WalletID = %q, want %q", wal.WalletID, w2WalletID)
	}
}

// ── Addresses ───────────────────────────────────────────

func TestInteg_Addresses_List(t *testing.T) {
	addrs, err := w2Wallet.Addresses.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(addrs) < 1 {
		t.Error("expected at least 1 address (generated)")
	}
}

func TestInteg_Addresses_Create(t *testing.T) {
	// Create a random (generated) address — no cost
	addr, err := w2Wallet.Addresses.Create(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(addr.Address, "@") {
		t.Errorf("Address = %q, expected to contain @", addr.Address)
	}
	if !addr.Generated {
		t.Error("expected Generated = true")
	}
	if addr.Cost != 0 {
		t.Errorf("Cost = %d, expected 0 for generated address", addr.Cost)
	}
}

func TestInteg_Addresses_TransferRejectsGenerated(t *testing.T) {
	addrs, err := w2Wallet.Addresses.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	var generated string
	for _, a := range addrs {
		if a.Generated {
			generated = a.Address
			break
		}
	}
	if generated == "" {
		t.Skip("no generated address found")
	}
	_, err = w2Wallet.Addresses.Transfer(context.Background(), generated, &TransferAddressParams{
		TargetWalletKey: "wk_dummy",
	})
	var badReq *BadRequestError
	if !errors.As(err, &badReq) {
		t.Errorf("expected BadRequestError, got %T: %v", err, err)
	}
}

// ── Invoices ────────────────────────────────────────────

func TestInteg_Invoices_Create(t *testing.T) {
	inv, err := w2Wallet.Invoices.Create(context.Background(), &CreateInvoiceParams{
		Amount: 2, Memo: Ptr("go-sdk-test"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if inv.Number < 1 {
		t.Errorf("Number = %d", inv.Number)
	}
	if inv.Status != "pending" {
		t.Errorf("Status = %q", inv.Status)
	}
	if !strings.HasPrefix(inv.Bolt11, "lnbc") {
		t.Errorf("Bolt11 = %q, want lnbc prefix", inv.Bolt11)
	}
}

func TestInteg_Invoices_ZeroAmountRejected(t *testing.T) {
	_, err := w2Wallet.Invoices.Create(context.Background(), &CreateInvoiceParams{Amount: 0})
	var badReq *BadRequestError
	if !errors.As(err, &badReq) {
		t.Errorf("expected BadRequestError, got %T", err)
	}
}

func TestInteg_Invoices_List(t *testing.T) {
	invs, err := w2Wallet.Invoices.List(context.Background(), &ListInvoicesParams{Limit: Ptr(5)})
	if err != nil {
		t.Fatal(err)
	}
	if len(invs) < 1 {
		t.Error("expected at least 1 invoice")
	}
}

func TestInteg_Invoices_Get(t *testing.T) {
	invs, err := w2Wallet.Invoices.List(context.Background(), &ListInvoicesParams{Limit: Ptr(1)})
	if err != nil {
		t.Fatal(err)
	}
	if len(invs) < 1 {
		t.Skip("no invoices")
	}
	inv, err := w2Wallet.Invoices.Get(context.Background(), invs[0].Number)
	if err != nil {
		t.Fatal(err)
	}
	if inv.Number != invs[0].Number {
		t.Errorf("Number = %d, want %d", inv.Number, invs[0].Number)
	}
}

func TestInteg_Invoices_GetNonexistent(t *testing.T) {
	_, err := w2Wallet.Invoices.Get(context.Background(), 999999)
	var notFound *NotFoundError
	if !errors.As(err, &notFound) {
		t.Errorf("expected NotFoundError, got %T", err)
	}
}

func TestInteg_PublicInvoice_ForWallet(t *testing.T) {
	anon := New("")
	inv, err := anon.Invoices.CreateForWallet(context.Background(), &CreateInvoiceForWalletParams{
		WalletID: w1WalletID, Amount: 5,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(inv.Bolt11, "lnbc") {
		t.Errorf("Bolt11 = %q", inv.Bolt11)
	}
}

func TestInteg_PublicInvoice_ForAddress(t *testing.T) {
	addrs, err := w1Wallet.Addresses.List(context.Background())
	if err != nil || len(addrs) == 0 {
		t.Skip("no addresses on w1")
	}
	anon := New("")
	inv, err := anon.Invoices.CreateForAddress(context.Background(), &CreateInvoiceForAddressParams{
		Address: addrs[0].Address, Amount: 5,
	})
	if err != nil {
		t.Fatal(err)
	}
	if inv.Amount != 5 {
		t.Errorf("Amount = %d", inv.Amount)
	}
}

func TestInteg_PublicInvoice_NonexistentWallet(t *testing.T) {
	anon := New("")
	_, err := anon.Invoices.CreateForWallet(context.Background(), &CreateInvoiceForWalletParams{
		WalletID: "wal_nonexistent", Amount: 1,
	})
	var badReq *BadRequestError
	if !errors.As(err, &badReq) {
		t.Errorf("expected BadRequestError, got %T", err)
	}
}

// ── Payments + Balance ──────────────────────────────────

func TestInteg_Payments_Resolve(t *testing.T) {
	addrs, err := w2Wallet.Addresses.List(context.Background())
	if err != nil || len(addrs) == 0 {
		t.Skip("no w2 addresses")
	}
	res, err := w1Wallet.Payments.Resolve(context.Background(), addrs[0].Address)
	if err != nil {
		t.Fatal(err)
	}
	if res.Type != "lightning_address" {
		t.Errorf("Type = %q, want lightning_address", res.Type)
	}
	if res.Min == nil || res.Max == nil || res.Fixed == nil {
		t.Error("expected Min, Max, Fixed to be set")
	}
}

func TestInteg_Payments_ResolveBolt11(t *testing.T) {
	inv, err := w2Wallet.Invoices.Create(context.Background(), &CreateInvoiceParams{
		Amount: 2, Memo: Ptr("resolve-bolt11-test"),
	})
	if err != nil {
		t.Fatal(err)
	}
	res, err := w1Wallet.Payments.Resolve(context.Background(), inv.Bolt11)
	if err != nil {
		t.Fatal(err)
	}
	if res.Type != "bolt11" {
		t.Errorf("Type = %q, want bolt11", res.Type)
	}
	if res.Amount == nil || *res.Amount != 2 {
		t.Errorf("Amount = %v, want 2", res.Amount)
	}
	if res.Fixed == nil || !*res.Fixed {
		t.Error("expected Fixed = true for bolt11 with amount")
	}
}

func waitForPayment(t *testing.T, wallet *WalletHandle, number int) *Payment {
	t.Helper()
	for i := 0; i < 30; i++ {
		time.Sleep(500 * time.Millisecond)
		p, err := wallet.Payments.Get(context.Background(), number)
		if err != nil {
			t.Fatal(err)
		}
		if p.Status == "settled" || p.Status == "failed" {
			return p
		}
	}
	p, _ := wallet.Payments.Get(context.Background(), number)
	return p
}

func TestInteg_Payments_CreateAndSettle(t *testing.T) {
	inv, err := w2Wallet.Invoices.Create(context.Background(), &CreateInvoiceParams{
		Amount: 2, Memo: Ptr("payment-settle-test"),
	})
	if err != nil {
		t.Fatal(err)
	}

	pay, err := w1Wallet.Payments.Create(context.Background(), &CreatePaymentParams{
		Target: inv.Bolt11,
	})
	if err != nil {
		t.Fatal(err)
	}
	if pay.Amount != 2 {
		t.Errorf("Amount = %d, want 2", pay.Amount)
	}

	settled := waitForPayment(t, w1Wallet, pay.Number)
	if settled.Status != "settled" {
		t.Fatalf("payment did not settle, status = %q", settled.Status)
	}
	if settled.Preimage == nil || *settled.Preimage == "" {
		t.Error("Preimage is empty after settlement")
	}
	if settled.SettledAt == nil {
		t.Error("SettledAt is nil after settlement")
	}
}

func TestInteg_Payments_BalanceUpdated(t *testing.T) {
	w1After, err := w1Wallet.Get(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if w1After.Balance >= w1BalanceBefore {
		t.Error("w1 balance should have decreased after payment")
	}

	w2After, err := w2Wallet.Get(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if w2After.Balance < 2 {
		t.Errorf("w2 balance = %d, expected >= 2", w2After.Balance)
	}
}

func TestInteg_Payments_InvoiceSettledOnReceiver(t *testing.T) {
	invs, err := w2Wallet.Invoices.List(context.Background(), &ListInvoicesParams{Limit: Ptr(10)})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, inv := range invs {
		if inv.Status == "settled" && inv.Amount == 2 {
			found = true
			if inv.SettledAt == nil {
				t.Error("settled invoice has nil SettledAt")
			}
			if inv.TxNumber == nil {
				t.Error("settled invoice has nil TxNumber")
			}
			break
		}
	}
	if !found {
		t.Error("expected to find a settled 2-sat invoice on w2")
	}
}

func TestInteg_Payments_List(t *testing.T) {
	ps, err := w1Wallet.Payments.List(context.Background(), &ListPaymentsParams{Limit: Ptr(5)})
	if err != nil {
		t.Fatal(err)
	}
	if len(ps) < 1 {
		t.Error("expected at least 1 payment")
	}

	// Pagination
	if len(ps) >= 2 {
		page, err := w1Wallet.Payments.List(context.Background(), &ListPaymentsParams{
			Limit: Ptr(1), After: Ptr(ps[0].Number),
		})
		if err != nil {
			t.Fatal(err)
		}
		if len(page) < 1 {
			t.Error("expected at least 1 result in paginated query")
		}
		if page[0].Number >= ps[0].Number {
			t.Errorf("pagination: got number %d, expected < %d", page[0].Number, ps[0].Number)
		}
	}
}

func TestInteg_Payments_Get(t *testing.T) {
	ps, err := w1Wallet.Payments.List(context.Background(), &ListPaymentsParams{Limit: Ptr(1)})
	if err != nil || len(ps) == 0 {
		t.Skip("no payments")
	}
	p, err := w1Wallet.Payments.Get(context.Background(), ps[0].Number)
	if err != nil {
		t.Fatal(err)
	}
	if p.Number != ps[0].Number {
		t.Errorf("Number = %d, want %d", p.Number, ps[0].Number)
	}
	if p.Status != "settled" {
		t.Errorf("Status = %q, want settled", p.Status)
	}
}

func TestInteg_Payments_GetNonexistent(t *testing.T) {
	_, err := w1Wallet.Payments.Get(context.Background(), 999999)
	var notFound *NotFoundError
	if !errors.As(err, &notFound) {
		t.Errorf("expected NotFoundError, got %T", err)
	}
}

func TestInteg_Payments_InsufficientBalance(t *testing.T) {
	inv, err := w1Wallet.Invoices.Create(context.Background(), &CreateInvoiceParams{
		Amount: 99999999,
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = w2Wallet.Payments.Create(context.Background(), &CreatePaymentParams{
		Target: inv.Bolt11,
	})
	var badReq *BadRequestError
	if !errors.As(err, &badReq) {
		t.Errorf("expected BadRequestError, got %T: %v", err, err)
	}
}

func TestInteg_Payments_IdempotencyKey(t *testing.T) {
	inv, err := w1Wallet.Invoices.Create(context.Background(), &CreateInvoiceParams{
		Amount: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	idemKey := fmt.Sprintf("idem-go-%d", time.Now().UnixNano())

	// First payment
	p1, err := w2Wallet.Payments.Create(context.Background(), &CreatePaymentParams{
		Target:         inv.Bolt11,
		IdempotencyKey: Ptr(idemKey),
	})
	if err != nil {
		t.Fatal(err)
	}
	waitForPayment(t, w2Wallet, p1.Number)

	// Second payment with same key — should return same payment
	p2, err := w2Wallet.Payments.Create(context.Background(), &CreatePaymentParams{
		Target:         inv.Bolt11,
		IdempotencyKey: Ptr(idemKey),
	})
	if err != nil {
		t.Fatal(err)
	}
	if p2.Number != p1.Number {
		t.Errorf("idempotency failed: p1.Number=%d, p2.Number=%d", p1.Number, p2.Number)
	}
}

// ── Transactions ────────────────────────────────────────

func TestInteg_Transactions_DebitOnSender(t *testing.T) {
	txs, err := w1Wallet.Transactions.List(context.Background(), &ListTransactionsParams{Limit: Ptr(10)})
	if err != nil {
		t.Fatal(err)
	}
	if len(txs) < 1 {
		t.Error("expected at least 1 transaction")
	}
	found := false
	for _, tx := range txs {
		if tx.Type == "debit" {
			found = true
			if tx.Amount <= 0 {
				t.Errorf("debit Amount = %d, expected > 0", tx.Amount)
			}
			if tx.CreatedAt == nil {
				t.Error("CreatedAt is nil")
			}
			break
		}
	}
	if !found {
		t.Error("expected at least one debit transaction on w1")
	}
}

func TestInteg_Transactions_CreditOnReceiver(t *testing.T) {
	txs, err := w2Wallet.Transactions.List(context.Background(), &ListTransactionsParams{Limit: Ptr(10)})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, tx := range txs {
		if tx.Type == "credit" {
			found = true
			if tx.Amount <= 0 {
				t.Errorf("credit Amount = %d, expected > 0", tx.Amount)
			}
			break
		}
	}
	if !found {
		t.Error("expected at least one credit transaction on w2")
	}
}

func TestInteg_Transactions_Pagination(t *testing.T) {
	txs, err := w1Wallet.Transactions.List(context.Background(), &ListTransactionsParams{Limit: Ptr(1)})
	if err != nil {
		t.Fatal(err)
	}
	if len(txs) != 1 {
		t.Skipf("expected 1 transaction, got %d", len(txs))
	}
	next, err := w1Wallet.Transactions.List(context.Background(), &ListTransactionsParams{
		Limit: Ptr(1), After: Ptr(txs[0].Number),
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(next) > 0 && next[0].Number >= txs[0].Number {
		t.Errorf("pagination: got number %d, expected < %d", next[0].Number, txs[0].Number)
	}
}

// ── Webhooks ────────────────────────────────────────────

func TestInteg_Webhooks_CRUD(t *testing.T) {
	wh, err := w2Wallet.Webhooks.Create(context.Background(), &CreateWebhookParams{
		URL: "https://example.com/webhook-test",
	})
	if err != nil {
		t.Fatal(err)
	}
	if wh.Secret == "" {
		t.Error("Secret is empty")
	}

	// List
	whs, err := w2Wallet.Webhooks.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, w := range whs {
		if w.ID == wh.ID {
			found = true
		}
	}
	if !found {
		t.Error("webhook not found in list")
	}

	// Delete
	err = w2Wallet.Webhooks.Delete(context.Background(), wh.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func TestInteg_Webhooks_DeleteNonexistent(t *testing.T) {
	err := w2Wallet.Webhooks.Delete(context.Background(), "wh_nonexistent")
	var notFound *NotFoundError
	if !errors.As(err, &notFound) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

// ── L402 ────────────────────────────────────────────────

func TestInteg_L402_CreateChallenge(t *testing.T) {
	ch, err := w2Wallet.L402.CreateChallenge(context.Background(), &CreateL402ChallengeParams{
		Amount:      1,
		Description: Ptr("test"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if ch.Macaroon == "" {
		t.Error("Macaroon is empty")
	}
	if ch.Invoice == "" {
		t.Error("Invoice is empty")
	}
	if ch.WwwAuthenticate == "" {
		t.Error("WwwAuthenticate is empty")
	}
}

func TestInteg_L402_InvalidTokenRejected(t *testing.T) {
	_, err := w2Wallet.L402.Verify(context.Background(), &VerifyL402Params{
		Authorization: "L402 invalid:invalid",
	})
	var badReq *BadRequestError
	if !errors.As(err, &badReq) {
		t.Errorf("expected BadRequestError, got %T: %v", err, err)
	}
}

func TestInteg_L402_FullFlow(t *testing.T) {
	// Create challenge on w2
	ch, err := w2Wallet.L402.CreateChallenge(context.Background(), &CreateL402ChallengeParams{
		Amount: 1,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Pay from w1
	res, err := w1Wallet.L402.Pay(context.Background(), &PayL402Params{
		WwwAuthenticate: ch.WwwAuthenticate,
		Wait:            Ptr(true),
		Timeout:         Ptr(30),
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Authorization == nil || *res.Authorization == "" {
		t.Fatal("Authorization is empty")
	}

	// Verify on w2
	v, err := w2Wallet.L402.Verify(context.Background(), &VerifyL402Params{
		Authorization: *res.Authorization,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !v.Valid {
		t.Error("expected valid L402 token")
	}
}

// ── SSE ─────────────────────────────────────────────────

func TestInteg_SSE_InvoiceWatch(t *testing.T) {
	// Create invoice on w2
	inv, err := w2Wallet.Invoices.Create(context.Background(), &CreateInvoiceParams{
		Amount: 1, Memo: Ptr("sse-test"),
	})
	if err != nil {
		t.Fatal(err)
	}

	// Watch with wallet key
	wkClient := New(w2WalletKey)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	events, errs := wkClient.Wallet(w2WalletID).Invoices.Watch(ctx, inv.Number, Ptr(10))

	// Pay from w1
	go func() {
		time.Sleep(500 * time.Millisecond)
		w1Wallet.Payments.Create(context.Background(), &CreatePaymentParams{
			Target: inv.Bolt11,
		})
	}()

	settled := false
	for ev := range events {
		if ev.Event == "settled" {
			settled = true
			cancel()
		}
	}
	if err := <-errs; err != nil && !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if !settled {
		t.Error("did not receive settled event")
	}
}

func TestInteg_SSE_PaymentWatch(t *testing.T) {
	inv, err := w1Wallet.Invoices.Create(context.Background(), &CreateInvoiceParams{
		Amount: 1, Memo: Ptr("payment-watch-test"),
	})
	if err != nil {
		t.Fatal(err)
	}

	// Start payment from w2
	pay, err := w2Wallet.Payments.Create(context.Background(), &CreatePaymentParams{
		Target: inv.Bolt11,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Watch with wallet key
	wkClient := New(w2WalletKey)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	events, errs := wkClient.Wallet(w2WalletID).Payments.Watch(ctx, pay.Number, Ptr(30))

	settled := false
	for ev := range events {
		if ev.Event == "settled" {
			settled = true
			cancel()
		}
	}
	if err := <-errs; err != nil && !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if !settled {
		t.Error("did not receive settled event for payment")
	}
}

func TestInteg_SSE_WalletEventStream(t *testing.T) {
	wkClient := New(w2WalletKey)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	events, errs := wkClient.Wallet(w2WalletID).Events.Stream(ctx)

	// Generate events: create and pay an invoice
	go func() {
		time.Sleep(1500 * time.Millisecond)
		inv, err := w2Wallet.Invoices.Create(context.Background(), &CreateInvoiceParams{
			Amount: 1, Memo: Ptr("event-stream-test"),
		})
		if err != nil {
			return
		}
		w1Wallet.Payments.Create(context.Background(), &CreatePaymentParams{
			Target: inv.Bolt11,
		})
	}()

	var collected []WalletEvent
	for ev := range events {
		collected = append(collected, ev)
		if len(collected) >= 2 {
			cancel()
			break
		}
	}
	if err := <-errs; err != nil && !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if len(collected) < 1 {
		t.Error("expected at least 1 wallet event")
	}
	hasInvoiceEvent := false
	for _, ev := range collected {
		if strings.HasPrefix(ev.Event, "invoice.") {
			hasInvoiceEvent = true
		}
	}
	if !hasInvoiceEvent {
		t.Errorf("expected invoice.* event, got events: %v", func() []string {
			names := make([]string, len(collected))
			for i, e := range collected {
				names[i] = e.Event
			}
			return names
		}())
	}
}

// ── Backup ──────────────────────────────────────────────

func TestInteg_Backup_Recovery(t *testing.T) {
	res, err := w2Client.Backup.Recovery(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	words := strings.Fields(res.Passphrase)
	if len(words) != 12 {
		t.Errorf("passphrase has %d words, want 12", len(words))
	}
}

// ── Error handling ──────────────────────────────────────

func TestInteg_Error_Unauthenticated(t *testing.T) {
	anon := New("")
	_, err := anon.Wallet(w1WalletID).Get(context.Background())
	var unauth *UnauthorizedError
	if !errors.As(err, &unauth) {
		t.Errorf("expected UnauthorizedError, got %T", err)
	}
}

func TestInteg_Error_WrongWalletID(t *testing.T) {
	_, err := w1Client.Wallet("wal_nonexistent").Get(context.Background())
	var notFound *NotFoundError
	if !errors.As(err, &notFound) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestInteg_Error_CrossUserAccess(t *testing.T) {
	// w2 client tries to access w1 wallet
	_, err := w2Client.Wallet(w1WalletID).Get(context.Background())
	if err == nil {
		t.Error("expected error for cross-user access")
	}
}

// ── Cleanup ─────────────────────────────────────────────

func TestInteg_Cleanup_ReturnFunds(t *testing.T) {
	wal, err := w2Wallet.Get(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if wal.Available <= 0 {
		t.Skip("no funds to return")
	}

	inv, err := w1Wallet.Invoices.Create(context.Background(), &CreateInvoiceParams{
		Amount: wal.Available,
	})
	if err != nil {
		t.Fatal(err)
	}
	p, err := w2Wallet.Payments.Create(context.Background(), &CreatePaymentParams{
		Target: inv.Bolt11,
	})
	if err != nil {
		t.Logf("return funds: %v (non-fatal)", err)
		return
	}
	settled := waitForPayment(t, w2Wallet, p.Number)
	if settled.Status != "settled" {
		t.Logf("return funds payment status: %s (non-fatal)", settled.Status)
	}

	// Verify w2 is empty
	after, err := w2Wallet.Get(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if after.Balance != 0 {
		t.Logf("w2 balance after return = %d (expected 0)", after.Balance)
	}
}

func TestInteg_Cleanup_BalanceRestored(t *testing.T) {
	w1After, err := w1Wallet.Get(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	// Balance may differ slightly due to service fees on internal transfers
	if w1After.Balance < w1BalanceBefore-10 {
		t.Errorf("w1 balance = %d, started at %d (lost more than 10 sats)", w1After.Balance, w1BalanceBefore)
	}
}

func TestInteg_Cleanup_DeleteAddress(t *testing.T) {
	addrs, err := w2Wallet.Addresses.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(addrs) == 0 {
		t.Skip("no addresses to delete")
	}
	// Delete one address, then verify double-delete returns NotFound
	target := addrs[0].Address
	err = w2Wallet.Addresses.Delete(context.Background(), target)
	if err != nil {
		t.Fatal(err)
	}

	// Verify it's gone
	after, _ := w2Wallet.Addresses.List(context.Background())
	for _, a := range after {
		if a.Address == target {
			t.Errorf("address %s still in list after delete", target)
		}
	}

	// Second delete should fail
	err = w2Wallet.Addresses.Delete(context.Background(), target)
	var notFound *NotFoundError
	if !errors.As(err, &notFound) {
		t.Errorf("expected NotFoundError on double-delete, got %T: %v", err, err)
	}
}

func TestInteg_Cleanup_DeleteWalletKey(t *testing.T) {
	err := w2Wallet.Key.Delete(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	// Verify key is revoked
	wkClient := New(w2WalletKey)
	_, err = wkClient.Me(context.Background())
	var unauth *UnauthorizedError
	if !errors.As(err, &unauth) {
		t.Errorf("expected UnauthorizedError after key deletion, got %T", err)
	}
}

func TestInteg_Cleanup_WalletKeyGetAfterDelete(t *testing.T) {
	_, err := w2Wallet.Key.Get(context.Background())
	var notFound *NotFoundError
	if !errors.As(err, &notFound) {
		t.Errorf("expected NotFoundError after key deletion, got %T: %v", err, err)
	}
}
