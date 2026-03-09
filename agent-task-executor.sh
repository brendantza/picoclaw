#!/bin/sh
# PicoClaw Agent Task Executor
# Polls for tasks from controller and executes them

set -e

TEAM_ID="${PICOCLAW_TEAM_ID:-Dev-Team-62747d75}"
GATEWAY="${PICOCLAW_GATEWAY_ADDRESS:-http://192.168.6.122:18790}"
AGENT_ID="${PICOCLAW_AGENT_ID:-dev-agent-docker}"
SESSION_FILE="/home/picoclaw/.picoclaw/agent_teams/$TEAM_ID.json"

echo "Task Executor started"
echo "Team: $TEAM_ID"
echo "Agent: $AGENT_ID"
echo "Gateway: $GATEWAY"
echo ""

# Get session ID
get_session() {
    if [ -f "$SESSION_FILE" ]; then
        grep '"session_id"' "$SESSION_FILE" | head -1 | cut -d'"' -f4
    fi
}

# Poll for tasks
poll_tasks() {
    SESSION_ID=$(get_session)
    if [ -z "$SESSION_ID" ]; then
        echo "No session available"
        return 1
    fi

    curl -s -X GET "$GATEWAY/api/teams/$TEAM_ID/agents/$AGENT_ID/tasks" \
        -H "X-Session-ID: $SESSION_ID" 2>/dev/null
}

# Submit task result
submit_result() {
    TASK_ID=$1
    STATUS=$2
    RESULT=$3
    SESSION_ID=$(get_session)

    curl -s -X POST "$GATEWAY/api/teams/$TEAM_ID/agents/$AGENT_ID/tasks/$TASK_ID/result" \
        -H "Content-Type: application/json" \
        -H "X-Session-ID: $SESSION_ID" \
        -d "{\"task_id\":\"$TASK_ID\",\"status\":\"$STATUS\",\"result\":$RESULT}" 2>/dev/null
}

# Execute a task
execute_task() {
    TASK_ID=$1
    TASK_TYPE=$2
    PAYLOAD=$3

    echo "Executing task $TASK_ID (type: $TASK_TYPE)"

    case "$TASK_TYPE" in
        "shell")
            # Execute shell command
            CMD=$(echo "$PAYLOAD" | grep -o '"command":"[^"]*"' | cut -d'"' -f4)
            echo "Running: $CMD"
            OUTPUT=$(sh -c "$CMD" 2>&1) || true
            submit_result "$TASK_ID" "completed" "{\"output\":\"$OUTPUT\"}"
            ;;
        "echo")
            # Simple echo task
            MESSAGE=$(echo "$PAYLOAD" | grep -o '"message":"[^"]*"' | cut -d'"' -f4)
            echo "Echo: $MESSAGE"
            submit_result "$TASK_ID" "completed" "{\"echo\":\"$MESSAGE\"}"
            ;;
        *)
            echo "Unknown task type: $TASK_TYPE"
            submit_result "$TASK_ID" "failed" "{\"error\":\"Unknown task type: $TASK_TYPE\"}"
            ;;
    esac
}

# Main loop
echo "Starting task polling loop..."
while true; do
    SESSION_ID=$(get_session)
    if [ -z "$SESSION_ID" ]; then
        echo "$(date '+%H:%M:%S') - No session, waiting..."
        sleep 10
        continue
    fi

    # Poll for tasks
    TASKS=$(poll_tasks)
    
    if [ -n "$TASKS" ] && [ "$TASKS" != "[]" ] && [ "$TASKS" != "null" ]; then
        echo "$(date '+%H:%M:%S') - Received tasks: $TASKS"
        
        # Parse and execute each task (simple parsing)
        # In production, use jq for proper JSON parsing
        echo "$TASKS" | grep -o '"id":"[^"]*"' | while read id_line; do
            TASK_ID=$(echo "$id_line" | cut -d'"' -f4)
            if [ -n "$TASK_ID" ]; then
                # Get full task details
                TASK_DETAILS=$(curl -s "$GATEWAY/api/teams/$TEAM_ID/tasks/$TASK_ID" -H "X-Session-ID: $SESSION_ID" 2>/dev/null)
                TASK_TYPE=$(echo "$TASK_DETAILS" | grep -o '"type":"[^"]*"' | head -1 | cut -d'"' -f4)
                
                echo "Processing task: $TASK_ID (type: $TASK_TYPE)"
                execute_task "$TASK_ID" "$TASK_TYPE" "$TASK_DETAILS"
            fi
        done
    fi

    sleep 5
done
