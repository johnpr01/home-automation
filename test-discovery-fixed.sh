#!/bin/bash

echo "🧪 Testing Asset Discovery Protocol - Fixed Version..."

# Kill any existing discovery processes
pkill -f discovery || true
sleep 1

echo ""
echo "📡 Step 1: Starting asset announcement in background..."
cd /home/philip/home-automation
./bin/discovery -mode=announce -type=gateway -name="Test Gateway" -room="office" -duration=30s &
ANNOUNCE_PID=$!

echo "   ✅ Announcement process started (PID: $ANNOUNCE_PID)"

echo ""
echo "⏱️  Step 2: Waiting 3 seconds for announcement to stabilize..."
sleep 3

echo ""
echo "🔍 Step 3: Running discovery to find announced asset..."
timeout 10s ./bin/discovery -mode=discover -duration=8s -verbose

echo ""
echo "📊 Step 4: Testing JSON output format..."
timeout 8s ./bin/discovery -mode=discover -duration=5s -json

echo ""
echo "❓ Step 5: Testing query mode for gateways..."
timeout 8s ./bin/discovery -mode=query -query-types="gateway" -duration=5s -verbose

echo ""
echo "🧹 Step 6: Cleanup..."
kill $ANNOUNCE_PID 2>/dev/null && echo "   ✅ Stopped announcement process" || echo "   ⚠️  Announcement process already stopped"

echo ""
echo "✅ Asset Discovery Protocol test completed!"
