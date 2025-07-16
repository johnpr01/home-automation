#!/bin/bash

# Fix logger calls in thermostat service
file="/home/philip/home-automation/internal/services/thermostat_service.go"

# Replace simple Printf calls
sed -i 's/ts\.logger\.Printf(\([^,)]*\))/ts.logger.Info(\1)/g' "$file"

# Replace Printf calls with format and args - convert to Info with fmt.Sprintf
sed -i 's/ts\.logger\.Printf(\([^,]*\), \(.*\))/ts.logger.Info(fmt.Sprintf(\1, \2))/g' "$file"

# Replace Println calls
sed -i 's/ts\.logger\.Println(\(.*\))/ts.logger.Info(\1)/g' "$file"

echo "Logger calls updated in thermostat service"
