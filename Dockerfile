FROM alpine:3.21

ARG TARGETPLATFORM
ARG TARGETARCH

RUN apk add --no-cache ca-certificates tzdata curl jq su-exec

# Create non-root user
RUN addgroup -g 1000 picoclaw && \
    adduser -D -u 1000 -G picoclaw picoclaw

# Create data directory
RUN mkdir -p /home/picoclaw/.picoclaw && \
    chown -R picoclaw:picoclaw /home/picoclaw

# Copy pre-built binaries from build context
COPY picoclaw /usr/local/bin/picoclaw
COPY picoclaw-launcher /usr/local/bin/picoclaw-launcher
RUN chmod +x /usr/local/bin/picoclaw /usr/local/bin/picoclaw-launcher

# Verify binaries work
RUN /usr/local/bin/picoclaw version

# Copy entrypoint script
COPY docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh
RUN chmod +x /usr/local/bin/docker-entrypoint.sh

# Set environment
ENV PICOCLAW_HOME=/home/picoclaw/.picoclaw
ENV HOME=/home/picoclaw

# Expose gateway port
EXPOSE 18790

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD wget -qO- http://localhost:18790/health || exit 1

ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]
CMD []
