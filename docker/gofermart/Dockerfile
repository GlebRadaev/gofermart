FROM golang:1.23 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o gofermart ./cmd/gofermart

FROM debian:bookworm-slim

WORKDIR /app

COPY --from=builder /app/gofermart /app/gofermart

COPY .env /app/.env

ENV RUN_ADDRESS=${RUN_ADDRESS}
ENV DATABASE_URI=${DATABASE_URI}
ENV ACCRUAL_SYSTEM_ADDRESS=${ACCRUAL_SYSTEM_ADDRESS}

CMD ["/app/gofermart"]
