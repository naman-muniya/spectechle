@echo off
echo Starting Spectechle...
echo.

echo Starting Python NLP service...
start "NLP Service" cmd /k "cd nlp && venv\Scripts\activate && python app.py"

echo Starting Go backend...
start "Go Backend" cmd /k "cd backend && go run main.go"

echo.
echo Services started! 
echo Web UI: http://localhost:8080
echo Go API: http://localhost:8080/api
echo Python NLP API: http://localhost:5000
echo.
pause
