# Swagger 和 Web UI 使用指南

## Swagger API 文档

项目已集成 Swagger 2.0 API 文档，提供交互式的 API 测试和文档浏览功能。

### 访问 Swagger UI

启动服务器后，访问：

```
http://localhost:8080/swagger/index.html
```

### Swagger 功能

1. **API 端点浏览**：查看所有可用的 API 端点
2. **参数说明**：每个端点的详细参数说明
3. **在线测试**：直接在浏览器中测试 API
4. **响应示例**：查看 API 响应的数据结构

### 重新生成文档

当修改 API 注释后，需要重新生成 Swagger 文档：

```bash
# 安装 swag 工具
go install github.com/swaggo/swag/cmd/swag@latest

# 生成文档
swag init -g cmd/server/main.go -o docs
```

## Web UI 前端界面

项目提供了基于 Vue 3 和 Element Plus 的 Web 管理界面，支持完全离线部署。

### 访问 Web UI

启动服务器后，访问：

```
http://localhost:8080/
```

### 功能模块

#### 1. 仓库管理

- **批量添加仓库**：一次性添加多个 Git 仓库
- **查看仓库列表**：显示所有仓库及其状态
- **切换分支**：切换仓库到不同分支
- **更新仓库**：拉取最新代码
- **重置仓库**：重置到干净状态
- **删除仓库**：从系统中删除仓库

#### 2. 统计管理

- **触发统计计算**：
  - 选择仓库和分支
  - 支持两种约束类型：
    - **日期范围**：统计指定日期区间的提交
    - **提交数限制**：统计最近 N 次提交
  
- **查询统计结果**：
  - 总提交数、贡献者数
  - 代码增加/删除行数
  - 统计周期
  - 贡献者详细列表（提交数、代码行数、首次/最后提交时间）

#### 3. 任务监控

通过仓库状态实时监控异步任务执行情况。

#### 4. API 文档

快速访问 Swagger API 文档的入口。

### 离线部署

Web UI 的所有外部资源（Vue、Element Plus、Axios）都已下载到本地：

```
web/
├── index.html          # 主页面
├── static/
│   ├── app.js         # 应用逻辑
│   └── lib/           # 第三方库
│       ├── vue.global.prod.js
│       ├── element-plus.min.js
│       ├── element-plus.css
│       └── axios.min.js
```

无需互联网连接即可正常使用所有功能。

### 配置

在 `configs/config.yaml` 中配置 Web UI：

```yaml
web:
  dir: ./web          # Web 文件目录
  enabled: true       # 是否启用 Web UI
```

设置 `enabled: false` 可以禁用 Web UI，仅保留 API 服务。

## 开发建议

### 添加新的 Swagger 注释

在 API handler 函数上方添加注释：

```go
// MethodName 方法描述
// @Summary 简短摘要
// @Description 详细描述
// @Tags 标签名
// @Accept json
// @Produce json
// @Param paramName path/query/body type true "参数说明"
// @Success 200 {object} Response{data=DataType}
// @Failure 400 {object} Response
// @Router /path [method]
func (h *Handler) MethodName(w http.ResponseWriter, r *http.Request) {
    // ...
}
```

### 扩展 Web UI

修改 `web/static/app.js` 添加新功能：

```javascript
// 在 methods 中添加新方法
methods: {
    async newFunction() {
        try {
            const response = await axios.get(`${API_BASE}/new-endpoint`);
            // 处理响应
        } catch (error) {
            ElMessage.error('请求失败: ' + error.message);
        }
    }
}
```

在 `web/index.html` 中添加新的 UI 组件：

```html
<el-tab-pane label="新功能" name="newFeature">
    <el-card>
        <!-- 添加 Element Plus 组件 -->
    </el-card>
</el-tab-pane>
```

## 故障排查

### Swagger 无法访问

1. 检查 `docs/` 目录是否存在生成的文件
2. 确认 `cmd/server/main.go` 中导入了 docs 包：
   ```go
   _ "github.com/gitcodestatic/gitcodestatic/docs"
   ```
3. 重新生成文档：`swag init -g cmd/server/main.go -o docs`

### Web UI 无法加载

1. 检查 `web/` 目录是否存在
2. 确认 `config.yaml` 中 `web.enabled` 为 `true`
3. 检查浏览器控制台是否有 JavaScript 错误
4. 确认所有静态资源文件都已下载

### API 请求失败

1. 检查浏览器控制台的网络请求
2. 确认 API 端点路径正确（/api/v1/...）
3. 查看服务器日志获取详细错误信息
4. 使用 Swagger UI 测试 API 是否正常工作

## 最佳实践

1. **文档同步**：修改 API 后立即更新 Swagger 注释并重新生成文档
2. **错误处理**：在前端添加适当的错误提示，提升用户体验
3. **加载状态**：使用 Element Plus 的 `v-loading` 指令显示加载状态
4. **确认操作**：对删除、重置等危险操作添加二次确认
5. **响应式布局**：使用 Element Plus 的栅格系统确保各种屏幕尺寸下都能正常显示

## 资源链接

- [Swagger 文档规范](https://swagger.io/specification/v2/)
- [swaggo/swag](https://github.com/swaggo/swag)
- [Vue 3 文档](https://cn.vuejs.org/)
- [Element Plus 文档](https://element-plus.org/)
