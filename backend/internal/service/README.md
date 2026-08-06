# service 目录说明

`service` 包承载 HTTP handler 之后的业务编排逻辑。根目录只保留包文档和结构测试；业务实现按入口端分为 `admin`、`client`、`dingtalkh5` 三类目录，handler 直接依赖对应业务域。

## 文件分组

- `service/admin/*`：后台登录、管理员、用户、内容、系统配置、权限、组织、字典、在线会话、FormKit 管理等后台管理能力。
- `service/client/*`：客户端登录、首页、资讯、收藏、报名、活动、问卷、考试、提交统计等前台业务能力。
- `service/dingtalkh5/performance`：钉钉 H5 登录、免登、权限快照、绩效流程、通知、模板、用户绑定和企业应用配置能力。
- `backend/internal/support`：数据范围、部门、媒体地址、发布范围、查询条件、排序等跨领域纯逻辑。

## 维护规则

1. 新增 service 文件必须使用明确的业务前缀。
2. 单个 service 根目录文件应保持在 300 行以内；超过时先按职责拆文件。
3. 通用纯逻辑优先放到 `backend/internal/support` 下，再由 service 包包装或调用。
4. 拆子包时优先让 handler 直接依赖领域子包；确需兼容包装时必须写清楚保留原因，并用结构测试约束迁移边界。
