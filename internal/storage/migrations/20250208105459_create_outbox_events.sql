-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS outbox_events(
  uuid uuid NOT NULL,
  send BOOLEAN DEFAULT false,
  name VARCHAR NOT NULL,
  message JSONB NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY(uuid)
)
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS outbox_events;
-- +goose StatementEnd
