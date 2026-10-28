export type SurveyStatus = 'published' | 'draft' | 'stopped'

export interface DemoSurvey {
  id: number
  title: string
  category: string
  status: SurveyStatus
  description: string
  responses: number
  questions: number
  updated: string
  sampleQuestion: string
  sampleOptions: string[]
  distribution: { label: string; count: number }[]
}

export const statusLabels: Record<SurveyStatus, string> = {
  published: '已发布', draft: '草稿', stopped: '已停用',
}

export function createDemoSurveys(): DemoSurvey[] {
  return [
    { id: 1008, title: 'DISC 职业性格测评', category: '人才测评', status: 'published', description: '了解工作风格，发现团队协作中的优势。', responses: 128, questions: 40, updated: '2026-10-12', sampleQuestion: '对于任务，我更愿意做：', sampleOptions: ['制定详细计划', '带领团队行动'], distribution: [{ label: 'D', count: 34 }, { label: 'I', count: 28 }, { label: 'S', count: 40 }, { label: 'C', count: 26 }] },
    { id: 1007, title: '新员工入职体验调查', category: '员工体验', status: 'published', description: '收集入职反馈，让每一位新同事更快融入。', responses: 36, questions: 12, updated: '2026-10-11', sampleQuestion: '入职引导是否清晰？', sampleOptions: ['非常清晰', '基本清晰'], distribution: [{ label: '非常清晰', count: 24 }, { label: '基本清晰', count: 10 }, { label: '需改进', count: 2 }] },
    { id: 1006, title: '第三季度培训反馈', category: '培训反馈', status: 'published', description: '回顾课程收获，优化下一阶段的培训安排。', responses: 84, questions: 8, updated: '2026-10-10', sampleQuestion: '本次培训对工作有帮助吗？', sampleOptions: ['很有帮助', '有一定帮助'], distribution: [{ label: '很有帮助', count: 52 }, { label: '有一定帮助', count: 27 }, { label: '帮助较少', count: 5 }] },
    { id: 1005, title: '团队协作满意度', category: '员工体验', status: 'draft', description: '从沟通、协作与支持三个维度了解团队。', responses: 0, questions: 16, updated: '2026-10-09', sampleQuestion: '团队内的信息传达及时吗？', sampleOptions: ['总是及时', '多数时候及时'], distribution: [] },
    { id: 1004, title: 'DISC 校园招聘测评', category: '人才测评', status: 'published', description: '发现个人特质，帮助应届生了解自己。', responses: 62, questions: 40, updated: '2026-10-08', sampleQuestion: '面对新挑战，我通常会：', sampleOptions: ['积极探索新的方法', '先了解目标和规则'], distribution: [{ label: 'D', count: 14 }, { label: 'I', count: 20 }, { label: 'S', count: 17 }, { label: 'C', count: 11 }] },
    { id: 1003, title: '办公环境改善建议', category: '员工体验', status: 'draft', description: '听取真实需求，共同改善日常办公体验。', responses: 0, questions: 6, updated: '2026-10-07', sampleQuestion: '你最希望改善哪一方面？', sampleOptions: ['办公设施', '公共空间'], distribution: [] },
    { id: 1002, title: '管理能力提升课程评价', category: '培训反馈', status: 'published', description: '评估课程内容与实际管理场景的匹配程度。', responses: 25, questions: 10, updated: '2026-10-06', sampleQuestion: '课程内容是否符合预期？', sampleOptions: ['超出预期', '符合预期'], distribution: [{ label: '超出预期', count: 10 }, { label: '符合预期', count: 13 }, { label: '低于预期', count: 2 }] },
    { id: 1001, title: '上半年员工满意度调查', category: '员工体验', status: 'stopped', description: '记录阶段性反馈，为后续改善提供依据。', responses: 156, questions: 20, updated: '2026-09-30', sampleQuestion: '你对目前的工作体验满意吗？', sampleOptions: ['非常满意', '比较满意'], distribution: [{ label: '非常满意', count: 96 }, { label: '比较满意', count: 48 }, { label: '一般', count: 12 }] },
  ]
}
