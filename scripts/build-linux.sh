#!/bin/bash

# GitCodeStatic Linux 打包脚本
# Bash 脚本用于在 Linux 平台构建和打包

set -e

VERSION=${1:-"latest"}
OUTPUT_DIR=${2:-"dist"}

echo "开始构建 GitCodeStatic for Linux..."

# 设置变量
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BINARY_NAME="gitcodestatic"
PACKAGE_NAME="gitcodestatic-linux-$VERSION"
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
export GOOS=linux
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

# 创建系统服务文件
cat > "$PACKAGE_PATH/gitcodestatic.service" << EOF
[Unit]
Description=GitCodeStatic Git Repository Statistics Service
After=network.target

[Service]
Type=simple
User=gitcodestatic
WorkingDirectory=/opt/gitcodestatic
ExecStart=/opt/gitcodestatic/gitcodestatic
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

# 创建安装脚本
cat > "$PACKAGE_PATH/install.sh" << 'EOF'
#!/bin/bash

# GitCodeStatic Linux 安装脚本

set -e

if [ "$EUID" -ne 0 ]; then
    echo "请使用 sudo 运行此安装脚本"
    exit 1
fi

echo "安装 GitCodeStatic..."

# 创建用户
if ! id "gitcodestatic" &>/dev/null; then
    echo "创建 gitcodestatic 用户..."
    useradd -r -s /bin/false gitcodestatic
fi

# 创建安装目录
INSTALL_DIR="/opt/gitcodestatic"
mkdir -p "$INSTALL_DIR"

# 复制文件
echo "复制文件到 $INSTALL_DIR..."
cp -r ./* "$INSTALL_DIR/"

# 设置权限
chown -R gitcodestatic:gitcodestatic "$INSTALL_DIR"
chmod +x "$INSTALL_DIR/gitcodestatic"

# 安装系统服务
echo "安装系统服务..."
cp "$INSTALL_DIR/gitcodestatic.service" /etc/systemd/system/
systemctl daemon-reload

echo "安装完成！"
echo ""
echo "使用以下命令管理服务："
echo "  启动服务: sudo systemctl start gitcodestatic"
echo "  停止服务: sudo systemctl stop gitcodestatic"
echo "  开机自启: sudo systemctl enable gitcodestatic"
echo "  查看状态: sudo systemctl status gitcodestatic"
echo ""
echo "服务将在 http://localhost:8080 上运行"
EOF

chmod +x "$PACKAGE_PATH/install.sh"

# 创建配置说明
cat > "$PACKAGE_PATH/README_Linux.md" << 'EOF'
# GitCodeStatic Linux 版本

## 快速启动

```bash
# 直接运行
./start.sh

# 或者直接运行二进制文件
./gitcodestatic
```

## 系统服务安装

```bash
# 以 root 权限安装
sudo ./install.sh

# 启动服务
sudo systemctl start gitcodestatic

# 开机自启
sudo systemctl enable gitcodestatic
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
- 系统服务: `sudo systemctl stop gitcodestatic`
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
FILE_SIZE=$(stat -c%s "$PACKAGE_NAME.tar.gz" 2>/dev/null || stat -f%z "$PACKAGE_NAME.tar.gz" 2>/dev/null)
echo "压缩包大小: $(echo "scale=2; $FILE_SIZE/1024/1024" | bc) MB"