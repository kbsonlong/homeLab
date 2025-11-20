# WAF管理后台

基于Go后端和React前端的Web应用防火墙(WAF)管理后台，专为家庭实验室环境设计。

## 功能特性

### 核心功能
- **模式管理**: 全局/按域名的WAF模式切换 (On/DetectionOnly/Off)
- **规则管理**: CRS开关、自定义规则片段管理
- **例外管理**: 按域名/路径/IP/CIDR的放行策略
- **审计与报表**: 拦截趋势、攻击类型分布统计
- **实时监控**: Ingress指标、vmalert告警状态
- **日志查询**: VictoriaLogs集成，支持LogSQL查询
- **变更审计**: 配置修改记录与回滚功能

### 技术架构
- **后端**: Go + Gin框架 + Kubernetes客户端
- **前端**: React + TypeScript + Tailwind CSS
- **监控**: VictoriaMetrics + vmalert + VictoriaLogs
- **部署**: Docker + Kubernetes

## 快速开始

### 开发环境

1. 启动后端服务:
```bash
cd backend
go mod download
go run cmd/main.go
```

2. 启动前端服务:
```bash
cd frontend
npm install
npm run dev
```

### Docker部署

1. 使用Docker Compose:
```bash
docker-compose up -d
```

2. 访问应用:
- 前端: http://localhost:3000
- 后端API: http://localhost:8080

### Kubernetes部署

1. 应用Kubernetes配置:
```bash
kubectl apply -f deployments/kubernetes.yaml
```

2. 配置Ingress:
```bash
# 修改deployments/kubernetes.yaml中的host配置
kubectl apply -f deployments/kubernetes.yaml
```

## API文档

### WAF管理API
- `GET /api/waf/status` - 获取WAF状态
- `POST /api/waf/mode` - 更新WAF模式
- `POST /api/waf/exceptions` - 更新例外规则
- `POST /api/waf/rules` - 更新自定义规则
- `POST /api/waf/apply` - 应用配置

### 监控API
- `GET /api/metrics/summary` - 获取指标汇总
- `POST /api/logs/search` - 搜索日志
- `GET /api/logs/filters` - 获取日志过滤器

## 配置说明

### 后端配置 (config/config.yaml)
```yaml
server:
  port: 8080
  host: "0.0.0.0"
  mode: "development"

kubernetes:
  namespace: "monitoring"

metrics:
  victoria_metrics_url: "http://victoria-metrics:8428"
  vmalert_url: "http://vmalert:8880"

logs:
  victoria_logs_url: "http://victoria-logs:9428"

security:
  enable_auth: true
  username: "admin"
  password: "admin123"
```

### 告警规则
查看 `deployments/alerts/waf-alerts.yaml` 获取预定义的告警规则。

## 安全考虑

1. **RBAC配置**: 使用最小权限原则配置Kubernetes RBAC
2. **认证**: 支持基础认证，生产环境建议使用OIDC
3. **网络策略**: 限制Pod间通信
4. **审计日志**: 记录所有配置变更

## 监控指标

应用暴露以下Prometheus指标:
- `waf_admin_requests_total` - 总请求数
- `waf_admin_request_duration_seconds` - 请求延迟
- `waf_admin_policy_changes_total` - 策略变更次数

## 故障排除

### 常见问题

1. **Kubernetes连接失败**
   - 检查ServiceAccount权限
   - 确认RBAC配置正确

2. **VictoriaMetrics连接失败**
   - 验证服务地址和端口
   - 检查网络连通性

3. **WAF配置未生效**
   - 检查ingress-nginx控制器配置
   - 确认ConfigMap更新成功

## 开发指南

### 项目结构
```
waf-admin/
├── backend/          # Go后端
│   ├── cmd/         # 主程序
│   ├── internal/    # 内部包
│   └── pkg/         # 公共包
├── frontend/        # React前端
│   ├── src/         # 源代码
│   └── public/      # 静态资源
└── deployments/     # K8s配置
```

### 贡献指南
1. Fork项目
2. 创建特性分支
3. 提交代码变更
4. 创建Pull Request

## 许可证

MIT License