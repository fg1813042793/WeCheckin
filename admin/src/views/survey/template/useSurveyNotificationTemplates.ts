import { ref } from 'vue'
import { adminApi, type SurveyNotificationTemplateScope, type TemplatePreset } from '../../../api'

export function useSurveyNotificationTemplates() {
  const builtinPresets = ref<TemplatePreset[]>([])
  const customPresets = ref<TemplatePreset[]>([])
  const loading = ref(false)
  const error = ref('')
  const savingScope = ref<SurveyNotificationTemplateScope | ''>('')

  async function loadTemplates() {
    loading.value = true
    error.value = ''
    try {
      const response = await adminApi.surveyNotificationTemplatesGet()
      builtinPresets.value = Array.isArray(response.data?.builtin) ? response.data.builtin : []
      customPresets.value = Array.isArray(response.data?.custom) ? response.data.custom : []
    } catch {
      builtinPresets.value = []
      customPresets.value = []
      error.value = '问卷通知模版加载失败'
    } finally {
      loading.value = false
    }
  }

  async function saveTemplates(scope: SurveyNotificationTemplateScope, presets: TemplatePreset[]) {
    savingScope.value = scope
    try {
      await adminApi.surveyNotificationTemplatesSave({ scope, presets })
      if (scope === 'builtin') builtinPresets.value = presets
      else customPresets.value = presets
    } finally {
      savingScope.value = ''
    }
  }

  return { builtinPresets, customPresets, loading, error, savingScope, loadTemplates, saveTemplates }
}
