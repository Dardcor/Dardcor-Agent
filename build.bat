@echo off
setlocal enabledelayedexpansion

echo ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
echo 🚀 Building Dardcor Agent Executable
echo ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

:: Step 1: Build Frontend
echo [1/3] Building frontend assets...
call npm install
call npm run build
if %ERRORLEVEL% neq 0 (
    echo ❌ Frontend build failed!
    exit /b %ERRORLEVEL%
)

:: Step 2: Build Backend (Go)
echo [2/3] Compiling Go binary...
:: Using -ldflags to reduce size and set version if needed
go build -ldflags="-s -w" -o dardcor.exe main.go
if %ERRORLEVEL% neq 0 (
    echo ❌ Go build failed!
    exit /b %ERRORLEVEL%
)

:: Step 3: Deployment Info
echo [3/3] Build complete: dardcor.exe
echo.
echo ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
echo ✅ Success!
echo.
echo To install globally:
echo 1. Create directory: mkdir %%USERPROFILE%%\.local\bin
echo 2. Move dardcor.exe to that directory
echo 3. Add %%USERPROFILE%%\.local\bin to your PATH environment variable
echo.
echo After installation, you can run 'dardcor' from any terminal.
echo ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

pause
