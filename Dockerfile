FROM golang:1.24.2-alpine AS builder

WORKDIR /app

RUN apk add --no-cache gcc musl-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 go build -ldflags="-s -w" -o /app/bin/poll-service ./cmd/main.go

FROM alpine:3.18

WORKDIR /app

RUN apk add --no-cache libc6-compat

COPY --from=builder /app/bin/poll-service ./poll-service
COPY config.yaml ./config.yaml

RUN mkdir -p /data && chmod 777 /data

ENV APP_ENV=production

VOLUME /data

CMD ["./poll-service"]
