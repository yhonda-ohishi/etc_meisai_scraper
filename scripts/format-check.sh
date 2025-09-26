#!/bin/bash
# Format and coverage check script

file=$1
if [[ -z "$file" ]]; then
    echo "Usage: $0 <file.go>"
    exit 1
fi

if [[ "$file" == *.go ]] && [[ "$file" != */pb/* ]] && [[ "$file" != */mocks/* ]]; then
    echo "✅ Checking Go file format..."
    if ! gofmt -d "$file" 2>&1 | grep -q '^'; then
        echo "✔️ Format OK"
    else
        echo "⚠️ FORMAT ERROR DETECTED:"
        gofmt -d "$file"
    fi

    echo "📊 Running coverage check..."
    pkg=$(dirname "$file" | sed 's|^src/||')
    go test -cover ./tests/unit/$pkg 2>/dev/null | grep -E 'coverage:|ok|PASS' | head -5 || echo "No related tests found"
fi