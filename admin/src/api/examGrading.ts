import request from '@/utils/request'
import { hasPerm } from '@/utils/permission'
import type { ExamGradingContext, ExamGradingSubmission } from '@/types/examGrading'

export function hasExamGradingPermission(): boolean {
  return hasPerm('admin:menu:exam:record-grade') && hasPerm('admin:api:exam-record:grade')
}

function gradingPath(examId: number, recordId: number): string {
  if (!hasExamGradingPermission()) throw new Error('无考试答卷批改权限')
  if (![examId, recordId].every(id => Number.isSafeInteger(id) && id > 0)) {
    throw new Error('考试或答卷参数无效')
  }
  return `/api/v2/admin/exams/${examId}/records/${recordId}/grading`
}

export const examGradingApi = {
  get(examId: number, recordId: number) {
    return request.get<ExamGradingContext>(gradingPath(examId, recordId))
  },
  submit(examId: number, recordId: number, data: ExamGradingSubmission) {
    return request.post<ExamGradingContext, ExamGradingSubmission>(gradingPath(examId, recordId), data, {
      headers: { 'Content-Type': 'application/json' },
      transformRequest: [(body: unknown) => JSON.stringify(body)],
    })
  },
}
