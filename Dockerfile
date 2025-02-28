#docker build -t go-worker-debit .

FROM golang:1.23.3 As builder

RUN apt-get update && apt-get install bash && apt-get install -y --no-install-recommends ca-certificates

WORKDIR /app
COPY . .

WORKDIR /app/cmd
RUN go build -o go-worker-debit -ldflags '-linkmode external -w -extldflags "-static"'
RUN go mod tidy

FROM alpine

WORKDIR /app
COPY --from=builder /app/cmd/go-worker-debit .
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

CMD ["/app/go-worker-debit"]