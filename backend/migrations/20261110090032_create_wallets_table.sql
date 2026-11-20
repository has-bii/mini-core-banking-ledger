-- +goose Up
CREATE TYPE wallet_type AS ENUM ('main', 'sub_wallet', 'savings');

CREATE TABLE wallets (
    id uuid PRIMARY KEY DEFAULT uuidv7(),
    user_id uuid NOT NULL,
    name text NOT NULL,
    type wallet_type NOT NULL,
    sort_order smallint NOT NULL CHECK (sort_order >= 0),
    balance bigint NOT NULL DEFAULT 0,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT wallets_user_id_sort_order_key UNIQUE (user_id, sort_order)
);

-- +goose StatementBegin
CREATE FUNCTION insert_new_main_wallet()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO wallets (user_id, name, type, sort_order)
    VALUES (NEW.id, 'Main Wallet', 'main', 0);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd
CREATE TRIGGER trigger_after_user_inserted
AFTER INSERT ON users
FOR EACH ROW
EXECUTE FUNCTION insert_new_main_wallet();

-- +goose Down
DROP TRIGGER IF EXISTS trigger_after_user_inserted ON users;
DROP FUNCTION IF EXISTS insert_new_main_wallet();
DROP TABLE IF EXISTS wallets;
DROP TYPE IF EXISTS wallet_type;
