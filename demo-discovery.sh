#!/bin/bash

echo "🏠 Home Automation Asset Discovery Demo"
echo "======================================"

cd /home/philip/home-automation

# Kill any existing discovery processes
pkill -f discovery || true
sleep 1

echo ""
echo "🚀 Starting Home Automation Discovery Demo..."

echo ""
echo "📡 Step 1: Starting various device announcements..."

# Start a gateway
echo "   🏠 Starting Main Gateway..."
timeout 120s ./bin/discovery -mode=announce -type=gateway -name="Main Gateway" -room="server-room" -duration=115s &
GATEWAY_PID=$!

# Start a smart plug
echo "   🔌 Starting Smart Plug..."
timeout 120s ./bin/discovery -mode=announce -type=smart_plug -name="Living Room Lamp" -room="living-room" -ip="192.168.1.101" -capabilities="power,energy_monitor,switch" -duration=115s &
PLUG_PID=$!

# Start a sensor
echo "   🌡️  Starting Environmental Sensor..."
timeout 120s ./bin/discovery -mode=announce -type=sensor -name="Living Room Sensor" -room="living-room" -capabilities="temperature,humidity,motion" -duration=115s &
SENSOR_PID=$!

echo "   ✅ All devices started"

echo ""
echo "⏱️  Step 2: Waiting 5 seconds for devices to initialize..."
sleep 5

echo ""
echo "🔍 Step 3: Discovering all devices on the network..."
./bin/discovery -mode=query -duration=10s -verbose

echo ""
echo "🏠 Step 4: Finding only gateway devices..."
./bin/discovery -mode=query -query-types="gateway" -duration=5s -verbose

echo ""
echo "🔌 Step 5: Finding smart plugs..."
./bin/discovery -mode=query -query-types="smart_plug" -duration=5s -verbose

echo ""
echo "🌡️  Step 6: Finding temperature sensors..."
./bin/discovery -mode=query -query-caps="temperature" -duration=5s -verbose

echo ""
echo "🏠 Step 7: Finding all devices in living room..."
./bin/discovery -mode=query -room="living-room" -duration=5s -verbose

echo ""
echo "📊 Step 8: Getting complete network inventory (JSON)..."
./bin/discovery -mode=query -duration=8s -json > network_inventory.json
echo "   📄 Network inventory saved to network_inventory.json"

if [ -f network_inventory.json ]; then
    asset_count=$(grep -o '"asset_id"' network_inventory.json | wc -l)
    echo "   📊 Total assets discovered: $asset_count"
fi

echo ""
echo "🧹 Step 9: Cleaning up..."
kill $GATEWAY_PID $PLUG_PID $SENSOR_PID 2>/dev/null
echo "   ✅ All announcement processes stopped"

echo ""
echo "🎉 Home Automation Asset Discovery Demo Complete!"
echo ""
echo "🔍 What we demonstrated:"
echo "   ✅ Multi-device asset announcement"
echo "   ✅ Comprehensive device discovery"
echo "   ✅ Filtering by device type"
echo "   ✅ Filtering by capabilities"
echo "   ✅ Filtering by room/location"
echo "   ✅ JSON output for automation"
echo ""
echo "📚 The asset discovery protocol successfully enables:"
echo "   🏠 Automatic home automation device discovery"
echo "   🔌 Smart plug and sensor identification"
echo "   📍 Location-based device organization"
echo "   ⚡ Real-time network topology mapping"
echo "   🔄 Dynamic device inventory management"
