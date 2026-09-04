import type { RegisteredAppPage } from '@/config/app-navigation'

function appContentRoute(key: string) {
  return `/pages/index/index?view=${encodeURIComponent(key)}`
}

export const performanceRootNavItem = {
  key: 'performance',
  label: '绩效管理',
  icon: 'list',
}
export const performanceMenuPages: RegisteredAppPage[] = [
  {
    key: 'performance:mine',
    contentKey: 'mine',
    route: appContentRoute('performance:mine'),
    title: '我的绩效',
    description: '填写和查看自己的月度绩效',
    icon: 'file-text',
    parentKey: 'performance',
    rootKey: 'performance',
  },
  {
    key: 'performance:history',
    contentKey: 'history',
    route: appContentRoute('performance:history'),
    title: '历史绩效',
    description: '查看已完成的历史绩效记录',
    icon: 'order',
    parentKey: 'performance',
    rootKey: 'performance',
  },
  {
    key: 'performance:manager',
    contentKey: 'manager',
    route: appContentRoute('performance:manager'),
    title: '我要评价',
    description: '处理直属上级待评价的绩效单',
    icon: 'account-fill',
    parentKey: 'performance',
    rootKey: 'performance',
  },
  {
    key: 'performance:hrbp',
    contentKey: 'hrbp',
    route: appContentRoute('performance:hrbp'),
    title: 'HRBP评价',
    description: '处理 HRBP 待评价的绩效单',
    icon: 'man-add',
    parentKey: 'performance',
    rootKey: 'performance',
  },
  {
    key: 'performance:summary',
    contentKey: 'summary',
    route: appContentRoute('performance:summary'),
    title: 'HRBP汇总',
    description: '按人员和月份查看进度、分档并导出结果',
    icon: 'order',
    parentKey: 'performance',
    rootKey: 'performance',
  },
  {
    key: 'performance:org',
    contentKey: 'org',
    route: appContentRoute('performance:org'),
    title: '流程制定',
    description: '维护流程人员、直属上级和 HRBP 关系',
    icon: 'setting',
    parentKey: 'performance',
    rootKey: 'performance',
  },
  {
    key: 'performance:template',
    contentKey: 'template',
    route: appContentRoute('performance:template'),
    title: '绩效参数',
    description: '维护目标模板、价值观标尺和绩效工资系数',
    icon: 'file-text',
    parentKey: 'performance',
    rootKey: 'performance',
  },
]
