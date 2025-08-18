@echo off
echo Starting Spectechle with Docker...
echo.

echo Current directory: %CD%
echo Checking for docker-compose.yml...
if not exist "docker-compose.yml" (
    echo ERROR: docker-compose.yml not found in current directory
    echo Please run this script from the project root directory
    pause
    exit /b 1
)

echo Building and starting services...
docker-compose up --build

echo.
echo Services started with Docker!
echo Web UI: http://localhost:8080
echo Go API: http://localhost:8080/api
echo Python NLP API: http://localhost:5000
echo.
echo Press Ctrl+C to stop services
pause
