@echo off
echo [1/2] Building Frontend...
cd frontend
call npm run build
if %ERRORLEVEL% NEQ 0 (
    echo Frontend build failed!
    exit /b %ERRORLEVEL%
)
cd ..

echo [2/2] Compiling LaraGal Go Binary with Wails Production Tags...
go build -tags "desktop,production" -ldflags "-H windowsgui" -o laragal.exe .
if %ERRORLEVEL% NEQ 0 (
    echo Go build failed!
    exit /b %ERRORLEVEL%
)

echo LaraGal build completed successfully! Output: laragal.exe
