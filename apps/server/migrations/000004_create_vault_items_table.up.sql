CREATE TABLE IF NOT EXISTS vault_items (
    id UUID PRIMARY KEY NOT NULL DEFAULT uuidv7(),
    vault_id UUID NOT NULL REFERENCES vaults(id) ON DELETE CASCADE,

    key_salt BYTEA NOT NULL,
    item_nonce BYTEA NOT NULL,
    ciphertext BYTEA NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_vault_items_vault_id ON vault_items(vault_id);
