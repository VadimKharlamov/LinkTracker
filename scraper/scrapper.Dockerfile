FROM golang:1.23.2 AS builder

WORKDIR /app

COPY ../go.mod ./
COPY ../go.sum ./
RUN go mod download

COPY .. .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o scrapper ./cmd/scrapper

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/scrapper .
COPY ../bot/config.yaml .
COPY .env .env

CMD ["./scrapper", "--config=./config.yaml"]
