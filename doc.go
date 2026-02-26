// Package lnbot is the official Go SDK for ln.bot — Bitcoin for AI Agents.
//
// Send and receive sats over the Lightning Network. Create wallets, generate
// invoices, make payments, and stream real-time settlement events via SSE.
//
// Zero dependencies beyond the standard library. Every method takes a
// context.Context as its first argument.
//
//	client := lnbot.New("key_...")
//	invoice, _ := client.Invoices.Create(ctx, &lnbot.CreateInvoiceParams{
//	    Amount: 1000,
//	    Memo:   lnbot.Ptr("Coffee"),
//	})
//
// Get your API key at https://ln.bot
//
// Also available as a TypeScript SDK (https://www.npmjs.com/package/@lnbot/sdk),
// Python SDK (https://pypi.org/project/lnbot/),
// and Rust SDK (https://crates.io/crates/lnbot).
package lnbot
