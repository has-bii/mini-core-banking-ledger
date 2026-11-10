-- +goose Up
CREATE TYPE account_type AS ENUM ('main', 'sub_account', 'savings');

CREATE TABLE accounts (
    id uuid PRIMARY KEY DEFAULT uuidv7(),
    user_id uuid NOT NULL,
    name text NOT NULL,
    type account_type NOT NULL,
    sort_order smallint NOT NULL CHECK (sort_order >= 0),
    balance bigint NOT NULL DEFAULT 0,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT accounts_user_id_sort_order_key UNIQUE (user_id, sort_order)
);

-- +goose StatementBegin
CREATE FUNCTION insert_new_main_account()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO accounts (user_id, name, type, sort_order)
    VALUES (NEW.id, 'Main Account', 'main', 0);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd
CREATE TRIGGER trigger_after_user_inserted
AFTER INSERT ON users
FOR EACH ROW
EXECUTE FUNCTION insert_new_main_account();

-- +goose Down
DROP TRIGGER IF EXISTS trigger_after_user_inserted ON users;
DROP FUNCTION IF EXISTS insert_new_main_account();
DROP TABLE IF EXISTS accounts;
DROP TYPE IF EXISTS account_type;
