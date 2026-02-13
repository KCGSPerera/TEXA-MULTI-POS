CREATE TABLE IF NOT EXISTS branch_bootstrap (
    branch_id UUID PRIMARY KEY,
    accounts_seeded BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT fk_branch_bootstrap_branch FOREIGN KEY (branch_id) REFERENCES branches(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_branch_bootstrap_seeded ON branch_bootstrap(accounts_seeded);
