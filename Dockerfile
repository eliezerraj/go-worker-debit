#docker build -t go-worker-debit .

FROM golang:1.22 As builder

WORKDIR /app
COPY . .

WORKDIR /app/cmd
RUN go build -o go-worker-debit -ldflags '-linkmode external -w -extldflags "-static"'

FROM alpine

WORKDIR /app
COPY --from=builder /app/cmd/go-worker-debit .

CMD ["/app/go-worker-debit"]