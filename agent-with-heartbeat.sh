#!/bin/sh
set -e

TEAM_ID="Development-Team-67d82e93"
TEAM_KEY="pk_team_MK2-SdlgjeX0Mwt0BrkHKPY1DPs2ZJ6tj9xfkMs5NUM="
GATEWAY="http://192.168.6.122:18790"
AGENT_ID="dev-agent-docker"
ROLE="fullstack"

mkdir -p /home/picoclaw/.picoclaw/agent_teams

echo "Joining team..."
JOIN_OUTPUT=$(picoclaw team join "$TEAM_ID" --key "$TEAM_KEY" --gateway "$GATEWAY" --role "$ROLE" --agent-id "$AGENT_ID" 2>&1)
echo "$JOIN_OUTPUT"

# Extract session from saved file
SESSION_FILE="/home/picoclaw/.picoclaw/agent_teams/$TEAM_ID.json"
if [ -f "$SESSION_FILE" ]; then
    SESSION_ID=$(grep '"session_id"' "$SESSION_FILE" | head -1 | cut -d'"' -f4)
fi

if [ -z "$SESSION_ID" ]; then
    echo "ERROR: Failed to get session ID"
    exit 1
fi

echo "Got session: $SESSION_ID"
echo "Starting heartbeat loop..."
echo ""

while true; do
    RESPONSE=$(curl -s -X POST "$GATEWAY/api/teams/heartbeat" \
        -H "Content-Type: application/json" \
        -d "{\"session_id\":\"$SESSION_ID\",\"team_id\":\"$TEAM_ID\",\"agent_id\":\"$AGENT_ID\"}" 2>&1)
    
    TIMESTAMP=$(date '+%H:%M:%S')
    
    if echo "$RESPONSE" | grep -q "Invalid or expired session"; then
        echo "[$TIMESTAMP] Session expired, rejoining..."
        JOIN_OUTPUT=$(picoclaw team join "$TEAM_ID" --key "$TEAM_KEY" --gateway "$GATEWAY" --role "$ROLE" --agent-id "$AGENT_ID" 2>&1)
        echo "$JOIN_OUTPUT"
        if [ -f "$SESSION_FILE" ]; then
            SESSION_ID=$(grep '"session_id"' "$SESSION_FILE" | head -1 | cut -d'"' -f4)
        fi
        echo "[$TIMESTAMP] New session: $SESSION_ID"
    else
        echo "[$TIMESTAMP] Heartbeat: $RESPONSE"
    fi
    
    sleep 10
done
