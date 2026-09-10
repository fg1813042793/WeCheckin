import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const currentDir = dirname(fileURLToPath(import.meta.url))
const rootDir = resolve(currentDir, '..')

// These legacy pages are intentionally frozen at their current size. New
// responsibilities must be extracted into typed components, composables or API modules.
const legacyBudgets = {
  'src/views/survey/SurveyDesigner.vue': 500,
  'src/views/survey/components/SurveyDesignerQuestionProperties.vue': 500,
  'src/views/survey/components/SurveyDesignerSidebar.vue': 400,
  'src/views/survey/components/SurveyDesignerResultSettings.vue': 120,
  'src/views/survey/composables/useSurveyDesignerQuestions.ts': 260,
  'src/views/survey/composables/useSurveyDesignerTransfer.ts': 320,
  'src/views/survey/composables/useSurveyDesignerDocument.ts': 520,
  'src/views/exam/ExamDesigner.vue': 950,
  'src/views/exam/components/ExamDesignerSidebar.vue': 260,
  'src/views/exam/components/ExamDesignerLogicPanel.vue': 240,
  'src/views/exam/components/ExamDesignerSettingsPanel.vue': 160,
  'src/views/exam/composables/useExamDesignerDocument.ts': 300,
  'src/views/exam/composables/useExamDesignerTransfer.ts': 200,
  'src/views/exam/exam-designer.css': 360,
  'src/views/user/index.vue': 1859,
  'src/views/workflow/designer/components/WorkflowFormDesigner.vue': 1479,
}

for (const [relativePath, maxLines] of Object.entries(legacyBudgets)) {
  const source = readFileSync(resolve(rootDir, relativePath), 'utf8')
  const lineCount = source.split(/\r?\n/).length - (source.endsWith('\n') ? 1 : 0)
  if (lineCount > maxLines) {
    throw new Error(
      `${relativePath} has ${lineCount} lines (budget ${maxLines}); extract the new responsibility instead of growing this legacy component`,
    )
  }
}

const workflowTypesBridge = readFileSync(resolve(rootDir, 'src/views/workflow/types.ts'), 'utf8')
if (!workflowTypesBridge.includes("export * from '../../types/workflow'")) {
  throw new Error('workflow contracts must live in src/types/workflow.ts')
}

const scheduledTaskTypesBridge = readFileSync(resolve(rootDir, 'src/views/scheduled-task/types.ts'), 'utf8')
if (!scheduledTaskTypesBridge.includes("export * from '../../types/scheduledTask'")) {
  throw new Error('scheduled task contracts must live in src/types/scheduledTask.ts')
}
