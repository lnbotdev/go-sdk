# ln.bot-go

[![Go Reference](https://pkg.go.dev/badge/github.com/lnbotdev/go-sdk.svg)](https://pkg.go.dev/github.com/lnbotdev/go-sdk)
[![License: MIT](https://img.shields.io/badge/license-MIT-green)](./LICENSE)

**The official Go SDK for [ln.bot](https://ln.bot)** — Bitcoin for AI Agents.

Give your AI agents, apps, and services access to Bitcoin over the Lightning Network. Create wallets, send and receive sats, and get real-time payment notifications.

```go
client := lnbot.New("uk_...")
w := client.Wallet("wal_...")

invoice, _ := w.Invoices.Create(ctx, &lnbot.CreateInvoiceParams{
    Amount: 1000,
    Memo:   lnbot.Ptr("Coffee"),
})
```

> ln.bot also ships a **[TypeScript SDK](https://www.npmjs.com/package/@lnbot/sdk)**, **[Python SDK](https://pypi.org/project/lnbot/)**, **[Rust SDK](https://crates.io/crates/lnbot)**, **[CLI](https://ln.bot/docs)**, and **[MCP server](https://ln.bot/docs)**.

---

## Install

```bash
go get github.com/lnbotdev/go-sdk
```

---

## Quick start

### Register an account

```go
package main

import (
    "context"
    "fmt"

    lnbot "github.com/lnbotdev/go-sdk"
)

func main() {
    ctx := context.Background()

    // Register (no auth needed)
    client := lnbot.New("")
    account, _ := client.Register(ctx)
    fmt.Println("User key:", account.PrimaryKey)
    fmt.Println("Recovery:", account.RecoveryPassphrase)

    // Create a wallet
    authed := lnbot.New(account.PrimaryKey)
    wallet, _ := authed.Wallets.Create(ctx)
    fmt.Println("Wallet:", wallet.WalletID)
    fmt.Println("Address:", wallet.Address)
}
```

### Receive sats

```go
client := lnbot.New("uk_...")
w := client.Wallet("wal_...")

invoice, _ := w.Invoices.Create(ctx, &lnbot.CreateInvoiceParams{
    Amount: 1000,
    Memo:   lnbot.Ptr("Payment for task #42"),
})
fmt.Println(invoice.Bolt11)
```

### Wait for payment (SSE)

SSE streams require a **wallet key** (`wk_`):

```go
// Create a wallet key (one-time setup)
wk, _ := w.Key.Create(ctx)

// Use the wallet key client for SSE
wkClient := lnbot.New(wk.Key)
events, errs := wkClient.Wallet("wal_...").Invoices.Watch(ctx, invoice.Number, nil)
for event := range events {
    if event.Event == "settled" {
        fmt.Println("Paid!")
    }
}
if err := <-errs; err != nil {
    log.Fatal(err)
}
```

### Send sats

```go
payment, _ := w.Payments.Create(ctx, &lnbot.CreatePaymentParams{
    Target: "alice@ln.bot",
    Amount: lnbot.Ptr(int64(500)),
})
```

### Resolve before sending

```go
res, _ := w.Payments.Resolve(ctx, "alice@ln.bot")
fmt.Printf("type=%s min=%d max=%d\n", res.Type, *res.Min, *res.Max)
```

### Check balance

```go
wal, _ := w.Get(ctx)
fmt.Printf("%d sats available\n", wal.Available)
```

### List wallets

```go
wallets, _ := client.Wallets.List(ctx)
for _, w := range wallets {
    fmt.Printf("%s — %s\n", w.WalletID, w.Name)
}
```

---

## Authentication

ln.bot uses two key types:

| Key type | Prefix | Scope | Created by |
|----------|--------|-------|------------|
| **User key** | `uk_` | All wallets under the account | `Register()` |
| **Wallet key** | `wk_` | Single wallet only | `wallet.Key.Create()` |

Use **user keys** for most operations. Use **wallet keys** for SSE streams or to give scoped access to a single wallet.

---

## L402 paywalls

```go
w := client.Wallet("wal_...")

// Create a challenge (server side)
challenge, _ := w.L402.CreateChallenge(ctx, &lnbot.CreateL402ChallengeParams{
    Amount:      100,
    Description: lnbot.Ptr("API access"),
})

// Pay the challenge (client side)
result, _ := w.L402.Pay(ctx, &lnbot.PayL402Params{
    WwwAuthenticate: challenge.WwwAuthenticate,
})

// Verify a token (server side, stateless)
v, _ := w.L402.Verify(ctx, &lnbot.VerifyL402Params{
    Authorization: *result.Authorization,
})
fmt.Println(v.Valid)
```

## Error handling

```go
import "errors"

wal, err := client.Wallet("wal_...").Get(ctx)
if err != nil {
    // Match any API error
    var apiErr *lnbot.APIError
    if errors.As(err, &apiErr) {
        fmt.Println(apiErr.StatusCode, apiErr.Message)
    }

    // Match specific error types
    var notFound *lnbot.NotFoundError
    var badReq *lnbot.BadRequestError
    var conflict *lnbot.ConflictError

    switch {
    case errors.As(err, &notFound):
        // 404
    case errors.As(err, &badReq):
        // 400
    case errors.As(err, &conflict):
        // 409
    default:
        // other error
    }
}
```

## Configuration

```go
client := lnbot.New("uk_...",
    lnbot.WithBaseURL("https://api.ln.bot"),
    lnbot.WithHTTPClient(customHTTPClient),
)
```

---

## API reference

### Client (top-level)

| Method | Description |
|--------|-------------|
| `Register(ctx)` | Create a new account (no auth) |
| `Me(ctx)` | Get authenticated identity |
| `Wallet(id)` | Get a wallet handle |
| `Wallets.Create(ctx)` | Create a new wallet |
| `Wallets.List(ctx)` | List all wallets |
| `Keys.Rotate(ctx, slot)` | Rotate user key (0=primary, 1=secondary) |
| `Invoices.CreateForWallet(ctx, params)` | Public invoice by wallet ID |
| `Invoices.CreateForAddress(ctx, params)` | Public invoice by Lightning address |
| `Backup.Recovery(ctx)` | Get recovery passphrase |
| `Backup.PasskeyBegin(ctx)` | Start passkey registration |
| `Backup.PasskeyComplete(ctx, params)` | Complete passkey registration |
| `Restore.Recovery(ctx, params)` | Restore via passphrase |
| `Restore.PasskeyBegin(ctx)` | Start passkey auth |
| `Restore.PasskeyComplete(ctx, params)` | Complete passkey restore |

### Wallet handle

| Method | Description |
|--------|-------------|
| `Get(ctx)` | Get wallet details |
| `Update(ctx, params)` | Update wallet (rename) |
| `Key.Create(ctx)` | Create wallet key (max 1) |
| `Key.Get(ctx)` | Get wallet key info |
| `Key.Delete(ctx)` | Revoke wallet key |
| `Key.Rotate(ctx)` | Rotate wallet key |
| `Invoices.Create(ctx, params)` | Create invoice |
| `Invoices.List(ctx, params)` | List invoices |
| `Invoices.Get(ctx, number)` | Get invoice by number |
| `Invoices.Watch(ctx, number, timeout)` | SSE stream (requires `wk_`) |
| `Payments.Create(ctx, params)` | Send payment |
| `Payments.List(ctx, params)` | List payments |
| `Payments.Get(ctx, number)` | Get payment by number |
| `Payments.Resolve(ctx, target)` | Inspect target before paying |
| `Payments.Watch(ctx, number, timeout)` | SSE stream (requires `wk_`) |
| `Addresses.Create(ctx, params)` | Create Lightning address |
| `Addresses.List(ctx)` | List addresses |
| `Addresses.Delete(ctx, address)` | Delete address |
| `Addresses.Transfer(ctx, address, params)` | Transfer address |
| `Transactions.List(ctx, params)` | List transactions |
| `Webhooks.Create(ctx, params)` | Register webhook |
| `Webhooks.List(ctx)` | List webhooks |
| `Webhooks.Delete(ctx, id)` | Delete webhook |
| `Events.Stream(ctx)` | SSE stream (requires `wk_`) |
| `L402.CreateChallenge(ctx, params)` | Create L402 challenge |
| `L402.Verify(ctx, params)` | Verify L402 token |
| `L402.Pay(ctx, params)` | Pay L402 challenge |

---

## Features

- **Zero dependencies** — stdlib `net/http` + `encoding/json` only
- **Context-first** — every method takes `context.Context` as the first argument
- **Wallet-scoped** — `client.Wallet(id)` returns a handle with all sub-resources
- **Typed errors** — `BadRequestError`, `NotFoundError`, `ConflictError`, `UnauthorizedError`, `ForbiddenError`
- **Generic helpers** — `Ptr[T]` for optional fields
- **SSE support** — `Watch` returns channels for real-time events

## Requirements

- Go 1.21+
- Get your API key at [ln.bot](https://ln.bot)

## Links

- [ln.bot](https://ln.bot) — website
- [Documentation](https://ln.bot/docs)
- [GitHub](https://github.com/lnbotdev)
- [pkg.go.dev](https://pkg.go.dev/github.com/lnbotdev/go-sdk)

## Other SDKs

- [TypeScript SDK](https://github.com/lnbotdev/typescript-sdk) · [npm](https://www.npmjs.com/package/@lnbot/sdk)
- [Python SDK](https://github.com/lnbotdev/python-sdk) · [pypi](https://pypi.org/project/lnbot/)
- [Rust SDK](https://github.com/lnbotdev/rust-sdk) · [crates.io](https://crates.io/crates/lnbot) · [docs.rs](https://docs.rs/lnbot)

## License

MIT
