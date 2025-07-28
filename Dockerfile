# -------- Stage 1: Build Go binary --------
  FROM golang:1.22 as builder

  WORKDIR /app
  
  COPY go.mod go.sum ./
  RUN go mod download
  
  COPY . .
  RUN CGO_ENABLED=0 GOOS=linux go build -o server ./main.go
  
  
  # -------- Stage 2: Runtime with yt-dlp --------
  FROM debian:bullseye-slim
  
  # Install dependencies: ffmpeg, curl, yt-dlp
  RUN apt-get update && apt-get install -y --no-install-recommends \
      ffmpeg \
      curl \
      python3-pip \
      ca-certificates \
      && pip3 install --no-cache-dir yt-dlp \
      && apt-get clean \
      && rm -rf /var/lib/apt/lists/*
  
  COPY --from=builder /app/server /server
  
  # Make binary executable just in case
  RUN chmod +x /server
  
  EXPOSE 9096
  ENTRYPOINT ["/server"]
  