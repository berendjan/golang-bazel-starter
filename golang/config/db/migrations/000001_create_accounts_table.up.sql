CREATE TABLE IF NOT EXISTS accounts (
    id BYTEA PRIMARY KEY,
    type INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_accounts_type ON accounts(type);
CREATE INDEX IF NOT EXISTS idx_accounts_created_at ON accounts(created_at DESC);
