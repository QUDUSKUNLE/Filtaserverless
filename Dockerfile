# -------- Stage 1: Build Go binary --------
FROM golang:1.24.3 as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
# Build static binary for linux
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./main.go


# -------- Stage 2: Runtime with yt-dlp and port proxy for Railway --------
FROM debian:bullseye-slim

# Install runtime dependencies: ffmpeg, curl, yt-dlp, socat, tini
RUN apt-get update && apt-get install -y --no-install-recommends \
    ffmpeg \
    curl \
    python3-pip \
    ca-certificates \
    socat \
    tini \
  && pip3 install --no-cache-dir yt-dlp \
  && apt-get clean \
  && rm -rf /var/lib/apt/lists/*

# Copy server binary
COPY --from=builder /app/server /server
RUN chmod +x /server

# Start script to proxy Railway's $PORT to internal 9096
COPY <<'EOF' /start.sh
#!/bin/sh
set -e
# Railway provides PORT; default to 9096 for local usage
PORT="${PORT:-9096}"

# Start the Go server in background (listening on 9096 internally)
/server &
SRV_PID=$!
echo "Server started with PID ${SRV_PID}. Forwarding 0.0.0.0:${PORT} -> 127.0.0.1:9096"

# Forward incoming connections on $PORT to internal 9096
exec socat TCP-LISTEN:${PORT},fork,reuseaddr TCP:127.0.0.1:9096
EOF
RUN chmod +x /start.sh

# Expose a conventional port for docs; Railway ignores EXPOSE and uses $PORT
EXPOSE 8080

# Use tini as init for proper signal handling
ENTRYPOINT ["/usr/bin/tini", "--"]
CMD ["/start.sh"]
