FROM golang:1.24.3 AS builder

WORKDIR /app

COPY . .

RUN make build

FROM debian:bookworm-slim
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/dist/server /server

ENTRYPOINT ["/server"]
