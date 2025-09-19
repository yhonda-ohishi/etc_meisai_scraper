@echo off
cd C:\go\etc_meisai
echo Starting Browser with ETC Site...
echo.
echo If browser doesn't open, check for errors below:
echo ========================================
go run cmd/interactive_scraper/main.go
pause