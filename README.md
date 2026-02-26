# lnbot-go

[![Go Reference](https://pkg.go.dev/badge/github.com/lnbotdev/go-sdk.svg)](https://pkg.go.dev/github.com/lnbotdev/go-sdk)
[![License: MIT](https://img.shields.io/badge/license-MIT-green)](./LICENSE)

**The official Go SDK for [LnBot](https://ln.bot)** — Bitcoin for AI Agents.

Give your AI agents, apps, and services access to Bitcoin over the Lightning Network. Create wallets, send and receive sats, and get real-time payment notifications.

```go
client := lnbot.New("key_...")

invoice, _ := client.Invoices.Create(ctx, &lnbot.CreateInvoiceParams{
    Amount: 1000,
    Memo:   lnbot.Ptr("Coffee"),
})
```

> LnBot also ships a **[TypeScript SDK](https://www.npmjs.com/package/@lnbot/sdk)**, **[Python SDK](https://pypi.org/project/lnbot/)**, **[Rust SDK](https://crates.io/crates/lnbot)**, **[CLI](https://ln.bot/docs)**, and **[MCP server](https://ln.bot/docs)**.

---

## Install

```bash
go get github.com/lnbotdev/go-sdk
```

---

## Quick start

### Create a wallet

```go
package main

import (
    "context"
    "fmt"

    lnbot "github.com/lnbotdev/go-sdk"
)

func main() {
    client := lnbot.New("")
    wallet, _ := client.Wallets.Create(context.Background(), &lnbot.CreateWalletParams{
        Name: lnbot.Ptr("my-agent"),
    })
    fmt.Println(wallet.PrimaryKey)
}
```

### Receive sats

```go
client := lnbot.New(wallet.PrimaryKey)

invoice, _ := client.Invoices.Create(ctx, &lnbot.CreateInvoiceParams{
    Amount: 1000,
    Memo:   lnbot.Ptr("Payment for task #42"),
})
fmt.Println(invoice.Bolt11)
```

### Wait for payment (SSE)

```go
events, errs := client.Invoices.Watch(ctx, invoice.Number, nil)
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
payment, _ := client.Payments.Create(ctx, &lnbot.CreatePaymentParams{
    Target: "alice@ln.bot",
    Amount: lnbot.Ptr(int64(500)),
})
```

### Check balance

```go
wallet, _ := client.Wallets.Current(ctx)
fmt.Printf("%d sats available\n", wallet.Available)
```

---

## Error handling

```go
import "errors"

wallet, err := client.Wallets.Current(ctx)
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
client := lnbot.New("key_...",
    lnbot.WithBaseURL("https://api.ln.bot"),
    lnbot.WithHTTPClient(customHTTPClient),
)
```

---

## Features

- **Zero dependencies** — stdlib `net/http` + `encoding/json` only
- **Context-first** — every method takes `context.Context` as the first argument
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
