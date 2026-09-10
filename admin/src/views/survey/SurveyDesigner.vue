<template>
  <div v-loading="designerLoading" class="survey-main">
    <!-- 最左侧导航: 概述 / 编辑 / 设置 / 数据 / 报表 / 项目 -->
    <div class="survey-main-navigator">
      <div class="nav-actions">
        <!--<button class="nav-btn" :class="{ active: activeView==='overview' }" @click="activeView='overview'" title="概述">
          <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="16" y1="13" x2="8" y2="13"/><line x1="16" y1="17" x2="8" y2="17"/><polyline points="10 9 9 9 8 9"/></svg>
        </button>-->
        <button class="nav-btn" :class="{ active: activeView==='edit' }" @click="activeView='edit'" title="编辑">
          <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
        </button>
        <button class="nav-btn" :class="{ active: activeView==='setting' }" @click="activeView='setting'" title="设置">
          <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"/></svg>
        </button>
        <button class="nav-btn" :class="{ active: activeView==='data' }" @click="activeView='data'" title="数据">
          <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2"><ellipse cx="12" cy="5" rx="9" ry="3"/><path d="M21 12c0 1.66-4 3-9 3s-9-1.34-9-3"/><path d="M3 5v14c0 1.66 4 3 9 3s9-1.34 9-3V5"/></svg>
        </button>
        <button class="nav-btn" :class="{ active: activeView==='report' }" @click="activeView='report'" title="报表">
          <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="20" x2="18" y2="10"/><line x1="12" y1="20" x2="12" y2="4"/><line x1="6" y1="20" x2="6" y2="14"/></svg>
        </button>
        <button class="nav-btn" @click="goBack" title="返回列表">
          <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2"><path d="M19 12H5M12 19l-7-7 7-7"/></svg>
        </button>
      </div>
    </div>

    <!-- 主内容 -->
    <div class="survey-main-content">
      <div v-if="designerLoadError" class="designer-load-error">
        <el-result icon="error" title="问卷加载失败" sub-title="未能读取当前问卷，请检查网络后重试">
          <template #extra><el-button @click="goBack">返回列表</el-button><el-button type="primary" @click="retryLoad">重新加载</el-button></template>
        </el-result>
      </div>
      <!-- 编辑视图 -->
      <div v-show="activeView==='edit'" id="editor" class="survey-editor survey-light survey-app pc" :class="{ 'has-open-settings': selected || showSurveySettings }">
        <SurveyDesignerSidebar
          ref="sidebarRef"
          v-model="middleTab"
          :form="form"
          :types="types"
          :questions="questions"
          :selected-id="selected?.id || null"
          :invalid-ids="invalidQuestionIds"
          :type-name="typeName"
          @add="addQuestion"
          @add-bank="addFromBank"
          @select="selectQuestion($event, true)"
          @save="save"
        />

        <SurveyDesignerLogicPanel
          v-if="middleTab === 'logic'"
          v-model="logicRuleList"
          :questions="questions"
          :survey-exists="Boolean(form.id)"
          @save="saveLogicRules"
        />

        <SurveyDesignerResultSettings
          v-if="middleTab === 'result'"
          :form="form"
          :saving="saving"
          @save="save"
          @open-report="goToStatReport"
        />

        <!-- 中: 画布 -->
        <div v-show="middleTab!=='logic' && middleTab!=='result'" class="survey-main-panel" @click.self="deselectQuestion">
          <div class="survey-main-panel-toolbar">
            <div class="toolbar-left">
              <el-button-group class="toolbar-btn-group">
                <el-tooltip content="撤销" placement="bottom">
                  <el-button text size="small" class="toolbar-btn" :disabled="!canUndo" @click="undoDesignerChange">
                    <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><polyline points="1 4 1 10 7 10"/><path d="M3.51 15a9 9 0 1 0 2.13-9.36L1 10"/></svg>
                  </el-button>
                </el-tooltip>
                <el-tooltip content="重做" placement="bottom">
                  <el-button text size="small" class="toolbar-btn" :disabled="!canRedo" @click="redoDesignerChange">
                    <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><polyline points="23 4 23 10 17 10"/><path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/></svg>
                  </el-button>
                </el-tooltip>
                <el-tooltip content="快捷键" placement="bottom">
                  <el-button text size="small" class="toolbar-btn" @click="showShortcutHelp">
                    <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><rect x="2" y="4" width="20" height="16" rx="2"/><path d="M6 8h.01M10 8h.01M14 8h.01M18 8h.01M8 12h.01M12 12h.01M16 12h.01M6 16h.01M10 16h.01M14 16h.01M18 16h.01"/></svg>
                  </el-button>
                </el-tooltip>
                <el-tooltip content="文本导入" placement="bottom">
                  <el-button text size="small" class="toolbar-btn" @click="openTextImport">
                    <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/></svg>
                  </el-button>
                </el-tooltip>
                <el-tooltip content="导出问卷" placement="bottom">
                  <el-button text size="small" class="toolbar-btn" @click="exportSurvey">
                    <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
                  </el-button>
                </el-tooltip>
              </el-button-group>
            </div>
            <div class="toolbar-center">
              <div class="toolbar-title-row">
                <span class="toolbar-title">{{ form.title || '未命名问卷' }}</span>
                <span class="toolbar-mode">{{ form.mode === 'exam' ? '考试' : '问卷' }}</span>
              </div>
              <div class="toolbar-meta">
                <span>总题数 {{ questions.length }}</span>
                <span>可答 {{ answerableQuestionCount }}</span>
                <span>{{ panelModeLabel }}</span>
                <span class="toolbar-save-status" :class="saveStatusClass">{{ saveStatusLabel }}</span>
              </div>
            </div>
            <div class="toolbar-right">
              <el-button-group class="toolbar-btn-group">
                <el-tooltip content="编辑" placement="bottom">
                  <el-button text size="small" class="toolbar-btn" :class="{ active: panelMode==='edit' }" @click="panelMode='edit'">
                    <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
                  </el-button>
                </el-tooltip>
                <el-tooltip content="JSON" placement="bottom">
                  <el-button text size="small" class="toolbar-btn" :class="{ active: panelMode==='json' }" @click="panelMode='json'; exportedJson = JSON.stringify({ version: '2.0', questions: questions, setting: {} }, null, 2)">
                    <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><polyline points="16 18 22 12 16 6"/><polyline points="8 6 2 12 8 18"/></svg>
                  </el-button>
                </el-tooltip>
                <el-tooltip content="预览" placement="bottom">
                  <el-button text size="small" class="toolbar-btn" :class="{ active: panelMode==='preview' }" @click="panelMode='preview'">
                    <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/><circle cx="12" cy="12" r="3"/></svg>
                  </el-button>
                </el-tooltip>
              </el-button-group>
              <span class="toolbar-divider" />
              <el-button type="primary" :loading="saving" size="small" @click="save">
                <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" style="margin-right:4px"><path d="M19 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11l5 5v11a2 2 0 0 1-2 2z"/><polyline points="17 21 17 13 7 13 7 21"/><polyline points="7 3 7 8 15 8"/></svg>
                保存
              </el-button>
            </div>
          </div>
          <SurveyDesignerIssues v-if="!designerLoading" :issues="designerIssues" @locate="locateDesignerIssue" />
          <div class="survey-main-panel-content">
            <div v-if="panelMode==='edit'" class="editor-wrapper">
              <div class="editor" :style="{ backgroundImage: form.backgroundImages?.length ? 'url('+form.backgroundImages[0].url+')' : 'none', backgroundSize: 'cover', backgroundPosition: 'center' }">
                <!-- 页眉图（撑满顶部） -->
                <div v-if="form.headerImages?.length" class="editor-header-img" @click="openSurveySettings">
                  <img :src="form.headerImages[0].url" />
                </div>
                <!-- 问卷头 -->
                <div class="header">
                  <div class="header-content">
                    <el-input v-model="form.title" placeholder="问卷标题" class="header-title-input" maxlength="100" @focus="openSurveySettings" />
                    <div v-if="form.description" class="header-desc">{{ form.description }}</div>
                  </div>
                </div>

                <!-- 题目列表 -->
                <div class="questions-area" @click.self="deselectQuestion">
                  <draggable-list :questions="questions" :invalid-ids="invalidQuestionIds" @update:questions="onQuestionsUpdate" @select="selectQuestion" :selected-id="selected?.id??null" editing @remove="removeQuestionById" @select-option="selectOption" @upload-bank="onUploadBank" />
                  <div v-if="!questions.length" class="designer-empty-canvas">
                    <div class="empty-canvas-icon">+</div>
                    <div class="empty-canvas-title">未添加题目</div>
                    <div class="empty-canvas-desc">可以从左侧题型库添加，也可以用文本导入批量生成。</div>
                    <div class="empty-canvas-actions">
                      <el-button size="small" type="primary" @click.stop="addQuestion({ type: 'radio', displayName: '单选题' })">添加单选题</el-button>
                      <el-button size="small" @click.stop="openTextImport">文本导入</el-button>
                    </div>
                  </div>
                </div>

                <!-- 底部 -->
                <div class="footer">
                  <div class="footer-inner">感谢您的参与！</div>
                </div>
              </div>
            </div>
            <div v-else-if="panelMode==='json'" class="json-panel" style="display:flex;flex-direction:column;height:100%">
              <div style="display:flex;gap:6px;padding:6px;flex-shrink:0">
                <el-button size="small" @click="exportSurveyKingJson">导出 SurveyKing JSON</el-button>
                <el-button size="small" @click="loadSurveyKingJson">导入 SurveyKing JSON</el-button>
                <el-button size="small" @click="downloadJson">下载 JSON</el-button>
              </div>
              <el-input v-model="exportedJson" type="textarea" :rows="20" readonly style="font-family:monospace;flex:1" />
            </div>
            <div v-else-if="panelMode==='preview'" class="survey-preview-panel" :style="{ backgroundImage: form.backgroundImages?.length ? 'url(\''+(typeof form.backgroundImages[0]==='string'?form.backgroundImages[0]:form.backgroundImages[0].url)+'\')' : 'none', backgroundSize: 'cover', backgroundPosition: 'center' }">
              <div v-if="form.headerImages?.length" class="preview-header-img"><img :src="typeof form.headerImages[0]==='string'?form.headerImages[0]:form.headerImages[0].url" /></div>
              <div class="survey-preview-header">
                <div class="preview-form-title">{{ form.title || '未命名问卷' }}</div>
                <div v-if="form.description" class="preview-form-desc">{{ form.description }}</div>
              </div>
              <div class="survey-preview-body">
                <template v-for="(q, i) in questions" :key="q.id">
                  <div v-if="!q.defaultHidden" class="preview-question-card">
                    <QuestionPreview :q="q" />
                  </div>
                </template>
                <el-empty v-if="!questions.length" description="暂无题目" :image-size="60" />
              </div>
              <div class="survey-preview-footer">感谢您的参与！</div>
            </div>
          </div>
        </div>

        <SurveyDesignerQuestionProperties
          v-show="middleTab !== 'logic' && middleTab !== 'result'"
          v-model="selected"
          v-model:selected-opt-idx="selectedOptIdx"
          v-model:show-survey-settings="showSurveySettings"
          :form="form"
          :context="selectedQuestionContext"
          :type-name="typeName"
          :generate-id="genId"
          @remove="removeSelected"
          @open-formula="openFormulaDialog"
          @save="save"
        />
      </div>

      <SurveyDesignerSettingsPanel
        v-show="activeView === 'setting'"
        v-model="completionAction"
        v-model:form="form"
        :active="activeView === 'setting'"
        :saving="saving"
        :save-status-label="saveStatusLabel"
        :save-status-tag-type="saveStatusTagType"
        :public-url="publicUrl"
        :qr-url="qrUrl"
        :fill-templates="fillTemplates"
        :has-invalid-time-range="hasInvalidTimeRange"
        @save="save"
        @copy-link="copyLink"
        @generate-qr="genQR"
        @download-qr="downloadQR"
      />

      <!-- 数据视图 -->
      <SurveyDesignerResponsesPanel
        v-show="activeView === 'data'"
        :active="activeView === 'data'"
        :survey-id="form.id"
        :questions="questions"
      />

      <!-- 报表视图 -->
      <SurveyDesignerReportPanel
        v-show="activeView === 'report'"
        :active="activeView === 'report'"
        :survey-id="form.id"
      />

      <!-- 项目视图（原投放） -->
      <div v-show="activeView==='project'" class="setting-wrapper">
        <div class="setting-scroll">
          <div v-if="!form.id" class="save-tip">请先保存问卷</div>
          <template v-else>
            <div class="share-card">
              <div class="share-title">访问链接</div>
              <div class="share-row"><el-input v-model="publicUrl" readonly><template #append><el-button @click="copyLink">复制</el-button></template></el-input></div>
              <div class="share-title">二维码</div>
              <el-button size="small" @click="genQR">生成二维码</el-button>
              <div v-if="qrUrl" class="qr-box"><el-image :src="qrUrl" style="width:180px;height:180px;border-radius:8px" /></div>
              <div class="share-title">嵌入代码</div>
              <el-input v-model="embedCode" type="textarea" :rows="3" readonly />
            </div>
          </template>
        </div>
      </div>
    </div>
  </div>

  <SurveyDesignerFormulaDialog ref="formulaDialogRef" :questions="questions" @save="saveFormula" />
  <SurveyDesignerBankUploadDialog ref="bankUploadDialogRef" @confirm="confirmUploadBank" />
  <SurveyDesignerTextImportDialog ref="textImportDialogRef" @import="importTextQuestions" />
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { useRoute, useRouter } from 'vue-router'
import { adminApi } from '../../api'
import DraggableList from './formkit/DraggableList.vue'

import QuestionPreview from './formkit/QuestionPreview.vue'
import SurveyDesignerIssues from './components/SurveyDesignerIssues.vue'
import SurveyDesignerQuestionProperties from './components/SurveyDesignerQuestionProperties.vue'
import SurveyDesignerSidebar from './components/SurveyDesignerSidebar.vue'
import SurveyDesignerResponsesPanel from './components/SurveyDesignerResponsesPanel.vue'
import SurveyDesignerReportPanel from './components/SurveyDesignerReportPanel.vue'
import SurveyDesignerLogicPanel from './components/SurveyDesignerLogicPanel.vue'
import SurveyDesignerSettingsPanel from './components/SurveyDesignerSettingsPanel.vue'
import SurveyDesignerResultSettings from './components/SurveyDesignerResultSettings.vue'
import SurveyDesignerFormulaDialog from './components/SurveyDesignerFormulaDialog.vue'
import SurveyDesignerTextImportDialog from './components/SurveyDesignerTextImportDialog.vue'
import SurveyDesignerBankUploadDialog from './components/SurveyDesignerBankUploadDialog.vue'
import { FALLBACK_TYPES, getSurveyQuestionTypeName as typeName } from './survey-designer-question-types'
import type { SurveyCompletionAction, SurveyFormulaPayload, SurveyLogicRule } from './survey-designer-types'
import { collectSurveyDesignerIssues, type SurveyDesignerIssue } from './survey-designer-validation'
import { createSurveyDesignerForm, useSurveyDesignerDocument } from './composables/useSurveyDesignerDocument'
import { useSurveyDesignerQuestions } from './composables/useSurveyDesignerQuestions'
import { useSurveyDesignerTransfer } from './composables/useSurveyDesignerTransfer'
import './survey-designer-responsive.css'

const route = useRoute()
const router = useRouter()

const form = createSurveyDesignerForm()

const completionAction = ref<SurveyCompletionAction>('default')

const activeView = ref('edit')
const panelMode = ref<'edit' | 'json' | 'preview'>('edit')
const panelModeLabel = computed(() => ({ edit: '编辑模式', json: 'JSON模式', preview: '预览模式' }[panelMode.value]))
const middleTab = ref('item')
const showSurveySettings = ref(false)

interface FormulaDialogHandle { open: (type: string, initialText?: string) => void }
interface SidebarHandle { loadBank: () => void; loadResources: () => void; showOutline: () => void }
const formulaDialogRef = ref<FormulaDialogHandle | null>(null)
const sidebarRef = ref<SidebarHandle | null>(null)

function openFormulaDialog(type: string) {
  if (!selected.value) return
  const fieldMap: Record<string, string> = {
    endFormula: 'endFormula', jumpFormula: 'jumpFormula',
    requiredFormula: 'requiredFormula', textReplace: 'replaceTextRule',
    optShow: 'optShowFormula', optAutoCheck: 'optAutoCheckFormula',
    calculate: 'calculateFormula',
  }
  const field = fieldMap[type] || ''
  formulaDialogRef.value?.open(type, selected.value[field] || selected.value.props?.[field] || '')
}
function saveFormula(payload: SurveyFormulaPayload) {
  if (!selected.value) return
  const fieldMap: Record<string, string> = {
    endFormula: 'endFormula', jumpFormula: 'jumpFormula',
    requiredFormula: 'requiredFormula', textReplace: 'replaceTextRule',
    optShow: 'optShowFormula', optAutoCheck: 'optAutoCheckFormula',
  }
  const field = fieldMap[payload.type] || ''
  if (['optShow', 'optAutoCheck', 'textReplace'].includes(payload.type)) {
    if (!selected.value.props) selected.value.props = {}
    selected.value.props[field] = payload.text
  } else if (payload.type === 'calculate') {
    if (!selected.value.props) selected.value.props = {}
    selected.value.props.calculateFormula = payload.text
  } else {
    selected.value[field] = payload.text
  }
  ElMessage.success('公式已保存')
}
const logicRuleList = ref<SurveyLogicRule[]>([])

const types = ref<any[]>([])
const {
  questions, selected, selectedOptIdx, selectedQuestionContext, answerableQuestionCount,
  genId, syncIdCounterFromQuestions, addQuestion, selectQuestion, selectOption,
  onQuestionsUpdate, removeSelected, clearAll, replaceQuestions, removeQuestionById,
  deselectQuestion, openSurveySettings,
} = useSurveyDesignerQuestions({ logicRules: logicRuleList, panelMode, showSurveySettings })

const designerIssues = computed(() => collectSurveyDesignerIssues({ form, questions: questions.value }))
const invalidQuestionIds = computed(() => Array.from(new Set(designerIssues.value.filter(issue => issue.severity === 'error' && issue.questionId).map(issue => issue.questionId!))))

const {
  textImportDialogRef, bankUploadDialogRef, exportedJson, fillTemplates,
  publicUrl, qrUrl, embedCode, importTextQuestions, openTextImport, exportSurvey,
  addFromBank, onUploadBank, confirmUploadBank, genQR, copyLink, downloadQR,
  exportSurveyKingJson, loadSurveyKingJson, downloadJson,
} = useSurveyDesignerTransfer({
  form,
  questions,
  selected,
  selectedOptIdx,
  generateId: genId,
  replaceQuestions,
  syncIdCounter: syncIdCounterFromQuestions,
  reloadBank: () => sidebarRef.value?.loadBank(),
})

const {
  saving, designerLoading, designerLoadError, hasInvalidTimeRange,
  saveStatusLabel, saveStatusClass, saveStatusTagType, canUndo, canRedo,
  load, retryLoad, save, saveLogicRules, undoDesignerChange,
  redoDesignerChange, showShortcutHelp,
} = useSurveyDesignerDocument({
  form,
  questions,
  selected,
  logicRules: logicRuleList,
  completionAction,
  issues: designerIssues,
  route,
  router,
  locateIssue: locateDesignerIssue,
  syncQuestionIds: syncIdCounterFromQuestions,
  reloadResources: () => sidebarRef.value?.loadResources(),
})

function locateDesignerIssue(issue: SurveyDesignerIssue) {
  const question = issue.questionId ? questions.value.find(item => item.id === issue.questionId) : questions.value[issue.questionIndex ?? -1]
  if (question) {
    activeView.value = 'edit'; panelMode.value = 'edit'; sidebarRef.value?.showOutline()
    selectQuestion(question.id, true)
    return
  }
  if (issue.section === 'edit') { activeView.value = 'edit'; panelMode.value = 'edit'; openSurveySettings() }
  else activeView.value = 'setting'
}
function goBack() { router.push('/survey') }
function goToStatReport() {
  const id = form.id || Number(route.query.id || 0)
  if (!id) {
    ElMessage.warning('请先保存问卷')
    return
  }
  window.open(`/survey/stat-report?surveyId=${id}&title=${encodeURIComponent(form.title || '')}`, '_blank')
}

onMounted(async () => {
  try {
    const res: any = await adminApi.formkitTypes()
    const apiTypes: any[] = res.data || []
    const apiMap = new Map(apiTypes.map((t:any) => [t.type, t]))
    // Merge: use API displayName if available, otherwise use fallback
    types.value = FALLBACK_TYPES.map(ft => {
      const api = apiMap.get(ft.type)
      return api || ft
    })
  } catch {
    types.value = FALLBACK_TYPES
  }
  await load()
})
</script>

<style scoped src="./survey-designer.css"></style>
