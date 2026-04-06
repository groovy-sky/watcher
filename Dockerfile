# ── Stage 1: build ───────────────────────────────────────────────────────────
FROM golang:1.26-alpine AS builder

WORKDIR /src

# Download dependencies first (cached layer)
COPY go.mod go.sum ./
RUN go mod download

# Build a fully static binary
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /watcher .

# ── Stage 2: runtime ─────────────────────────────────────────────────────────
FROM alpine:3.21

# chromium brings in all required shared libraries (NSS, fontconfig, etc.).
# chromium-chromedriver is not needed; we drive the browser via CDP.
RUN apk add --no-cache \
        chromium \
        ttf-freefont \
    && ln -sf /usr/bin/chromium-browser /usr/local/bin/google-chrome

COPY --from=builder /watcher /usr/local/bin/watcher

# watcher is not a server – no ports to expose.
# Run as a non-root user; --no-sandbox is already set in main.go so this is safe.
RUN adduser -D appuser
USER appuser

ENTRYPOINT ["/usr/local/bin/watcher"]