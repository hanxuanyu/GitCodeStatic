#!/bin/bash

# GitCodeStatic macOS 打包脚本
# Bash 脚本用于在 macOS 平台构建和打包

set -e

VERSION=${1:-"latest"}
OUTPUT_DIR=${2:-"dist"}

echo "开始构建 GitCodeStatic for macOS..."

# 设置变量
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BINARY_NAME="gitcodestatic"
PACKAGE_NAME="gitcodestatic-macos-$VERSION"
OUTPUT_PATH="$PROJECT_ROOT/$OUTPUT_DIR"
PACKAGE_PATH="$OUTPUT_PATH/$PACKAGE_NAME"

# 清理输出目录
if [ -d "$PACKAGE_PATH" ]; then
    echo "清理旧的输出目录..."
    rm -rf "$PACKAGE_PATH"
fi

# 创建输出目录
mkdir -p "$PACKAGE_PATH"

# 设置构建环境
export GOOS=darwin
export GOARCH=amd64
export CGO_ENABLED=1

echo "构建 Go 二进制文件..."
cd "$PROJECT_ROOT"

# 构建二进制文件
go build -ldflags "-s -w -X main.Version=$VERSION" -o "$PACKAGE_PATH/$BINARY_NAME" cmd/server/main.go

echo "复制必需的文件..."

# 复制 web 静态文件
cp -r web "$PACKAGE_PATH/"

# 复制配置文件
cp -r configs "$PACKAGE_PATH/"

# 复制文档文件
cp README.md "$PACKAGE_PATH/"
cp QUICKSTART.md "$PACKAGE_PATH/"

# 创建启动脚本
cat > "$PACKAGE_PATH/start.sh" << 'EOF'
#!/bin/bash

echo "Starting GitCodeStatic Server..."
echo ""
echo "Web UI: http://localhost:8080"
echo "API Docs: http://localhost:8080/swagger/"
echo ""
echo "Press Ctrl+C to stop the server"
echo ""

# 设置可执行权限
chmod +x "./gitcodestatic"

# 启动服务器
./gitcodestatic
EOF

chmod +x "$PACKAGE_PATH/start.sh"

# 创建 LaunchAgent plist 文件
cat > "$PACKAGE_PATH/com.gitcodestatic.server.plist" << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.gitcodestatic.server</string>
    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/gitcodestatic</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>WorkingDirectory</key>
    <string>/usr/local/opt/gitcodestatic</string>
    <key>StandardOutPath</key>
    <string>/usr/local/var/log/gitcodestatic.log</string>
    <key>StandardErrorPath</key>
    <string>/usr/local/var/log/gitcodestatic.error.log</string>
</dict>
</plist>
EOF

# 创建 Homebrew 安装脚本
cat > "$PACKAGE_PATH/install-macos.sh" << 'EOF'
#!/bin/bash

# GitCodeStatic macOS 安装脚本

set -e

echo "安装 GitCodeStatic for macOS..."

# 检查是否有管理员权限
if [ "$EUID" -ne 0 ]; then
    echo "注意：某些操作可能需要管理员权限"
fi

# 创建安装目录
INSTALL_DIR="/usr/local/opt/gitcodestatic"
BIN_DIR="/usr/local/bin"

echo "创建安装目录: $INSTALL_DIR"
sudo mkdir -p "$INSTALL_DIR"
sudo mkdir -p "/usr/local/var/log"

# 复制文件
echo "复制文件..."
sudo cp -r ./* "$INSTALL_DIR/"

# 创建符号链接
echo "创建符号链接..."
sudo ln -sf "$INSTALL_DIR/gitcodestatic" "$BIN_DIR/gitcodestatic"

# 设置权限
sudo chmod +x "$INSTALL_DIR/gitcodestatic"
sudo chmod +x "$BIN_DIR/gitcodestatic"

# 安装 LaunchAgent
echo "安装 LaunchAgent..."
LAUNCH_AGENTS_DIR="$HOME/Library/LaunchAgents"
mkdir -p "$LAUNCH_AGENTS_DIR"
cp "$INSTALL_DIR/com.gitcodestatic.server.plist" "$LAUNCH_AGENTS_DIR/"

echo "安装完成！"
echo ""
echo "使用方法："
echo "  直接运行: gitcodestatic"
echo "  启动服务: launchctl load ~/Library/LaunchAgents/com.gitcodestatic.server.plist"
echo "  停止服务: launchctl unload ~/Library/LaunchAgents/com.gitcodestatic.server.plist"
echo ""
echo "服务将在 http://localhost:8080 上运行"
echo "配置文件位置: $INSTALL_DIR/configs/config.yaml"
EOF

chmod +x "$PACKAGE_PATH/install-macos.sh"

# 创建卸载脚本
cat > "$PACKAGE_PATH/uninstall-macos.sh" << 'EOF'
#!/bin/bash

# GitCodeStatic macOS 卸载脚本

set -e

echo "卸载 GitCodeStatic..."

# 停止服务
launchctl unload ~/Library/LaunchAgents/com.gitcodestatic.server.plist 2>/dev/null || true

# 删除 LaunchAgent
rm -f ~/Library/LaunchAgents/com.gitcodestatic.server.plist

# 删除符号链接
sudo rm -f /usr/local/bin/gitcodestatic

# 删除安装目录
sudo rm -rf /usr/local/opt/gitcodestatic

# 删除日志文件
sudo rm -f /usr/local/var/log/gitcodestatic.log
sudo rm -f /usr/local/var/log/gitcodestatic.error.log

echo "卸载完成！"
EOF

chmod +x "$PACKAGE_PATH/uninstall-macos.sh"

# 创建配置说明
cat > "$PACKAGE_PATH/README_macOS.md" << 'EOF'
# GitCodeStatic macOS 版本

## 快速启动

```bash
# 直接运行
./start.sh

# 或者直接运行二进制文件
./gitcodestatic
```

## 系统安装

```bash
# 安装到系统目录
./install-macos.sh

# 全局使用
gitcodestatic
```

## 服务管理

```bash
# 启动后台服务
launchctl load ~/Library/LaunchAgents/com.gitcodestatic.server.plist

# 停止后台服务
launchctl unload ~/Library/LaunchAgents/com.gitcodestatic.server.plist

# 卸载
./uninstall-macos.sh
```

## 配置文件

- `configs/config.yaml`: 主配置文件
- 可以修改端口、数据库路径等配置

## 数据目录

- 日志文件: `logs/app.log`
- 数据库文件: `workspace/gitcodestatic.db`
- 仓库缓存: `workspace/repos/`

## 访问地址

- Web UI: http://localhost:8080
- API 文档: http://localhost:8080/swagger/

## 停止服务

- 直接运行: 按 `Ctrl+C`
- 后台服务: `launchctl unload ~/Library/LaunchAgents/com.gitcodestatic.server.plist`
EOF

# 设置二进制文件权限
chmod +x "$PACKAGE_PATH/$BINARY_NAME"

# 创建 tar.gz 包
echo "创建压缩包..."
cd "$OUTPUT_PATH"
tar -czf "$PACKAGE_NAME.tar.gz" "$PACKAGE_NAME"

echo "构建完成！"
echo "输出路径: $OUTPUT_PATH/$PACKAGE_NAME.tar.gz"

# 显示文件大小
FILE_SIZE=$(stat -f%z "$PACKAGE_NAME.tar.gz" 2>/dev/null || stat -c%s "$PACKAGE_NAME.tar.gz" 2>/dev/null)
echo "压缩包大小: $(echo "scale=2; $FILE_SIZE/1024/1024" | bc) MB"