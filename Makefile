.PHONY: build run test clean install help

# 变量定义
APP_NAME=gitcodestatic
BUILD_DIR=./bin
CMD_DIR=./cmd/server
CONFIG_DIR=./configs
WORKSPACE_DIR=./workspace

# 默认目标
help:
	@echo "GitCodeStatic - Makefile Commands"
	@echo ""
	@echo "Usage:"
	@echo "  make install    - 安装依赖"
	@echo "  make build      - 编译项目"
	@echo "  make run        - 运行服务"
	@echo "  make test       - 运行测试"
	@echo "  make test-cover - 运行测试并生成覆盖率报告"
	@echo "  make clean      - 清理构建文件"
	@echo "  make fmt        - 格式化代码"
	@echo "  make lint       - 代码检查"
	@echo "  make init-dirs  - 初始化工作目录"
	@echo ""

# 安装依赖
install:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# 编译项目
build:
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(APP_NAME) $(CMD_DIR)/main.go
	@echo "Build complete: $(BUILD_DIR)/$(APP_NAME)"

# 运行服务
run:
	@echo "Starting $(APP_NAME)..."
	go run $(CMD_DIR)/main.go

# 运行测试
test:
	@echo "Running tests..."
	go test ./... -v

# 测试覆盖率
test-cover:
	@echo "Running tests with coverage..."
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# 清理构建文件
clean:
	@echo "Cleaning..."
	rm -rf $(BUILD_DIR)
	rm -rf $(WORKSPACE_DIR)
	rm -f coverage.out coverage.html
	@echo "Clean complete"

# 格式化代码
fmt:
	@echo "Formatting code..."
	go fmt ./...
	@echo "Format complete"

# 代码检查（需要安装golangci-lint）
lint:
	@echo "Linting code..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# 初始化工作目录
init-dirs:
	@echo "Initializing workspace directories..."
	@mkdir -p $(WORKSPACE_DIR)/cache
	@mkdir -p $(WORKSPACE_DIR)/stats
	@echo "Directories created"

# 开发模式（热重载，需要安装air）
dev:
	@echo "Starting development mode..."
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "air not installed. Run: go install github.com/cosmtrek/air@latest"; \
		echo "Falling back to normal run..."; \
		make run; \
	fi

# Docker相关（可选）
docker-build:
	@echo "Building Docker image..."
	docker build -t $(APP_NAME):latest .

docker-run:
	@echo "Running Docker container..."
	docker run -p 8080:8080 -v $(PWD)/workspace:/app/workspace $(APP_NAME):latest

# 生产构建（优化）
build-prod:
	@echo "Building for production..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
		-ldflags="-w -s" \
		-o $(BUILD_DIR)/$(APP_NAME) \
		$(CMD_DIR)/main.go
	@echo "Production build complete: $(BUILD_DIR)/$(APP_NAME)"
