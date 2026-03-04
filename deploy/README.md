# 🐳 PicoClaw Docker Deployment Guide

This directory contains Docker deployment configurations for PicoClaw with Kimi 2.5 support.

## Quick Start

### Using Docker Compose

```bash
# Clone the repository
git clone https://github.com/sipeed/picoclaw.git
cd picoclaw

# Create a .env file with your Kimi API key
cat > .env << EOF
PICOCLAW_PROVIDER=moonshot
PICOCLAW_MODEL=k2p5
KIMI_API_KEY=sk-kimi-your-api-key-here
PICOCLAW_PORT=18790
EOF

# Start the container
docker-compose up -d

# Check logs
docker-compose logs -f picoclaw
```

### Using Docker Run

```bash
docker run -d \
  --name picoclaw \
  -p 18790:18790 \
  -e PICOCLAW_AGENTS_DEFAULTS_MODEL_NAME=k2p5 \
  -v picoclaw-data:/app/.picoclaw \
  --restart unless-stopped \
  picoclaw/picoclaw:latest
```

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PICOCLAW_PROVIDER` | Default LLM provider | `moonshot` |
| `PICOCLAW_MODEL` | Model name | `k2p5` |
| `PICOCLAW_MAX_TOKENS` | Max tokens | `32768` |
| `KIMI_API_KEY` | Your Kimi API key | (required) |

### Volume Mounts

| Path | Description |
|------|-------------|
| `/app/.picoclaw` | Configuration, state, and workspace |

## Coolify Deployment

### Method 1: Using Coolify Service Template

1. In Coolify, go to **Services** → **Add Service**
2. Choose **Docker Compose** as the source
3. Paste the contents of `coolify-service.yaml`
4. Configure the environment variables
5. Deploy

### Method 2: Using Git Repository

1. In Coolify, go to **Services** → **Add Service** → **Git Repository**
2. Repository URL: `https://github.com/sipeed/picoclaw`
3. Branch: `main`
4. Docker Compose file: `docker-compose.yml`
5. Configure environment variables in Coolify UI
6. Deploy

### Coolify Environment Variables

Set these in your Coolify service configuration:

```
PICOCLAW_PROVIDER=moonshot
PICOCLAW_MODEL=k2p5
PICOCLAW_MAX_TOKENS=32768
KIMI_API_KEY=sk-kimi-your-api-key
```

## Building Custom Image

```bash
# Build locally
cd deploy
chmod +x docker-build.sh
./docker-build.sh

# Or manually
docker build -t picoclaw/picoclaw:latest ..
```

## Health Checks

The container includes a health check that monitors the gateway endpoint:

```bash
# Check health
curl http://localhost:18790/health
```

## Troubleshooting

### Container won't start

Check logs:
```bash
docker-compose logs picoclaw
```

### API key not working

Ensure the config.json is properly mounted:
```bash
docker exec picoclaw cat /app/.picoclaw/config.json
```

### Port already in use

Change the port mapping in docker-compose.yml:
```yaml
ports:
  - "8080:18790"  # Map to port 8080 instead
```

## Resource Requirements

- **CPU**: 0.5 cores minimum, 2 cores recommended
- **Memory**: 128 MB minimum, 512 MB recommended
- **Storage**: 1 GB for data volume
- **Network**: Outbound HTTPS access required

## Security Notes

- The container runs as non-root user (`picoclaw`)
- API keys should be passed via environment variables or mounted config
- No new privileges are granted to the container
- All capabilities are dropped except `NET_BIND_SERVICE`
