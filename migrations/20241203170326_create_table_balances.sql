-- +goose Up
-- +goose StatementBegin
CREATE TABLE balances (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    current_balance NUMERIC(10, 2) DEFAULT 0,
    withdrawn_total NUMERIC(10, 2) DEFAULT 0
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS balances;
-- +goose StatementEnd


