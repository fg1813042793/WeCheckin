# API 性能基线

本文档记录 `npm run check:performance` 的关键接口性能结果。每轮优化后追加一条记录，便于确认性能变化不是主观感受。

## 记录模板

```text
日期：2026-07-28
环境：本地 MySQL / Redis / 后端 8083
命令：npm run check:performance
备注：未开启 WECHECKIN_PERF_STRICT，接口不可访问时只告警

接口                                 min     avg     p95     max
/api/v2/admin/users                 -       -       -       -
/api/v2/admin/managers              -       -       -       -
/api/v2/dingtalk/h5/bootstrap       -       -       -       -
/api/v2/dingtalk/h5/workbench       -       -       -       -
/api/v2/dingtalk/h5/reviews         -       -       -       -
```

## 当前基线

当前已在本地 MySQL / Redis / 临时后端 18083 上完成 strict 基线验证。后续每轮优化后继续追加记录：

```bash
npm run check:performance
```

发布前建议执行：

```bash
WECHECKIN_PERF_STRICT=1 npm run check:performance
```

如发现接口明显变慢，先查看 MySQL slow query log，再对慢 SQL 执行 `EXPLAIN`。索引命中和回滚策略记录在 `docs/performance/mysql-indexes.md`。

## 2026-07-28 strict 真实接口基线

```text
环境：本地 MySQL / Redis / 临时后端 18083
命令：自动登录测试账号后执行 WECHECKIN_PERF_STRICT=1 npm run check:performance
结果：通过，所有计划内接口 P95 均低于阈值

接口                              阈值      min       avg       p95       max
后台用户列表                      250ms     186.5ms   193.7ms   208.7ms   208.7ms
后台管理员列表                    250ms     136.4ms   160.6ms   181.0ms   181.0ms
后台角色列表                      250ms     74.6ms    84.2ms    90.2ms    90.2ms
钉钉 H5 启动信息                  150ms     110.4ms   117.6ms   122.6ms   122.6ms
钉钉 H5 工作台                    250ms     122.5ms   131.2ms   144.7ms   144.7ms
钉钉 H5 绩效列表                  300ms     139.5ms   148.1ms   157.2ms   157.2ms
通知列表                          300ms     131.8ms   141.0ms   153.9ms   153.9ms
打卡任务列表                      300ms     125.9ms   150.9ms   243.9ms   243.9ms
赛事活动列表                      300ms     12.0ms    16.6ms    20.5ms    20.5ms
问卷列表                          300ms     20.3ms    29.1ms    46.8ms    46.8ms
考试列表                          300ms     16.9ms    19.4ms    23.2ms    23.2ms
```

本轮额外修正：性能脚本会解析响应 JSON，只有业务 `code = 0` 才计入成功；未登录、权限不足等 `code != 0` 响应会在 strict 模式下阻断。

角色列表优化对比：在旧 8083 进程上，`/api/v2/admin/roles?page=1&pageSize=20` 曾出现 P95 约 321.0ms，超过 250ms 阈值；当前代码将角色列表的多类授权读取合并为一次授权快照查询后，临时 18083 strict 基线 P95 降至 90.2ms。

## 2026-07-28 结构基线

- 已增加慢请求日志 `[SLOW_REQUEST]`，默认阈值由 `WECHECKIN_SLOW_REQUEST_MS` 控制。
- 已增加性能脚本 `scripts/check-performance.mjs`，默认访问 `http://127.0.0.1:8083`。
- 已将 `CHECK_PERFORMANCE=1 bash scripts/check.sh` 作为本地质量门禁入口。
- MySQL slow query 和 EXPLAIN 排查流程见 `docs/performance/README.md`。

## 2026-07-28 本地脚本可运行性记录

```text
环境：本机，后端服务未启动，未提供后台/客户端/钉钉 token
命令：npm run check:performance
结果：命令退出码 0；受保护接口显示 SKIP，公开列表接口显示 fetch failed
说明：该记录只证明脚本和非 strict 模式可运行，不代表接口真实耗时。

命令：WECHECKIN_PERF_STRICT=1 npm run check:performance
结果：命令退出码 1；公开列表接口 fetch failed 时会阻断
说明：严格模式阻断链路正常；后续真实 P95 以“2026-07-28 strict 真实接口基线”为准。
```

## 2026-07-28 结构与前端性能守护记录

```text
命令：npm --prefix frontend run check:all
结果：通过
覆盖：请求层 GET in-flight 去重、固定搜索区、搜索确认触发、列表图片懒加载、分页追加、防重复触底。

命令：npm --prefix admin run build && npm --prefix admin run check:bundle
结果：通过
覆盖：管理后台生产构建和拆包预算。

命令：npm --prefix dingtalk-h5 run check:scaffold && npm --prefix dingtalk-h5 run check:bootstrap
结果：通过
覆盖：钉钉 H5 脚手架和 bootstrap 轻量化守护。
```

后续需要在 MySQL、Redis、后端服务均启动后，携带 token 重新执行：

```bash
WECHECKIN_PERF_BASE_URL=http://127.0.0.1:8083 \
WECHECKIN_ADMIN_TOKEN='your-admin-token' \
WECHECKIN_USER_TOKEN='your-user-token' \
WECHECKIN_DINGTALK_TOKEN='your-dingtalk-token' \
npm run check:performance
```
