package models

import (
	"time"

	"github.com/google/uuid"
)

type SolTxStatus string

const (
	SolTxPending   SolTxStatus = "pending"
	SolTxConfirmed SolTxStatus = "confirmed"
	SolTxFailed    SolTxStatus = "failed"
)

type SolTxType string

const (
	SolTxTopup         SolTxType = "topup"
	SolTxPurchase      SolTxType = "purchase"
	SolTxSellerPayout  SolTxType = "seller_payout"
	SolTxEscrowDeposit SolTxType = "escrow_deposit"
	SolTxEscrowRelease SolTxType = "escrow_release"
	SolTxEscrowRefund  SolTxType = "escrow_refund"
)

type EscrowStatus string

const (
	EscrowActive    EscrowStatus = "active"
	EscrowCompleted EscrowStatus = "completed"
	EscrowRefunded  EscrowStatus = "refunded"
)

// SolTransaction represents an on-chain Solana transaction tracked by the backend.
type SolTransaction struct {
	ID             uuid.UUID   `json:"id"`
	UserID         uuid.UUID   `json:"user_id"`
	TxSignature    string      `json:"tx_signature"`
	TxType         SolTxType   `json:"tx_type"`
	AmountLamports int64       `json:"amount_lamports"`
	FromWallet     string      `json:"from_wallet"`
	ToWallet       string      `json:"to_wallet"`
	Status         SolTxStatus `json:"status"`
	ReferenceID    *uuid.UUID  `json:"reference_id,omitempty"`
	ConfirmedAt    *time.Time  `json:"confirmed_at,omitempty"`
	CreatedAt      time.Time   `json:"created_at"`
}

// SolEscrow tracks the on-chain escrow state for a dataset purchase.
type SolEscrow struct {
	ID               uuid.UUID    `json:"id"`
	BuyerID          uuid.UUID    `json:"buyer_id"`
	DatasetID        uuid.UUID    `json:"dataset_id"`
	EscrowPDA        string       `json:"escrow_pda"`
	VaultPDA         string       `json:"vault_pda"`
	AmountLamports   int64        `json:"amount_lamports"`
	ReleasedLamports int64        `json:"released_lamports"`
	Status           EscrowStatus `json:"status"`
	DepositSignature string       `json:"deposit_signature"`
	CreatedAt        time.Time    `json:"created_at"`
}
