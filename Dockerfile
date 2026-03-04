FROM alpine:3.21

ARG TARGETPLATFORM
ARG TARGETARCH

RUN apk add --no-cache ca-certificates tzdata curl jq

# Create non-root user
RUN addgroup -g 1000 picoclaw && \
    adduser -D -u 1000 -G picoclaw picoclaw

# Create data directory
RUN mkdir -p /home/picoclaw/.picoclaw && \
    chown -R picoclaw:picoclaw /home/picoclaw

# Copy pre-built binary from build context
COPY picoclaw /usr/local/bin/picoclaw
RUN chmod +x /usr/local/bin/picoclaw

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
