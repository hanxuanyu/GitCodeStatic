@echo off
setlocal enabledelayedexpansion

REM GitCodeStatic Windows 全平台构建脚本
REM 使用 PowerShell 的交叉编译功能构建所有平台的包

set "VERSION=%1"
set "OUTPUT_DIR=%2"

if "%VERSION%"=="" set "VERSION=latest"
if "%OUTPUT_DIR%"=="" set "OUTPUT_DIR=dist"

echo 开始全平台构建 GitCodeStatic v%VERSION%...

REM 设置变量
set "PROJECT_ROOT=%~dp0\.."
set "OUTPUT_PATH=%PROJECT_ROOT%\%OUTPUT_DIR%"

REM 清理输出目录
if exist "%OUTPUT_PATH%" (
    echo 清理旧的输出目录...
    rmdir /s /q "%OUTPUT_PATH%"
)
mkdir "%OUTPUT_PATH%"

REM 平台列表
set PLATFORMS=windows/amd64 linux/amd64 darwin/amd64 linux/arm64 darwin/arm64

cd /d "%PROJECT_ROOT%"

for %%P in (%PLATFORMS%) do (
    echo.
    echo 构建 %%P...
    
    REM 解析平台字符串
    for /f "tokens=1,2 delims=/" %%A in ("%%P") do (
        set "GOOS=%%A"
        set "GOARCH=%%B"
    )
    
    REM 设置输出文件名
    set "BINARY_NAME=gitcodestatic"
    if "!GOOS!"=="windows" set "BINARY_NAME=gitcodestatic.exe"
    
    REM 设置包名
    set "PACKAGE_NAME=gitcodestatic-!GOOS!-!GOARCH!-!VERSION!"
    set "PACKAGE_PATH=!OUTPUT_PATH!\!PACKAGE_NAME!"
    
    mkdir "!PACKAGE_PATH!"
    
    REM 设置构建环境
    set "GOOS=!GOOS!"
    set "GOARCH=!GOARCH!"
    set "CGO_ENABLED=1"
    
    REM ARM64 和交叉编译时禁用 CGO
    if "!GOARCH!"=="arm64" set "CGO_ENABLED=0"
    if not "!GOOS!"=="windows" set "CGO_ENABLED=0"
    
    echo   构建二进制文件...
    go build -ldflags "-s -w -X main.Version=!VERSION!" -o "!PACKAGE_PATH!\!BINARY_NAME!" cmd\server\main.go
    
    if !errorlevel! neq 0 (
        echo   构建 %%P 失败！
        continue
    )
    
    echo   复制文件...
    
    REM 复制通用文件
    xcopy /E /I /Y web "!PACKAGE_PATH!\web\" > nul
    xcopy /E /I /Y configs "!PACKAGE_PATH!\configs\" > nul
    copy /Y README.md "!PACKAGE_PATH!\" > nul
    copy /Y QUICKSTART.md "!PACKAGE_PATH!\" > nul
    
    REM 根据平台创建特定的启动脚本
    if "!GOOS!"=="windows" (
        REM Windows 启动脚本
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
        
        REM Windows 配置说明
        (
            echo GitCodeStatic Windows 版本
            echo.
            echo ## 使用方法
            echo 1. 双击 start.bat 启动服务器
            echo 2. 打开浏览器访问 http://localhost:8080
            echo 3. 查看 API 文档: http://localhost:8080/swagger/
            echo.
            echo ## 配置文件
            echo - configs/config.yaml: 主配置文件
            echo.
            echo ## 停止服务
            echo - 在命令窗口中按 Ctrl+C 停止服务器
        ) > "!PACKAGE_PATH!\使用说明.txt"
        
    ) else (
        REM Unix 启动脚本
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
        
        REM Unix 说明文件
        (
            echo # GitCodeStatic !GOOS! 版本
            echo.
            echo ## 快速启动
            echo ```bash
            echo ./start.sh
            echo # 或
            echo ./gitcodestatic
            echo ```
            echo.
            echo ## 配置文件
            echo - configs/config.yaml: 主配置文件
            echo.
            echo ## 访问地址
            echo - Web UI: http://localhost:8080
            echo - API 文档: http://localhost:8080/swagger/
            echo.
            echo ## 停止服务
            echo 按 Ctrl+C 停止服务器
        ) > "!PACKAGE_PATH!\README_!GOOS!.md"
    )
    
    echo   创建压缩包...
    cd /d "!OUTPUT_PATH!"
    
    if "!GOOS!"=="windows" (
        REM Windows 使用内置压缩
        powershell -command "Compress-Archive -Path '!PACKAGE_NAME!' -DestinationPath '!PACKAGE_NAME!.zip' -Force" 2>nul
        if !errorlevel! equ 0 (
            echo   ✓ 已创建 !PACKAGE_NAME!.zip
        ) else (
            echo   ⚠ 创建压缩包失败
        )
    ) else (
        REM Unix 平台压缩包（如果有 tar 命令）
        where tar >nul 2>nul
        if !errorlevel! equ 0 (
            tar -czf "!PACKAGE_NAME!.tar.gz" "!PACKAGE_NAME!"
            if !errorlevel! equ 0 (
                echo   ✓ 已创建 !PACKAGE_NAME!.tar.gz
            ) else (
                echo   ⚠ 创建 tar.gz 失败
            )
        ) else (
            powershell -command "Compress-Archive -Path '!PACKAGE_NAME!' -DestinationPath '!PACKAGE_NAME!.zip' -Force" 2>nul
            if !errorlevel! equ 0 (
                echo   ✓ 已创建 !PACKAGE_NAME!.zip
            ) else (
                echo   ⚠ 创建压缩包失败
            )
        )
    )
    
    cd /d "%PROJECT_ROOT%"
)

echo.
echo ===================================
echo 全平台构建完成！
echo ===================================
echo 输出目录: %OUTPUT_PATH%
echo.
echo 构建的平台：

cd /d "%OUTPUT_PATH%"

REM 显示构建结果
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
            for %%F in ("!ARCHIVE!") do echo   ✓ !GOOS!/!GOARCH! - !ARCHIVE! ^(%%~zF 字节^)
        ) else (
            echo   ✗ !GOOS!/!GOARCH! - 构建失败
        )
    )
)

echo.

REM 统计文件
set /a COUNT=0
for %%F in (*.zip *.tar.gz) do set /a COUNT+=1
echo 总文件数: !COUNT!

pause