@echo off
setlocal

set "DARDCOR_DIR=%~dp0"
set "PORT=25000"
if defined DARDCOR_PORT set "PORT=%DARDCOR_PORT%"

if "%~1"=="run"     goto run_mode
if "%~1"=="cli"     goto cli_mode

echo Dardcor Agent
echo ------------------
echo dardcor run      - Gateway + WebUI Dashboard
echo dardcor cli      - CLI Agent (headless TUI)
exit /b 1

:run_mode
echo [*] Stopping old instances...
call :kill_old
cd /d "%DARDCOR_DIR%"
if not exist "dist\index.html" (
    echo [*] UI build missing. Running build...
    call npm run build
)
echo [*] Starting Gateway...
call :start_backend "backend.log"
timeout /t 3 /nobreak >nul
echo [OK] Gateway running on http://127.0.0.1:%PORT%
start "" "http://127.0.0.1:%PORT%"
goto keepalive

:cli_mode
echo [*] Stopping old instances...
call :kill_old
cd /d "%DARDCOR_DIR%"
echo [*] Starting CLI Agent...
call :start_backend "backend_cli.log"
timeout /t 2 /nobreak >nul
echo [OK] CLI Agent ready.
goto keepalive

:start_backend
cd /d "%DARDCOR_DIR%"
if exist "dardcor-agent.exe" (
    start /B "" "dardcor-agent.exe" >"%~1" 2>&1
) else (
    start /B "Dardcor-Backend" go run main.go >"%~1" 2>&1
)
exit /b 0

:kill_old
taskkill /F /IM "dardcor-agent.exe" /T >nul 2>&1
for /f "tokens=5" %%p in ('netstat -ano 2^>nul ^| findstr /C:":%PORT% " ^| findstr /C:"LISTENING"') do (
    taskkill /PID %%p /T /F >nul 2>&1
)
exit /b 0

:keepalive
timeout /t 10 /nobreak >nul
netstat -ano 2>nul | findstr /C:":%PORT% " | findstr /C:"LISTENING" >nul 2>&1
if errorlevel 1 (
    echo [!] Server down. Restarting...
    call :start_backend "backend.log"
)
goto keepalive
