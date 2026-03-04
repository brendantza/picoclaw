FROM alpine:3.21

ARG TARGETPLATFORM
ARG TARGETARCH
ARG VERSION
ARG GITHUB_REPO=brendantza/picoclaw

RUN apk add --no-cache ca-certificates tzdata curl jq

# Create non-root user
RUN addgroup -g 1000 picoclaw && \
    adduser -D -u 1000 -G picoclaw picoclaw

# Create data directory
RUN mkdir -p /home/picoclaw/.picoclaw && \
    chown -R picoclaw:picoclaw /home/picoclaw

# Download binary from GitHub release based on architecture
RUN set -eux; \
    case "${TARGETARCH}" in \
        amd64) ARCH='amd64' ;; \
        arm64) ARCH='arm64' ;; \
        arm) ARCH='arm' ;; \
        *) echo "Unsupported architecture: ${TARGETARCH}"; exit 1 ;; \
    esac; \
    if [ -z "${VERSION}" ]; then \
        echo "VERSION build arg is required"; exit 1; \
    fi; \
    echo "Downloading picoclaw ${VERSION} for linux-${ARCH}..."; \
    curl -fsSL \
        "https://github.com/${GITHUB_REPO}/releases/download/${VERSION}/picoclaw-linux-${ARCH}" \
        -o /usr/local/bin/picoclaw; \
    chmod +x /usr/local/bin/picoclaw

# Verify binary works
RUN /usr/local/bin/picoclaw version

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
