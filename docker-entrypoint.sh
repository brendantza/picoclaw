#!/bin/sh
# PicoClaw Docker Entrypoint
# Generates config.json from environment variables

set -e

CONFIG_DIR="${PICOCLAW_HOME:-/home/picoclaw/.picoclaw}"
CONFIG_FILE="${CONFIG_DIR}/config.json"

# Create config directory if it doesn't exist
mkdir -p "$CONFIG_DIR"

# Generate config.json from environment variables
if [ ! -f "$CONFIG_FILE" ] || [ "${PICOCLAW_OVERWRITE_CONFIG:-false}" = "true" ]; then
    echo "Generating config.json from environment variables..."
    
    # Get values from env vars with defaults
    MODEL_NAME="${PICOCLAW_AGENTS_DEFAULTS_MODEL_NAME:-kimi-k2.5}"
    KIMI_API_KEY="${PICOCLAW_PROVIDERS_KIMI_CODING_API_KEY:-}"
    KIMI_API_BASE="${PICOCLAW_PROVIDERS_KIMI_CODING_API_BASE:-https://api.kimi.com/coding}"
    KIMI_PROXY="${PICOCLAW_PROVIDERS_KIMI_CODING_PROXY:-}"
    GATEWAY_HOST="${PICOCLAW_GATEWAY_HOST:-0.0.0.0}"
    GATEWAY_PORT="${PICOCLAW_GATEWAY_PORT:-18790}"
    
    # Build model_list JSON
    MODEL_LIST=""
    
    # Add Kimi Coding model if API key is provided
    if [ -n "$KIMI_API_KEY" ]; then
        MODEL_LIST="${MODEL_LIST}{
      \"model_name\": \"${MODEL_NAME}\",
      \"model\": \"kimi-coding/k2p5\",
      \"api_key\": \"${KIMI_API_KEY}\",
      \"api_base\": \"${KIMI_API_BASE}\""
        if [ -n "$KIMI_PROXY" ]; then
            MODEL_LIST="${MODEL_LIST},
      \"proxy\": \"${KIMI_PROXY}\""
        fi
        MODEL_LIST="${MODEL_LIST}
    }"
    fi
    
    # If no models, add a placeholder that will fail gracefully
    if [ -z "$MODEL_LIST" ]; then
        MODEL_LIST="{
      \"model_name\": \"${MODEL_NAME}\",
      \"model\": \"anthropic/claude-sonnet-4.5\",
      \"api_key\": \"\"
    }"
        echo "WARNING: No API key provided. Please set PICOCLAW_PROVIDERS_KIMI_CODING_API_KEY"
    fi
    
    cat > "$CONFIG_FILE" << EOF
{
  "agents": {
    "defaults": {
      "model_name": "${MODEL_NAME}",
      "max_tokens": 4096,
      "max_tool_iterations": 10
    }
  },
  "model_list": [
    ${MODEL_LIST}
  ],
  "gateway": {
    "host": "${GATEWAY_HOST}",
    "port": ${GATEWAY_PORT}
  },
  "tools": {
    "web": {
      "brave": {
        "enabled": false
      },
      "duckduckgo": {
        "enabled": true
      }
    }
  },
  "heartbeat": {
    "enabled": false
  },
  "devices": {
    "enabled": false
  }
}
EOF
    
    echo "Config file generated at: $CONFIG_FILE"
else
    echo "Using existing config: $CONFIG_FILE"
fi

# Execute the main command
exec "$@"
