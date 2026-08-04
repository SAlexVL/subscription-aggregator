FROM golang:1.23-alpine AS builder

WORKDIR /src

ENV GOTOOLCHAIN=auto

RUN apk add --no-cache git ca-certificates

COPY go.mod ./
COPY go.sum* ./
RUN go mod download || true

COPY . .
RUN go mod tidy && CGO_ENABLED=0 GOOS=linux go build -o /bin/api ./cmd/api

FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /bin/api /app/api
COPY config.yaml /app/config.yaml
COPY migrations /app/migrations

EXPOSE 8080

ENTRYPOINT ["/app/api"]
