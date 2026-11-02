# 考试答卷人工批改

目标：补齐待人工评分答卷的批改入口、独立页面和服务端提交链路。此处按考试人工批改处理，不引入审批流的通过/驳回节点。

- 状态 1 的答卷可批改；状态 0 不可批改，状态 2 可查看批改结果，不重复改分。
- 新权限：`admin:menu:exam:record-grade` 和 `admin:api:exam-record:grade`，服务端 `exam-record:grade`；维护原有考试数据范围。
- `GET /api/v2/admin/exams/{id}/records/{recordId}/grading` 返回批改上下文；`POST` 同路径提交 `{expectedVersion, scores:[{questionId,score}], comment}`。
- 返回 `{record,examTitle,passScore,version,canGrade,notice,questions:[{questionId,title,type,answer,referenceAnswer,fullScore,gotScore,needManual}]}`。
- 必须为全部待人工题评分，分数为 0 至该题提交时满分之间的整数；禁止改动自动题得分。服务端重新计算总分和通过结果。
- 事务锁定答卷，校验状态和版本，避免重复提交或覆盖其他人的批改。保存批改人、时间、评语，写审计日志。
- 新提交答卷保存私有题目/及格分快照，批改评分不受后续编辑影响。无快照的历史答卷兼容当前题目资料并明确提示，评分上限仍以已存逐题结果为准。
- 数据库迁移添加 `exam_r_review_snapshot`（MEDIUMTEXT、可空，JSON 不公开）、`exam_r_graded_by`、`exam_r_graded_at`、`exam_r_grading_comment`，并补权限声明及授权。迁移只生成，不在生产执行。
- 不重发已有交卷通知，不增加提交计数。

实现分工：主代理实现领域校验、持久化事务与快照；后台代理实现页面/入口/API/类型；权限代理实现迁移、权限、HTTP 适配、路由和 Swagger 声明。各自仅修改分配文件，保留工作区已有改动。

验证：合法与非法分数、缺漏/重复题目、状态/版本冲突、越权保护、持久化回滚；Admin 编译及模拟接口浏览器批改成功/失败/无权限/已完成场景；完成后运行相关和全量检查并记录基线失败。

## 已完成

- 列表与答卷详情入口、独立批改页、GET/POST 接口、数据范围、权限目录、迁移和审计已接通。
- 评分规则、事务提交/回滚、重复提交、答卷归属/数据范围、交卷后答案保护、私有快照与复杂题材料测试通过。
- `npm run check:exam-grading`、`check:exam-management`、`check:admin-ui-contract` 和 `npm run build` 通过。
- 本地浏览器模拟接口验证通过：查看入口进入批改、空评分拦截、0分、失败草稿保留、版本冲突重载并核对、成功只读、返回列表刷新、无权限按钮隐藏与路由拦截；未出现页面运行时错误。
- 全量 Go 测试仍有既有考试复制 Swagger 断言、钉钉配置、通知审计分类、生产配置检查、旧考试导出原始错误响应检查失败；Admin `check:all` 仍停在既有导航断言 `props.depth > 0`。
- Swagger 已通过规定命令生成。未执行生产迁移、未部署、未提交生产批改；迁移必须先于新后端上线。
