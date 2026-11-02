<script setup lang="ts">
import type { WorkflowSummaryFormFieldGroup } from '../workflow-summary-columns'
import type { SummaryColumnSource, SummaryMappedColumn, SummaryVersionGroups } from '../workflow-summary-mappings'
import type { WorkflowInstanceSummary } from '@/types/workflow'
import { useLocale } from 'uview-pro'
import { computed, ref, watch } from 'vue'
import { summaryMappedTypeMismatch, validateSummaryMappedColumn } from '../workflow-summary-mappings'

const props = defineProps<{
  modelValue: boolean
  value: SummaryMappedColumn | null
  groups: WorkflowSummaryFormFieldGroup[]
  instances: WorkflowInstanceSummary[]
  versionGroups: SummaryVersionGroups
  versionLoading: Record<string, boolean>
  versionErrors: Record<string, string>
}>()
const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  'save': [value: SummaryMappedColumn, replaceSources: boolean]
  'loadVersion': [definitionId: number, version: number]
}>()
const { t } = useLocale()
const visible = computed({ get: () => props.modelValue, set: value => emit('update:modelValue', value) })
const draft = ref<SummaryMappedColumn>({ id: '', label: '', enabled: true, sources: [] })
const submitted = ref(false)
const replaceSources = ref(true)
const validation = computed(() => submitted.value ? validateSummaryMappedColumn(draft.value) : [])
const typeMismatch = computed(() => summaryMappedTypeMismatch(draft.value, props.groups, props.versionGroups))

watch(() => props.modelValue, (open) => {
  if (!open)
    return
  submitted.value = false
  replaceSources.value = !props.value
  draft.value = props.value
    ? { ...props.value, sources: props.value.sources.map(source => ({ ...source })) }
    : { id: `column-${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 8)}`, label: '', enabled: true, sources: [{ definitionId: 0, fieldKey: '' }] }
  for (const source of draft.value.sources) {
    if (source.version)
      emit('loadVersion', source.definitionId, source.version)
  }
}, { immediate: true })

function currentGroup(source: SummaryColumnSource) {
  return props.groups.find(group => group.definitionId === source.definitionId)
}
function sourceGroup(source: SummaryColumnSource) {
  const group = currentGroup(source)
  return !source.version || group?.version === source.version ? group : props.versionGroups[`${source.definitionId}:${source.version}`]
}
function versionOptions(source: SummaryColumnSource) {
  return [...new Set([
    currentGroup(source)?.version,
    source.version,
    ...props.instances.filter(item => item.definitionId === source.definitionId).map(item => item.definitionVersion),
  ].filter((version): version is number => typeof version === 'number' && Number.isInteger(version) && version > 0))].sort((a, b) => b - a)
}
function eventValue(event: Event) {
  return (event.target as HTMLSelectElement).value
}
function changeFlow(source: SummaryColumnSource, event: Event) {
  source.definitionId = Number(eventValue(event)) || 0
  source.fieldKey = ''
  delete source.fieldLabel
  delete source.version
}
function changeVersion(source: SummaryColumnSource, event: Event) {
  const version = Number(eventValue(event)) || undefined
  source.version = version
  source.fieldKey = ''
  delete source.fieldLabel
  if (version)
    emit('loadVersion', source.definitionId, version)
}
function changeField(source: SummaryColumnSource, event: Event) {
  source.fieldKey = eventValue(event)
  source.fieldLabel = sourceGroup(source)?.columns.find(column => column.field?.key === source.fieldKey)?.label.slice(0, 50)
}
function hasSelectedField(source: SummaryColumnSource) {
  return Boolean(sourceGroup(source)?.columns.some(column => column.field?.key === source.fieldKey))
}
function versionMessage(source: SummaryColumnSource) {
  if (!source.version)
    return ''
  const key = `${source.definitionId}:${source.version}`
  if (props.versionLoading[key])
    return t('summaryMappings.versionLoading')
  return props.versionErrors[key] ? t(`summaryMappings.${props.versionErrors[key]}`) : ''
}
function save() {
  submitted.value = true
  if (validateSummaryMappedColumn(draft.value).length)
    return
  emit('save', { ...draft.value, label: draft.value.label.trim(), sources: draft.value.sources.map(source => ({ ...source })) }, replaceSources.value)
  visible.value = false
}
</script>

<template>
  <u-popup v-model="visible" mode="center" width="min(740px, 92vw)" :border-radius="8" :mask-close-able="false" custom-class="summary-mapping-popup app-pc-control-scope">
    <view class="summary-mapping-editor">
      <view class="summary-mapping-editor__head">
        <text>{{ value ? t('summaryMappings.editTitle') : t('summaryMappings.addTitle') }}</text>
        <u-button custom-class="summary-mapping-editor__close" size="mini" :aria-label="t('summaryMappings.cancel')" @click="visible = false">
          <u-icon name="close" size="18px" />
        </u-button>
      </view>
      <scroll-view scroll-y class="summary-mapping-editor__body">
        <view class="summary-mapping-editor__content">
          <text class="summary-mapping-editor__label">
            {{ t('summaryMappings.columnName') }}
          </text>
          <u-input v-model="draft.label" :maxlength="50" :border="true" :placeholder="t('summaryMappings.namePlaceholder')" />
          <text class="summary-mapping-editor__hint">
            {{ t('summaryMappings.mappingHint') }}
          </text>
          <view v-for="(source, index) in draft.sources" :key="index" class="summary-mapping-editor__source">
            <view class="summary-mapping-editor__source-head">
              <text>{{ t('summaryMappings.sourceNumber', { n: index + 1 }) }}</text>
              <u-button size="mini" type="error" plain :aria-label="t('summaryMappings.removeSource', { n: index + 1 })" @click="draft.sources.splice(index, 1)">
                {{ t('summaryMappings.remove') }}
              </u-button>
            </view>
            <!-- H5 native selects open beside the field on PC and use the system picker on phones. -->
            <view class="summary-mapping-editor__fields">
              <view>
                <text class="summary-mapping-editor__label">
                  {{ t('summaryMappings.flow') }}
                </text>
                <select class="summary-mapping-editor__select" :aria-label="t('summaryMappings.flowNumber', { n: index + 1 })" :value="source.definitionId || ''" @change="changeFlow(source, $event)">
                  <option value="">
                    {{ t('summaryMappings.chooseFlow') }}
                  </option>
                  <option v-if="source.definitionId && !currentGroup(source)" :value="source.definitionId">
                    {{ t('summaryMappings.unavailableFlow') }}
                  </option>
                  <option v-for="group in groups" :key="group.definitionId" :value="group.definitionId">
                    {{ group.label }}
                  </option>
                </select>
              </view>
              <view>
                <text class="summary-mapping-editor__label">
                  {{ t('summaryMappings.version') }}
                </text>
                <select class="summary-mapping-editor__select" :aria-label="t('summaryMappings.versionNumber', { n: index + 1 })" :value="source.version || ''" :disabled="!source.definitionId" @change="changeVersion(source, $event)">
                  <option value="">
                    {{ t('summaryMappings.defaultVersion') }}
                  </option>
                  <option v-for="version in versionOptions(source)" :key="version" :value="version">
                    v{{ version }}
                  </option>
                </select>
              </view>
              <view>
                <text class="summary-mapping-editor__label">
                  {{ t('summaryMappings.field') }}
                </text>
                <select class="summary-mapping-editor__select" :aria-label="t('summaryMappings.fieldNumber', { n: index + 1 })" :value="source.fieldKey" :disabled="!source.definitionId || Boolean(source.version && versionLoading[`${source.definitionId}:${source.version}`])" @change="changeField(source, $event)">
                  <option value="">
                    {{ t('summaryMappings.chooseField') }}
                  </option>
                  <option v-if="source.fieldKey && !hasSelectedField(source)" :value="source.fieldKey">
                    {{ source.fieldLabel || t('summaryMappings.unavailableField') }}
                  </option>
                  <option v-for="column in sourceGroup(source)?.columns || []" :key="column.key" :value="column.field?.key">
                    {{ column.label }}
                  </option>
                </select>
              </view>
            </view>
            <text v-if="versionMessage(source)" class="summary-mapping-editor__hint">
              {{ versionMessage(source) }}
            </text>
            <u-button v-if="source.version && versionErrors[`${source.definitionId}:${source.version}`] === 'versionLoadFailed'" size="mini" @click="emit('loadVersion', source.definitionId, source.version)">
              {{ t('summaryMappings.retry') }}
            </u-button>
          </view>
          <u-button size="small" type="primary" plain :disabled="draft.sources.length >= 50" @click="draft.sources.push({ definitionId: 0, fieldKey: '' })">
            {{ t('summaryMappings.addSource') }}
          </u-button>
          <text class="summary-mapping-editor__hint">
            {{ t('summaryMappings.versionHint') }}
          </text>
          <u-checkbox v-model="replaceSources" :label="t('summaryMappings.replaceSources')" />
          <text v-if="typeMismatch" class="summary-mapping-editor__warning">
            {{ t('summaryMappings.typeWarning') }}
          </text>
          <view v-if="validation.length" class="summary-mapping-editor__errors" role="alert">
            <text v-for="error in validation" :key="error">
              {{ t(`summaryMappings.errors.${error}`) }}
            </text>
          </view>
        </view>
      </scroll-view>
      <view class="summary-mapping-editor__footer">
        <u-button size="small" @click="visible = false">
          {{ t('summaryMappings.cancel') }}
        </u-button>
        <u-button size="small" type="primary" @click="save">
          {{ t('summaryMappings.save') }}
        </u-button>
      </view>
    </view>
  </u-popup>
</template>

<style scoped lang="scss">
.summary-mapping-editor { width: min(740px, 92vw); max-height: 88vh; display: flex; flex-direction: column; background: #fff; color: #1f2329; }
.summary-mapping-editor__head { display: flex; align-items: center; justify-content: space-between; padding: 16px 20px; border-bottom: 1px solid #e5e9ef; font-size: 17px; font-weight: 600; }
:deep(.summary-mapping-editor__close) { width: 28px; height: 28px; padding: 0; margin: 0; border: 0; }
.summary-mapping-editor__body { min-height: 0; max-height: 65vh; }
.summary-mapping-editor__content { padding: 20px; box-sizing: border-box; }
.summary-mapping-editor__label { display: block; margin-bottom: 8px; color: #4e5969; font-size: 13px; }
.summary-mapping-editor__hint { display: block; margin: 12px 0; color: #86909c; font-size: 12px; line-height: 1.6; }
.summary-mapping-editor__source { margin: 12px 0; padding: 12px; border: 1px solid #e5e9ef; border-radius: 6px; background: #f8fafb; }
.summary-mapping-editor__source-head { display: flex; align-items: center; justify-content: space-between; margin-bottom: 12px; font-size: 13px; }
.summary-mapping-editor__source-head :deep(.u-btn) { margin: 0; }
.summary-mapping-editor__fields { display: grid; grid-template-columns: minmax(0, 1.15fr) 112px minmax(0, 1fr); gap: 10px; }
.summary-mapping-editor__fields > view { min-width: 0; }
.summary-mapping-editor__select { width: 100%; height: 34px; padding: 0 8px; border: 1px solid #d9e1e8; border-radius: 4px; background: #fff; color: #1f2329; font-size: 13px; box-sizing: border-box; }
.summary-mapping-editor__select:focus { border-color: var(--u-primary, #0f766e); outline: none; }
.summary-mapping-editor__select:disabled { color: #a9b0bb; background: #f2f4f7; }
.summary-mapping-editor__footer { display: flex; justify-content: flex-end; gap: 10px; padding: 14px 20px; border-top: 1px solid #e5e9ef; }
.summary-mapping-editor__footer :deep(.u-btn) { margin: 0; min-width: 72px; }
.summary-mapping-editor__warning { display: block; margin: 12px 0; color: #b7791f; font-size: 12px; line-height: 1.6; }
.summary-mapping-editor__errors { display: grid; gap: 6px; margin-top: 12px; color: #e45656; font-size: 13px; }
@media (max-width: 768px) { .summary-mapping-editor__fields { grid-template-columns: minmax(0, 1fr); } .summary-mapping-editor__content { padding: 14px; } }
</style>
