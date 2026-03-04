# ============================================================
# PicoClaw Docker Image - Using Release Binary
# ============================================================
# This Dockerfile downloads the pre-built release binary from GitHub
# instead of compiling from source.
#
# Usage:
#   docker build -t picoclaw:latest .
#   docker build --build-arg VERSION=v0.3.0-kimi -t picoclaw:v0.3.0-kimi .
# ============================================================

FROM alpine:3.21

# Build arguments
ARG VERSION=latest
ARG GITHUB_OWNER=brendantza
ARG GITHUB_REPO=picoclaw

# Labels
LABEL org.opencontainers.image.title="PicoClaw"
LABEL org.opencontainers.image.description="Ultra-lightweight personal AI Assistant with Kimi 2.5 support"
LABEL org.opencontainers.image.source="https://github.com/${GITHUB_OWNER}/${GITHUB_REPO}"
LABEL org.opencontainers.image.version="${VERSION}"

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    curl \
    jq

# Create non-root user
RUN addgroup -g 1000 picoclaw && \
    adduser -D -u 1000 -G picoclaw picoclaw

# Create data directory
RUN mkdir -p /home/picoclaw/.picoclaw && \
    chown -R picoclaw:picoclaw /home/picoclaw

# Detect architecture and download appropriate binary
RUN set -eux; \
    ARCH=$(uname -m); \
    case "$ARCH" in \
        x86_64)  BINARY_ARCH="amd64" ;; \
        aarch64) BINARY_ARCH="arm64" ;; \
        armv7l)  BINARY_ARCH="arm" ;; \
        *)       echo "Unsupported architecture: $ARCH"; exit 1 ;; \
    esac; \
    \
    if [ "$VERSION" = "latest" ]; then \
        echo "Fetching latest release version..."; \
        VERSION=$(curl -sL "https://api.github.com/repos/${GITHUB_OWNER}/${GITHUB_REPO}/releases/latest" | jq -r '.tag_name'); \
        echo "Latest version: $VERSION"; \
    fi; \
    \
    DOWNLOAD_URL="https://github.com/${GITHUB_OWNER}/${GITHUB_REPO}/releases/download/${VERSION}/picoclaw-linux-${BINARY_ARCH}"; \
    echo "Downloading PicoClaw ${VERSION} for linux-${BINARY_ARCH}..."; \
    echo "URL: ${DOWNLOAD_URL}"; \
    \
    curl -L --progress-bar -o /usr/local/bin/picoclaw "${DOWNLOAD_URL}"; \
    chmod +x /usr/local/bin/picoclaw; \
    \
    # Verify binary works
    /usr/local/bin/picoclaw version || { \
        echo "ERROR: Binary verification failed"; \
        exit 1; \
    }; \
    \
    echo "PicoClaw ${VERSION} installed successfully for linux-${BINARY_ARCH}"

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
