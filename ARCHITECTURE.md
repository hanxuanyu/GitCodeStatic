# Git 仓库统计与缓存系统 - 架构设计文档

## 1. 总体架构

### 1.1 模块划分

```
┌─────────────────────────────────────────────────────────────┐
│                         API Layer                            │
│  ┌────────────┬────────────┬────────────┬─────────────┐    │
│  │  Repo APIs │ Stats APIs │ Task APIs  │ Health APIs │    │
│  └────────────┴────────────┴────────────┴─────────────┘    │
└─────────────────────────────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────┐
│                       Service Layer                          │
│  ┌──────────────────┬──────────────────┬─────────────────┐ │
│  │  RepoService     │  StatsService    │  TaskService    │ │
│  │  - AddRepos      │  - Calculate     │  - Submit       │ │
│  │  - UpdateRepo    │  - QueryCache    │  - Query        │ │
│  │  - SwitchBranch  │  - CountCommits  │  - Cancel       │ │
│  │  - SetCreds      │                  │                 │ │
│  │  - Reset         │                  │                 │ │
│  └──────────────────┴──────────────────┴─────────────────┘ │
└─────────────────────────────────────────────────────────────┘
                             │
        ┌────────────────────┼────────────────────┐
        ▼                    ▼                    ▼
┌──────────────┐    ┌──────────────┐    ┌──────────────┐
│ Git Manager  │    │ Cache Layer  │    │ Task Queue   │
│ - Clone      │    │ - Get/Set    │    │ - Enqueue    │
│ - Pull       │    │ - Invalidate │    │ - Dequeue    │
│ - Checkout   │    │ - KeyGen     │    │ - Dedupe     │
│ - Stats      │    └──────────────┘    └──────────────┘
│ (cmd/go-git) │                                │
└──────────────┘                                ▼
        │                              ┌──────────────────┐
        │                              │   Worker Pool    │
        │                              │  ┌────────────┐  │
        │                              │  │ Clone      │  │
        │                              │  │ Pull       │  │
        │                              │  │ Switch     │  │
        │                              │  │ Stats      │  │
        │                              │  │ Reset      │  │
        │                              │  └────────────┘  │
        │                              └──────────────────┘
        ▼
┌─────────────────────────────────────────────────────────────┐
│                       Storage Layer                          │
│  ┌──────────────┬──────────────┬──────────────────────────┐ │
│  │  Repo Store  │  Task Store  │  StatsCache Store       │ │
│  │  (SQLite/PG) │  (SQLite/PG) │  (SQLite/PG + Disk)     │ │
│  └──────────────┴──────────────┴──────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
                             │
                             ▼
                    ┌──────────────────┐
                    │  File System     │
                    │  workspace/cache/│
                    │  workspace/stats/│
                    └──────────────────┘
```

### 1.2 目录结构

```
GitCodeStatic/
├── cmd/
│   └── server/
│       └── main.go                 # 主程序入口
├── internal/
│   ├── api/                        # API层
│   │   ├── handlers/               # HTTP handlers
│   │   │   ├── repo.go            # 仓库相关API
│   │   │   ├── stats.go           # 统计相关API
│   │   │   ├── task.go            # 任务相关API
│   │   │   └── health.go          # 健康检查API
│   │   ├── middleware/            # 中间件
│   │   │   ├── logger.go          # 日志中间件
│   │   │   ├── recovery.go        # 恢复中间件
│   │   │   └── metrics.go         # 指标中间件
│   │   └── router.go              # 路由配置
│   ├── service/                    # 服务层
│   │   ├── repo_service.go        # 仓库服务
│   │   ├── stats_service.go       # 统计服务
│   │   └── task_service.go        # 任务服务
│   ├── worker/                     # 异步任务处理
│   │   ├── queue.go               # 任务队列
│   │   ├── worker.go              # Worker实现
│   │   ├── pool.go                # Worker池
│   │   └── handlers.go            # 任务处理器
│   ├── git/                        # Git操作抽象
│   │   ├── manager.go             # Git管理器接口
│   │   ├── cmd_git.go             # Git命令实现
│   │   └── go_git.go              # go-git实现
│   ├── stats/                      # 统计模块
│   │   ├── calculator.go          # 统计计算器
│   │   ├── parser.go              # Git日志解析
│   │   └── models.go              # 统计数据模型
│   ├── cache/                      # 缓存模块
│   │   ├── cache.go               # 缓存接口
│   │   ├── key.go                 # 缓存key生成
│   │   └── file_cache.go          # 文件+DB缓存实现
│   ├── storage/                    # 存储层
│   │   ├── interface.go           # 存储接口定义
│   │   ├── sqlite/                # SQLite实现
│   │   │   ├── repo.go
│   │   │   ├── task.go
│   │   │   └── stats_cache.go
│   │   └── postgres/              # PostgreSQL实现（可选）
│   │       ├── repo.go
│   │       ├── task.go
│   │       └── stats_cache.go
│   ├── models/                     # 数据模型
│   │   ├── repo.go                # 仓库模型
│   │   ├── task.go                # 任务模型
│   │   └── stats.go               # 统计模型
│   ├── config/                     # 配置
│   │   └── config.go              # 配置结构和加载
│   ├── logger/                     # 日志
│   │   └── logger.go              # 结构化日志
│   ├── metrics/                    # 指标
│   │   └── metrics.go             # 基础指标收集
│   └── security/                   # 安全
│       ├── credentials.go         # 凭据管理
│       └── validator.go           # 输入校验
├── pkg/                            # 公共库
│   └── utils/
│       ├── hash.go                # 哈希工具
│       └── path.go                # 路径工具
├── test/                           # 测试
│   ├── unit/                      # 单元测试
│   └── integration/               # 集成测试
├── configs/                        # 配置文件
│   └── config.yaml
├── scripts/                        # 脚本
│   └── init_db.sql                # 数据库初始化
├── go.mod
├── go.sum
├── Makefile
├── README.md
└── ARCHITECTURE.md                 # 本文档
```

## 2. 数据模型

### 2.1 表结构设计 (PostgreSQL/SQLite)

#### 2.1.1 仓库表 (repositories)

```sql
CREATE TABLE repositories (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT,  -- PG: SERIAL PRIMARY KEY
    url                 TEXT NOT NULL UNIQUE,                -- 仓库URL
    name                TEXT NOT NULL,                       -- 仓库名称（从URL解析）
    current_branch      TEXT,                                -- 当前分支
    local_path          TEXT NOT NULL UNIQUE,                -- 本地缓存路径
    status              TEXT NOT NULL DEFAULT 'pending',     -- pending/cloning/ready/failed
    error_message       TEXT,                                -- 错误信息
    last_pull_at        TIMESTAMP,                           -- 最后拉取时间
    last_commit_hash    TEXT,                                -- 最后commit哈希
    credential_id       TEXT,                                -- 凭据ID（引用加密存储）
    created_at          TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_repositories_status ON repositories(status);
CREATE INDEX idx_repositories_updated_at ON repositories(updated_at);
```

#### 2.1.2 任务表 (tasks)

```sql
CREATE TABLE tasks (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT,  -- PG: SERIAL PRIMARY KEY
    task_type           TEXT NOT NULL,                       -- clone/pull/switch/stats/reset/count_commits
    repo_id             INTEGER NOT NULL,                    -- 关联仓库ID
    status              TEXT NOT NULL DEFAULT 'pending',     -- pending/running/completed/failed/cancelled
    priority            INTEGER NOT NULL DEFAULT 0,          -- 优先级（数字越大优先级越高）
    parameters          TEXT,                                -- JSON格式参数（分支名、统计条件等）
    result              TEXT,                                -- JSON格式结果
    error_message       TEXT,                                -- 错误信息
    retry_count         INTEGER NOT NULL DEFAULT 0,          -- 重试次数
    started_at          TIMESTAMP,                           -- 开始时间
    completed_at        TIMESTAMP,                           -- 完成时间
    created_at          TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (repo_id) REFERENCES repositories(id) ON DELETE CASCADE
);

CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_tasks_repo_id ON tasks(repo_id);
CREATE INDEX idx_tasks_type_repo ON tasks(task_type, repo_id, status);
CREATE INDEX idx_tasks_created_at ON tasks(created_at);

-- 任务去重：同一仓库+同一类型+相同参数的任务，pending状态下只允许存在一个
CREATE UNIQUE INDEX idx_tasks_dedup ON tasks(repo_id, task_type, parameters) 
    WHERE status IN ('pending', 'running');
```

#### 2.1.3 统计缓存表 (stats_cache)

```sql
CREATE TABLE stats_cache (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT,  -- PG: SERIAL PRIMARY KEY
    repo_id             INTEGER NOT NULL,                    -- 仓库ID
    branch              TEXT NOT NULL,                       -- 分支名
    constraint_type     TEXT NOT NULL,                       -- date_range/commit_limit
    constraint_value    TEXT NOT NULL,                       -- JSON: {"from":"2024-01-01","to":"2024-12-31"} 或 {"limit":100}
    commit_hash         TEXT NOT NULL,                       -- 统计截止的commit hash
    result_path         TEXT NOT NULL,                       -- 统计结果文件路径
    result_size         INTEGER NOT NULL,                    -- 结果文件大小(bytes)
    cache_key           TEXT NOT NULL UNIQUE,                -- 缓存键（用于快速查询）
    hit_count           INTEGER NOT NULL DEFAULT 0,          -- 缓存命中次数
    created_at          TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_hit_at         TIMESTAMP,                           -- 最后命中时间
    FOREIGN KEY (repo_id) REFERENCES repositories(id) ON DELETE CASCADE
);

CREATE INDEX idx_stats_cache_key ON stats_cache(cache_key);
CREATE INDEX idx_stats_cache_repo ON stats_cache(repo_id, branch);
CREATE INDEX idx_stats_cache_created_at ON stats_cache(created_at);

-- 唯一约束：同一仓库+分支+约束类型+约束值+commit_hash只能有一条记录
CREATE UNIQUE INDEX idx_stats_cache_unique ON stats_cache(
    repo_id, branch, constraint_type, constraint_value, commit_hash
);
```

#### 2.1.4 凭据表 (credentials) - 加密存储

```sql
CREATE TABLE credentials (
    id                  TEXT PRIMARY KEY,                    -- UUID
    username            TEXT,                                -- 用户名（加密）
    password            TEXT,                                -- 密码/Token（加密）
    auth_type           TEXT NOT NULL DEFAULT 'basic',       -- basic/token/ssh
    encrypted_data      BLOB NOT NULL,                       -- AES加密后的JSON数据
    created_at          TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

## 3. API 设计

### 3.1 RESTful API 路由

```
Base URL: /api/v1
```

#### 3.1.1 仓库管理 API

**批量添加仓库**
```
POST /repos/batch
Content-Type: application/json

Request:
{
  "urls": [
    "https://github.com/user/repo1.git",
    "https://github.com/user/repo2.git"
  ]
}

Response: 200 OK
{
  "code": 0,
  "message": "success",
  "data": {
    "total": 2,
    "succeeded": [
      {
        "repo_id": 1,
        "url": "https://github.com/user/repo1.git",
        "task_id": 101
      }
    ],
    "failed": [
      {
        "url": "https://github.com/user/repo2.git",
        "error": "repository already exists"
      }
    ]
  }
}
```

**获取仓库列表**
```
GET /repos?status=ready&page=1&page_size=20

Response: 200 OK
{
  "code": 0,
  "message": "success",
  "data": {
    "total": 50,
    "page": 1,
    "page_size": 20,
    "repositories": [
      {
        "id": 1,
        "url": "https://github.com/user/repo1.git",
        "name": "repo1",
        "current_branch": "main",
        "status": "ready",
        "last_pull_at": "2025-12-31T10:00:00Z",
        "last_commit_hash": "abc123...",
        "created_at": "2025-12-30T08:00:00Z"
      }
    ]
  }
}
```

**获取仓库详情**
```
GET /repos/:id

Response: 200 OK
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "url": "https://github.com/user/repo1.git",
    "name": "repo1",
    "current_branch": "main",
    "local_path": "/workspace/cache/repo1",
    "status": "ready",
    "error_message": null,
    "last_pull_at": "2025-12-31T10:00:00Z",
    "last_commit_hash": "abc123...",
    "has_credentials": true,
    "created_at": "2025-12-30T08:00:00Z",
    "updated_at": "2025-12-31T10:00:00Z"
  }
}
```

**切换分支**
```
POST /repos/:id/switch-branch
Content-Type: application/json

Request:
{
  "branch": "develop"
}

Response: 200 OK
{
  "code": 0,
  "message": "branch switch task submitted",
  "data": {
    "task_id": 102,
    "repo_id": 1,
    "task_type": "switch",
    "status": "pending"
  }
}
```

**更新仓库（pull）**
```
POST /repos/:id/update

Response: 200 OK
{
  "code": 0,
  "message": "update task submitted",
  "data": {
    "task_id": 103,
    "repo_id": 1,
    "task_type": "pull",
    "status": "pending"
  }
}
```

**设置凭据**
```
POST /repos/:id/credentials
Content-Type: application/json

Request:
{
  "auth_type": "basic",  // basic/token
  "username": "user",
  "password": "token_or_password"
}

Response: 200 OK
{
  "code": 0,
  "message": "credentials set successfully",
  "data": {
    "credential_id": "uuid-here"
  }
}
```

**重置仓库**
```
POST /repos/:id/reset

Response: 200 OK
{
  "code": 0,
  "message": "reset task submitted",
  "data": {
    "task_id": 104,
    "repo_id": 1,
    "task_type": "reset",
    "status": "pending"
  }
}
```

**删除仓库**
```
DELETE /repos/:id

Response: 200 OK
{
  "code": 0,
  "message": "repository deleted successfully"
}
```

#### 3.1.2 统计 API

**触发统计**
```
POST /stats/calculate
Content-Type: application/json

Request:
{
  "repo_id": 1,
  "branch": "main",
  "constraint": {
    "type": "date_range",  // date_range 或 commit_limit (互斥)
    "from": "2024-01-01",  // type=date_range时必填
    "to": "2024-12-31"     // type=date_range时必填
  }
}

OR

{
  "repo_id": 1,
  "branch": "main",
  "constraint": {
    "type": "commit_limit",
    "limit": 100           // type=commit_limit时必填
  }
}

Response: 200 OK
{
  "code": 0,
  "message": "statistics task submitted",
  "data": {
    "task_id": 105,
    "repo_id": 1,
    "task_type": "stats",
    "status": "pending"
  }
}

Error: 400 Bad Request (参数互斥校验)
{
  "code": 40001,
  "message": "constraint type and parameters mismatch: date_range requires from/to, commit_limit requires limit",
  "data": null
}
```

**查询统计结果**
```
GET /stats/result?repo_id=1&branch=main&constraint_type=date_range&from=2024-01-01&to=2024-12-31

Response: 200 OK
{
  "code": 0,
  "message": "success",
  "data": {
    "cache_hit": true,
    "cached_at": "2025-12-30T15:00:00Z",
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
          "modifications": 150,  // 口径: min(additions, deletions)
          "net_additions": 800   // additions - deletions
        }
      ]
    }
  }
}

Response: 404 Not Found (未统计)
{
  "code": 40400,
  "message": "statistics not found, please submit calculation task first",
  "data": null
}
```

**查询某日期到当前的提交次数（辅助查询）**
```
GET /stats/commit-count?repo_id=1&branch=main&from=2024-01-01

Response: 200 OK
{
  "code": 0,
  "message": "success",
  "data": {
    "repo_id": 1,
    "branch": "main",
    "from": "2024-01-01",
    "to": "HEAD",
    "commit_count": 150,
    "queried_at": "2025-12-31T12:00:00Z"
  }
}
```

#### 3.1.3 任务管理 API

**获取任务列表**
```
GET /tasks?repo_id=1&status=running&page=1&page_size=20

Response: 200 OK
{
  "code": 0,
  "message": "success",
  "data": {
    "total": 3,
    "page": 1,
    "page_size": 20,
    "tasks": [
      {
        "id": 105,
        "task_type": "stats",
        "repo_id": 1,
        "status": "running",
        "parameters": "{\"branch\":\"main\",\"constraint\":{...}}",
        "started_at": "2025-12-31T12:00:00Z",
        "created_at": "2025-12-31T11:59:00Z"
      }
    ]
  }
}
```

**获取任务详情**
```
GET /tasks/:id

Response: 200 OK
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 105,
    "task_type": "stats",
    "repo_id": 1,
    "status": "completed",
    "parameters": "{\"branch\":\"main\",\"constraint\":{...}}",
    "result": "{\"cache_key\":\"...\",\"stats_cache_id\":10}",
    "error_message": null,
    "retry_count": 0,
    "started_at": "2025-12-31T12:00:00Z",
    "completed_at": "2025-12-31T12:05:00Z",
    "created_at": "2025-12-31T11:59:00Z",
    "duration_ms": 300000
  }
}
```

**取消任务**
```
POST /tasks/:id/cancel

Response: 200 OK
{
  "code": 0,
  "message": "task cancelled successfully"
}

Response: 400 Bad Request (任务已完成)
{
  "code": 40002,
  "message": "task cannot be cancelled: already completed",
  "data": null
}
```

#### 3.1.4 健康检查 API

```
GET /health

Response: 200 OK
{
  "status": "healthy",
  "timestamp": "2025-12-31T12:00:00Z",
  "components": {
    "database": "ok",
    "worker_pool": "ok",
    "git_available": true
  }
}
```

### 3.2 错误码设计

```
0      - 成功
40001  - 参数校验失败（互斥参数、缺失参数等）
40002  - 操作不允许（任务状态不正确等）
40400  - 资源未找到
40900  - 资源冲突（仓库已存在等）
50000  - 内部服务器错误
50001  - 数据库错误
50002  - Git操作失败
50003  - 任务队列错误
```

## 4. 异步任务与并发设计

### 4.1 任务类型

```go
const (
    TaskTypeClone         = "clone"          // 克隆仓库
    TaskTypePull          = "pull"           // 更新仓库
    TaskTypeSwitch        = "switch"         // 切换分支
    TaskTypeReset         = "reset"          // 重置仓库
    TaskTypeStats         = "stats"          // 统计代码
    TaskTypeCountCommits  = "count_commits"  // 计数提交
)
```

### 4.2 任务队列架构

```
┌─────────────┐
│   Submit    │
│   Task      │
└──────┬──────┘
       │
       ▼
┌─────────────────────────────────┐
│      Task Deduplication         │
│  (Check unique index in DB)     │
└──────┬──────────────────────────┘
       │
       ▼
┌─────────────────────────────────┐
│      In-Memory Queue            │
│    (Buffered Channel)           │
│   - Priority Queue              │
│   - FIFO within same priority   │
└──────┬──────────────────────────┘
       │
       ▼
┌─────────────────────────────────┐
│      Worker Pool                │
│  ┌──────────┐  ┌──────────┐    │
│  │ Worker 1 │  │ Worker 2 │... │
│  └────┬─────┘  └────┬─────┘    │
└───────┼─────────────┼───────────┘
        │             │
        ▼             ▼
   ┌────────────────────────┐
   │   Task Handlers        │
   │  - CloneHandler        │
   │  - PullHandler         │
   │  - StatsHandler        │
   │  ...                   │
   └────────────────────────┘
```

### 4.3 幂等与去重策略

1. **数据库层去重**：通过唯一索引 `idx_tasks_dedup` 实现
   - 同一 `repo_id` + `task_type` + `parameters` 的 pending/running 任务只能存在一个
   - 提交任务时先查询，若存在则返回已有任务ID

2. **任务合并**：
   - 相同参数的任务自动合并为一个
   - 返回相同的 task_id 给所有提交者

3. **幂等性保证**：
   - Clone: 检查本地目录是否已存在，存在则跳过
   - Pull: 可重复执行，git pull 本身幂等
   - Switch: 检查当前分支是否已是目标分支
   - Stats: 缓存命中则跳过计算
   - Reset: 删除目录+缓存后重新 clone

### 4.4 并发控制

```yaml
worker_pool:
  clone_workers: 2      # Clone 并发度（IO密集型，限制较小）
  pull_workers: 2       # Pull 并发度
  stats_workers: 2      # Stats 并发度（CPU密集型，根据CPU核心数配置）
  general_workers: 4    # 其他任务并发度
```

### 4.5 超时策略

```go
const (
    CloneTimeout        = 10 * time.Minute   // 克隆超时
    PullTimeout         = 5 * time.Minute    // 拉取超时
    SwitchTimeout       = 1 * time.Minute    // 切换分支超时
    StatsTimeout        = 30 * time.Minute   // 统计超时（大仓库可能很慢）
    CountCommitsTimeout = 2 * time.Minute    // 计数超时
)
```

### 4.6 重试策略

- 网络错误：最多重试 3 次，指数退避（1s, 2s, 4s）
- 认证错误：不重试，直接失败
- 超时：不重试，直接失败
- 其他错误：重试 1 次

## 5. 统计实现细节

### 5.1 Git 命令方案（优先）

#### 统计命令
```bash
# 统计所有贡献者的代码变更
git log --no-merges --numstat --pretty=format:"COMMIT:%H|AUTHOR:%an|EMAIL:%ae|DATE:%ai" \
  --since="2024-01-01" --until="2024-12-31"

# 输出格式：
COMMIT:abc123|AUTHOR:Alice|EMAIL:alice@example.com|DATE:2024-01-15 10:00:00 +0800
100     50      src/main.go
200     30      src/utils.go
COMMIT:def456|AUTHOR:Bob|EMAIL:bob@example.com|DATE:2024-01-16 11:00:00 +0800
50      10      src/test.go
```

#### 解析逻辑
```
对于每个文件变更：
  additions: 新增行数
  deletions: 删除行数
  modifications: min(additions, deletions)  # 修改的定义：被替换的行数
  net_additions: additions - deletions       # 净增加

按作者聚合：
  total_additions = sum(additions)
  total_deletions = sum(deletions)
  total_modifications = sum(modifications)
  total_net_additions = total_additions - total_deletions
```

#### 提交次数统计
```bash
# 按日期范围
git rev-list --count --since="2024-01-01" --until="2024-12-31" HEAD

# 按提交数限制
git log --oneline -n 100 | wc -l
```

### 5.2 go-git 方案（Fallback）

```go
// 伪代码
repo, _ := git.PlainOpen(repoPath)
ref, _ := repo.Head()
commits, _ := repo.Log(&git.LogOptions{From: ref.Hash()})

contributors := make(map[string]*ContributorStats)

commits.ForEach(func(c *object.Commit) error {
    if len(c.ParentHashes) > 1 {
        return nil // Skip merge commits
    }
    
    parent, _ := c.Parent(0)
    patch, _ := parent.Patch(c)
    
    stats := patch.Stats()
    for _, fileStat := range stats {
        contributors[c.Author.Email].Additions += fileStat.Addition
        contributors[c.Author.Email].Deletions += fileStat.Deletion
    }
    
    return nil
})
```

**限制说明**：
- go-git 的 diff 性能比 git 命令慢（特别是大仓库）
- 作为 fallback 方案，功能等价但性能可能差 10-100 倍
- 建议生产环境保证 git 命令可用

### 5.3 互斥参数校验

```go
func ValidateStatsConstraint(req *StatsRequest) error {
    c := req.Constraint
    
    if c.Type == "date_range" {
        if c.From == "" || c.To == "" {
            return errors.New("date_range requires both from and to")
        }
        if c.Limit != 0 {
            return errors.New("date_range cannot be used with limit")
        }
    } else if c.Type == "commit_limit" {
        if c.Limit <= 0 {
            return errors.New("commit_limit requires positive limit value")
        }
        if c.From != "" || c.To != "" {
            return errors.New("commit_limit cannot be used with date range")
        }
    } else {
        return errors.New("constraint type must be date_range or commit_limit")
    }
    
    return nil
}
```

## 6. 缓存策略

### 6.1 缓存 Key 设计

```go
func GenerateCacheKey(repoID int64, branch string, constraint Constraint, commitHash string) string {
    var constraintStr string
    if constraint.Type == "date_range" {
        constraintStr = fmt.Sprintf("dr_%s_%s", constraint.From, constraint.To)
    } else {
        constraintStr = fmt.Sprintf("cl_%d", constraint.Limit)
    }
    
    data := fmt.Sprintf("repo:%d|branch:%s|constraint:%s|commit:%s",
        repoID, branch, constraintStr, commitHash)
    
    hash := sha256.Sum256([]byte(data))
    return hex.EncodeToString(hash[:])
}
```

### 6.2 缓存失效策略

触发失效的操作：
1. **仓库更新（pull）**: 如果有新提交，则 `commit_hash` 变化，旧缓存自然失效
2. **切换分支（switch）**: 分支变化，缓存 key 不同
3. **重置仓库（reset）**: 删除该仓库的所有统计缓存

查询时：
```go
// 1. 先获取当前 HEAD 的 commit hash
currentHash := getHeadCommitHash(repo, branch)

// 2. 生成缓存 key
cacheKey := GenerateCacheKey(repoID, branch, constraint, currentHash)

// 3. 查询缓存
cache, found := queryCacheByKey(cacheKey)
if found {
    cache.HitCount++
    cache.LastHitAt = time.Now()
    return cache.LoadResult()
}

// 4. 缓存未命中，执行统计
...
```

### 6.3 存储方案

```
1. 元数据存储: 数据库 (stats_cache 表)
   - cache_key, repo_id, branch, constraint, commit_hash
   - result_path, result_size, hit_count, created_at, last_hit_at

2. 结果数据存储: 文件系统
   - Path: workspace/stats/{cache_key}.json.gz
   - Format: gzip 压缩的 JSON
   - 清理策略: LRU（最近最少使用），保留最近 30 天或最多 10GB
```

### 6.4 大小控制

```yaml
cache:
  max_total_size: 10GB           # 总缓存大小限制
  max_single_result: 100MB       # 单个结果文件大小限制
  retention_days: 30             # 保留天数
  cleanup_interval: 1h           # 清理检查间隔
```

## 7. 安全与凭据

### 7.1 凭据存储

```go
// 使用 AES-256-GCM 加密
type CredentialManager struct {
    encryptionKey []byte // 从环境变量或配置文件读取
}

func (cm *CredentialManager) EncryptCredential(cred *Credential) ([]byte, error) {
    plaintext, _ := json.Marshal(cred)
    
    block, _ := aes.NewCipher(cm.encryptionKey)
    gcm, _ := cipher.NewGCM(block)
    nonce := make([]byte, gcm.NonceSize())
    io.ReadFull(rand.Reader, nonce)
    
    ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
    return ciphertext, nil
}
```

### 7.2 日志脱敏

```go
func SanitizeURL(url string) string {
    // 移除 URL 中的用户名密码
    re := regexp.MustCompile(`(https?://)[^@]+@`)
    return re.ReplaceAllString(url, "${1}***@")
}

// 日志输出示例
log.Info("cloning repository",
    "repo_id", repoID,
    "url", SanitizeURL(repoURL),  // https://***@github.com/user/repo.git
)
```

### 7.3 Git 凭据注入

#### Git 命令方案
```go
// 方式1: 使用 credential helper
os.Setenv("GIT_ASKPASS", "/path/to/credential-helper.sh")

// 方式2: URL 重写（临时使用）
func InjectCredentials(url, username, password string) string {
    u, _ := neturl.Parse(url)
    u.User = neturl.UserPassword(username, password)
    return u.String()
}

// 执行命令时
cmd := exec.Command("git", "clone", credentialURL, localPath)
cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0") // 禁止交互式提示
```

#### go-git 方案
```go
auth := &http.BasicAuth{
    Username: username,
    Password: password,
}

_, err := git.PlainClone(localPath, false, &git.CloneOptions{
    URL:      url,
    Auth:     auth,
    Progress: os.Stdout,
})
```

### 7.4 命令注入防护

```go
// 禁止直接拼接用户输入到命令中
// ❌ 错误示例
cmd := exec.Command("sh", "-c", "git log "+userInput)

// ✅ 正确示例
cmd := exec.Command("git", "log", userInput)  // 使用参数数组

// 路径隔离
func ValidateRepoPath(path string) error {
    abs, _ := filepath.Abs(path)
    workspace, _ := filepath.Abs(config.WorkspaceDir)
    
    if !strings.HasPrefix(abs, workspace) {
        return errors.New("path outside workspace")
    }
    return nil
}
```

## 8. 可观测性

### 8.1 结构化日志

```go
// 使用 zerolog 或 logrus
log.Info().
    Int64("repo_id", repoID).
    Str("task_id", taskID).
    Str("operation", "clone").
    Int64("duration_ms", duration.Milliseconds()).
    Str("status", "success").
    Msg("repository cloned successfully")
```

### 8.2 关键指标

```go
// 使用 Prometheus 风格的指标
var (
    // 任务指标
    taskTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{Name: "tasks_total"},
        []string{"type", "status"}, // clone/pull/stats, success/failed
    )
    
    taskDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "task_duration_seconds",
            Buckets: []float64{1, 5, 10, 30, 60, 300, 600, 1800},
        },
        []string{"type"},
    )
    
    // 缓存指标
    cacheHits = prometheus.NewCounter(
        prometheus.CounterOpts{Name: "stats_cache_hits_total"},
    )
    
    cacheMisses = prometheus.NewCounter(
        prometheus.CounterOpts{Name: "stats_cache_misses_total"},
    )
    
    // Worker 指标
    workerBusy = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{Name: "worker_busy"},
        []string{"type"},  // clone/stats/general
    )
    
    queueLength = prometheus.NewGauge(
        prometheus.GaugeOpts{Name: "task_queue_length"},
    )
)

// 暴露指标端点
http.Handle("/metrics", promhttp.Handler())
```

### 8.3 错误分类

```go
const (
    ErrCategoryNetwork      = "network"       // 网络错误
    ErrCategoryAuth         = "auth"          // 认证错误
    ErrCategoryNotFound     = "not_found"     // 仓库/分支不存在
    ErrCategoryTimeout      = "timeout"       // 超时
    ErrCategoryInternal     = "internal"      // 内部错误
    ErrCategoryValidation   = "validation"    // 参数校验错误
)

func ClassifyGitError(err error) string {
    errMsg := err.Error()
    
    if strings.Contains(errMsg, "authentication") || strings.Contains(errMsg, "401") {
        return ErrCategoryAuth
    }
    if strings.Contains(errMsg, "not found") || strings.Contains(errMsg, "404") {
        return ErrCategoryNotFound
    }
    if strings.Contains(errMsg, "timeout") || strings.Contains(errMsg, "deadline exceeded") {
        return ErrCategoryTimeout
    }
    if strings.Contains(errMsg, "connection refused") || strings.Contains(errMsg, "network") {
        return ErrCategoryNetwork
    }
    
    return ErrCategoryInternal
}
```

## 9. 假设与默认配置

### 9.1 部署假设
- 单机部署优先（可扩展到多实例，需引入分布式锁/消息队列）
- 运行环境：Linux (Ubuntu 20.04+)
- Go 版本：1.21+
- Git 版本：2.30+（推荐）

### 9.2 默认配置

```yaml
server:
  host: 0.0.0.0
  port: 8080
  read_timeout: 30s
  write_timeout: 30s

workspace:
  base_dir: ./workspace
  cache_dir: ./workspace/cache      # 仓库缓存目录
  stats_dir: ./workspace/stats      # 统计结果目录

storage:
  type: sqlite                       # sqlite/postgres
  sqlite:
    path: ./workspace/data.db
  postgres:
    host: localhost
    port: 5432
    database: gitcodestatic
    user: postgres
    password: ""
    sslmode: disable

worker:
  clone_workers: 2
  pull_workers: 2
  stats_workers: 2
  general_workers: 4
  queue_buffer: 100                  # 内存队列缓冲大小

cache:
  max_total_size: 10737418240        # 10GB
  max_single_result: 104857600       # 100MB
  retention_days: 30
  cleanup_interval: 3600             # 1 hour

security:
  encryption_key: ""                 # 从环境变量 ENCRYPTION_KEY 读取

git:
  command_path: /usr/bin/git         # Git 命令路径（为空则从 PATH 查找）
  fallback_to_gogit: true            # 是否 fallback 到 go-git

log:
  level: info                        # debug/info/warn/error
  format: json                       # json/text
  output: stdout                     # stdout/file path

metrics:
  enabled: true
  path: /metrics
```

### 9.3 资源限制假设
- 仓库规模：单仓库最大 5GB
- 并发请求：50 QPS
- 同时处理的仓库数：10 个
- 单次批量添加仓库数：最多 20 个

---

## 附录：运行流程示例

### 流程1：批量添加仓库
```
1. POST /api/v1/repos/batch
   └─> RepoService.AddRepos()
       ├─> 校验 URL 格式
       ├─> 检查是否已存在（去重）
       ├─> 创建 Repository 记录（status=pending）
       ├─> 提交 Clone 任务到队列
       └─> 返回 task_id 列表

2. Worker 异步处理 Clone 任务
   └─> CloneHandler()
       ├─> 更新任务状态为 running
       ├─> 更新仓库状态为 cloning
       ├─> 调用 GitManager.Clone()
       │   ├─> 优先使用 git command
       │   └─> fallback to go-git（如果配置允许）
       ├─> 获取当前分支和 HEAD commit hash
       ├─> 更新仓库状态为 ready
       └─> 更新任务状态为 completed

3. GET /api/v1/repos/:id
   └─> 查询仓库状态（ready）
```

### 流程2：统计代码并缓存
```
1. POST /api/v1/stats/calculate
   └─> StatsService.Calculate()
       ├─> 校验参数（互斥检查）
       ├─> 检查仓库状态（必须是 ready）
       ├─> 提交 Stats 任务到队列
       └─> 返回 task_id

2. Worker 异步处理 Stats 任务
   └─> StatsHandler()
       ├─> 更新任务状态为 running
       ├─> 生成缓存 key（基于 repo/branch/constraint/commit_hash）
       ├─> 检查缓存是否存在
       │   └─> 如果存在，直接返回
       ├─> 调用 StatsCalculator.Calculate()
       │   ├─> 执行 git log --numstat
       │   ├─> 解析输出，按作者聚合
       │   └─> 计算 additions/deletions/modifications/net
       ├─> 保存结果到文件（gzip压缩）
       ├─> 创建 stats_cache 记录
       ├─> 更新任务状态为 completed
       └─> 任务结果中记录 cache_id

3. GET /api/v1/stats/result?...
   └─> StatsService.QueryResult()
       ├─> 生成缓存 key
       ├─> 查询 stats_cache 表
       ├─> 如果命中，更新 hit_count 和 last_hit_at
       ├─> 读取结果文件
       └─> 返回（cache_hit=true）
```

---

**下一步：代码实现**

接下来我将生成完整的可运行代码骨架。
