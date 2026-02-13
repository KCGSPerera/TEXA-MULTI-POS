CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    branch_id UUID NOT NULL,
    code VARCHAR(50) NOT NULL,
    name VARCHAR(150) NOT NULL,
    type VARCHAR(20) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT fk_accounts_branch FOREIGN KEY (branch_id) REFERENCES branches(id),
    CONSTRAINT uq_accounts_branch_code UNIQUE (branch_id, code),
    CONSTRAINT ck_accounts_type CHECK (type IN ('ASSET', 'LIABILITY', 'EQUITY', 'INCOME', 'EXPENSE'))
);

CREATE INDEX IF NOT EXISTS idx_accounts_branch_id ON accounts(branch_id);

CREATE TABLE IF NOT EXISTS journal_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    branch_id UUID NOT NULL,
    reference_type VARCHAR(30) NOT NULL,
    reference_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT fk_journal_entries_branch FOREIGN KEY (branch_id) REFERENCES branches(id)
);

CREATE INDEX IF NOT EXISTS idx_journal_entries_branch_created ON journal_entries(branch_id, created_at);

CREATE TABLE IF NOT EXISTS journal_lines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    journal_entry_id UUID NOT NULL,
    account_id UUID NOT NULL,
    debit NUMERIC(14,2) NOT NULL DEFAULT 0,
    credit NUMERIC(14,2) NOT NULL DEFAULT 0,
    CONSTRAINT fk_journal_lines_entry FOREIGN KEY (journal_entry_id) REFERENCES journal_entries(id) ON DELETE CASCADE,
    CONSTRAINT fk_journal_lines_account FOREIGN KEY (account_id) REFERENCES accounts(id),
    CONSTRAINT ck_journal_lines_debit_non_negative CHECK (debit >= 0),
    CONSTRAINT ck_journal_lines_credit_non_negative CHECK (credit >= 0)
);

CREATE INDEX IF NOT EXISTS idx_journal_lines_entry_id ON journal_lines(journal_entry_id);
CREATE INDEX IF NOT EXISTS idx_journal_lines_account_id ON journal_lines(account_id);
