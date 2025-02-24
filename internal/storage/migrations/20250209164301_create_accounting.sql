-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS accounting(
  uuid UUID NOT NULL,
  order_number VARCHAR NOT NULL,
  account_id NUMERIC NOT NULL,
  amount NUMERIC NOT NULL,
  operation VARCHAR NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

  PRIMARY KEY(uuid)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS accounting;
-- +goose StatementEnd
