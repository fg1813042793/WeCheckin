# Docker 运维可观测性优化设计

## 背景

Docker 部署已经具备 `.env.example`、健康依赖、备份和恢复脚本。剩余生产部署优化中，优先补齐低风险的日志保留和请求追踪能力，避免容器日志无限增长，并让排查单次请求时能从 Nginx 访问日志关联到后端请求。

## 目标

1. Docker Compose 为 MySQL、Redis、backend 和 Nginx 配置日志轮转。
2. `.env.example` 暴露日志大小和保留文件数。
3. Nginx 访问日志记录 `request_id`，并向后端透传 `X-Request-ID`。
4. 提供 `docker-logs.sh` 辅助脚本，便于查看服务日志。
5. 将上述能力纳入部署静态检查和中文部署文档。

## 非目标

- 不引入 Prometheus、Grafana 或第三方日志平台。
- 不改变后端日志框架。
- 不自动申请 HTTPS 证书。
- 不改变当前容器网络和端口映射。

## 方案

Compose 顶部新增 `x-logging` 锚点，统一使用 Docker `json-file` 日志驱动和 `max-size`、`max-file` 轮转配置。四个服务统一复用该锚点。

Nginx `log_format` 增加 `request_id=$request_id`，所有代理到后端的 location 增加 `proxy_set_header X-Request-ID $request_id`，server 级别增加 `add_header X-Request-ID $request_id always`。

新增 `backend/scripts/docker-logs.sh`，默认跟踪 backend 日志，也支持 `bash scripts/docker-logs.sh nginx` 和 `TAIL=500` 形式。

## 验证

- `node scripts/check-deploy-config.mjs`
- `docker compose -f backend/docker-compose.yml config`
- `bash scripts/check.sh`
