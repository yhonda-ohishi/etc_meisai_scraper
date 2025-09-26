#!/bin/bash
# File watcher that validates test file placement on every save
# Runs continuously and checks for test files in src/ directory

echo "Starting file watcher for test file validation..."
echo "Press Ctrl+C to stop"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Use fswatch, inotifywait, or entr depending on what's available
if command -v fswatch &> /dev/null; then
    echo "Using fswatch..."
    fswatch -r src/ | while read file; do
        if [[ "$file" == *_test.go ]] || [[ "$file" == *mock*.go ]]; then
            echo -e "${RED}❌ ERROR: Test/mock file detected in src/: $file${NC}"
            echo -e "${YELLOW}Move to tests/ directory!${NC}"
            # Optionally play a sound or send notification
            # echo -e "\a" # Terminal bell
        fi
    done
elif command -v inotifywait &> /dev/null; then
    echo "Using inotifywait..."
    while true; do
        inotifywait -r -e create,modify,moved_to src/ 2>/dev/null | while read path action file; do
            if [[ "$file" == *_test.go ]] || [[ "$file" == *mock*.go ]]; then
                echo -e "${RED}❌ ERROR: Test/mock file detected in src/: ${path}${file}${NC}"
                echo -e "${YELLOW}Move to tests/ directory!${NC}"
            fi
        done
    done
else
    echo "No file watcher available. Install fswatch or inotify-tools."
    echo "Alternative: Use editor plugin (see .vscode/settings.json)"
    exit 1
fi