-- +goose Up
CREATE TABLE credentials (
    id uuid PRIMARY KEY DEFAULT uuidv7(),
    user_id uuid NOT NULL,
    name text NOT NULL,
    counter bigint NOT NULL,
    public_key text NOT NULL,
    credential_id text NOT NULL,
    transports text,
    backed_up boolean NOT NULL DEFAULT FALSE,
    created_at timestamptz NOT NULL DEFAULT now(),
    last_used_at timestamptz NOT NULL DEFAULT now(),
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

CREATE INDEX credentials_user_id_idx ON credentials (user_id);

-- +goose Down
DROP TABLE IF EXISTS credentials;
