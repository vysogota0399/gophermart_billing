-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS inbox_events(
  uuid uuid NOT NULL,
  state VARCHAR NOT NULL,
  name VARCHAR NOT NULL,
  message JSONB NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

  PRIMARY KEY(uuid)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS inbox_events;
-- +goose StatementEnd
