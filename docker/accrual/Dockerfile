FROM alpine:latest

RUN apk add --no-cache libc6-compat

WORKDIR /app

RUN chmod -R 777 /app

COPY cmd/accrual/accrual_linux_amd64 /app/accrual

RUN chmod +x /app/accrual

RUN chmod -R 777 /app

ENV RUN_ADDRESS=${RUN_ADDRESS}

CMD ["./accrual"]