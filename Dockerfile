# ================= Build =================
FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o efm_sub_api ./cmd/server/main.go

# ================= Stage =================
FROM alpine:3.18

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/efm_sub_api .
RUN chmod +x /app/efm_sub_api

COPY --from=builder /app/.env .env

EXPOSE 8080

ENTRYPOINT ["/app/efm_sub_api"]
