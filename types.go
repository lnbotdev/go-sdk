package lnbot

import "time"

// Ptr returns a pointer to v. Useful for optional fields in request params.
func Ptr[T any](v T) *T { return &v }

// ---------------------------------------------------------------------------
// Wallet
// ---------------------------------------------------------------------------

// Wallet represents a LnBot wallet.
type Wallet struct {
	WalletID  string `json:"walletId"`
	Name      string `json:"name"`
	Balance   int64  `json:"balance"`
	OnHold    int64  `json:"onHold"`
	Available int64  `json:"available"`
}

// CreateWalletParams are the parameters for creating a new wallet.
type CreateWalletParams struct {
	Name *string `json:"name,omitempty"`
}

// WalletCredentials is returned when a new wallet is created.
// It contains the API keys and recovery passphrase.
type WalletCredentials struct {
	WalletID           string `json:"walletId"`
	PrimaryKey         string `json:"primaryKey"`
	SecondaryKey       string `json:"secondaryKey"`
	Name               string `json:"name"`
	Address            string `json:"address"`
	RecoveryPassphrase string `json:"recoveryPassphrase"`
}

// UpdateWalletParams are the parameters for updating a wallet.
type UpdateWalletParams struct {
	Name string `json:"name"`
}

// ---------------------------------------------------------------------------
// API Keys
// ---------------------------------------------------------------------------

// APIKey represents an API key's metadata.
type APIKey struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	Hint       string     `json:"hint"`
	CreatedAt  *time.Time `json:"createdAt"`
	LastUsedAt *time.Time `json:"lastUsedAt"`
}

// RotatedAPIKey is returned after rotating an API key.
type RotatedAPIKey struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

// ---------------------------------------------------------------------------
// Invoices
// ---------------------------------------------------------------------------

// CreateInvoiceParams are the parameters for creating a new invoice.
type CreateInvoiceParams struct {
	Amount    int64   `json:"amount"`
	Reference *string `json:"reference,omitempty"`
	Memo      *string `json:"memo,omitempty"`
}

// Invoice represents a Lightning invoice.
type Invoice struct {
	Number    int        `json:"number"`
	Status    string     `json:"status"`
	Amount    int64      `json:"amount"`
	Bolt11    string     `json:"bolt11"`
	Reference *string    `json:"reference"`
	Memo      *string    `json:"memo"`
	TxNumber  *int       `json:"txNumber"`
	CreatedAt *time.Time `json:"createdAt"`
	SettledAt *time.Time `json:"settledAt"`
	ExpiresAt *time.Time `json:"expiresAt"`
}

// ListInvoicesParams are the pagination parameters for listing invoices.
type ListInvoicesParams struct {
	Limit *int
	After *int
}

// InvoiceEvent represents a server-sent event for an invoice.
type InvoiceEvent struct {
	Event string
	Data  Invoice
}

// CreateInvoiceForWalletParams are the parameters for creating an invoice for a specific wallet.
// No authentication required. Rate limited by IP.
type CreateInvoiceForWalletParams struct {
	WalletID  string  `json:"walletId"`
	Amount    int64   `json:"amount"`
	Reference *string `json:"reference,omitempty"`
	Comment   *string `json:"comment,omitempty"`
}

// CreateInvoiceForAddressParams are the parameters for creating an invoice for a Lightning address.
// No authentication required. Rate limited by IP.
type CreateInvoiceForAddressParams struct {
	Address string  `json:"address"`
	Amount  int64   `json:"amount"`
	Tag     *string `json:"tag,omitempty"`
	Comment *string `json:"comment,omitempty"`
}

// AddressInvoice is an invoice created via wallet ID or Lightning address.
type AddressInvoice struct {
	Bolt11    string     `json:"bolt11"`
	Amount    int64      `json:"amount"`
	ExpiresAt *time.Time `json:"expiresAt"`
}

// ---------------------------------------------------------------------------
// Payments
// ---------------------------------------------------------------------------

// CreatePaymentParams are the parameters for creating a new payment.
// Target accepts a Lightning address (user@domain), LNURL, or BOLT11 invoice.
type CreatePaymentParams struct {
	Target         string  `json:"target"`
	Amount         *int64  `json:"amount,omitempty"`
	IdempotencyKey *string `json:"idempotencyKey,omitempty"`
	MaxFee         *int64  `json:"maxFee,omitempty"`
	Reference      *string `json:"reference,omitempty"`
}

// Payment represents a Lightning payment.
type Payment struct {
	Number        int        `json:"number"`
	Status        string     `json:"status"`
	Amount        int64      `json:"amount"`
	MaxFee        int64      `json:"maxFee"`
	ActualFee     *int64     `json:"actualFee"`
	Address       string     `json:"address"`
	Reference     *string    `json:"reference"`
	TxNumber      *int       `json:"txNumber"`
	FailureReason *string    `json:"failureReason"`
	CreatedAt     *time.Time `json:"createdAt"`
	SettledAt     *time.Time `json:"settledAt"`
}

// ListPaymentsParams are the pagination parameters for listing payments.
type ListPaymentsParams struct {
	Limit *int
	After *int
}

// PaymentEvent represents a server-sent event for a payment.
type PaymentEvent struct {
	Event string
	Data  Payment
}

// ---------------------------------------------------------------------------
// Addresses
// ---------------------------------------------------------------------------

// CreateAddressParams are the parameters for creating a Lightning address.
type CreateAddressParams struct {
	Address *string `json:"address,omitempty"`
}

// Address represents a Lightning address.
type Address struct {
	Address   string     `json:"address"`
	Generated bool       `json:"generated"`
	Cost      int64      `json:"cost"`
	CreatedAt *time.Time `json:"createdAt"`
}

// TransferAddressParams are the parameters for transferring an address.
type TransferAddressParams struct {
	TargetWalletKey string `json:"targetWalletKey"`
}

// AddressTransfer is returned after transferring an address.
type AddressTransfer struct {
	Address       string `json:"address"`
	TransferredTo string `json:"transferredTo"`
}

// ---------------------------------------------------------------------------
// Transactions
// ---------------------------------------------------------------------------

// Transaction represents a wallet transaction.
type Transaction struct {
	Number       int        `json:"number"`
	Type         string     `json:"type"`
	Amount       int64      `json:"amount"`
	BalanceAfter int64      `json:"balanceAfter"`
	NetworkFee   int64      `json:"networkFee"`
	ServiceFee   int64      `json:"serviceFee"`
	PaymentHash  *string    `json:"paymentHash"`
	Preimage     *string    `json:"preimage"`
	Reference    *string    `json:"reference"`
	Note         *string    `json:"note"`
	CreatedAt    *time.Time `json:"createdAt"`
}

// ListTransactionsParams are the pagination parameters for listing transactions.
type ListTransactionsParams struct {
	Limit *int
	After *int
}

// ---------------------------------------------------------------------------
// Webhooks
// ---------------------------------------------------------------------------

// CreateWebhookParams are the parameters for creating a webhook.
type CreateWebhookParams struct {
	URL string `json:"url"`
}

// WebhookWithSecret is returned when a webhook is created.
// It includes the secret used for signature verification.
type WebhookWithSecret struct {
	ID        string     `json:"id"`
	URL       string     `json:"url"`
	Secret    string     `json:"secret"`
	CreatedAt *time.Time `json:"createdAt"`
}

// Webhook represents a webhook endpoint.
type Webhook struct {
	ID        string     `json:"id"`
	URL       string     `json:"url"`
	Active    bool       `json:"active"`
	CreatedAt *time.Time `json:"createdAt"`
}

// ---------------------------------------------------------------------------
// Backup / Restore
// ---------------------------------------------------------------------------

// RecoveryPassphrase is returned when backup via recovery passphrase is initiated.
type RecoveryPassphrase struct {
	Passphrase string `json:"passphrase"`
}

// RecoveryRestoreParams are the parameters for restoring via recovery passphrase.
type RecoveryRestoreParams struct {
	Passphrase string `json:"passphrase"`
}

// RestoredWallet is returned after a successful wallet restore.
type RestoredWallet struct {
	WalletID     string `json:"walletId"`
	Name         string `json:"name"`
	PrimaryKey   string `json:"primaryKey"`
	SecondaryKey string `json:"secondaryKey"`
}

// PasskeyRegistrationChallenge is returned when beginning passkey backup.
type PasskeyRegistrationChallenge struct {
	SessionID string         `json:"sessionId"`
	Options   map[string]any `json:"options"`
}

// PasskeyAttestationParams are the parameters for completing passkey backup.
type PasskeyAttestationParams struct {
	SessionID   string         `json:"sessionId"`
	Attestation map[string]any `json:"attestation"`
}

// PasskeyAuthenticationChallenge is returned when beginning passkey restore.
type PasskeyAuthenticationChallenge struct {
	SessionID string         `json:"sessionId"`
	Options   map[string]any `json:"options"`
}

// PasskeyAssertionParams are the parameters for completing passkey restore.
type PasskeyAssertionParams struct {
	SessionID string         `json:"sessionId"`
	Assertion map[string]any `json:"assertion"`
}
