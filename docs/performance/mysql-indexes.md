# MySQL 索引记录

本文档记录性能优化相关索引、命中场景、EXPLAIN 检查方式和回滚建议。新增索引均应优先使用幂等 migration，避免历史环境重复执行时报错。

## 用户与管理员列表

| 索引 | 表 | 命中场景 |
|---|---|---|
| `idx_users_status_role_id` | `users` | 用户列表按状态、角色过滤 |
| `idx_users_add_time_id` | `users` | 用户列表按创建时间和 id 分页排序 |
| `idx_users_mobile` | `users` | 用户手机号搜索 |
| `idx_users_name` | `users` | 用户姓名搜索 |
| `idx_users_admin_list` | `users` | 后台管理员列表筛选 `user_role_id > 0` |
| `idx_users_admin_login_time` | `users` | 管理员最近登录排序 |
| `idx_user_depts_dept_user` | `user_depts` | 通过部门批量查用户 |
| `idx_user_depts_user_dept` | `user_depts` | 通过用户批量查部门 |

推荐 EXPLAIN：

```sql
EXPLAIN SELECT id, user_name, user_mobile, user_status, user_role_id
FROM users
WHERE user_status = 1
ORDER BY user_add_time DESC, id DESC
LIMIT 20;
```

## 客户端列表

| 索引 | 表 | 命中场景 |
|---|---|---|
| `idx_news_status_order_time` | `news` | 通知列表按发布状态、排序和时间分页 |
| `idx_news_title` | `news` | 通知标题搜索 |
| `idx_surveys_status_order_id` | `surveys` | 问卷列表已发布数据排序 |
| `idx_exams_status_order_id` | `exams` | 考试列表已发布数据排序 |
| `idx_enrolls_status_order_time` | `enrolls` | 打卡任务列表排序 |
| `idx_enrolls_title` | `enrolls` | 打卡任务标题搜索 |
| `idx_events_status_type_order_time` | `events` | 赛事活动列表按状态、类型、时间排序 |

列表接口不要加载 schema、富文本、题目详情、大图集合。慢接口先跑 `npm run check:performance` 找到具体 route，再用后端 `[SLOW_REQUEST]` 日志定位 SQL。

## 钉钉 H5 绩效

| 索引 | 表 | 命中场景 |
|---|---|---|
| `idx_dt_h5_review_employee_period` | `dingtalk_h5_perf_reviews` | 员工查看自己的绩效 |
| `idx_dt_h5_review_manager_status` | `dingtalk_h5_perf_reviews` | 主管按状态查看待处理绩效 |
| `idx_dt_h5_review_hrbp_status` | `dingtalk_h5_perf_reviews` | HRBP 按状态查看绩效 |
| `idx_dt_h5_review_hrbp_reviewer_status` | `dingtalk_h5_perf_reviews` | HRBP 评价视角 |
| `idx_dt_h5_review_status_period` | `dingtalk_h5_perf_reviews` | 工作台状态统计 |
| `idx_dt_h5_history_review_time` | `dingtalk_h5_perf_histories` | 批量查询流程历史 |

推荐 EXPLAIN：

```sql
EXPLAIN SELECT id, review_title, employee_account, status, period
FROM dingtalk_h5_perf_reviews
WHERE employee_account = 'user-openid'
ORDER BY period DESC, id DESC
LIMIT 20;
```

## slow query 处理规则

1. 先确认 `npm run check:performance` 中哪个接口 P95 超阈值。
2. 再查后端 `[SLOW_REQUEST]` 的 method、path、requestId。
3. 对应 SQL 执行 `EXPLAIN`，确认 `key` 和 `rows`。
4. 如果新增索引，补 migration、结构测试和本文件记录。
5. 如果索引影响写入性能，回滚时使用 `DROP INDEX index_name ON table_name`。
