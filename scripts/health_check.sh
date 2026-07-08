#!/usr/bin/env bash

URL="$1"
BOT_TOKEN="$2"
CHAT_ID="$3"
STATE_FILE="/tmp/tivri_health_state"

if [ -z "$URL" ] || [ -z "$BOT_TOKEN" ] || [ -z "$CHAT_ID" ]; then
    echo "Usage: $0 <url> <bot_token> <chat_id>"
    exit 1
fi

send_telegram() {
    local text="$1"
    curl -s -X POST "https://api.telegram.org/bot${BOT_TOKEN}/sendMessage" \
        -d "chat_id=${CHAT_ID}" \
        -d "text=${text}" \
        -d "parse_mode=Markdown" > /dev/null
}

RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" --max-time 10 "${URL}/healthz")

if [ "$RESPONSE" -eq 200 ]; then
    CURRENT_STATE="up"
else
    CURRENT_STATE="down"
fi

if [ -f "$STATE_FILE" ]; then
    LAST_STATE=$(cat "$STATE_FILE")
else
    LAST_STATE="up"
fi

if [ "$CURRENT_STATE" != "$LAST_STATE" ]; then
    if [ "$CURRENT_STATE" = "down" ]; then
        send_telegram "🔴 *System Down Alert*

The system health check at ${URL}/healthz failed with HTTP status \`${RESPONSE}\`."
    else
        send_telegram "🟢 *System Recovery Alert*

The system health check at ${URL}/healthz has recovered and is returning HTTP status 200."
    fi
    echo "$CURRENT_STATE" > "$STATE_FILE"
fi
