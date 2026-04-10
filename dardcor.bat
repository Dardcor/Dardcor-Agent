@echo off
setlocal
set "DARDCOR_DIR=%~dp0"

if "%~1"=="" goto help
if "%~1"=="help" goto help

cd /d "%DARDCOR_DIR%"
node "%DARDCOR_DIR%cli.js" %*
exit /b %ERRORLEVEL%

:help
node "%DARDCOR_DIR%cli.js" help
exit /b 0




