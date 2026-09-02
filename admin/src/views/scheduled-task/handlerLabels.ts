import type { HandlerType } from './types'

const handlerTypeLabels: Record<HandlerType, string> = {
  go: 'Go 注册任务',
  workflow: '发起流程',
  http: 'HTTP / Webhook 请求',
  shell: '受控 Shell 命令',
  sql: '受控 SQL 任务',
}

export function handlerTypeLabel(type: HandlerType) {
  return handlerTypeLabels[type] || type
}
