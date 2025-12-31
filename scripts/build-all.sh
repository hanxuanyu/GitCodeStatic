#!/bin/bash

# GitCodeStatic 全平台构建脚本
# 在支持交叉编译的环境中构建所有平台的包

set -e

VERSION=${1:-"latest"}
OUTPUT_DIR=${2:-"dist"}

echo "开始全平台构建 GitCodeStatic v$VERSION..."

# 设置变量
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUTPUT_PATH="$PROJECT_ROOT/$OUTPUT_DIR"

# 清理输出目录
if [ -d "$OUTPUT_PATH" ]; then
    echo "清理旧的输出目录..."
    rm -rf "$OUTPUT_PATH"
fi

mkdir -p "$OUTPUT_PATH"

# 平台列表: OS/ARCH
PLATFORMS=(
    "windows/amd64"
    "linux/amd64"
    "darwin/amd64"
    "linux/arm64"
    "darwin/arm64"
)

cd "$PROJECT_ROOT"

for platform in "${PLATFORMS[@]}"; do
    IFS='/' read -r GOOS GOARCH <<< "$platform"
    
    echo ""
    echo "构建 $GOOS/$GOARCH..."
    
    # 设置输出文件名
    BINARY_NAME="gitcodestatic"
    if [ "$GOOS" = "windows" ]; then
        BINARY_NAME="gitcodestatic.exe"
    fi
    
    # 设置包名
    PACKAGE_NAME="gitcodestatic-$GOOS-$GOARCH-$VERSION"
    PACKAGE_PATH="$OUTPUT_PATH/$PACKAGE_NAME"
    
    mkdir -p "$PACKAGE_PATH"
    
    # 设置构建环境
    export GOOS="$GOOS"
    export GOARCH="$GOARCH"
    export CGO_ENABLED=1
    
    # 特殊处理：ARM64 和交叉编译时禁用 CGO
    if [ "$GOARCH" = "arm64" ] || [ "$(uname)" != "$(echo $GOOS | tr '[:lower:]' '[:upper:]')" ]; then
        export CGO_ENABLED=0
    fi
    
    echo "  构建二进制文件..."
    go build -ldflags "-s -w -X main.Version=$VERSION" -o "$PACKAGE_PATH/$BINARY_NAME" cmd/server/main.go
    
    echo "  复制文件..."
    
    # 复制通用文件
    cp -r web "$PACKAGE_PATH/"
    cp -r configs "$PACKAGE_PATH/"
    cp README.md "$PACKAGE_PATH/"
    cp QUICKSTART.md "$PACKAGE_PATH/"
    
    # 根据平台创建特定的启动脚本
    if [ "$GOOS" = "windows" ]; then
        # Windows 启动脚本
        cat > "$PACKAGE_PATH/start.bat" << 'EOF'
@echo off
echo Starting GitCodeStatic Server...
echo.
echo Web UI: http://localhost:8080
echo API Docs: http://localhost:8080/swagger/
echo.
gitcodestatic.exe
pause
EOF
        
        # Windows 配置说明
        cat > "$PACKAGE_PATH/使用说明.txt" << 'EOF'
GitCodeStatic Windows 版本

## 使用方法
1. 双击 start.bat 启动服务器
2. 打开浏览器访问 http://localhost:8080
3. 查看 API 文档: http://localhost:8080/swagger/

## 配置文件
- configs/config.yaml: 主配置文件

## 停止服务
- 在命令窗口中按 Ctrl+C 停止服务器
EOF
        
    else
        # Unix 启动脚本
        cat > "$PACKAGE_PATH/start.sh" << 'EOF'
#!/bin/bash
echo "Starting GitCodeStatic Server..."
echo ""
echo "Web UI: http://localhost:8080"
echo "API Docs: http://localhost:8080/swagger/"
echo ""
echo "Press Ctrl+C to stop the server"
echo ""

chmod +x "./gitcodestatic"
./gitcodestatic
EOF
        chmod +x "$PACKAGE_PATH/start.sh"
        
        # Unix 说明文件
        cat > "$PACKAGE_PATH/README_$GOOS.md" << EOF
# GitCodeStatic $GOOS 版本

## 快速启动
\`\`\`bash
./start.sh
# 或
./gitcodestatic
\`\`\`

## 配置文件
- configs/config.yaml: 主配置文件

## 访问地址
- Web UI: http://localhost:8080  
- API 文档: http://localhost:8080/swagger/

## 停止服务
按 Ctrl+C 停止服务器
EOF
    fi
    
    # 设置可执行权限
    chmod +x "$PACKAGE_PATH/$BINARY_NAME"
    
    # 创建压缩包
    echo "  创建压缩包..."
    cd "$OUTPUT_PATH"
    
    if [ "$GOOS" = "windows" ]; then
        # Windows 使用 zip
        if command -v zip >/dev/null 2>&1; then
            zip -r "$PACKAGE_NAME.zip" "$PACKAGE_NAME/" >/dev/null
            echo "  ✓ 已创建 $PACKAGE_NAME.zip"
        else
            echo "  ⚠ 未找到 zip 命令，跳过压缩"
        fi
    else
        # Unix 使用 tar.gz
        tar -czf "$PACKAGE_NAME.tar.gz" "$PACKAGE_NAME"
        echo "  ✓ 已创建 $PACKAGE_NAME.tar.gz"
    fi
    
    cd "$PROJECT_ROOT"
done

echo ""
echo "==================================="
echo "全平台构建完成！"
echo "==================================="
echo "输出目录: $OUTPUT_PATH"
echo ""
echo "构建的平台："

# 显示构建结果
cd "$OUTPUT_PATH"
for platform in "${PLATFORMS[@]}"; do
    IFS='/' read -r GOOS GOARCH <<< "$platform"
    PACKAGE_NAME="gitcodestatic-$GOOS-$GOARCH-$VERSION"
    
    if [ "$GOOS" = "windows" ]; then
        ARCHIVE="$PACKAGE_NAME.zip"
    else
        ARCHIVE="$PACKAGE_NAME.tar.gz"
    fi
    
    if [ -f "$ARCHIVE" ]; then
        SIZE=$(du -h "$ARCHIVE" | cut -f1)
        echo "  ✓ $GOOS/$GOARCH - $ARCHIVE ($SIZE)"
    else
        echo "  ✗ $GOOS/$GOARCH - 构建失败"
    fi
done

echo ""
echo "总文件数: $(ls -1 *.tar.gz *.zip 2>/dev/null | wc -l)"
TOTAL_SIZE=$(du -sh . | cut -f1)
echo "总大小: $TOTAL_SIZE"