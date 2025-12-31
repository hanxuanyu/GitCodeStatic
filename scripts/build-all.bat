@echo off
chcp 65001 > nul
setlocal enabledelayedexpansion

REM GitCodeStatic Windows Multi-Platform Build Script
REM Cross-compile and package for all platforms using PowerShell

set "VERSION=%1"
set "OUTPUT_DIR=%2"

if "%VERSION%"=="" set "VERSION=latest"
if "%OUTPUT_DIR%"=="" set "OUTPUT_DIR=dist"

echo Starting GitCodeStatic v%VERSION% multi-platform build...

REM Setup variables
set "PROJECT_ROOT=%~dp0\.."
set "OUTPUT_PATH=%PROJECT_ROOT%\%OUTPUT_DIR%"

REM Clean output directory
if exist "%OUTPUT_PATH%" (
    echo Cleaning old output directory...
    rmdir /s /q "%OUTPUT_PATH%"
)
mkdir "%OUTPUT_PATH%"

REM Platform list
set PLATFORMS=windows/amd64 linux/amd64 darwin/amd64 linux/arm64 darwin/arm64

cd /d "%PROJECT_ROOT%"

for %%P in (%PLATFORMS%) do (
    echo.
    echo Building %%P...
    
    REM Parse platform string
    for /f "tokens=1,2 delims=/" %%A in ("%%P") do (
        set "GOOS=%%A"
        set "GOARCH=%%B"
    )
    
    REM Set output filename
    set "BINARY_NAME=gitcodestatic"
    if "!GOOS!"=="windows" set "BINARY_NAME=gitcodestatic.exe"
    
    REM Set package name
    set "PACKAGE_NAME=gitcodestatic-!GOOS!-!GOARCH!-!VERSION!"
    set "PACKAGE_PATH=!OUTPUT_PATH!\!PACKAGE_NAME!"
    
    mkdir "!PACKAGE_PATH!"
    
    REM Set build environment
    set "GOOS=!GOOS!"
    set "GOARCH=!GOARCH!"
    set "CGO_ENABLED=1"
    
    REM Disable CGO for ARM64 and cross-compilation
    if "!GOARCH!"=="arm64" set "CGO_ENABLED=0"
    if not "!GOOS!"=="windows" set "CGO_ENABLED=0"
    
    echo   Building binary...
    go build -ldflags "-s -w -X main.Version=!VERSION!" -o "!PACKAGE_PATH!\!BINARY_NAME!" cmd\server\main.go
    
    if !errorlevel! neq 0 (
        echo   Build failed for %%P!
        continue
    )
    
    echo   Copying files...
    
    REM Copy common files
    xcopy /E /I /Y web "!PACKAGE_PATH!\web\" > nul
    xcopy /E /I /Y configs "!PACKAGE_PATH!\configs\" > nul
    copy /Y README.md "!PACKAGE_PATH!\" > nul
    copy /Y QUICKSTART.md "!PACKAGE_PATH!\" > nul
    
    REM Create platform-specific startup scripts
    if "!GOOS!"=="windows" (
        REM Windows startup script
        (
            echo @echo off
            echo echo Starting GitCodeStatic Server...
            echo echo.
            echo echo Web UI: http://localhost:8080
            echo echo API Docs: http://localhost:8080/swagger/
            echo echo.
            echo gitcodestatic.exe
            echo pause
        ) > "!PACKAGE_PATH!\start.bat"
        
        REM Windows configuration guide
        (
            echo GitCodeStatic Windows Version
            echo.
            echo ## Usage
            echo 1. Double-click start.bat to start the server
            echo 2. Open browser and visit http://localhost:8080
            echo 3. View API docs: http://localhost:8080/swagger/
            echo.
            echo ## Configuration
            echo - configs/config.yaml: Main configuration file
            echo.
            echo ## Stop Service
            echo - Press Ctrl+C in the command window to stop the server
        ) > "!PACKAGE_PATH!\README-Windows.txt"
        
    ) else (
        REM Unix startup script
        (
            echo #!/bin/bash
            echo echo "Starting GitCodeStatic Server..."
            echo echo ""
            echo echo "Web UI: http://localhost:8080"
            echo echo "API Docs: http://localhost:8080/swagger/"
            echo echo ""
            echo echo "Press Ctrl+C to stop the server"
            echo echo ""
            echo.
            echo chmod +x "./gitcodestatic"
            echo ./gitcodestatic
        ) > "!PACKAGE_PATH!\start.sh"
        
        REM Unix documentation file
        (
            echo # GitCodeStatic !GOOS! Version
            echo.
            echo ## Quick Start
            echo ```bash
            echo ./start.sh
            echo # or
            echo ./gitcodestatic
            echo ```
            echo.
            echo ## Configuration
            echo - configs/config.yaml: Main configuration file
            echo.
            echo ## Access URLs
            echo - Web UI: http://localhost:8080
            echo - API Docs: http://localhost:8080/swagger/
            echo.
            echo ## Stop Service
            echo Press Ctrl+C to stop the server
        ) > "!PACKAGE_PATH!\README_!GOOS!.md"
    )
    
    echo   Creating archive...
    cd /d "!OUTPUT_PATH!"
    
    if "!GOOS!"=="windows" (
        REM Windows uses built-in compression
        powershell -command "Compress-Archive -Path '!PACKAGE_NAME!' -DestinationPath '!PACKAGE_NAME!.zip' -Force" 2>nul
        if !errorlevel! equ 0 (
            echo   [OK] Created !PACKAGE_NAME!.zip
        ) else (
            echo   [WARN] Failed to create archive
        )
    ) else (
        REM Unix platform archive (if tar command is available)
        where tar >nul 2>nul
        if !errorlevel! equ 0 (
            tar -czf "!PACKAGE_NAME!.tar.gz" "!PACKAGE_NAME!"
            if !errorlevel! equ 0 (
                echo   [OK] Created !PACKAGE_NAME!.tar.gz
            ) else (
                echo   [WARN] Failed to create tar.gz
            )
        ) else (
            powershell -command "Compress-Archive -Path '!PACKAGE_NAME!' -DestinationPath '!PACKAGE_NAME!.zip' -Force" 2>nul
            if !errorlevel! equ 0 (
                echo   [OK] Created !PACKAGE_NAME!.zip
            ) else (
                echo   [WARN] Failed to create archive
            )
        )
    )
    
    cd /d "%PROJECT_ROOT%"
)

echo.
echo ===================================
echo Multi-Platform Build Complete!
echo ===================================
echo Output directory: %OUTPUT_PATH%
echo.
echo Built platforms:

cd /d "%OUTPUT_PATH%"

REM Display build results
for %%P in (%PLATFORMS%) do (
    for /f "tokens=1,2 delims=/" %%A in ("%%P") do (
        set "GOOS=%%A"
        set "GOARCH=%%B"
        set "PACKAGE_NAME=gitcodestatic-!GOOS!-!GOARCH!-!VERSION!"
        
        if "!GOOS!"=="windows" (
            set "ARCHIVE=!PACKAGE_NAME!.zip"
        ) else (
            set "ARCHIVE=!PACKAGE_NAME!.tar.gz"
            if not exist "!ARCHIVE!" set "ARCHIVE=!PACKAGE_NAME!.zip"
        )
        
        if exist "!ARCHIVE!" (
            for %%F in ("!ARCHIVE!") do echo   [OK] !GOOS!/!GOARCH! - !ARCHIVE! ^(%%~zF bytes^)
        ) else (
            echo   [FAIL] !GOOS!/!GOARCH! - Build failed
        )
    )
)

echo.

REM File statistics
set /a COUNT=0
for %%F in (*.zip *.tar.gz) do set /a COUNT+=1
echo Total files: !COUNT!

pause