# 家庭实验室（Kind + Ingress-NGINX + ModSecurity + VictoriaMetrics/vmalert + VictoriaLogs + Alloy）完整指引

## 架构与目标
- 单节点 `kind` 集群，宿主机端口映射：`8080`→HTTP、`8443`→HTTPS（对应 NodePort `30080/30443`）
- 入口：`ingress-nginx` 控制器，启用 `ModSecurity` 与 `OWASP CRS` 基线规则
- 指标：`VictoriaMetrics` 单节点存储，`vmalert` 规则告警
- 日志：`VictoriaLogs` 单节点存储，`Grafana Alloy` 统一采集 Kubernetes 容器日志并写入 `VictoriaLogs`
- 存储：外置 USB 磁盘通过 `kind.extraMounts` 映射到节点 `/mnt/usb`，各组件使用 `hostPath` PV/PVC 持久化

## 目录结构
- `k8s/kind/cluster.yaml`：kind 单节点配置（端口映射、USB 挂载）
- `k8s/storage/pv-vm.yaml`、`k8s/storage/pv-vl.yaml`：VictoriaMetrics / VictoriaLogs 的 `hostPath` PV/PVC
- `k8s/monitoring/namespace.yaml`：监控命名空间
- `k8s/ingress-nginx/values.yaml`：ingress-nginx Helm values（开启 WAF、NodePort、Metrics）
- `k8s/ingress/app.yaml`：示例 echo 应用与 Ingress（含 WAF 注解）
- `k8s/victoria-metrics/deployment.yaml`：VictoriaMetrics Deployment/Service
- `k8s/victoria-logs/deployment.yaml`：VictoriaLogs Deployment/Service
- `k8s/alloy/configmap.yaml`、`k8s/alloy/daemonset.yaml`：Alloy 配置与 DaemonSet
- `k8s/vmalert/deployment.yaml`：vmalert Deployment 与示例规则 ConfigMap

## 前置准备
- MacOS 已安装：`Docker Desktop`、`kind`、`kubectl`、`helm`
- 外置 USB 设备已挂载，例如：`/Volumes/USB_DISK/k8s-storage`
- 如需直接使用仓库内目录持久化（示例），已创建 `usb-data`：
  - `mkdir -p /Users/zengshenglong/Code/HandBooks/homeLab/usb-data`

## 配置 USB 挂载
- 在 `k8s/kind/cluster.yaml` 中，修改 `extraMounts.hostPath` 为你的 USB 路径（推荐）：
  - 示例：`hostPath: /Volumes/USB_DISK/k8s-storage`
- 保持容器内映射路径不变：`containerPath: /mnt/usb`
- PV 使用的 `hostPath` 已固定为容器内路径子目录：
  - `pv-vm.yaml`：`/mnt/usb/victoria-metrics`
  - `pv-vl.yaml`：`/mnt/usb/victoria-logs`

## 创建集群
```bash
kind create cluster --name homelab --config k8s/kind/cluster.yaml
kubectl cluster-info --context kind-homelab
```

## 安装 ingress-nginx 控制器
```bash
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm repo update
helm upgrade --install ingress-nginx ingress-nginx/ingress-nginx \
  -n ingress-nginx --create-namespace \
  -f k8s/ingress-nginx/values.yaml
```
- `values.yaml` 已设置：
  - `service.type: NodePort`，`nodePorts.http: 30080`、`nodePorts.https: 30443`
  - `enableModsecurity: true`、`enableOWASPCoreRules: true`
  - `metrics.enabled: true`
- 控制器就绪后再创建 Ingress（Admission Webhook 启动前创建会报错）

## 创建监控命名空间与存储 PV/PVC
```bash
kubectl apply -f k8s/monitoring/namespace.yaml
kubectl apply -f k8s/storage/pv-vm.yaml -n monitoring
kubectl apply -f k8s/storage/pv-vl.yaml -n monitoring
```

## 部署 VictoriaMetrics 与 VictoriaLogs
```bash
kubectl apply -f k8s/victoria-metrics/deployment.yaml
kubectl apply -f k8s/victoria-logs/deployment.yaml
```
- VictoriaMetrics 默认保留：`-retentionPeriod=90d`
- VictoriaLogs 默认端口：`9428`，数据目录：`/victoria-logs-data`

## 部署 Alloy（统一日志采集 + 指标抓取）
```bash
kubectl apply -f k8s/alloy/configmap.yaml
kubectl apply -f k8s/alloy/daemonset.yaml
```
- 日志收集：`/var/log/containers/*.log` → `VictoriaLogs` Loki Push API
  - Alloy 写入端点：`http://victoria-logs.monitoring.svc:9428/insert/loki/api/v1/push`
  - 参考：VictoriaLogs Loki 接入路径 `/insert/loki/api/v1/push`（官方文档）
- 指标抓取：`ingress-nginx-controller` metrics `:10254`
- Remote Write 到 VictoriaMetrics：`http://victoria-metrics.monitoring.svc:8428/api/v1/write`

## 部署 vmalert 与规则
```bash
kubectl apply -f k8s/vmalert/deployment.yaml
```
- 示例规则：`IngressHigh5xxRatio`（5xx 比例 > 5% 触发告警）
- 可扩展：节点资源、Pod 重启、WAF 命中率、Ingress 4xx/5xx 比例等

## 部署示例应用与 Ingress（含 WAF）
```bash
kubectl apply -f k8s/ingress/app.yaml
```
- Ingress 注解：
  - `nginx.ingress.kubernetes.io/enable-modsecurity: "true"`
  - `nginx.ingress.kubernetes.io/enable-owasp-core-rules: "true"`
  - `nginx.ingress.kubernetes.io/modsecurity-snippet: |\n  SecRuleEngine On`

## 访问与验证
- 控制器服务：
  ```bash
  kubectl get svc -n ingress-nginx
  # 期望：ingress-nginx-controller 80:30080/TCP, 443:30443/TCP
  ```
- 访问 echo 应用：
  ```bash
  # HTTP（使用 NodePort 经宿主机端口转发）
  curl -H 'Host: echo.local' http://localhost:8080/
  # HTTPS（自签或未配置证书时可忽略证书验证）
  curl -k -H 'Host: echo.local' https://localhost:8443/
  ```
- WAF 验证（模拟 SQLi）：
  ```bash
  curl -i -H 'Host: echo.local' 'http://localhost:8080/?q=1%20OR%201=1'
  # 预期：返回 403 被阻断
  ```
  - SQLi ' OR '1'='1'
  ```bash
  curl -i -H 'Host: echo.local' 'http://localhost:8080/?q=%27%20OR%20%271%27=%271%27'
  ```
  - SQLi UNION SELECT ： 
  ```bash
  curl -i -H 'Host: echo.local' 'http://localhost:8080/?q=UNION%20SELECT%201,2,3'
  ```
  - XSS ： 
  ```bash
  curl -i -H 'Host: echo.local' 'http://localhost:8080/?q=<script>alert(1)</script>'
  ``` 

- 指标验证：
  ```bash
  # VictoriaMetrics /metrics 与写入健康
  kubectl port-forward -n monitoring svc/victoria-metrics 8428:8428 &
  curl 'http://localhost:8428/api/v1/query?query=up'
  ```
- 日志验证（VictoriaLogs 查询）：
  ```bash
  kubectl port-forward -n monitoring svc/victoria-logs 9428:9428 &
  # 简单查询全部日志（LogSQL）
  curl -X POST 'http://localhost:9428/select/logsql/query' -d 'query=_msg:*'
  ```

## 常见问题排查
- Ingress Admission Webhook 拒绝：
  - 原因：`ingress-nginx-controller-admission` 尚未就绪
  - 处理：等待控制器 Pod `READY 1/1` 再创建 Ingress
- NodePort 未映射到宿主机端口：
  - 检查 `k8s/kind/cluster.yaml` 中 `containerPort: 30080/30443` → `hostPort: 8080/8443`
- USB 路径权限问题：
  - 确认 MacOS 上 USB 路径可读写；如为外置 NTFS/exFAT，需确保 Docker 进程具有读写权限
  - PV 使用 `hostPath` 指向容器内 `/mnt/usb/...`，请确保 `extraMounts` 正确映射
- Alloy 无法写入 VictoriaLogs：
  - 检查 Alloy 容器日志与 VictoriaLogs `/metrics`：`vl_http_errors_total{path="/insert/loki/api/v1/push"}`
- WAF 误拦截：
  - 在具体 Ingress 上使用例外注解或 `modsecurity-snippet` 定制白名单

## 清理
```bash
helm uninstall ingress-nginx -n ingress-nginx
kubectl delete -f k8s/ingress/app.yaml
kubectl delete -f k8s/vmalert/deployment.yaml
kubectl delete -f k8s/alloy/daemonset.yaml
kubectl delete -f k8s/alloy/configmap.yaml
kubectl delete -f k8s/victoria-logs/deployment.yaml
kubectl delete -f k8s/victoria-metrics/deployment.yaml
kubectl delete -f k8s/storage/pv-vl.yaml
kubectl delete -f k8s/storage/pv-vm.yaml
kubectl delete -f k8s/monitoring/namespace.yaml
kind delete cluster --name homelab
```

## 说明
- 端口占用：如宿主机 80/443 已占用，保留 `8080/8443` 映射即可
- kind 节点镜像：当前使用 `registry.cn-hangzhou.aliyuncs.com/seam/node:v1.33.1`，可按需调整为官方镜像
- Alloy 配置中 Loki Push 路径参考官方文档（VictoriaLogs 支持 Loki JSON API 的 `/insert/loki/api/v1/push`）