@echo off
set DARDCOR_DIR=%~dp0

if "%~1"=="run" goto run_agent
if "%~1"=="dev" goto dev_mode
if "%~1"=="build" goto build_agent

echo.
echo ==========================================
echo    DARDCOR AGENT COMMAND LINE
echo ==========================================
echo.
echo Perintah yang tersedia:
echo   dardcor run     - Menjalankan Agent (Versi Produksi)
echo   dardcor dev     - Menjalankan Mode Development (Real-time UI)
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
    echo [*] UI Build not found. Building React UI first...
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
start http://127.0.0.1:25000
echo [*] Meluncurkan Dashboard di browser default Anda...
echo.

:: Keep the batch script alive to prevent closing
goto keepalive

:dev_mode
echo.
echo ==========================================
echo    STARTING DARDCOR AGENT (DEV MODE)
echo ==========================================
echo.
echo [*] Checking and safely closing old instances...
taskkill /F /IM dardcor-agent.exe /T >nul 2>&1
FOR /F "tokens=5" %%P IN ('netstat -ano ^| findstr :25000') DO TaskKill.exe /PID %%P /T /F >nul 2>&1

cd /d "%DARDCOR_DIR%"

echo [*] Checking node_modules...
if not exist node_modules (
    echo [*] Installing frontend dependencies...
    call npm install
)

echo [*] Starting Frontend Dev Server (Vite - internal)...
start /B "" cmd /c "npx vite --port 25099 --host 127.0.0.1 --strictPort > frontend_dev.log 2>&1"

echo [*] Waiting for Frontend Dev Server...
set /a count=0
:wait_vite
set /a count+=1
if %count% geq 30 (
    echo [!] Frontend dev server failed to start.
    exit /b 1
)
timeout /t 1 >nul
netstat -ano | findstr :25099 >nul
if errorlevel 1 goto wait_vite

echo [*] Starting Core Agent on port 25000 (with UI proxy)...
set PORT=25000
set DARDCOR_DEV_URL=http://127.0.0.1:25099
if exist dardcor-agent.exe (
    start /B "" cmd /c "dardcor-agent.exe > backend_dev.log 2>&1"
) else (
    start /B "" cmd /c "go run main.go > backend_dev.log 2>&1"
)

echo [*] Waiting for Backend to be ready...
set /a count=0
:wait_dev
set /a count+=1
if %count% geq 30 (
    echo [!] Development mode failed to start.
    exit /b 1
)
timeout /t 1 >nul
netstat -ano | findstr :25000 >nul
if errorlevel 1 goto wait_dev

echo.
echo [OK] MODE DEVELOPMENT BERJALAN! (Real-time Aktif)
echo [!] Portal Utama: http://127.0.0.1:25000
echo [!] Semua akses melalui satu port: 25000
echo.
start http://127.0.0.1:25000
echo [*] Meluncurkan Dashboard di browser default Anda...
echo.

goto keepalive

:keepalive
timeout /t 10 >nul
goto keepalive

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
