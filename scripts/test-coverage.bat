@echo off
REM Test with coverage for Windows

set TEST_DIR=%1
if "%TEST_DIR%"=="" set TEST_DIR=./tests/...

echo 📊 Running tests with coverage analysis...
echo ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

go test -v -cover %TEST_DIR%

echo ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
echo 📊 Coverage analysis complete!