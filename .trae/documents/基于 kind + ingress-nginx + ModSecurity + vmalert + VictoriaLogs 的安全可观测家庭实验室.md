## 目标与范围
- 单节点 kind 集群（Mac 上通过 Docker），对外暴露 80/443（或 8080/8443）
- 使用外置 USB 存储承载持久化数据（指标、日志），通过 kind `extraMounts` 映射到节点
- 入口控制器为 ingress-nginx，启用 ModSecurity + OWASP CRS 基线
- 指标存储使用 VictoriaMetrics（单节点），告警使用 vmalert
- 日志存储使用 VictoriaLogs，日志采集与指标抓取统一使用 Grafana Alloy

## 架构概览
- 数据面：
  - 指标：Alloy -> VictoriaMetrics；vmalert 读取规则 -> Alertmanager/通知渠道
  - 日志：Alloy（Loki 协议）-> VictoriaLogs
- 控制面：
  - 入口：ingress-nginx + ModSecurity（按 Ingress 注解/全局策略）
  - 监控组件：kube-state-metrics、node-exporter、cadvisor
- 存储：外置 USB 驱动器在 Mac 挂载到 `/Volumes/<USB>`；通过 kind `extraMounts` 映射到节点 `/mnt/usb`；以 hostPath PV 绑定到具体组件目录

## 集群与存储设计
- kind 单节点配置：
  - `extraPortMappings`: 80/443（或 8080/8443）映射到宿主机
  - `extraMounts`: hostPath=`/Volumes/<USB>/k8s-storage` -> containerPath=`/mnt/usb`
- 持久化卷：
  - PV/PVC（hostPath 指向 `/mnt/usb/victoria-metrics` 与 `/mnt/usb/victoria-logs`）
  - 保留策略：VictoriaMetrics 设置 `-retentionPeriod=90d`；VictoriaLogs 设置相应保留与压缩参数

## 入口与 WAF
- 安装 ingress-nginx（Helm）：
  - 开启 ModSecurity：在 Ingress 资源添加注解 `nginx.ingress.kubernetes.io/enable-modsecurity: "true"`、`nginx.ingress.kubernetes.io/enable-owasp-core-rules: "true"`
  - 全局策略：通过 `controller.config`/`modsecurity-snippet` 配置基线规则与例外白名单
- Ingress TLS：
  - 自签证书或从现有证书创建 `Secret`；本地开发可优先使用 8443 端口

## 指标与告警
- VictoriaMetrics（单节点）部署，挂载 PVC
- Alloy 作为统一抓取与转发：
  - Prometheus 抓取：`kube-state-metrics`、`node-exporter`、`cadvisor`、`ingress-nginx` 指标端点
  - Remote Write：`http://victoriametrics:8428/api/v1/write`
- vmalert：
  - 规则以 ConfigMap 管理（目录 `k8s/vmalert/rules/*.yaml`）
  - 数据源指向 VictoriaMetrics；通知目标对接 Alertmanager（或直接 Webhook）
  - 初始规则：节点资源、Pod 重启、Ingress 5xx/4xx 比例、WAF 命中率与阻断数

## 日志与采集
- VictoriaLogs 部署（单节点），挂载 PVC
- Alloy DaemonSet：
  - 文件源：`/var/log/containers/*.log`（hostPath 到 Pod）
  - Loki 写入端：`http://victorialogs:9428/loki/api/v1/push`
  - 标签：`namespace`、`pod`、`container`、`app`、`env`
- Ingress-NGINX 日志：开启 access/error 日志到 stdout，随容器日志收集

## 目录与清单布局
- `k8s/kind/cluster.yaml`（单节点 + 端口映射 + USB 挂载）
- `k8s/storage/pv-vm.yaml`、`k8s/storage/pv-vl.yaml`（hostPath PV/PVC）
- `k8s/ingress-nginx/values.yaml`（开启 ModSecurity/CRS、必要 controller 配置）
- `k8s/ingress/app-ingress.yaml`（示例 Ingress + WAF 注解）
- `k8s/victoria-metrics/values.yaml` 或 `deployment.yaml`（单节点）
- `k8s/victoria-logs/deployment.yaml`（单节点）
- `k8s/alloy/daemonset.yaml` + `configmap.yaml`（同时配置 metrics + logs）
- `k8s/monitoring/kube-state-metrics.yaml`、`node-exporter.yaml`、`cadvisor.yaml`
- `k8s/vmalert/deployment.yaml`、`k8s/vmalert/rules/*.yaml`
- 可选：`k8s/alertmanager/*`

## 实施步骤
1. 准备 USB：在 Mac 挂载到 `/Volumes/<USB>/k8s-storage` 并确保可读写
2. 创建 kind 集群：应用 `cluster.yaml`（含 `extraMounts` 与端口映射）
3. 创建 PV/PVC：应用 `storage/*` 清单，实际在节点目录建立数据子目录
4. 安装 ingress-nginx：应用 `values.yaml`；验证 controller Ready
5. 部署 VictoriaMetrics、VictoriaLogs：绑定 PVC，设置资源与保留参数
6. 部署 Alloy：加载统一配置，验证到 VM/Logs 的写入成功
7. 部署 kube-state-metrics、node-exporter、cadvisor：验证指标被抓取
8. 部署 vmalert：加载规则并连接 Alertmanager/通知渠道
9. 创建示例 Ingress：验证 WAF 拦截与放行策略

## 校验清单
- 端口暴露：本机能访问 `http://localhost:8080`/`https://localhost:8443`
- 存储：USB 路径内生成 `victoria-metrics`、`victoria-logs` 数据文件
- 指标：VictoriaMetrics `up` 指标为 1；Alloy remote_write 队列无积压
- 日志：在 VictoriaLogs 中能按 `namespace/pod` 查询容器日志
- WAF：模拟常见攻击载荷（如 SQLi）被拦截；正常业务请求放行
- 告警：手动制造阈值触发，vmalert 能产生并下发到通知渠道

## 说明与取舍
- 遵循“第三方组件用 Docker”原则：kind/Helm 组件均运行在 Docker 容器内；USB 通过 `extraMounts` 暴露到节点，再用 hostPath PV
- 合并采集代理：使用 Alloy 统一替代 Promtail + Prometheus Agent，减少组件数量
- 本地端口占用：如宿主机 80/443 不便，可用 8080/8443

## 确认后将执行
- 创建上述目录与清单文件，按步骤应用到集群
- 提供关键 `values.yaml` 与 Alloy 配置示例，运行验证与自动 git commit
- 根据实际 USB 路径与端口选择微调配置