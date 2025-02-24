FROM golang:1.23-alpine AS builder

WORKDIR /build
RUN apk update --no-cache && apk add --no-cache tzdata
COPY ../ .
RUN go mod tidy
RUN go build -o billing_sync_commands cmd/server/main.go
RUN go build -o billing_accruals_consumer cmd/transaction_inbox/accruals/main.go

FROM alpine:3.14 AS production
RUN apk update --no-cache && apk add --no-cache ca-certificates

COPY --from=builder /usr/share/zoneinfo/Europe/Moscow /usr/share/zoneinfo/Europe/Moscow
ENV TZ=Europe/Moscow
WORKDIR /build
COPY --from=builder build/billing_sync_commands /build/billing_sync_commands
COPY --from=builder build/billing_accruals_consumer /build/billing_accruals_consumer

