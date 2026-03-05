package lnbot

import (
	"context"
	"testing"
)

// ---------------------------------------------------------------------------
// Wallets
// ---------------------------------------------------------------------------

func TestWallets_Create(t *testing.T) {
	h, cap := jsonHandler(200, map[string]any{
		"walletId": "wal_abc", "name": "Test", "address": "test@ln.bot",
	})
	c, _ := testServer(t, h)

	res, err := c.Wallets.Create(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if cap.Method != "POST" {
		t.Errorf("Method = %q, want POST", cap.Method)
	}
	if cap.Path != "/v1/wallets" {
		t.Errorf("Path = %q", cap.Path)
	}
	if res.WalletID != "wal_abc" {
		t.Errorf("WalletID = %q", res.WalletID)
	}
	if res.Address != "test@ln.bot" {
		t.Errorf("Address = %q", res.Address)
	}
}

func TestWallets_List(t *testing.T) {
	h, cap := jsonHandler(200, []map[string]any{
		{"walletId": "wal_1", "name": "Wallet One"},
		{"walletId": "wal_2", "name": "Wallet Two"},
	})
	c, _ := testServer(t, h)

	items, err := c.Wallets.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if cap.Method != "GET" {
		t.Errorf("Method = %q, want GET", cap.Method)
	}
	if cap.Path != "/v1/wallets" {
		t.Errorf("Path = %q", cap.Path)
	}
	if len(items) != 2 {
		t.Fatalf("len = %d, want 2", len(items))
	}
	if items[0].Name != "Wallet One" {
		t.Errorf("Name = %q", items[0].Name)
	}
}

// ---------------------------------------------------------------------------
// Wallet Handle (Get / Update)
// ---------------------------------------------------------------------------

func TestWallet_Get(t *testing.T) {
	h, cap := jsonHandler(200, map[string]any{
		"walletId": "wal_1", "name": "My Wallet", "balance": 1000, "onHold": 50, "available": 950,
	})
	c, _ := testServer(t, h)

	w, err := c.Wallet("wal_1").Get(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if cap.Method != "GET" {
		t.Errorf("Method = %q, want GET", cap.Method)
	}
	if cap.Path != "/v1/wallets/wal_1" {
		t.Errorf("Path = %q", cap.Path)
	}
	if w.Name != "My Wallet" {
		t.Errorf("Name = %q", w.Name)
	}
	if w.Available != 950 {
		t.Errorf("Available = %d", w.Available)
	}
}

func TestWallet_Update(t *testing.T) {
	h, cap := jsonHandler(200, map[string]any{
		"walletId": "wal_1", "name": "Renamed", "balance": 0, "onHold": 0, "available": 0,
	})
	c, _ := testServer(t, h)

	w, err := c.Wallet("wal_1").Update(context.Background(), &UpdateWalletParams{Name: "Renamed"})
	if err != nil {
		t.Fatal(err)
	}
	if cap.Method != "PATCH" {
		t.Errorf("Method = %q, want PATCH", cap.Method)
	}
	if cap.Path != "/v1/wallets/wal_1" {
		t.Errorf("Path = %q", cap.Path)
	}
	if w.Name != "Renamed" {
		t.Errorf("Name = %q", w.Name)
	}
}

// ---------------------------------------------------------------------------
// Wallet Key
// ---------------------------------------------------------------------------

func TestWalletKey_Create(t *testing.T) {
	h, cap := jsonHandler(200, map[string]any{"key": "wk_abc", "hint": "wk_a..."})
	c, _ := testServer(t, h)

	k, err := c.Wallet("wal_1").Key.Create(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if cap.Method != "POST" {
		t.Errorf("Method = %q", cap.Method)
	}
	if cap.Path != "/v1/wallets/wal_1/key" {
		t.Errorf("Path = %q", cap.Path)
	}
	if k.Key != "wk_abc" {
		t.Errorf("Key = %q", k.Key)
	}
}

func TestWalletKey_Get(t *testing.T) {
	h, cap := jsonHandler(200, map[string]any{"hint": "wk_a..."})
	c, _ := testServer(t, h)

	k, err := c.Wallet("wal_1").Key.Get(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if cap.Path != "/v1/wallets/wal_1/key" {
		t.Errorf("Path = %q", cap.Path)
	}
	if k.Hint != "wk_a..." {
		t.Errorf("Hint = %q", k.Hint)
	}
}

func TestWalletKey_Delete(t *testing.T) {
	h, cap := jsonHandler(204, nil)
	c, _ := testServer(t, h)

	err := c.Wallet("wal_1").Key.Delete(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if cap.Method != "DELETE" {
		t.Errorf("Method = %q", cap.Method)
	}
	if cap.Path != "/v1/wallets/wal_1/key" {
		t.Errorf("Path = %q", cap.Path)
	}
}

func TestWalletKey_Rotate(t *testing.T) {
	h, cap := jsonHandler(200, map[string]any{"key": "wk_new", "hint": "wk_n..."})
	c, _ := testServer(t, h)

	k, err := c.Wallet("wal_1").Key.Rotate(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if cap.Method != "POST" {
		t.Errorf("Method = %q", cap.Method)
	}
	if cap.Path != "/v1/wallets/wal_1/key/rotate" {
		t.Errorf("Path = %q", cap.Path)
	}
	if k.Key != "wk_new" {
		t.Errorf("Key = %q", k.Key)
	}
}

// ---------------------------------------------------------------------------
// Keys (user key rotation)
// ---------------------------------------------------------------------------

func TestKeys_Rotate(t *testing.T) {
	h, cap := jsonHandler(200, map[string]any{"key": "uk_new", "name": "primary"})
	c, _ := testServer(t, h)

	k, err := c.Keys.Rotate(context.Background(), 0)
	if err != nil {
		t.Fatal(err)
	}
	if cap.Method != "POST" {
		t.Errorf("Method = %q, want POST", cap.Method)
	}
	if cap.Path != "/v1/keys/0/rotate" {
		t.Errorf("Path = %q", cap.Path)
	}
	if k.Key != "uk_new" {
		t.Errorf("Key = %q", k.Key)
	}
	if k.Name != "primary" {
		t.Errorf("Name = %q", k.Name)
	}
}

// ---------------------------------------------------------------------------
// Invoices (wallet-scoped)
// ---------------------------------------------------------------------------

func TestInvoices_Create(t *testing.T) {
	h, cap := jsonHandler(200, map[string]any{
		"number": 1, "status": "pending", "amount": 100, "bolt11": "lnbc1...",
	})
	c, _ := testServer(t, h)

	inv, err := c.Wallet("wal_1").Invoices.Create(context.Background(), &CreateInvoiceParams{
		Amount: 100, Memo: Ptr("test"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if cap.Method != "POST" {
		t.Errorf("Method = %q", cap.Method)
	}
	if cap.Path != "/v1/wallets/wal_1/invoices" {
		t.Errorf("Path = %q", cap.Path)
	}
	if inv.Number != 1 {
		t.Errorf("Number = %d", inv.Number)
	}
	if inv.Bolt11 != "lnbc1..." {
		t.Errorf("Bolt11 = %q", inv.Bolt11)
	}
}

func TestInvoices_List(t *testing.T) {
	h, cap := jsonHandler(200, []map[string]any{
		{"number": 1, "status": "settled", "amount": 100, "bolt11": "lnbc1..."},
		{"number": 2, "status": "pending", "amount": 200, "bolt11": "lnbc2..."},
	})
	c, _ := testServer(t, h)

	invs, err := c.Wallet("wal_1").Invoices.List(context.Background(), &ListInvoicesParams{
		Limit: Ptr(10), After: Ptr(0),
	})
	if err != nil {
		t.Fatal(err)
	}
	if cap.Method != "GET" {
		t.Errorf("Method = %q", cap.Method)
	}
	if cap.Path != "/v1/wallets/wal_1/invoices" {
		t.Errorf("Path = %q", cap.Path)
	}
	if cap.Query != "after=0&limit=10" {
		t.Errorf("Query = %q", cap.Query)
	}
	if len(invs) != 2 {
		t.Fatalf("len = %d, want 2", len(invs))
	}
	if invs[0].Amount != 100 {
		t.Errorf("invs[0].Amount = %d", invs[0].Amount)
	}
}

func TestInvoices_List_NilParams(t *testing.T) {
	h, cap := jsonHandler(200, []map[string]any{})
	c, _ := testServer(t, h)

	_, err := c.Wallet("wal_1").Invoices.List(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if cap.Query != "" {
		t.Errorf("Query = %q, want empty", cap.Query)
	}
}

func TestInvoices_Get(t *testing.T) {
	h, cap := jsonHandler(200, map[string]any{
		"number": 42, "status": "settled", "amount": 500, "bolt11": "lnbc42...",
	})
	c, _ := testServer(t, h)

	inv, err := c.Wallet("wal_1").Invoices.Get(context.Background(), 42)
	if err != nil {
		t.Fatal(err)
	}
	if cap.Path != "/v1/wallets/wal_1/invoices/42" {
		t.Errorf("Path = %q", cap.Path)
	}
	if inv.Number != 42 {
		t.Errorf("Number = %d", inv.Number)
	}
}

func TestInvoices_GetByHash(t *testing.T) {
	h, cap := jsonHandler(200, map[string]any{
		"number": 1, "status": "settled", "amount": 100, "bolt11": "lnbc1...",
	})
	c, _ := testServer(t, h)

	_, err := c.Wallet("wal_1").Invoices.GetByHash(context.Background(), "abc123")
	if err != nil {
		t.Fatal(err)
	}
	if cap.Path != "/v1/wallets/wal_1/invoices/abc123" {
		t.Errorf("Path = %q", cap.Path)
	}
}

// ---------------------------------------------------------------------------
// Public Invoices
// ---------------------------------------------------------------------------

func TestPublicInvoices_CreateForWallet(t *testing.T) {
	h, cap := jsonHandler(200, map[string]any{
		"bolt11": "lnbc1...", "amount": 100,
	})
	c, _ := testServer(t, h)

	inv, err := c.Invoices.CreateForWallet(context.Background(), &CreateInvoiceForWalletParams{
		WalletID: "wal_abc", Amount: 100,
	})
	if err != nil {
		t.Fatal(err)
	}
	if cap.Method != "POST" {
		t.Errorf("Method = %q", cap.Method)
	}
	if cap.Path != "/v1/invoices/for-wallet" {
		t.Errorf("Path = %q", cap.Path)
	}
	if inv.Bolt11 != "lnbc1..." {
		t.Errorf("Bolt11 = %q", inv.Bolt11)
	}
}

func TestPublicInvoices_CreateForAddress(t *testing.T) {
	h, cap := jsonHandler(200, map[string]any{
		"bolt11": "lnbc1...", "amount": 200,
	})
	c, _ := testServer(t, h)

	inv, err := c.Invoices.CreateForAddress(context.Background(), &CreateInvoiceForAddressParams{
		Address: "user@ln.bot", Amount: 200,
	})
	if err != nil {
		t.Fatal(err)
	}
	if cap.Path != "/v1/invoices/for-address" {
		t.Errorf("Path = %q", cap.Path)
	}
	if inv.Amount != 200 {
		t.Errorf("Amount = %d", inv.Amount)
	}
}

// ---------------------------------------------------------------------------
// Payments (wallet-scoped)
// ---------------------------------------------------------------------------

func TestPayments_Create(t *testing.T) {
	h, cap := jsonHandler(200, map[string]any{
		"number": 1, "status": "pending", "amount": 50, "maxFee": 10,
		"serviceFee": 0, "address": "user@ln.bot",
	})
	c, _ := testServer(t, h)

	p, err := c.Wallet("wal_1").Payments.Create(context.Background(), &CreatePaymentParams{
		Target: "user@ln.bot", Amount: Ptr(int64(50)),
	})
	if err != nil {
		t.Fatal(err)
	}
	if cap.Method != "POST" {
		t.Errorf("Method = %q", cap.Method)
	}
	if cap.Path != "/v1/wallets/wal_1/payments" {
		t.Errorf("Path = %q", cap.Path)
	}
	if p.Number != 1 {
		t.Errorf("Number = %d", p.Number)
	}
	if p.Address != "user@ln.bot" {
		t.Errorf("Address = %q", p.Address)
	}
}

func TestPayments_List(t *testing.T) {
	h, cap := jsonHandler(200, []map[string]any{
		{"number": 1, "status": "settled", "amount": 50, "maxFee": 10, "serviceFee": 0, "address": "user@ln.bot"},
	})
	c, _ := testServer(t, h)

	ps, err := c.Wallet("wal_1").Payments.List(context.Background(), &ListPaymentsParams{Limit: Ptr(5)})
	if err != nil {
		t.Fatal(err)
	}
	if cap.Path != "/v1/wallets/wal_1/payments" {
		t.Errorf("Path = %q", cap.Path)
	}
	if cap.Query != "limit=5" {
		t.Errorf("Query = %q", cap.Query)
	}
	if len(ps) != 1 {
		t.Fatalf("len = %d", len(ps))
	}
}

func TestPayments_Get(t *testing.T) {
	h, cap := jsonHandler(200, map[string]any{
		"number": 7, "status": "settled", "amount": 100, "maxFee": 10, "serviceFee": 0, "address": "bob@ln.bot",
	})
	c, _ := testServer(t, h)

	p, err := c.Wallet("wal_1").Payments.Get(context.Background(), 7)
	if err != nil {
		t.Fatal(err)
	}
	if cap.Path != "/v1/wallets/wal_1/payments/7" {
		t.Errorf("Path = %q", cap.Path)
	}
	if p.Number != 7 {
		t.Errorf("Number = %d", p.Number)
	}
}

func TestPayments_GetByHash(t *testing.T) {
	h, cap := jsonHandler(200, map[string]any{
		"number": 1, "status": "settled", "amount": 100, "maxFee": 10, "serviceFee": 0, "address": "a@b.com",
	})
	c, _ := testServer(t, h)

	_, err := c.Wallet("wal_1").Payments.GetByHash(context.Background(), "hash456")
	if err != nil {
		t.Fatal(err)
	}
	if cap.Path != "/v1/wallets/wal_1/payments/hash456" {
		t.Errorf("Path = %q", cap.Path)
	}
}

func TestPayments_Resolve(t *testing.T) {
	h, cap := jsonHandler(200, map[string]any{
		"type": "lnaddress", "min": 1, "max": 1000000,
	})
	c, _ := testServer(t, h)

	res, err := c.Wallet("wal_1").Payments.Resolve(context.Background(), "user@ln.bot")
	if err != nil {
		t.Fatal(err)
	}
	if cap.Method != "GET" {
		t.Errorf("Method = %q", cap.Method)
	}
	if cap.Path != "/v1/wallets/wal_1/payments/resolve" {
		t.Errorf("Path = %q", cap.Path)
	}
	if res.Type != "lnaddress" {
		t.Errorf("Type = %q", res.Type)
	}
}

// ---------------------------------------------------------------------------
// Addresses (wallet-scoped)
// ---------------------------------------------------------------------------

func TestAddresses_Create(t *testing.T) {
	h, cap := jsonHandler(200, map[string]any{
		"address": "user@ln.bot", "generated": false, "cost": 0,
	})
	c, _ := testServer(t, h)

	a, err := c.Wallet("wal_1").Addresses.Create(context.Background(), &CreateAddressParams{Address: Ptr("user@ln.bot")})
	if err != nil {
		t.Fatal(err)
	}
	if cap.Method != "POST" {
		t.Errorf("Method = %q", cap.Method)
	}
	if cap.Path != "/v1/wallets/wal_1/addresses" {
		t.Errorf("Path = %q", cap.Path)
	}
	if a.Address != "user@ln.bot" {
		t.Errorf("Address = %q", a.Address)
	}
}

func TestAddresses_List(t *testing.T) {
	h, cap := jsonHandler(200, []map[string]any{
		{"address": "a@ln.bot", "generated": true, "cost": 0},
		{"address": "b@ln.bot", "generated": false, "cost": 100},
	})
	c, _ := testServer(t, h)

	addrs, err := c.Wallet("wal_1").Addresses.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if cap.Path != "/v1/wallets/wal_1/addresses" {
		t.Errorf("Path = %q", cap.Path)
	}
	if len(addrs) != 2 {
		t.Fatalf("len = %d", len(addrs))
	}
	if !addrs[0].Generated {
		t.Error("addrs[0].Generated = false")
	}
}

func TestAddresses_Delete(t *testing.T) {
	h, cap := jsonHandler(200, nil)
	c, _ := testServer(t, h)

	err := c.Wallet("wal_1").Addresses.Delete(context.Background(), "user@ln.bot")
	if err != nil {
		t.Fatal(err)
	}
	if cap.Method != "DELETE" {
		t.Errorf("Method = %q", cap.Method)
	}
	if cap.Path != "/v1/wallets/wal_1/addresses/user@ln.bot" {
		t.Errorf("Path = %q", cap.Path)
	}
}

func TestAddresses_Transfer(t *testing.T) {
	h, cap := jsonHandler(200, map[string]any{
		"address": "user@ln.bot", "transferredTo": "wal_target",
	})
	c, _ := testServer(t, h)

	tr, err := c.Wallet("wal_1").Addresses.Transfer(context.Background(), "user@ln.bot", &TransferAddressParams{
		TargetWalletKey: "wk_target",
	})
	if err != nil {
		t.Fatal(err)
	}
	if cap.Method != "POST" {
		t.Errorf("Method = %q", cap.Method)
	}
	if cap.Path != "/v1/wallets/wal_1/addresses/user@ln.bot/transfer" {
		t.Errorf("Path = %q", cap.Path)
	}
	if tr.TransferredTo != "wal_target" {
		t.Errorf("TransferredTo = %q", tr.TransferredTo)
	}
}

// ---------------------------------------------------------------------------
// Transactions (wallet-scoped)
// ---------------------------------------------------------------------------

func TestTransactions_List(t *testing.T) {
	h, cap := jsonHandler(200, []map[string]any{
		{"number": 1, "type": "credit", "amount": 100, "balanceAfter": 100, "networkFee": 0, "serviceFee": 0},
	})
	c, _ := testServer(t, h)

	txs, err := c.Wallet("wal_1").Transactions.List(context.Background(), &ListTransactionsParams{
		Limit: Ptr(20), After: Ptr(0),
	})
	if err != nil {
		t.Fatal(err)
	}
	if cap.Path != "/v1/wallets/wal_1/transactions" {
		t.Errorf("Path = %q", cap.Path)
	}
	if cap.Query != "after=0&limit=20" {
		t.Errorf("Query = %q", cap.Query)
	}
	if len(txs) != 1 {
		t.Fatalf("len = %d", len(txs))
	}
	if txs[0].Type != "credit" {
		t.Errorf("Type = %q", txs[0].Type)
	}
}

func TestTransactions_List_NilParams(t *testing.T) {
	h, cap := jsonHandler(200, []map[string]any{})
	c, _ := testServer(t, h)

	_, err := c.Wallet("wal_1").Transactions.List(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if cap.Query != "" {
		t.Errorf("Query = %q, want empty", cap.Query)
	}
}

// ---------------------------------------------------------------------------
// Webhooks (wallet-scoped)
// ---------------------------------------------------------------------------

func TestWebhooks_Create(t *testing.T) {
	h, cap := jsonHandler(200, map[string]any{
		"id": "wh_1", "url": "https://example.com/hook", "secret": "sec_abc",
	})
	c, _ := testServer(t, h)

	wh, err := c.Wallet("wal_1").Webhooks.Create(context.Background(), &CreateWebhookParams{
		URL: "https://example.com/hook",
	})
	if err != nil {
		t.Fatal(err)
	}
	if cap.Method != "POST" {
		t.Errorf("Method = %q", cap.Method)
	}
	if cap.Path != "/v1/wallets/wal_1/webhooks" {
		t.Errorf("Path = %q", cap.Path)
	}
	if wh.Secret != "sec_abc" {
		t.Errorf("Secret = %q", wh.Secret)
	}
}

func TestWebhooks_List(t *testing.T) {
	h, cap := jsonHandler(200, []map[string]any{
		{"id": "wh_1", "url": "https://example.com/hook", "active": true},
	})
	c, _ := testServer(t, h)

	whs, err := c.Wallet("wal_1").Webhooks.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if cap.Path != "/v1/wallets/wal_1/webhooks" {
		t.Errorf("Path = %q", cap.Path)
	}
	if len(whs) != 1 {
		t.Fatalf("len = %d", len(whs))
	}
	if !whs[0].Active {
		t.Error("Active = false")
	}
}

func TestWebhooks_Delete(t *testing.T) {
	h, cap := jsonHandler(200, nil)
	c, _ := testServer(t, h)

	err := c.Wallet("wal_1").Webhooks.Delete(context.Background(), "wh_1")
	if err != nil {
		t.Fatal(err)
	}
	if cap.Method != "DELETE" {
		t.Errorf("Method = %q", cap.Method)
	}
	if cap.Path != "/v1/wallets/wal_1/webhooks/wh_1" {
		t.Errorf("Path = %q", cap.Path)
	}
}

// ---------------------------------------------------------------------------
// Backup
// ---------------------------------------------------------------------------

func TestBackup_Recovery(t *testing.T) {
	h, cap := jsonHandler(200, map[string]any{"passphrase": "word1 word2 word3"})
	c, _ := testServer(t, h)

	r, err := c.Backup.Recovery(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if cap.Method != "POST" {
		t.Errorf("Method = %q", cap.Method)
	}
	if cap.Path != "/v1/backup/recovery" {
		t.Errorf("Path = %q", cap.Path)
	}
	if r.Passphrase != "word1 word2 word3" {
		t.Errorf("Passphrase = %q", r.Passphrase)
	}
}

func TestBackup_PasskeyBegin(t *testing.T) {
	h, cap := jsonHandler(200, map[string]any{
		"sessionId": "sess_1", "options": map[string]any{"challenge": "abc"},
	})
	c, _ := testServer(t, h)

	ch, err := c.Backup.PasskeyBegin(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if cap.Path != "/v1/backup/passkey/begin" {
		t.Errorf("Path = %q", cap.Path)
	}
	if ch.SessionID != "sess_1" {
		t.Errorf("SessionID = %q", ch.SessionID)
	}
}

func TestBackup_PasskeyComplete(t *testing.T) {
	h, cap := jsonHandler(200, nil)
	c, _ := testServer(t, h)

	err := c.Backup.PasskeyComplete(context.Background(), &PasskeyAttestationParams{
		SessionID:   "sess_1",
		Attestation: map[string]any{"id": "cred_1"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if cap.Method != "POST" {
		t.Errorf("Method = %q", cap.Method)
	}
	if cap.Path != "/v1/backup/passkey/complete" {
		t.Errorf("Path = %q", cap.Path)
	}
}

// ---------------------------------------------------------------------------
// Restore
// ---------------------------------------------------------------------------

func TestRestore_Recovery(t *testing.T) {
	h, cap := jsonHandler(200, map[string]any{
		"walletId": "wal_1", "name": "Restored", "primaryKey": "uk_1", "secondaryKey": "uk_2",
	})
	c, _ := testServer(t, h)

	w, err := c.Restore.Recovery(context.Background(), &RecoveryRestoreParams{
		Passphrase: "word1 word2 word3",
	})
	if err != nil {
		t.Fatal(err)
	}
	if cap.Method != "POST" {
		t.Errorf("Method = %q", cap.Method)
	}
	if cap.Path != "/v1/restore/recovery" {
		t.Errorf("Path = %q", cap.Path)
	}
	if w.WalletID != "wal_1" {
		t.Errorf("WalletID = %q", w.WalletID)
	}
}

func TestRestore_PasskeyBegin(t *testing.T) {
	h, cap := jsonHandler(200, map[string]any{
		"sessionId": "sess_2", "options": map[string]any{"challenge": "xyz"},
	})
	c, _ := testServer(t, h)

	ch, err := c.Restore.PasskeyBegin(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if cap.Path != "/v1/restore/passkey/begin" {
		t.Errorf("Path = %q", cap.Path)
	}
	if ch.SessionID != "sess_2" {
		t.Errorf("SessionID = %q", ch.SessionID)
	}
}

func TestRestore_PasskeyComplete(t *testing.T) {
	h, cap := jsonHandler(200, map[string]any{
		"walletId": "wal_1", "name": "Restored", "primaryKey": "uk_1", "secondaryKey": "uk_2",
	})
	c, _ := testServer(t, h)

	w, err := c.Restore.PasskeyComplete(context.Background(), &PasskeyAssertionParams{
		SessionID: "sess_2",
		Assertion: map[string]any{"id": "cred_1"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if cap.Method != "POST" {
		t.Errorf("Method = %q", cap.Method)
	}
	if cap.Path != "/v1/restore/passkey/complete" {
		t.Errorf("Path = %q", cap.Path)
	}
	if w.PrimaryKey != "uk_1" {
		t.Errorf("PrimaryKey = %q", w.PrimaryKey)
	}
}

// ---------------------------------------------------------------------------
// L402 (wallet-scoped)
// ---------------------------------------------------------------------------

func TestL402_CreateChallenge(t *testing.T) {
	h, cap := jsonHandler(200, map[string]any{
		"macaroon": "mac_abc", "invoice": "lnbc1...", "paymentHash": "hash_1",
		"expiresAt": "2024-01-01T00:00:00Z", "wwwAuthenticate": "L402 mac:inv",
	})
	c, _ := testServer(t, h)

	ch, err := c.Wallet("wal_1").L402.CreateChallenge(context.Background(), &CreateL402ChallengeParams{
		Amount:  100,
		Caveats: []string{"service=api"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if cap.Method != "POST" {
		t.Errorf("Method = %q", cap.Method)
	}
	if cap.Path != "/v1/wallets/wal_1/l402/challenges" {
		t.Errorf("Path = %q", cap.Path)
	}
	if ch.Macaroon != "mac_abc" {
		t.Errorf("Macaroon = %q", ch.Macaroon)
	}
	if ch.WwwAuthenticate != "L402 mac:inv" {
		t.Errorf("WwwAuthenticate = %q", ch.WwwAuthenticate)
	}
}

func TestL402_Verify(t *testing.T) {
	h, cap := jsonHandler(200, map[string]any{
		"valid": true, "paymentHash": "hash_1", "caveats": []string{"service=api"},
	})
	c, _ := testServer(t, h)

	resp, err := c.Wallet("wal_1").L402.Verify(context.Background(), &VerifyL402Params{
		Authorization: "L402 token:preimage",
	})
	if err != nil {
		t.Fatal(err)
	}
	if cap.Path != "/v1/wallets/wal_1/l402/verify" {
		t.Errorf("Path = %q", cap.Path)
	}
	if !resp.Valid {
		t.Error("Valid = false")
	}
}

func TestL402_Pay(t *testing.T) {
	h, cap := jsonHandler(200, map[string]any{
		"authorization": "L402 token:preimage", "paymentHash": "hash_1",
		"preimage": "pre_1", "amount": 100, "fee": 1, "paymentNumber": 1, "status": "settled",
	})
	c, _ := testServer(t, h)

	resp, err := c.Wallet("wal_1").L402.Pay(context.Background(), &PayL402Params{
		WwwAuthenticate: "L402 mac:inv",
		MaxFee:          Ptr(int64(10)),
	})
	if err != nil {
		t.Fatal(err)
	}
	if cap.Path != "/v1/wallets/wal_1/l402/pay" {
		t.Errorf("Path = %q", cap.Path)
	}
	if resp.Status != "settled" {
		t.Errorf("Status = %q", resp.Status)
	}
	if resp.Amount != 100 {
		t.Errorf("Amount = %d", resp.Amount)
	}
}
