#!/bin/bash
# Test with coverage helper script

# デフォルトのテストディレクトリ
TEST_DIR="${1:-./tests/...}"

echo "📊 Running tests with coverage analysis..."
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# テストを実行してカバレッジを表示
go test -v -cover "$TEST_DIR" 2>&1 | while IFS= read -r line; do
    # カバレッジ情報を強調
    if echo "$line" | grep -q "coverage:"; then
        coverage=$(echo "$line" | grep -oE "[0-9]+\.[0-9]+%" | head -1)
        pkg=$(echo "$line" | awk '{print $1}')

        # カバレッジ率を数値として取得
        coverage_num=$(echo "$coverage" | sed 's/%//')

        # カバレッジレベルに応じて表示を変更
        if (( $(echo "$coverage_num >= 80" | bc -l) )); then
            echo "✅ $pkg: $coverage"
        elif (( $(echo "$coverage_num >= 60" | bc -l) )); then
            echo "⚠️  $pkg: $coverage (needs improvement)"
        else
            echo "❌ $pkg: $coverage (low coverage!)"
        fi
    elif echo "$line" | grep -q "PASS"; then
        echo "✅ $line"
    elif echo "$line" | grep -q "FAIL"; then
        echo "❌ $line"
    else
        echo "$line"
    fi
done

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📊 Coverage analysis complete!"