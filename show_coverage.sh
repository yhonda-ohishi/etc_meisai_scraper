#!/bin/bash

echo "📊 テストカバレッジレポート (Generated Codeを除く)"
echo "================================================"
echo ""

# Run tests and generate coverage (suppress test output)
go test -coverprofile=coverage.out -coverpkg=./src/... ./tests/... 2>/dev/null | grep -v "^?" | grep -v "^ok"

# Process coverage data with awk
go tool cover -func=coverage.out | grep -v "\.pb\.go" | grep -v "\.pb\.gw\.go" | grep -v "_grpc\.pb\.go" | awk '
function color_code(pct) {
    val = substr(pct, 1, length(pct)-1) + 0
    if (pct == "100.0%") return "\033[32m✅\033[0m"
    if (val >= 80) return "\033[33m🔶\033[0m"
    if (val >= 50) return "\033[34m🔷\033[0m"
    if (val > 0) return "\033[31m⚠️\033[0m"
    return "\033[90m⏹️\033[0m"
}

function color_pct(pct) {
    val = substr(pct, 1, length(pct)-1) + 0
    if (pct == "100.0%") return "\033[32m" pct "\033[0m"
    if (val >= 80) return "\033[33m" pct "\033[0m"
    if (val >= 50) return "\033[34m" pct "\033[0m"
    if (val > 0) return "\033[31m" pct "\033[0m"
    return "\033[90m" pct "\033[0m"
}

/^total:/ {
    # Skip the wrong total that includes PB files
    next
}

/.go:[0-9]+:/ && $NF ~ /%$/ {
    split($1, parts, "/")
    filename = parts[length(parts)]
    split(filename, fileparts, ":")
    file = fileparts[1]
    line = fileparts[2]
    func_name = $2
    pct = $NF

    if (pct == "100.0%") count_100++
    else if (pct == "0.0%") count_zero++
    else count_partial++

    printf "%s %-50s %s\n", color_code(pct), file ":" func_name, color_pct(pct)
}

END {
    total = count_100 + count_partial + count_zero
    if (total > 0) {
        real_coverage = (count_100 * 100.0) / total
    }

    print "\n============================================"
    if (count_100) printf "✅ 100%%カバレッジ: %d 関数\n", count_100
    if (count_partial) printf "📊 部分カバレッジ: %d 関数\n", count_partial
    if (count_zero) printf "⏹️  未テスト: %d 関数\n", count_zero

    print "\n============================================"
    printf "\033[1m📊 総合カバレッジ (PB除外): \033[32m%.1f%%\033[0m\033[1m\033[0m\n", real_coverage
    print "============================================"
    print "\n凡例: ✅ 100% | 🔶 80-99% | 🔷 50-79% | ⚠️ <50% | ⏹️ 0%"
}'