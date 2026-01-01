#!/bin/sh
set -e
# Railway provides PORT dynamically; default to 9096 for local usage
PORT="${PORT:-9096}"

# Start the Go server in background (listening on 9096 internally)
/server &
SRV_PID=$!
echo "Server started with PID ${SRV_PID}. Forwarding 0.0.0.0:${PORT} -> 127.0.0.1:9096"

# Forward incoming connections on $PORT to internal 9096
exec socat TCP-LISTEN:${PORT},fork,reuseaddr TCP:127.0.0.1:9096
