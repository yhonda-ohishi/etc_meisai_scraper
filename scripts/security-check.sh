#!/bin/bash

# セキュリティチェックスクリプト
# ハードコードされた認証情報を検出

echo "========================================="
echo "セキュリティチェック: ハードコード認証情報検出"
echo "========================================="
echo ""

# 検出するパターン
PATTERNS=(
    "kikuraku"
    "password.*=.*\"[^\"]*\""
    "Password.*=.*\"[^\"]*\""
    "APIKey.*=.*\"[^\"]*\""
    "secret.*=.*\"[^\"]*\""
    "172\.18\.21\.35"
    "pbi"
)

# 除外するファイル/ディレクトリ
EXCLUDE_DIRS=(
    ".git"
    "node_modules"
    "vendor"
    ".env"
    "*.log"
)

# 検出結果カウンター
ISSUES_FOUND=0

# 各パターンを検索
for pattern in "${PATTERNS[@]}"; do
    echo "検索パターン: $pattern"
    echo "-------------------"

    # grepで検索（バイナリファイルと除外ディレクトリを無視）
    results=$(grep -r -n -i "$pattern" . \
        --exclude-dir=.git \
        --exclude-dir=node_modules \
        --exclude-dir=vendor \
        --exclude="*.log" \
        --exclude=".env" \
        --exclude="*.exe" \
        --exclude="*.dll" \
        --exclude="security-check.sh" \
        --exclude="constitution.md" \
        --exclude=".env.example" \
        2>/dev/null)

    if [ -n "$results" ]; then
        echo "$results"
        ISSUES_FOUND=$((ISSUES_FOUND + 1))
    else
        echo "✓ 問題なし"
    fi
    echo ""
done

echo "========================================="
if [ $ISSUES_FOUND -eq 0 ]; then
    echo "✅ セキュリティチェック: 合格"
    echo "ハードコードされた認証情報は見つかりませんでした。"
else
    echo "⚠️  セキュリティチェック: 警告"
    echo "$ISSUES_FOUND 個のパターンで潜在的な問題が見つかりました。"
    echo "上記のファイルを確認し、必要に応じて修正してください。"
    exit 1
fi
echo "========================================="