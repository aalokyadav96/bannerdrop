package models

import (
	"time"
)

// Meta is a generic key-value map for transaction metadata
type Meta map[string]interface{}

// Transaction represents a wallet or payment transaction
type Transaction struct {
	ID             string    `bson:"_id,omitempty" json:"id"`
	UserID         string    `bson:"userid,omitempty" json:"userid,omitempty"`
	ParentTxn      string    `bson:"parent_txn,omitempty" json:"parent_txn,omitempty"`
	Type           string    `bson:"type" json:"type"` // credit, debit, topup, payment, transfer, refund
	Amount         float64   `bson:"amount" json:"amount"`
	Method         string    `bson:"method" json:"method"` // wallet, card, upi, cod, topup, transfer, refund
	EntityID       string    `bson:"entity_id,omitempty" json:"entity_id,omitempty"`
	EntityType     string    `bson:"entity_type,omitempty" json:"entity_type,omitempty"`
	FromAccount    string    `bson:"from_account,omitempty" json:"from_account,omitempty"`
	ToAccount      string    `bson:"to_account,omitempty" json:"to_account,omitempty"`
	Status         string    `bson:"state" json:"state"` // initiated, success, failed, reversed
	Currency       string    `bson:"currency" json:"currency"`
	CreatedAt      time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time `bson:"updated_at" json:"updated_at"`
	IdempotencyKey string    `bson:"external_ref,omitempty" json:"external_ref,omitempty"`
	Meta           Meta      `bson:"meta,omitempty" json:"meta,omitempty"`
}

// JournalEntry represents a ledger double-entry record
type JournalEntry struct {
	ID            string    `bson:"_id,omitempty" json:"id"`
	TxnID         string    `bson:"txn_id" json:"txn_id"`
	DebitAccount  string    `bson:"debit_account" json:"debit_account"`
	CreditAccount string    `bson:"credit_account" json:"credit_account"`
	Amount        float64   `bson:"amount" json:"amount"`
	Currency      string    `bson:"currency" json:"currency"`
	CreatedAt     time.Time `bson:"created_at" json:"created_at"`
	Meta          Meta      `bson:"meta,omitempty" json:"meta,omitempty"`
}

// Account represents a user's wallet/account
type Account struct {
	ID            string    `bson:"_id,omitempty" json:"id"`
	UserID        string    `bson:"userid" json:"userid"`
	Currency      string    `bson:"currency" json:"currency"`
	Status        string    `bson:"status" json:"status"` // active, inactive
	CachedBalance float64   `bson:"cached_balance" json:"cached_balance"`
	Version       int       `bson:"version" json:"version"`
	CreatedAt     time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time `bson:"updated_at" json:"updated_at"`
}

// Price represents a resolvable price for an entity
type Price struct {
	EntityType string  `bson:"entity_type" json:"entity_type"`
	EntityID   string  `bson:"entity_id" json:"entity_id"`
	Amount     float64 `bson:"amount" json:"amount"`
	Currency   string  `bson:"currency" json:"currency"`
}

// PayRequest is the request payload for a payment
type PayRequest struct {
	EntityType string  `bson:"entity_type" json:"entityType"`
	EntityID   string  `bson:"entity_id" json:"entityId"`
	Method     string  `bson:"method" json:"method"` // wallet, card, upi, cod
	Amount     float64 `bson:"amount" json:"amount"` // optional for user input
}

// IdempotencyRecord represents an idempotency key record stored in Mongo.
type IdempotencyRecord struct {
	Key         string                 `bson:"key" json:"key"`
	Method      string                 `bson:"method" json:"method"`
	Path        string                 `bson:"path" json:"path"`
	UserID      string                 `bson:"userid" json:"userid"`
	RequestHash string                 `bson:"request_hash" json:"request_hash"`
	Response    map[string]interface{} `bson:"response,omitempty" json:"response,omitempty"`
	CreatedAt   time.Time              `bson:"created_at" json:"created_at"`
	ExpiresAt   time.Time              `bson:"expires_at" json:"expires_at"`
}
