# GitCodeStatic Windows 打包脚本
# PowerShell 脚本用于在 Windows 平台构建和打包

param(
    [string]$Version = "latest",
    [string]$OutputDir = "dist"
)

Write-Host "开始构建 GitCodeStatic for Windows..." -ForegroundColor Green

# 设置变量
$ProjectRoot = Split-Path -Parent $PSScriptRoot
$BinaryName = "gitcodestatic.exe"
$PackageName = "gitcodestatic-windows-$Version"
$OutputPath = Join-Path $ProjectRoot $OutputDir
$PackagePath = Join-Path $OutputPath $PackageName

# 清理输出目录
if (Test-Path $PackagePath) {
    Write-Host "清理旧的输出目录..." -ForegroundColor Yellow
    Remove-Item -Recurse -Force $PackagePath
}

# 创建输出目录
New-Item -ItemType Directory -Path $PackagePath -Force | Out-Null

# 设置构建环境
$env:GOOS = "windows"
$env:GOARCH = "amd64"
$env:CGO_ENABLED = "1"

Write-Host "构建 Go 二进制文件..." -ForegroundColor Blue
Set-Location $ProjectRoot

# 构建二进制文件
$BuildCmd = "go build -ldflags `"-s -w -X main.Version=$Version`" -o `"$PackagePath\$BinaryName`" cmd/server/main.go"
Invoke-Expression $BuildCmd

if ($LASTEXITCODE -ne 0) {
    Write-Error "构建失败！"
    exit 1
}

Write-Host "复制必需的文件..." -ForegroundColor Blue

# 复制 web 静态文件
Copy-Item -Recurse -Path "web" -Destination "$PackagePath\web"

# 复制配置文件
Copy-Item -Recurse -Path "configs" -Destination "$PackagePath\configs"

# 复制文档文件
Copy-Item -Path "README.md" -Destination $PackagePath
Copy-Item -Path "QUICKSTART.md" -Destination $PackagePath

# 创建启动脚本
$StartScript = @"
@echo off
echo Starting GitCodeStatic Server...
echo.
echo Web UI: http://localhost:8080
echo API Docs: http://localhost:8080/swagger/
echo.
gitcodestatic.exe
pause
"@

$StartScript | Out-File -FilePath "$PackagePath\start.bat" -Encoding ASCII

# 创建配置说明
$ConfigInfo = @"
GitCodeStatic Windows 版本

## 使用方法

1. 双击 start.bat 启动服务器
2. 打开浏览器访问 http://localhost:8080
3. 查看 API 文档: http://localhost:8080/swagger/

## 配置文件

- configs/config.yaml: 主配置文件
- 可以修改端口、数据库路径等配置

## 日志和数据

- 日志文件: logs/app.log
- 数据库文件: workspace/gitcodestatic.db
- 仓库缓存: workspace/repos/

## 停止服务

- 在命令窗口中按 Ctrl+C 停止服务器
"@

$ConfigInfo | Out-File -FilePath "$PackagePath\使用说明.txt" -Encoding UTF8

# 创建 ZIP 包
Write-Host "创建压缩包..." -ForegroundColor Blue
$ZipPath = "$OutputPath\$PackageName.zip"
if (Test-Path $ZipPath) {
    Remove-Item $ZipPath
}

# 使用 PowerShell 5.0+ 的压缩功能
Compress-Archive -Path "$PackagePath\*" -DestinationPath $ZipPath

Write-Host "构建完成！" -ForegroundColor Green
Write-Host "输出路径: $ZipPath" -ForegroundColor Green

# 显示文件大小
$ZipSize = (Get-Item $ZipPath).Length
Write-Host "压缩包大小: $([math]::Round($ZipSize/1MB, 2)) MB" -ForegroundColor Green