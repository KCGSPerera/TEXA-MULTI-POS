package models

import "time"

type Account struct {
	ID       string `json:"id"`
	BranchID string `json:"branch_id"`
	Code     string `json:"code"`
	Name     string `json:"name"`
	Type     string `json:"type"`
}

type JournalEntry struct {
	ID            string        `json:"id"`
	BranchID      string        `json:"branch_id"`
	ReferenceType string        `json:"reference_type"`
	ReferenceID   string        `json:"reference_id"`
	CreatedAt     time.Time     `json:"created_at"`
	Lines         []JournalLine `json:"lines"`
}

type JournalLine struct {
	ID             string  `json:"id"`
	JournalEntryID string  `json:"journal_entry_id"`
	AccountID      string  `json:"account_id"`
	Debit          float64 `json:"debit"`
	Credit         float64 `json:"credit"`
}

type JournalLineInput struct {
	AccountID string
	Debit     float64
	Credit    float64
}
