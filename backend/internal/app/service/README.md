# service 目录说明

`service` 包承载 HTTP handler 之后的业务编排逻辑。根目录只保留包文档和结构测试；后台、首页、问卷、考试、登录、报名、活动等领域都拆成子包，handler 直接依赖对应业务域。

## 文件分组

- `service/adminauth`、`service/adminmgr`、`service/adminuser`、`service/admincontent`、`service/adminlog`、`service/dashboard`：后台登录、管理员、用户、内容、操作日志和首页统计。
- `service/role`、`service/department`、`service/dict`、`service/menu`、`service/online`：后台角色、部门、字典、菜单权限和在线会话。
- `service/survey`：问卷、答卷；提交后统计和通知已拆到 `service/poststat`。
- `service/exam`：考试服务。
- `service/setup`：系统设置。
- `service/event`：活动列表、详情、报名、动态、评分和后台管理。
- `service/enroll`：打卡报名的客户端、详情、记录和提交。
- `service/passport`：客户端登录、注册、个人资料。
- `service/news`、`service/favorite`：资讯和收藏。
- `service/home`：首页聚合逻辑。
- `backend/internal/app/support`：数据范围、部门、媒体地址、发布范围、排序等跨领域纯逻辑。

## 维护规则

1. 新增 service 文件必须使用明确的业务前缀。
2. 单个 service 根目录文件应保持在 300 行以内；超过时先按职责拆文件。
3. 通用纯逻辑优先放到 `backend/internal/app/support` 下，再由 service 包包装或调用。
4. 拆子包时优先让 handler 直接依赖领域子包；确需兼容包装时必须写清楚保留原因，并用结构测试约束迁移边界。
