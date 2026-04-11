#!/bin/bash

PORT=8080

echo "🔪 Killing anything on port $PORT..."
PID=$(lsof -ti :$PORT)

if [ -n "$PID" ]; then
  kill -9 $PID
  echo "✅ Killed process $PID"
else
  echo "ℹ️ Nothing running on port $PORT"
fi

echo "🚀 Starting server..."
go run .