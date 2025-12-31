# GitCodeStatic - Git仓库统计与缓存系统

一个用Go实现的高性能Git仓库代码统计与缓存系统，支持批量仓库管理、异步任务处理、智能缓存、多种统计维度，提供 Swagger API 文档和 Web 管理界面。

## 功能特性

### 核心功能
- ✅ **批量仓库管理**：支持批量添加、更新、切换分支、重置仓库
- ✅ **异步任务处理**：基于队列的Worker池，支持并发控制和任务去重
- ✅ **代码统计**：按分支、贡献者维度统计代码变更（新增/删除/修改/净增加）
- ✅ **智能缓存**：基于文件+数据库的双层缓存，自动失效机制
- ✅ **灵活约束**：支持日期范围或提交次数限制（互斥校验）
- ✅ **辅助查询**：查询指定日期到当前的提交次数
- ✅ **凭据管理**：支持私有仓库（用户名/密码/Token）
- ✅ **Git双引擎**：优先使用git命令，可fallback到go-git

### 技术特性
- 📊 **可观测**：结构化日志（zerolog）、基础指标收集
- 🔒 **安全**：凭据加密存储、URL脱敏、命令注入防护
- 🧪 **可测试**：关键逻辑提供单元测试示例
- 🎯 **RESTful API**：统一响应格式、完善错误码
- 📚 **Swagger 文档**：完整的 API 文档和交互式测试界面
- 🖥️ **Web UI**：基于 Vue 3 + Element Plus 的管理界面，支持离线部署
- 🗄️ **存储灵活**：默认SQLite，可扩展PostgreSQL
- ⚡ **高性能**：任务去重、缓存命中、并发控制

## 快速体验

### 启动服务

```bash
# 构建
make build

# 运行
./bin/gitcodestatic

# 或直接运行
go run cmd/server/main.go
```

服务启动后，可以通过以下方式访问：

- **Web UI**: http://localhost:8080/
- **Swagger API 文档**: http://localhost:8080/swagger/index.html
- **Health Check**: http://localhost:8080/health
cp configs/config.yaml configs/config.local.yaml
```

主要配置项：

```yaml
server:
  port: 8080

workspace:
  cache_dir: ./workspace/cache  # 仓库本地缓存
  stats_dir: ./workspace/stats  # 统计结果存储

worker:
  clone_workers: 2   # 克隆并发数
  stats_workers: 2   # 统计并发数

cache:
  max_total_size: 10737418240  # 10GB
  retention_days: 30

git:
  command_path: ""   # 空表示使用PATH中的git
  fallback_to_gogit: true
```

### 运行

```bash
# 开发模式
go run cmd/server/main.go

# 编译
go build -o gitcodestatic cmd/server/main.go

# 运行
./gitcodestatic
```

服务启动后访问：
- API: `http://localhost:8080/api/v1`
- Health: `http://localhost:8080/health`

## Web UI 使用

启动服务后访问 http://localhost:8080/ 进入 Web 管理界面。

### 主要功能

1. **仓库管理**
   - 批量添加仓库（支持多行输入）
   - 查看仓库列表和状态
   - 切换分支、更新、重置、删除操作

2. **统计管理**
   - 触发统计计算（支持日期范围和提交数限制）
   - 查询统计结果（可视化展示）
   - 查看贡献者详情

3. **API 文档**
   - 快速访问 Swagger 文档
   - API 使用示例

详细使用说明请参考 [WEBUI_GUIDE.md](docs/WEBUI_GUIDE.md)

## API 使用示例

### 1. 批量添加仓库

```bash
curl -X POST http://localhost:8080/api/v1/repos/batch \
  -H "Content-Type: application/json" \
  -d '{
    "urls": [
      "https://github.com/golang/go.git",
      "https://github.com/kubernetes/kubernetes.git"
    ]
  }'
```

响应：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total": 2,
    "succeeded": [
      {
        "repo_id": 1,
        "url": "https://github.com/golang/go.git",
        "task_id": 101
      }
    ],
    "failed": []
  }
}
```

### 2. 查询仓库列表

```bash
curl http://localhost:8080/api/v1/repos?status=ready&page=1&page_size=20
```

### 3. 触发代码统计

**按日期范围统计：**
```bash
curl -X POST http://localhost:8080/api/v1/stats/calculate \
  -H "Content-Type: application/json" \
  -d '{
    "repo_id": 1,
    "branch": "main",
    "constraint": {
      "type": "date_range",
      "from": "2024-01-01",
      "to": "2024-12-31"
    }
  }'
```

**按提交次数统计：**
```bash
curl -X POST http://localhost:8080/api/v1/stats/calculate \
  -H "Content-Type: application/json" \
  -d '{
    "repo_id": 1,
    "branch": "main",
    "constraint": {
      "type": "commit_limit",
      "limit": 100
    }
  }'
```

### 4. 查询统计结果

```bash
curl "http://localhost:8080/api/v1/stats/result?repo_id=1&branch=main&constraint_type=date_range&from=2024-01-01&to=2024-12-31"
```

响应：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "cache_hit": true,
    "cached_at": "2024-12-31T10:00:00Z",
    "commit_hash": "abc123...",
    "statistics": {
      "summary": {
        "total_commits": 150,
        "total_contributors": 5,
        "date_range": {
          "from": "2024-01-01",
          "to": "2024-12-31"
        }
      },
      "by_contributor": [
        {
          "author": "Alice",
          "email": "alice@example.com",
          "commits": 50,
          "additions": 1000,
          "deletions": 200,
          "modifications": 200,
          "net_additions": 800
        }
      ]
    }
  }
}
```

### 5. 辅助查询：统计提交次数

```bash
curl "http://localhost:8080/api/v1/stats/commit-count?repo_id=1&branch=main&from=2024-01-01"
```

响应：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "repo_id": 1,
    "branch": "main",
    "from": "2024-01-01",
    "to": "HEAD",
    "commit_count": 150
  }
}
```

### 6. 其他操作

**切换分支：**
```bash
curl -X POST http://localhost:8080/api/v1/repos/1/switch-branch \
  -H "Content-Type: application/json" \
  -d '{"branch": "develop"}'
```

**更新仓库：**
```bash
curl -X POST http://localhost:8080/api/v1/repos/1/update
```

**重置仓库：**
```bash
curl -X POST http://localhost:8080/api/v1/repos/1/reset
```

## 数据模型

### 统计指标说明

| 字段 | 说明 | 计算方式 |
|------|------|----------|
| `additions` | 新增行数 | git log --numstat 的additions |
| `deletions` | 删除行数 | git log --numstat 的deletions |
| `modifications` | 修改行数 | min(additions, deletions) |
| `net_additions` | 净增加行数 | additions - deletions |

**修改行数定义**：一行代码被替换时，同时计入additions和deletions，`modifications`取两者最小值表示真正被修改的行数。

### 约束类型互斥

`date_range` 和 `commit_limit` 互斥使用：

- ✅ `{"type": "date_range", "from": "2024-01-01", "to": "2024-12-31"}`
- ✅ `{"type": "commit_limit", "limit": 100}`
- ❌ `{"type": "date_range", "from": "2024-01-01", "to": "2024-12-31", "limit": 100}` - 错误

## 缓存策略

### 缓存Key生成

```
SHA256(repo_id | branch | constraint_type | constraint_value | commit_hash)
```

### 缓存失效时机

1. 仓库更新（pull）：commit_hash变化，旧缓存自然失效
2. 切换分支：branch变化，缓存key不同
3. 重置仓库：主动删除该仓库所有缓存

### 存储位置

- **元数据**：SQLite `stats_cache` 表
- **结果数据**：文件系统 `workspace/stats/{cache_key}.json.gz`（gzip压缩）

## 任务系统

### 任务类型

- `clone`: 克隆仓库
- `pull`: 拉取更新
- `switch`: 切换分支
- `reset`: 重置仓库
- `stats`: 统计代码

### 任务状态

- `pending`: 等待处理
- `running`: 执行中
- `completed`: 完成
- `failed`: 失败
- `cancelled`: 已取消

### 去重机制

相同仓库+相同任务类型+相同参数的待处理任务只会存在一个，重复提交返回已有任务ID。

## 测试

### 运行单元测试

```bash
# 运行所有测试
go test ./...

# 运行特定测试
go test ./test/unit -v

# 测试覆盖率
go test ./... -cover
```

### 测试示例

见 `test/unit/` 目录：
- `service_test.go` - 参数校验测试
- `cache_test.go` - 缓存key生成测试

## 开发指南

### 添加新的 API 端点

1. 在 `internal/api/handlers/` 创建handler方法
2. 添加 Swagger 注释
3. 在 `internal/api/router.go` 注册路由
4. 重新生成 Swagger 文档：
   ```bash
   swag init -g cmd/server/main.go -o docs
   ```

### 添加新的任务类型

1. 在 `internal/models/task.go` 定义任务类型常量
2. 在 `internal/worker/handlers.go` 实现 `TaskHandler` 接口
3. 在 `cmd/server/main.go` 注册handler

### 扩展存储层

实现 `internal/storage/interface.go` 中的接口即可，参考 `sqlite/` 实现。

### 扩展 Web UI

1. 修改 `web/index.html` 添加新的页面组件
2. 在 `web/static/app.js` 添加相应的方法和数据
3. 参考 [WEBUI_GUIDE.md](docs/WEBUI_GUIDE.md) 了解详细开发流程

## 错误码

| Code | 说明 |
|------|------|
| 0 | 成功 |
| 40001 | 参数校验失败 |
| 40002 | 操作不允许 |
| 40400 | 资源未找到 |
| 40900 | 资源冲突 |
| 50000 | 内部错误 |
| 50001 | 数据库错误 |
| 50002 | Git操作失败 |

## 文档

- [架构设计](ARCHITECTURE.md) - 系统架构和技术选型
- [快速开始](QUICKSTART.md) - 快速上手指南
- [Web UI 使用指南](docs/WEBUI_GUIDE.md) - 前端和 Swagger 文档使用
- [项目总结](SUMMARY.md) - 项目完整总结

## 性能优化建议

1. **Git命令模式**：确保安装git命令，性能比go-git快10-100倍
2. **并发调优**：根据CPU核心数和IO性能调整worker数量
3. **缓存预热**：对常用仓库/分支提前触发统计
4. **定期清理**：配置缓存保留天数和总大小限制

## 已知限制

1. 单机部署，不支持分布式（可扩展）
2. go-git模式性能较差，仅作为fallback
3. 大仓库（>5GB）统计可能耗时较长
4. SSH认证暂未完整实现（仅支持https）

## 技术栈

- **后端**: Go 1.21+, Chi Router, zerolog
- **存储**: SQLite (可扩展 PostgreSQL)
- **Git**: git CLI + go-git fallback
- **文档**: Swagger 2.0 (swaggo/swag)
- **前端**: Vue 3, Element Plus, Axios
- **特性**: 完全离线部署支持

## 贡献

欢迎提Issue和PR！

## License

MIT License

## 作者

Created by Senior Backend/Full-stack Engineer (Go专家)
