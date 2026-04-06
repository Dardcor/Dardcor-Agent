@echo off
set DARDCOR_DIR=%~dp0

if "%~1"=="run" goto run_agent
if "%~1"=="build" goto build_agent

echo.
echo ==========================================
echo    DARDCOR AGENT COMMAND LINE
echo ==========================================
echo.
echo Perintah yang tersedia:
echo   dardcor run     - Menjalankan Agent (Backend + Frontend)
echo   dardcor build   - Melakukan kompilasi sistem siap produksi
echo.
exit /b 1

:run_agent
echo.
echo ==========================================
echo    STARTING DARDCOR AGENT
echo ==========================================
echo.
echo [*] Checking and safely closing old instances...
taskkill /F /IM dardcor-agent.exe /T >nul 2>&1
FOR /F "tokens=5" %%P IN ('netstat -ano ^| findstr :25000') DO TaskKill.exe /PID %%P /T /F >nul 2>&1

cd /d "%DARDCOR_DIR%"

if not exist dist\index.html (
    echo [*] UI Build not found. Building Reach UI first...
    call npm run build
)

echo [*] Starting Core Agent on port 25000...
set PORT=25000
if exist dardcor-agent.exe (
    start /B "" cmd /c "dardcor-agent.exe > backend.log 2>&1"
) else (
    start /B "" cmd /c "go run main.go > backend.log 2>&1"
)

echo [*] Waiting for Agent to be ready...
set /a count=0
:wait_backend
set /a count+=1
if %count% geq 20 (
    echo [!] Agent failed to start after 20 seconds.
    echo [!] Cek "backend.log" untuk detail error.
    exit /b 1
)
timeout /t 1 >nul
netstat -ano | findstr :25000 >nul
if errorlevel 1 goto wait_backend

echo.
echo [OK] DARDCOR AGENT SEDANG BERJALAN!
echo [!] Portal Utama: http://127.0.0.1:25000
echo [!] Tekan Ctrl+C di sini untuk mematikan sistem.
echo.

:: Keep the batch script alive to prevent closing
:keepalive
timeout /t 10 >nul
goto keepalive
exit /b 0

:build_agent
echo.
echo ==========================================
echo    BUILDING DARDCOR AGENT
echo ==========================================
echo.
cd /d "%DARDCOR_DIR%"
echo [*] Building React UI (TypeScript)...
call npm run build
echo [*] Building Go Core Agent Binary...
go build -o dardcor-agent.exe .
echo.
echo [OK] BUILD SELESAI! Anda sekarang memiliki "dardcor-agent.exe" terkompilasi penuh.
exit /b 0
