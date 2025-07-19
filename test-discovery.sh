#!/bin/bash

echo "🧪 Testing Asset Discovery Protocol..."

# Kill any existing discovery processes
pkill -f discovery || true
sleep 1

echo ""
echo "📡 Step 1: Starting asset announcement..."
cd /home/philip/home-automation
./bin/discovery -mode=announce -type=gateway -name="Test Gateway" -room="office" -duration=20s &
ANNOUNCE_PID=$!

echo "   Announcement process started (PID: $ANNOUNCE_PID)"

echo ""
echo "⏱️  Step 2: Waiting 3 seconds for announcement to start..."
sleep 3

echo ""
echo "🔍 Step 3: Starting discovery to find announced asset..."
./bin/discovery -mode=discover -duration=10s -verbose

echo ""
echo "📊 Step 4: Testing JSON output..."
./bin/discovery -mode=discover -duration=5s -json

echo ""
echo "❓ Step 5: Testing query mode..."
./bin/discovery -mode=query -query-types="gateway" -duration=5s -verbose

echo ""
echo "🧹 Cleanup: Stopping announcement process..."
kill $ANNOUNCE_PID 2>/dev/null || true

echo ""
echo "✅ Asset Discovery Protocol test completed!"
