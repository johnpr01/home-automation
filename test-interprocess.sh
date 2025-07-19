#!/bin/bash

echo "🧪 Testing Inter-Process Asset Discovery..."

# Kill any existing discovery processes
pkill -f discovery || true
sleep 1

echo ""
echo "📡 Step 1: Starting continuous announcement in background..."
cd /home/philip/home-automation

# Run announcement for 60 seconds
timeout 60s ./bin/discovery -mode=announce -type=gateway -name="Test Gateway" -room="office" -duration=55s &
ANNOUNCE_PID=$!

echo "   ✅ Announcement process started (PID: $ANNOUNCE_PID)"

echo ""
echo "⏱️  Step 2: Waiting 5 seconds for announcement to start..."
sleep 5

echo ""
echo "🔍 Step 3: Running discovery in separate process..."
timeout 15s ./bin/discovery -mode=discover -duration=10s -verbose

if [ $? -eq 0 ]; then
    echo "   ✅ Discovery completed successfully"
else
    echo "   ⚠️  Discovery timed out or failed"
fi

echo ""
echo "❓ Step 4: Testing query mode..."
timeout 10s ./bin/discovery -mode=query -query-types="gateway" -duration=5s -verbose

if [ $? -eq 0 ]; then
    echo "   ✅ Query completed successfully"
else
    echo "   ⚠️  Query timed out or failed"
fi

echo ""
echo "🧹 Step 5: Cleanup..."
kill $ANNOUNCE_PID 2>/dev/null && echo "   ✅ Stopped announcement process" || echo "   ⚠️  Announcement process already stopped"
wait 2>/dev/null

echo ""
echo "✅ Inter-process test completed!"
