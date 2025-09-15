@echo off
echo Running ETC Scraper Test...

REM Create directories
mkdir downloads 2>nul
mkdir downloads\screenshots 2>nul

REM Check for .env file
if not exist .env (
    echo Creating .env file from .env.example
    copy .env.example .env
    echo Please edit .env file with your ETC credentials
    pause
    exit /b 1
)

REM Run the scraper test
echo Starting scraper with debug mode...
go run cmd/scraper/main.go -debug %*

echo.
echo Scraper test completed.
echo Check downloads\screenshots directory for results.
pause