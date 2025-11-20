## 目标
- 为家庭实验室提供可视化与可控的WAF管理后台：开关模式、规则与白名单管理、审计与报表、联动告警。
- 基于现有集成：ingress-nginx + ModSecurity（CRS）、VictoriaMetrics/vmalert、VictoriaLogs、Alloy。

## 核心功能
- 模式管理：全局/按域名的 `SecRuleEngine`（On/DetectionOnly/Off）。
- 规则管理：全局CRS启用、自定义规则片段（含白名单/排除）、限制允许的指令以降低风险。
- 例外管理：按域名/路径/IP/CIDR 的放行策略；常见误报模板（搜索、富文本、支付回调等）。
- 审计与报表：
  - 拦截趋势、拦截原因Top、攻击类型分布（SQLi/XSS/RCE/PathTraversal）。
  - 受影响域名/路径/客户端IP统计。
- 实时监控：
  - Ingress指标（请求总量、5xx比率、命中WAF的403占比）。
  - vmalert告警状态与最近告警。
- 日志查询：按 `namespace/pod/container/host`、时间范围、关键词（LogSQL）查询；保存查询。
- 变更审计：配置修改记录（时间、操作者、资源差异），支持回滚。
- 多租视角（简版）：按域名分组展示与授权（家庭实验室可选）。

## 架构设计
- 前端：React/Vite（或SvelteKit）单页应用，组件：Dashboard、Policies、Rules、Exceptions、Logs、Alerts、Settings。
- 后端API：Go（Gin/Fiber）或 Node.js（Fastify）轻服务，运行在K8s；
  - K8s客户端：读写 ConfigMap、Ingress、Deployment；
  - 观测连接：
    - VictoriaMetrics API：`/api/v1/query`、`/api/v1/query_range`；
    - VictoriaLogs API：`/select/logsql/query`；
    - vmalert：服务 `:8880` 提供规则/健康数据。
- 存储：
  - 策略模型存储在 K8s ConfigMap `waf-policies`（JSON/YAML），后端负责渲染至 ingress-nginx 可用配置；
  - 变更审计记录存储在 `ConfigMap` 或本地文件卷（家庭实验室场景即可）。
- 运行：以 `ClusterIP` + Ingress 暴露后台；绑定 ServiceAccount 与最小化 RBAC。

## 数据源与整合
- 指标：
  - `nginx_ingress_controller_requests`（按 `status`/`host` 聚合）
  - `vl_http_errors_total{path="/insert/loki/api/v1/push"}`（校验日志接入）
  - vmalert评估结果（通过其API或在VictoriaMetrics中查询规则序列）。
- 日志：
  - 通过 Alloy 将 `/var/log/containers/*.log` 与 `/var/log/pods/*/*/*.log` 推送至 VictoriaLogs；
  - ModSecurity审计：在控制器层设置 `SecAuditEngine On` 并输出至stdout，解析审计条目（message/tag/ruleid）。

## 策略模型
- Policy（域名/应用维度）：
  - `mode`: On | DetectionOnly | Off
  - `enable_crs`: bool
  - `exceptions`: {
    - `paths`: [prefix/regex]
    - `methods`: [GET/POST/...]
    - `ip_allow`: [CIDR]
    - `headers_allow`: [name/value规则]
  }
  - `custom_rules`: 受限指令白名单（如 `SecRule`, `SecAction`），不允许 `Include`/文件写入类指令。

## 后端API草案
- `GET /api/waf/status`：返回全局与各域策略、控制器有效配置。
- `POST /api/waf/mode`：入参 `{host, mode}`；更新策略并下发。
- `POST /api/waf/exceptions`：追加/删除例外；支持测试模式（仅写策略，不下发）。
- `POST /api/waf/rules`：新增/更新规则片段，服务端进行指令校验。
- `POST /api/waf/apply`：将策略渲染与发布：
  - 选项A（推荐/细粒度）：开启 `allow-snippet-annotations`，为目标 Ingress 打注解 `enable-modsecurity/enable-owasp-core-rules` + `modsecurity-snippet`；
  - 选项B（安全/粗粒度）：仅通过控制器 ConfigMap 的 `modsecurity-snippet` 作全局策略渲染（基于host匹配），不启用注解。 
- `GET /api/metrics/summary`：聚合 ingress 指标（QPS、4xx/5xx、403/WAF占比）。
- `GET /api/logs/search`：透传 VictoriaLogs 查询（受限语法/时间范围）；提供多个预置查询模板。
- `GET /api/alerts`：列举 vmalert 规则与告警；
- `POST /api/alerts/rules`：管理 vmalert 规则（更新 ConfigMap `vmalert-rules`）。
- `GET /api/audit`：配置变更历史；

## UI页面
- Dashboard：今日拦截数、同比趋势、5xx/403占比、Top域名/路径/规则ID。
- Policies（按域名）：模式开关、CRS开关、例外编辑、规则片段预览/差异、发布按钮；测试运行（DetectionOnly）。
- Logs：快速筛选（时间、host、status、ruleid/tag）、详情弹窗（请求头、payload截断、原因标签）。
- Alerts：vmalert规则列表、开关与阈值、最近告警；一键创建Ingress 5xx/WAF命中率告警。
- Settings：控制器级别设置（允许注解片段开关）、审计开关、后端连接配置（VM/VL/vmalert）。

## 下发机制
- 变更写入 `waf-policies` ConfigMap；后端根据选项选择渲染路径：
  - 选项A：Patch Ingress 注解（需要 `controller.config.allow-snippet-annotations: "true"`），并重载控制器；
  - 选项B：Patch 控制器 ConfigMap 的 `modsecurity-snippet`（按host构建 `SecRule` 条件），滚动控制器。
- 所有下发动作记录审计（旧值/新值/diff）。

## 安全与RBAC
- ServiceAccount（命名空间：`monitoring` 或独立 `ops`）：
  - 允许 `get/list/patch`：`configmaps`（控制器、waf-policies、vmalert-rules）、`ingresses`、`deployments`（controller滚动）；
- 后端限制规则指令白名单；对外仅暴露内部网络；基础鉴权（Basic或OIDC简版）。

## 告警联动
- 规则模板：
  - Ingress 5xx比例、WAF 403占比（WAF阻断监控）、单IP短期命中数爆发。
- 后端提供规则生成器，写入 `vmalert-rules`，并支持选择通知器（目前为 `blackhole`，可引导用户安装 Alertmanager）。

## 指标与日志查询示例
- 指标（VM）：
  - `sum(rate(nginx_ingress_controller_requests{status=~"4..|5.."}[5m])) by (status)`
  - `sum(rate(nginx_ingress_controller_requests{status="403"}[5m])) / sum(rate(nginx_ingress_controller_requests[5m]))`
- 日志（VL LogSQL）：
  - `query=_msg:* AND status:403` （拦截请求）
  - 按 `host`/`path`/`remote_addr` 聚合，提供Top统计。

## 实施步骤
1. 后端与前端工程脚手架；定义策略模型与API；
2. K8s RBAC与部署清单（ServiceAccount、Deployment、Service、Ingress）；
3. 控制器配置接入：读取/更新 `ingress-nginx` 控制器 ConfigMap；可选打开 `allow-snippet-annotations`；
4. 策略渲染器：支持选项A/B的下发路径；
5. 指标/日志客户端与图表；
6. vmalert规则管理（读取/写入ConfigMap + 调用vmalert API查看状态）；
7. 审计与回滚；
8. 集成测试：
   - 制造SQLi/XSS载荷，验证On/DetectionOnly切换；
   - 观察指标与日志曲线；
   - 触发vmalert告警并在后台展示。

## 验证与上线
- 在当前kind集群验证：
  - 切换模式→立刻验证403阻断；
  - 新增例外→目标请求放行；
  - 报表与查询→看到对应数据；
- 文档：更新 `docs/homelab-guide.md` 增补“WAF后台使用”章节：操作路径、注意事项、风险提示。

## 取舍与提示
- 家庭实验室推荐先用选项B（全局渲染）保障安全与简洁；如需更细粒度再开启注解片段（选项A）。
- 所有自定义规则做语法与指令白名单校验，避免注入危险指令；提供规则模板库（常见白名单/例外/自定义检测）。
