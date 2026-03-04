# ============================================================
# PicoClaw Docker Image with Kimi 2.5 Support
# ============================================================

# Stage 1: Build the picoclaw binary
FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git make ca-certificates tzdata

WORKDIR /src

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o picoclaw ./cmd/picoclaw

# ============================================================
# Stage 2: Minimal runtime image
# ============================================================
FROM alpine:3.21

LABEL org.opencontainers.image.title="PicoClaw"
LABEL org.opencontainers.image.description="Ultra-lightweight personal AI Assistant with Kimi 2.5 support"
LABEL org.opencontainers.image.source="https://github.com/sipeed/picoclaw"

RUN apk add --no-cache ca-certificates tzdata curl

# Create non-root user
RUN addgroup -g 1000 picoclaw && \
    adduser -D -u 1000 -G picoclaw picoclaw

# Create data directory
RUN mkdir -p /home/picoclaw/.picoclaw && \
    chown -R picoclaw:picoclaw /home/picoclaw

# Copy binary
COPY --from=builder /src/picoclaw /usr/local/bin/picoclaw
RUN chmod +x /usr/local/bin/picoclaw

# Switch to non-root user
USER picoclaw
WORKDIR /home/picoclaw

# Set environment
ENV PICOCLAW_HOME=/home/picoclaw/.picoclaw
ENV HOME=/home/picoclaw

# Expose gateway port
EXPOSE 18790

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD wget -qO- http://localhost:18790/health || exit 1

ENTRYPOINT ["picoclaw"]
CMD ["gateway"]
