-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS orders(
  uuid UUID UNIQUE NOT NULL,
  number VARCHAR UNIQUE NOT NULL,
  State INT NOT NULL,
  account_id NUMERIC NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  uploaded_at TIMESTAMP WITH TIME ZONE NOT NULL,

  PRIMARY KEY(uuid)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
