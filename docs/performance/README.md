# 性能排查说明

本文档用于记录 WeCheckin 的接口性能排查方法。性能优化统一从可观测开始：先跑 `npm run check:performance` 获取接口耗时，再结合后端 `[SLOW_REQUEST]` 日志和 MySQL slow query log 定位 SQL。

## 本地基线

启动后端服务后，在项目根目录执行：

```bash
npm run check:performance
```

默认模式只输出告警，不会阻断命令。发布前或重点验证时使用 strict 模式：

```bash
WECHECKIN_PERF_BASE_URL=http://127.0.0.1:8083 \
WECHECKIN_ADMIN_TOKEN='your-admin-token' \
WECHECKIN_USER_TOKEN='your-user-token' \
WECHECKIN_DINGTALK_TOKEN='your-dingtalk-token' \
WECHECKIN_PERF_STRICT=1 \
npm run check:performance
```

脚本会输出 `min / avg / p95 / max`，慢接口记录到 `docs/performance/api-baseline.md`，便于对比优化前后变化。

注意：当前项目前端请求头传递的是原始 token，性能脚本会自动去掉手工输入的 `Bearer ` 前缀后再请求接口。脚本同时会解析响应 JSON，只有业务 `code = 0` 才计入成功；未登录、权限不足等业务错误在 strict 模式下会阻断。

## MySQL slow query

本地或测试环境可以临时开启 MySQL slow query log：

```sql
SET GLOBAL slow_query_log = 'ON';
SET GLOBAL long_query_time = 0.2;
SET GLOBAL log_queries_not_using_indexes = 'ON';
```

排查完成后建议恢复阈值或关闭未使用索引日志，避免日志过大。生产环境开启前需要确认日志路径、磁盘空间和采集策略。

## EXPLAIN 检查

慢 SQL 需要执行 `EXPLAIN`，重点查看：

- `type`：优先达到 `ref`、`range` 或更好，避免大表 `ALL`。
- `key`：确认命中预期索引。
- `rows`：估算扫描行数是否和分页/筛选条件匹配。
- `Extra`：关注 `Using filesort`、`Using temporary`、`Using where`。

索引说明和回滚策略记录在 `docs/performance/mysql-indexes.md`。

## 本地质量门禁

常规检查：

```bash
bash scripts/check.sh
```

带性能基线：

```bash
CHECK_PERFORMANCE=1 bash scripts/check.sh
```

如果接口不可访问但没有开启 strict，性能脚本只会输出错误和告警。需要阻断发布时，请同时设置 `WECHECKIN_PERF_STRICT=1`。
