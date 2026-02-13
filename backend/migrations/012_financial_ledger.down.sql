DROP INDEX IF EXISTS idx_journal_lines_account_id;
DROP INDEX IF EXISTS idx_journal_lines_entry_id;
DROP TABLE IF EXISTS journal_lines;

DROP INDEX IF EXISTS idx_journal_entries_branch_created;
DROP TABLE IF EXISTS journal_entries;

DROP INDEX IF EXISTS idx_accounts_branch_id;
DROP TABLE IF EXISTS accounts;
