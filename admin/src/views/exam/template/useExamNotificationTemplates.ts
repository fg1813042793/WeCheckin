import { ref } from 'vue'
import { adminApi, type ExamNotificationTemplateScope, type TemplatePreset } from '../../../api'

export function useExamNotificationTemplates() {
  const builtinPresets = ref<TemplatePreset[]>([])
  const customPresets = ref<TemplatePreset[]>([])
  const loading = ref(false)
  const error = ref('')
  const savingScope = ref<ExamNotificationTemplateScope | ''>('')

  async function loadTemplates() {
    loading.value = true
    error.value = ''
    try {
      const response = await adminApi.examNotificationTemplatesGet()
      builtinPresets.value = Array.isArray(response.data?.builtin) ? response.data.builtin : []
      customPresets.value = Array.isArray(response.data?.custom) ? response.data.custom : []
    } catch {
      builtinPresets.value = []
      customPresets.value = []
      error.value = '考试通知模版加载失败'
    } finally { loading.value = false }
  }

  async function saveTemplates(scope: ExamNotificationTemplateScope, presets: TemplatePreset[]) {
    savingScope.value = scope
    try {
      await adminApi.examNotificationTemplatesSave({ scope, presets })
      if (scope === 'builtin') builtinPresets.value = presets
      else customPresets.value = presets
    } finally { savingScope.value = '' }
  }

  return { builtinPresets, customPresets, loading, error, savingScope, loadTemplates, saveTemplates }
}
