@echo off
REM Set script directory as current working directory
cd /d "%~dp0"

REM Terminate existing Laragal application processes
taskkill /F /T /IM laragal.exe >nul 2>&1
taskkill /F /T /IM laragal_app.exe >nul 2>&1

REM Terminate all php-cgi processes (intended per user request to free bound ports)
taskkill /F /T /IM php-cgi.exe >nul 2>&1

REM Wait 1 second for socket cleanup
timeout /t 1 /nobreak >nul

REM Launch a clean instance of laragal
start "" "%~dp0laragal.exe"
