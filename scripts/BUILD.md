# GitCodeStatic 打包脚本说明文档

## 打包脚本概述

本项目提供了多个构建脚本，支持在不同平台下构建和打包 GitCodeStatic 项目：

- `build-all.sh` - Unix/Linux/macOS 下的全平台构建脚本
- `build-all.bat` - Windows 下的全平台构建脚本  
- `build.ps1` - Windows PowerShell 构建脚本（单平台）
- `build-linux.sh` - Linux 专用构建脚本
- `build-macos.sh` - macOS 专用构建脚本

## 使用方法

### 快速构建全平台包

**Linux/macOS:**
```bash
cd scripts
chmod +x build-all.sh
./build-all.sh [版本号] [输出目录]
```

**Windows:**
```batch
cd scripts
build-all.bat [版本号] [输出目录]
```

### 参数说明

- `版本号`: 可选，默认为 "latest"
- `输出目录`: 可选，默认为 "dist"

### 示例

```bash
# 使用默认参数构建
./build-all.sh

# 指定版本号
./build-all.sh v1.2.3

# 指定版本号和输出目录
./build-all.sh v1.2.3 releases
```

## 支持的平台

脚本会为以下平台构建二进制文件和安装包：

| 平台 | 架构 | 二进制文件 | 压缩包格式 |
|------|------|-----------|-----------|
| Windows | amd64 | gitcodestatic.exe | .zip |
| Linux | amd64 | gitcodestatic | .tar.gz |
| Linux | arm64 | gitcodestatic | .tar.gz |
| macOS | amd64 | gitcodestatic | .tar.gz |
| macOS | arm64 | gitcodestatic | .tar.gz |

## 输出结构

构建完成后，每个平台的包都包含以下文件：

```
gitcodestatic-{平台}-{架构}-{版本}/
├── gitcodestatic[.exe]      # 主程序
├── web/                     # Web 前端文件
│   ├── index.html
│   └── static/
├── configs/                 # 配置文件
│   └── config.yaml
├── README.md               # 项目说明
├── QUICKSTART.md           # 快速开始指南
├── start.[sh|bat]          # 启动脚本
└── [使用说明.txt|README_{平台}.md]  # 平台特定说明
```

## 启动脚本

每个包都包含平台特定的启动脚本：

**Windows (`start.bat`):**
```batch
@echo off
echo Starting GitCodeStatic Server...
echo Web UI: http://localhost:8080
gitcodestatic.exe
pause
```

**Unix/Linux (`start.sh`):**
```bash
#!/bin/bash
echo "Starting GitCodeStatic Server..."
echo "Web UI: http://localhost:8080"
chmod +x "./gitcodestatic"
./gitcodestatic
```

## 特殊说明

### CGO 处理

- **Windows amd64**: 启用 CGO (CGO_ENABLED=1) 用于 SQLite 支持
- **其他平台**: 禁用 CGO (CGO_ENABLED=0) 以简化交叉编译

### 压缩格式

- **Windows**: 使用 ZIP 格式压缩
- **Unix/Linux/macOS**: 使用 tar.gz 格式压缩

### 兼容性

- 构建脚本自动检测可用的压缩工具
- 如果系统缺少特定工具，会尝试使用替代方案

## 前置要求

### 必需：
- Go 1.21+ 
- Git (用于版本信息)

### 可选（用于压缩）：
- **Linux/macOS**: tar, gzip
- **Windows**: PowerShell (内置压缩) 或 zip 命令

## 故障排除

### 常见问题

1. **权限错误**
   ```bash
   chmod +x scripts/*.sh
   ```

2. **Go 模块错误**
   ```bash
   go mod tidy
   go mod download
   ```

3. **交叉编译失败**
   - 确保 Go 版本 >= 1.21
   - 检查网络连接（可能需要下载工具链）

4. **压缩失败**
   - Windows: 确保 PowerShell 可用
   - Unix: 安装 tar 和 gzip

### 调试模式

设置环境变量启用详细输出：
```bash
export GOOS_DEBUG=1
./build-all.sh
```

## 自定义构建

如需自定义构建过程，可以修改脚本中的以下部分：

1. **平台列表**: 编辑 `PLATFORMS` 数组
2. **构建标志**: 修改 `go build -ldflags` 参数  
3. **包含文件**: 调整文件复制逻辑
4. **压缩设置**: 更改压缩格式或参数

## 部署建议

构建完成后的包可以直接部署到目标服务器：

1. **解压包文件**到目标目录
2. **运行启动脚本**或直接执行二进制文件
3. **访问 Web 界面**: http://localhost:8080
4. **查看 API 文档**: http://localhost:8080/swagger/

## 版本管理

脚本会在构建时自动注入版本信息：

```bash
go build -ldflags "-X main.Version=v1.2.3"
```

版本号可通过程序参数或环境变量查看：

```bash
./gitcodestatic --version
```