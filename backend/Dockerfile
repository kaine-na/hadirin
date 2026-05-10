# ---- Builder ----
# Catatan: go.mod mendeklarasikan `go 1.25.0`. Image ini menggunakan
# golang:1.22-alpine sesuai spesifikasi; Go toolchain auto-upgrade
# (GOTOOLCHAIN=auto, default sejak Go 1.21) akan otomatis mengunduh
# toolchain 1.25 yang dibutuhkan saat build.
FROM golang:1.22-alpine AS builder

WORKDIR /src

# Install build deps (git needed for VCS-based dependencies + toolchain auto-download)
RUN apk add --no-cache git

ENV GOTOOLCHAIN=auto

# Cache deps
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build static binary
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
RUN go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

# ---- Runtime ----
FROM alpine:latest AS runtime

# Root certs + timezone data, no cache
RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -S appgroup && \
    adduser -S -G appgroup -h /app appuser

WORKDIR /app

# Copy binary + migrations (runtime reads ./migrations)
COPY --from=builder /out/server /app/server
COPY --from=builder /src/migrations /app/migrations

# Prepare uploads dir with proper ownership
RUN mkdir -p /app/uploads && chown -R appuser:appgroup /app

USER appuser

ENV PORT=8080 \
    UPLOAD_DIR=/app/uploads \
    TZ=Asia/Jakarta

EXPOSE 8080

ENTRYPOINT ["/app/server"]
