#!/bin/sh
# Combined launcher + heartbeat for agent node

set -e

TEAM_ID="${PICOCLAW_TEAM_ID:-Development-Team-67d82e93}"
TEAM_KEY="${PICOCLAW_TEAM_KEY:-pk_team_MK2-SdlgjeX0Mwt0BrkHKPY1DPs2ZJ6tj9xfkMs5NUM=}"
GATEWAY="${PICOCLAW_GATEWAY_ADDRESS:-http://192.168.6.122:18790}"
AGENT_ID="${PICOCLAW_AGENT_ID:-dev-agent-docker}"
ROLE="${PICOCLAW_AGENT_ROLE:-fullstack}"

mkdir -p /home/picoclaw/.picoclaw/agent_teams /home/picoclaw/.picoclaw/teams

# Start launcher in background
echo "Starting PicoClaw Launcher..."
/usr/local/bin/picoclaw-launcher -public &
LAUNCHER_PID=$!

# Wait for launcher to be ready
echo "Waiting for launcher to start..."
for i in 1 2 3 4 5; do
    if wget -qO- http://localhost:18800/health >/dev/null 2>&1; then
        echo "Launcher is ready!"
        break
    fi
    sleep 2
done

# Join team
echo "Joining team..."
/usr/local/bin/picoclaw team join "$TEAM_ID" \
    --key "$TEAM_KEY" \
    --gateway "$GATEWAY" \
    --role "$ROLE" \
    --agent-id "$AGENT_ID" 2>&1 || true

# Extract session
SESSION_FILE="/home/picoclaw/.picoclaw/agent_teams/$TEAM_ID.json"
SESSION_ID=""
if [ -f "$SESSION_FILE" ]; then
    SESSION_ID=$(grep '"session_id"' "$SESSION_FILE" | head -1 | cut -d'"' -f4)
fi

if [ -z "$SESSION_ID" ]; then
    echo "WARNING: Failed to get session ID, will retry..."
fi

echo "Got session: $SESSION_ID"
echo "Starting heartbeat loop..."

# Heartbeat loop - runs forever
while true; do
    # Check if launcher is still running
    if ! kill -0 $LAUNCHER_PID 2>/dev/null; then
        echo "Launcher died, restarting..."
        /usr/local/bin/picoclaw-launcher -public &
        LAUNCHER_PID=$!
        sleep 3
    fi

    # Send heartbeat if we have a session
    if [ -n "$SESSION_ID" ]; then
        RESPONSE=$(curl -s -X POST "$GATEWAY/api/teams/heartbeat" \
            -H "Content-Type: application/json" \
            -d "{\"session_id\":\"$SESSION_ID\",\"team_id\":\"$TEAM_ID\",\"agent_id\":\"$AGENT_ID\"}" 2>&1)
        
        TIMESTAMP=$(date '+%H:%M:%S')
        
        if echo "$RESPONSE" | grep -q "Invalid or expired session"; then
            echo "[$TIMESTAMP] Session expired, rejoining..."
            /usr/local/bin/picoclaw team join "$TEAM_ID" \
                --key "$TEAM_KEY" \
                --gateway "$GATEWAY" \
                --role "$ROLE" \
                --agent-id "$AGENT_ID" 2>&1 || true
            if [ -f "$SESSION_FILE" ]; then
                SESSION_ID=$(grep '"session_id"' "$SESSION_FILE" | head -1 | cut -d'"' -f4)
            fi
            echo "[$TIMESTAMP] New session: $SESSION_ID"
        else
            echo "[$TIMESTAMP] Heartbeat: $RESPONSE"
        fi
    else
        # Try to join if no session
        echo "[$(date '+%H:%M:%S')] No session, attempting to join..."
        /usr/local/bin/picoclaw team join "$TEAM_ID" \
            --key "$TEAM_KEY" \
            --gateway "$GATEWAY" \
            --role "$ROLE" \
            --agent-id "$AGENT_ID" 2>&1 || true
        if [ -f "$SESSION_FILE" ]; then
            SESSION_ID=$(grep '"session_id"' "$SESSION_FILE" | head -1 | cut -d'"' -f4)
        fi
    fi
    
    sleep 10
done
