# GitCodeStatic - 快速启动指南

## 🚀 5分钟快速上手

### 1. 编译并运行

```bash
# 安装依赖
go mod tidy

# 运行服务
go run cmd/server/main.go
```

服务将在 `http://localhost:8080` 启动

### 2. 添加第一个仓库

```bash
curl -X POST http://localhost:8080/api/v1/repos/batch \
  -H "Content-Type: application/json" \
  -d '{
    "urls": ["https://github.com/gin-gonic/gin.git"]
  }'
```

响应示例：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total": 1,
    "succeeded": [{
      "repo_id": 1,
      "url": "https://github.com/gin-gonic/gin.git",
      "task_id": 1
    }],
    "failed": []
  }
}
```

### 3. 等待克隆完成

```bash
# 查看仓库状态
curl http://localhost:8080/api/v1/repos/1
```

等待 `status` 变为 `"ready"`

### 4. 触发代码统计

```bash
curl -X POST http://localhost:8080/api/v1/stats/calculate \
  -H "Content-Type: application/json" \
  -d '{
    "repo_id": 1,
    "branch": "master",
    "constraint": {
      "type": "commit_limit",
      "limit": 100
    }
  }'
```

### 5. 查询统计结果

```bash
curl "http://localhost:8080/api/v1/stats/result?repo_id=1&branch=master&constraint_type=commit_limit&limit=100"
```

你将看到：
- 总提交数
- 贡献者列表
- 每个贡献者的代码变更统计（新增/删除/修改/净增加）

## 📊 完整工作流示例

```bash
# 1. 添加多个仓库
curl -X POST http://localhost:8080/api/v1/repos/batch \
  -H "Content-Type: application/json" \
  -d '{
    "urls": [
      "https://github.com/gin-gonic/gin.git",
      "https://github.com/go-chi/chi.git"
    ]
  }'

# 2. 查看所有ready状态的仓库
curl "http://localhost:8080/api/v1/repos?status=ready"

# 3. 先查询某个日期到现在有多少提交（辅助决策）
curl "http://localhost:8080/api/v1/stats/commit-count?repo_id=1&branch=master&from=2024-01-01"

# 4. 根据提交数选择合适的约束类型
# 如果提交数少（<1000），用日期范围
curl -X POST http://localhost:8080/api/v1/stats/calculate \
  -H "Content-Type: application/json" \
  -d '{
    "repo_id": 1,
    "branch": "master",
    "constraint": {
      "type": "date_range",
      "from": "2024-01-01",
      "to": "2024-12-31"
    }
  }'

# 5. 查询结果（会自动命中缓存）
curl "http://localhost:8080/api/v1/stats/result?repo_id=1&branch=master&constraint_type=date_range&from=2024-01-01&to=2024-12-31"

# 6. 切换分支
curl -X POST http://localhost:8080/api/v1/repos/1/switch-branch \
  -H "Content-Type: application/json" \
  -d '{"branch": "develop"}'

# 7. 更新仓库（获取最新代码）
curl -X POST http://localhost:8080/api/v1/repos/1/update

# 8. 重置仓库（清除缓存+重新克隆）
curl -X POST http://localhost:8080/api/v1/repos/1/reset
```

## 🔧 常见问题

### Q: 如何处理私有仓库？
A: 暂不支持通过API设置凭据，需要手动在数据库中添加或使用https://username:token@github.com/repo.git格式

### Q: 统计任务一直pending？
A: 检查worker是否正常启动，查看日志：
```bash
# 日志会显示worker pool启动信息
# 确认没有错误
```

### Q: 如何加速统计？
A: 
1. 确保安装了git命令（比go-git快很多）
2. 增加stats_workers数量
3. 使用commit_limit而不是date_range（如果适用）

### Q: 缓存占用空间过大？
A: 修改配置：
```yaml
cache:
  max_total_size: 5368709120  # 改为5GB
  retention_days: 7            # 只保留7天
```

## 🎯 API完整列表

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/v1/repos/batch` | POST | 批量添加仓库 |
| `/api/v1/repos` | GET | 获取仓库列表 |
| `/api/v1/repos/:id` | GET | 获取仓库详情 |
| `/api/v1/repos/:id/switch-branch` | POST | 切换分支 |
| `/api/v1/repos/:id/update` | POST | 更新仓库 |
| `/api/v1/repos/:id/reset` | POST | 重置仓库 |
| `/api/v1/repos/:id` | DELETE | 删除仓库 |
| `/api/v1/stats/calculate` | POST | 触发统计 |
| `/api/v1/stats/result` | GET | 查询统计结果 |
| `/api/v1/stats/commit-count` | GET | 查询提交次数 |
| `/health` | GET | 健康检查 |

## 📝 日志查看

```bash
# 开发模式：日志输出到stdout
go run cmd/server/main.go

# 查看结构化日志
# 示例：
{"level":"info","time":"2024-12-31T12:00:00+08:00","message":"worker started","worker_id":1}
{"level":"info","time":"2024-12-31T12:00:01+08:00","message":"task started","worker_id":1,"task_id":1,"task_type":"clone","repo_id":1}
```

## 🎉 下一步

- 阅读完整 [README.md](README.md) 了解所有功能
- 查看 [ARCHITECTURE.md](ARCHITECTURE.md) 理解系统架构
- 查看单元测试示例学习如何测试：`test/unit/`
- 根据需求调整 `configs/config.yaml` 配置

Happy Coding! 🚀
