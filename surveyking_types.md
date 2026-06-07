# SurveyKing 题型属性数据

> 数据来源: https://s.surveyking.cn/survey/m15gCH/edit?mode=survey
> 采集方式: Playwright 连接 Chrome 调试端口，从 JSON 视图和 UI 抽屉面板提取

---

## 一、题型分类与列表

### 1. 选择题
| 类型 | 内部 type | 图标 |
|------|-----------|------|
| 单选题 | `Radio` | radio |
| 多选题 | `Checkbox` | checkbox |
| 下拉题 | `Select` | select |
| 级联题 | `Cascader` | cascader |
| 上传文件 | `Upload` | upload |

### 2. 填空题
| 类型 | 内部 type |
|------|-----------|
| 单行文本题 | `FillBlank` |
| 多行文本题 | `Textarea` |
| 多项填空 | `MultipleBlank` |
| 电子签名 | `Signature` |
| 横向填空 | `HorzBlank` |
| 扫码题 | `Barcode` |

### 3. 打分题
| 类型 | 内部 type |
|------|-----------|
| 打分题 | `Score` |
| NPS题 | `Nps` |

### 4. 矩阵题
| 类型 | 内部 type |
|------|-----------|
| 矩阵填空 | `MatrixFillBlank` |
| 矩阵单选 | `MatrixRadio` |
| 矩阵多选 | `MatrixCheckbox` |
| 表格自增 | `MatrixAuto` |

### 5. 辅助布局
| 类型 | 内部 type |
|------|-----------|
| 问题组 | `QuestionSet` |
| 文字描述 | `Remark` |
| 分页 | `Pagination` |
| 分割线 | `SplitLine` |

### 6. 高级题型
| 类型 | 内部 type |
|------|-----------|
| 成员 | `User` |
| 部门 | `Dept` |
| 富文本 | `RichText` |

### 7. 个人信息
| 类型 | 内部 type |
|------|-----------|
| 地理位置 | `Address` |
| 省市县 | `(Cascader)` |
| 民族 | `(Select)` |
| 邮箱 | `(FillBlank)` |
| 日期 | `(FillBlank)` |
| 时间 | `(FillBlank)` |
| 手机 | `(FillBlank)` |
| 身份证 | `(FillBlank)` |

---

## 二、JSON 数据结构（通用模板）

```json
{
  "id": "xxx",
  "title": "题目标题",
  "description": "题目说明（可选）",
  "type": "TypeName",
  "attribute": {
    // 问题级别的属性
  },
  "children": [
    {
      "id": "xxx",
      "title": "选项/子项标题（可选）",
      "attribute": {
        // 选项/子项级别的属性
      }
    }
  ]
}
```

---

## 三、各题型属性详情

### 1. 单选题 (Radio)

#### 问题设置（抽屉面板）
| 属性 | 类型 | 说明 |
|------|------|------|
| 必填 | switch | 是否必填 |
| 只读 | switch | 是否只读 |
| 结束公式 | link → 点击设置 | 满足条件后结束问卷 |
| 跳转公式 | link → 点击设置 | 满足条件跳转到指定题 |
| 默认隐藏 | switch | 初始是否隐藏 |
| 选项布局 | number | 每行显示选项数 |
| 必填公式 | link → 点击设置 | 满足条件时才必填 |
| 文本替换 | link → 点击设置 | 替换题目中的文本 |
| 校验规则 | link → 点击设置 | 自定义校验 |
| 题干说明 | textarea | 题干补充说明 |

#### JSON 属性
```json
{
  "type": "Radio",
  "attribute": {
    "required": true
  },
  "children": [
    { "id": "opt1", "attribute": { "配额": "数值", "显示隐藏公式": "...", "自动勾选公式": "..." } },
    { "id": "opt2", "attribute": {} }
  ]
}
```

#### 选项设置（抽屉面板-点击选项后）
| 属性 | 说明 |
|------|------|
| 选项配额 | 该选项可被选中的次数上限 |
| 显示隐藏公式 | 条件公式控制选项显隐 |
| 选项自动勾选公式 | 条件公式自动勾选 |
| ID | 选项唯一标识 |

#### 选项编辑（主面板）
- 添加单个选项 按钮
- 批量添加选项 按钮
- 每个选项可拖拽排序
- 选项文本支持富文本编辑（加粗、斜体、清除格式、文本引用、插入填空）

---

### 2. 多选题 (Checkbox)、下拉题 (Select)

同单选题结构，`type` 分别为 `Checkbox` / `Select`。

- **问题属性**: `required`
- **选项属性**: 同单选题（配额、公式）
- **选项编辑**: 同单选题

---

### 3. 级联题 (Cascader)

- **问题属性**: `required`
- JSON 属性同 Radio
- 选项为级联层级结构

---

### 4. 上传文件 (Upload)

- **问题属性**: `required`
- 选项设置不同（文件上传相关配置）

---

### 5. 单行文本题 (FillBlank)

#### 问题设置（抽屉面板）
| 属性 | 说明 |
|------|------|
| 必填 | switch |
| 只读 | switch |
| 结束公式 | 条件结束问卷 |
| 跳转公式 | 条件跳转 |
| 默认隐藏 | switch |
| 选项布局 | 每行列数 |
| 必填公式 | 条件必填 |
| 文本替换 | 文本替换公式 |
| 校验规则 | 自定义校验 |
| 题干说明 | 补充说明 |

#### 选项/字段设置（抽屉面板-点击输入框）
| 属性 | 说明 |
|------|------|
| 必填 | switch |
| 只读 | switch |
| 计算公式 | 自动计算公式（Excel 风格） |
| 内容限制 | 数字/文本/手机/邮箱/身份证等 |
| 小数位数限制 | 数字类型时有效 |
| 最少填写 | 最少字符数 |
| 最多填写 | 最多字符数 |
| 选项后文字 | 输入框后面的文字（如"cm"、"kg"） |
| 答案唯一 | 禁止重复提交相同内容 |
| 问卷关联 | 关联其他问卷数据 |
| 提示语 | 输入框占位文本 |
| ID | 字段唯一标识 |

#### JSON 属性
```json
{
  "type": "FillBlank",
  "attribute": {
    "required": true,
    "examAnswerMode": "none"
  },
  "children": [
    {
      "id": "field1",
      "attribute": {
        "dataType": "number",
        "calculate": "#{otherField}*2"
      }
    }
  ]
}
```

#### 完整 JSON 属性字典
- **问题属性**: `required`, `examAnswerMode`
- **子项属性**: `dataType` (number/idCard等), `calculate` (公式), `dateTimeFormat`

---

### 6. 多行文本题 (Textarea)

- **问题属性**: `required`
- 子项属性同 FillBlank

---

### 7. 多项填空 (MultipleBlank)
- **问题属性**: `required`
- 结构：一个题目包含多个填空字段

---

### 8. 电子签名 (Signature)
- **问题属性**: `required`, `examAnswerMode`
- 子项无特殊属性

---

### 9. 横向填空 (HorzBlank)
- **问题属性**: `required`, `content`
- 横向排列的多个填空

---

### 10. 扫码题 (Barcode)
- **问题属性**: `required`

---

### 11. 打分题 (Score)
- **问题属性**: `required`
- UI 设置同单行文本模板

### 12. NPS题 (Nps)
- **问题属性**: `required`
- NPS 评分（0-10分）

---

### 13. 矩阵填空 (MatrixFillBlank)、矩阵单选 (MatrixRadio)、矩阵多选 (MatrixCheckbox)
- **问题属性**: `required`, `width`
- **子项属性**: `width`
- 结构：行标题 + 列标题 + 单元格

### 14. 表格自增 (MatrixAuto)
- **问题设置**: 必填, 默认隐藏, 最少行数, 最多行数, 文本替换, 题干说明
- **JSON**: `{ required, examAnswerMode }`
- **子项属性**: `width`, `dataType`
- 用户可动态添加行

---

### 15. 文字描述 (Remark)
- **设置**: 文本替换（点击设置）
- **JSON**: `{ replaceTextRule: "公式字符串" }`
- 只展示文本，不可输入

### 16. 问题组 (QuestionSet)
- **设置**: 默认隐藏, 文本替换
- 用于对题目分组

### 17. 分页 (Pagination)
- **JSON**: `{ currentPage: 1, totalPage: N }`
- 多页问卷的分页节点

### 18. 分割线 (SplitLine)
- 无特殊属性

---

### 19. 成员 (User)、部门 (Dept)、富文本 (RichText)
- **问题属性**: `required`
- UI 同单行文本模板
- User：选择组织成员
- Dept：选择部门
- RichText：富文本编辑区

### 20. 地理位置 (Address)
- **问题属性**: 未捕捉到特殊属性
- UI 同单行文本模板

---

## 四、UI 布局结构

```
.survey-editor (flex row)
├── .survey-sidebar-panel (左)
│   ├── .survey-sidebar-panel-tabs: [题目] [外观] [逻辑]
│   │   ├── 题目 → .ant-tabs: [题型] [题库] [大纲]
│   │   │   └── .menu-group (题型分类)
│   │   │       ├── 选择题 → Radio/Checkbox/Select/Cascader/Upload
│   │   │       ├── 填空题 → FillBlank/Textarea/MultipleBlank/Signature/HorzBlank/Barcode
│   │   │       ├── 打分题 → Score/Nps
│   │   │       ├── 矩阵题 → MatrixFillBlank/MatrixRadio/MatrixCheckbox/MatrixAuto
│   │   │       ├── 辅助布局 → QuestionSet/Remark/Pagination/SplitLine
│   │   │       ├── 高级题型 → User/Dept/RichText
│   │   │       └── 个人信息 → Address/省市县/民族/邮箱/日期/时间/手机/身份证
│   │   ├── 外观 → 背景图/页眉图
│   │   └── 逻辑 → (空)
├── .survey-main-panel (中)
│   ├── toolbar: 撤销/恢复/快捷键/文本导入 | [编 辑] [JSON] [预 览] [保 存]
│   └── content: 问卷预览/编辑区
├── .survey-setting-panel (右 - 占位)
└── .ant-drawer.setting-drawer (浮动抽屉)
    └── .drawerTitle → "Q1 问题设置" 或 "Q1 选项1设置"
        └── .setting-item → 具体设置行
```

---

## 五、侧边栏标签页

| 标签 | 内容 |
|------|------|
| **题目** | 题型分类列表（可拖拽添加到问卷） |
| **外观** | 背景图、页眉图设置 |
| **逻辑** | 问卷逻辑规则列表 |

题型子标签: [题型] [题库] [大纲]

---

## 六、注意事项

1. **问题设置 vs 选项设置**: 点击题目标题区域打开"问题设置"，点击选项/输入框区域打开"选项设置"
2. **选择题选项**: 主面板通过"添加单个选项"/"批量添加选项"管理选项列表，每个选项可独立打开抽屉设置属性
3. **JSON 中未出现的属性**: 部分属性（如只读、结束公式、跳转公式等）在本次采集的 JSON 中未出现，是因为这些属性只在用户主动配置后才序列化
4. **个人信息类型复用**: 省市县实际为 Cascader，民族实际为 Select，邮箱/日期/时间/手机/身份证为 FillBlank，UI 上归类到"个人信息"便于用户查找
5. **问卷整体属性**: `suffix` (结束语), `submitButton` (提交按钮文字), `backgroundImage`, `globalRule`
