CREATE TABLE IF NOT EXISTS keyrings (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    vault_id UUID NOT NULL REFERENCES vaults(id) ON DELETE CASCADE,

    encryption_salt BYTEA NOT NULL,
    encryption_nonce BYTEA NOT NULL,
    encrypted_master_key BYTEA NOT NULL,
    encryption_time_cost INTEGER NOT NULL CHECK (encryption_time_cost >= 1),
    encryption_memory_cost INTEGER NOT NULL CHECK (encryption_memory_cost >= 1),
    encryption_parallelism INTEGER NOT NULL CHECK (encryption_parallelism >= 1),
    encryption_key_size INTEGER NOT NULL CHECK (encryption_key_size >= 1),

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (user_id, vault_id)
);

CREATE INDEX IF NOT EXISTS idx_keyrings_vault_id ON keyrings(vault_id);
