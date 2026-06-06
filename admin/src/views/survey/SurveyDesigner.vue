<template>
  <div class="survey-main">
    <!-- 左侧导航 -->
    <div class="survey-main-navigator">
      <div class="nav-actions">
        <el-tooltip content="返回列表" placement="right">
          <el-button text @click="goBack" class="nav-btn">
            <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2"><path d="M19 12H5M12 19l-7-7 7-7"/></svg>
          </el-button>
        </el-tooltip>
        <div class="nav-divider" />
        <el-tooltip content="编辑" placement="right">
          <el-button text class="nav-btn" :class="{ active: activeView==='edit' }" @click="activeView='edit'">
            <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
          </el-button>
        </el-tooltip>
        <el-tooltip content="设置" placement="right">
          <el-button text class="nav-btn" :class="{ active: activeView==='setting' }" @click="activeView='setting'">
            <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"/></svg>
          </el-button>
        </el-tooltip>
        <el-tooltip content="投放" placement="right">
          <el-button text class="nav-btn" :class="{ active: activeView==='share' }" @click="activeView='share'">
            <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2"><circle cx="18" cy="5" r="3"/><circle cx="6" cy="12" r="3"/><circle cx="18" cy="19" r="3"/><line x1="8.59" y1="13.51" x2="15.42" y2="17.49"/><line x1="15.41" y1="6.51" x2="8.59" y2="10.49"/></svg>
          </el-button>
        </el-tooltip>
      </div>
      <div class="nav-bottom">
        <el-tooltip content="首页" placement="right">
          <el-button text class="nav-btn" @click="goBack">
            <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2"><path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/><polyline points="9 22 9 12 15 12 15 22"/></svg>
          </el-button>
        </el-tooltip>
      </div>
    </div>

    <!-- 主内容 -->
    <div class="survey-main-content">
      <!-- 编辑视图 -->
      <div v-show="activeView==='edit'" id="editor" class="survey-editor survey-light survey-app pc">
        <!-- 左: 题型面板侧边栏 -->
        <div class="survey-sidebar-panel">
          <div class="survey-sidebar-panel-tabs">
            <div class="survey-sidebar-panel-tabs-pane active" title="题目">
              <svg viewBox="0 0 1024 1024" width="20" height="20" fill="currentColor"><path d="M810.666667 128H213.333333c-46.933333 0-85.333333 38.4-85.333333 85.333333v597.333334c0 46.933333 38.4 85.333333 85.333333 85.333333h597.333334c46.933333 0 85.333333-38.4 85.333333-85.333333V213.333333c0-46.933333-38.4-85.333333-85.333333-85.333333z m-42.666667 554.666667H256v-85.333334h512v85.333334z m0-170.666667H256v-85.333333h512v85.333333z m0-170.666667H256V256h512v85.333333z"/></svg>
            </div>
            <div class="survey-sidebar-panel-tabs-pane" title="外观">
              <svg viewBox="0 0 1024 1024" width="20" height="20" fill="currentColor"><path d="M512 128c-212.053333 0-384 171.946667-384 384s171.946667 384 384 384 384-171.946667 384-384-171.946667-384-384-384z m0 682.666667c-164.693333 0-298.666667-133.973333-298.666667-298.666667s133.973333-298.666667 298.666667-298.666667 298.666667 133.973333 298.666667 298.666667-133.973333 298.666667-298.666667 298.666667zM512 341.333333c-94.293333 0-170.666667 76.373333-170.666667 170.666667s76.373333 170.666667 170.666667 170.666667 170.666667-76.373333 170.666667-170.666667-76.373333-170.666667-170.666667-170.666667z"/></svg>
            </div>
            <div class="survey-sidebar-panel-tabs-pane" title="逻辑">
              <svg viewBox="0 0 1024 1024" width="20" height="20" fill="currentColor"><path d="M632.888889 56.888889c0-31.288889-25.6-56.888889-56.888889-56.888889H455.111111c-31.288889 0-56.888889 25.6-56.888889 56.888889v120.888889H170.666667c-31.288889 0-56.888889 25.6-56.888889 56.888889v113.777777c0 31.288889 25.6 56.888889 56.888889 56.888889h120.888889v369.777778c0 31.288889 25.6 56.888889 56.888889 56.888889h113.777777c31.288889 0 56.888889-25.6 56.888889-56.888889V462.222222h113.777778c31.288889 0 56.888889-25.6 56.888889-56.888889V291.555556c0-31.288889-25.6-56.888889-56.888889-56.888889H632.888889V56.888889z"/></svg>
            </div>
          </div>
          <div class="survey-sidebar-panel-tabs-content">
            <div class="question-panel">
              <el-tabs v-model="activeCategory" class="type-tabs-card" :class="{ 'mode-exam': form.mode==='exam' }">
                <el-tab-pane v-for="cat in categories" :key="cat.name" :label="cat.label" :name="cat.name">
                  <div class="question-type">
                    <dl class="menu-group" v-for="(items, group) in groupedBySub" :key="group">
                      <dt class="menu-group-title">{{ group }}</dt>
                      <dd v-for="t in items" :key="t.type" class="menu-group-item" @click="addQuestion(t)">
                        <span class="itemIcon"><question-icon :type="t.type" /></span>
                        <span class="item-label">{{ t.displayName }}</span>
                        <span class="item-type">{{ t.type }}</span>
                      </dd>
                    </dl>
                  </div>
                </el-tab-pane>
              </el-tabs>
            </div>
          </div>
        </div>

        <!-- 中: 画布 -->
        <div class="survey-main-panel">
          <div class="survey-main-panel-toolbar">
            <el-input v-model="form.title" placeholder="问卷标题" class="title-input" maxlength="100" />
            <div class="toolbar-actions">
              <el-tag v-if="form.id" type="info" effect="plain" size="small">ID: {{ form.id }}</el-tag>
              <el-button :type="form.statusBool===1?'success':'danger'" plain size="small" @click="form.statusBool=form.statusBool===1?0:1">
                {{ form.statusBool===1?'已发布':'已停用' }}
              </el-button>
              <el-button type="primary" :loading="saving" size="small" @click="save">
                <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" style="margin-right:4px"><path d="M19 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11l5 5v11a2 2 0 0 1-2 2z"/><polyline points="17 21 17 13 7 13 7 21"/><polyline points="7 3 7 8 15 8"/></svg>
                保存
              </el-button>
            </div>
          </div>
          <div class="survey-main-panel-content">
            <div class="editor-wrapper">
              <div class="editor">
                <!-- 问卷头 -->
                <div class="header">
                  <div class="prefix">问卷</div>
                  <div class="header-content">
                    <div class="title">{{ form.title || '未命名问卷' }}</div>
                  </div>
                </div>

                <!-- 题目列表 -->
                <div class="questions-area">
                  <draggable-list :questions="questions" @update:questions="onQuestionsUpdate" @select="selectQuestion" :selected-id="selected?.id??null" />
                </div>

                <!-- 底部 -->
                <div class="footer">
                  <div class="footer-inner">感谢您的参与！</div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- 右: 属性面板 -->
        <div class="survey-setting-panel">
          <div v-if="selected" class="props-panel">
            <h3>{{ selected.title || '题目属性' }}</h3>
            <el-form label-position="top" size="small">
              <el-form-item label="ID"><el-input v-model="selected.id" :disabled="!!selected._existing" /></el-form-item>
              <el-form-item label="标题"><el-input v-model="selected.title" /></el-form-item>
              <el-form-item label="说明"><el-input v-model="selected.description" type="textarea" :rows="2" /></el-form-item>
              <el-form-item label="占位提示"><el-input v-model="selected.placeholder" /></el-form-item>
              <el-row :gutter="8">
                <el-col :span="12"><el-form-item label="必填"><el-switch v-model="selected.required" /></el-form-item></el-col>
                <el-col :span="12"><el-form-item label="只读"><el-switch v-model="selected.readOnly" /></el-form-item></el-col>
              </el-row>
              <el-form-item v-if="hasDataType(selected)" label="数据类型">
                <el-select v-model="selected.dataType" placeholder="不限制" clearable style="width:100%">
                  <el-option label="不限制" value="" />
                  <el-option label="数字" value="number" />
                  <el-option label="邮箱" value="email" />
                  <el-option label="手机号" value="mobile" />
                  <el-option label="身份证" value="idCard" />
                  <el-option label="日期" value="date" />
                  <el-option label="日期时间" value="dateTime" />
                  <el-option label="时间" value="time" />
                  <el-option label="中文" value="chinese" />
                  <el-option label="字母" value="alphabet" />
                </el-select>
              </el-form-item>
              <el-form-item v-if="hasOptions(selected)&&selected.props" label="选项">
                <OptionsEditor :model-value="selected.props.options||[]" @update:model-value="onOptionsUpdate" />
              </el-form-item>
              <el-divider border-style="dashed" style="margin:12px 0" />
              <el-collapse v-model="collapseActive" accordion>
                <el-collapse-item title="校验规则" name="validate">
                  <ValidateEditor :model-value="selected.validate||[]" @update:model-value="onValidateUpdate" />
                </el-collapse-item>
                <el-collapse-item title="计算表达式" name="calc">
                  <CalcEditor :model-value="selected.calcValue" :env="envFromAnswers" @update:model-value="onCalcUpdate" />
                </el-collapse-item>
                <el-collapse-item title="显示逻辑" name="logic">
                  <LogicEditor :model-value="selected.logic||[]" :all-questions="questions" @update:model-value="onLogicUpdate" />
                </el-collapse-item>
                <el-collapse-item v-if="form.mode==='exam'" title="考试属性" name="exam">
                  <el-form-item label="分值">
                    <el-input-number v-model="selected.examScore" :min="0" :step="0.5" style="width:100%" />
                  </el-form-item>
                  <el-form-item v-if="selected.examScore" label="正确答案">
                    <el-input v-model="selected.examCorrectAnswer" placeholder="填写正确答案" />
                  </el-form-item>
                  <el-form-item v-if="selected.examScore" label="答案解析">
                    <el-input v-model="selected.examAnalysis" type="textarea" :rows="2" placeholder="选填" />
                  </el-form-item>
                </el-collapse-item>
              </el-collapse>
              <el-button type="danger" text style="margin-top:12px;width:100%" @click="removeSelected">删除此题</el-button>
            </el-form>
          </div>
          <el-empty v-else description="选择题目编辑属性" :image-size="60" />
        </div>
      </div>

      <!-- 设置视图 -->
      <div v-show="activeView==='setting'" class="setting-wrapper">
        <div class="setting-scroll">
          <div class="setting-group">
            <div class="group-title">基础信息</div>
            <el-form :model="form" label-width="110px">
              <el-form-item label="描述"><el-input v-model="form.description" type="textarea" :rows="3" /></el-form-item>
              <el-row :gutter="16">
                <el-col :span="12"><el-form-item label="模式">
                  <el-radio-group v-model="form.mode">
                    <el-radio value="survey" border>问卷</el-radio>
                    <el-radio value="exam" border>考试</el-radio>
                  </el-radio-group>
                </el-form-item></el-col>
                <el-col :span="12"><el-form-item label="分类"><el-input v-model="form.category" /></el-form-item></el-col>
              </el-row>
              <el-form-item label="标签"><el-input v-model="form.tags" placeholder="逗号分隔" /></el-form-item>
            </el-form>
          </div>
          <div class="setting-group">
            <div class="group-title">答题设置</div>
            <el-form label-width="120px">
              <el-row :gutter="16">
                <el-col :span="8"><el-form-item label="显示题号"><el-switch v-model="form.questionNumber" /></el-form-item></el-col>
                <el-col :span="8"><el-form-item label="显示进度条"><el-switch v-model="form.progressBar" /></el-form-item></el-col>
                <el-col :span="8"><el-form-item label="自动暂存"><el-switch v-model="form.autoSave" /></el-form-item></el-col>
              </el-row>
              <el-row :gutter="16">
                <el-col :span="8"><el-form-item label="一页一题"><el-switch v-model="form.onePageOneQuestion" /></el-form-item></el-col>
                <el-col :span="8"><el-form-item label="显示答题卡"><el-switch v-model="form.answerSheetVisible" /></el-form-item></el-col>
                <el-col :span="8"><el-form-item label="允许复制"><el-switch v-model="form.copyEnabled" /></el-form-item></el-col>
              </el-row>
              <el-form-item label="填写密码"><el-input v-model="form.password" placeholder="留空不设密码" style="max-width:300px" /></el-form-item>
              <el-form-item label="校验触发">
                <el-radio-group v-model="form.triggerType">
                  <el-radio value="onInput">输入时</el-radio>
                  <el-radio value="onBlur">失焦时</el-radio>
                  <el-radio value="onSubmit">提交时</el-radio>
                </el-radio-group>
              </el-form-item>
            </el-form>
          </div>
          <div class="setting-group">
            <div class="group-title">提交后设置</div>
            <el-form label-width="120px">
              <el-row :gutter="16">
                <el-col :span="8"><el-form-item label="允许改答案"><el-switch v-model="form.enableUpdate" /></el-form-item></el-col>
                <el-col :span="8"><el-form-item label="显示成绩单"><el-switch v-model="form.transcriptVisible" /></el-form-item></el-col>
                <el-col :span="8"><el-form-item label="显示排行榜"><el-switch v-model="form.rankVisible" /></el-form-item></el-col>
              </el-row>
              <el-form-item label="跳转链接"><el-input v-model="form.redirectUrl" placeholder="提交后跳转的自定义 URL" /></el-form-item>
              <el-form-item label="结束语"><el-input v-model="form.endContent" type="textarea" :rows="3" placeholder="提交后显示的 HTML 内容" /></el-form-item>
            </el-form>
          </div>
          <div v-if="form.mode==='exam'" class="setting-group">
            <div class="group-title">考试设置</div>
            <el-form label-width="120px">
              <el-row :gutter="16">
                <el-col :span="8"><el-form-item label="显示排名"><el-switch v-model="form.examRankingEnabled" /></el-form-item></el-col>
                <el-col :span="8"><el-form-item label="练习模式"><el-switch v-model="form.exerciseMode" /></el-form-item></el-col>
                <el-col :span="8"><el-form-item label="随机顺序"><el-switch v-model="form.randomOrder" /></el-form-item></el-col>
              </el-row>
              <el-row :gutter="16">
                <el-col :span="12"><el-form-item label="最短交卷(分钟)"><el-input-number v-model="form.minSubmitMinutes" :min="0" style="width:100%" /></el-form-item></el-col>
                <el-col :span="12"><el-form-item label="最长交卷(分钟)"><el-input-number v-model="form.maxSubmitMinutes" :min="0" style="width:100%" /></el-form-item></el-col>
              </el-row>
            </el-form>
          </div>
          <div class="setting-group">
            <div class="group-title">访问权限</div>
            <el-form label-width="120px">
              <el-form-item label="可见性">
                <el-radio-group v-model="form.visibility">
                  <el-radio :value="0" border>公开链接</el-radio>
                  <el-radio :value="1" border>登录可见</el-radio>
                  <el-radio :value="2" border>部门限定</el-radio>
                </el-radio-group>
              </el-form-item>
              <el-form-item v-if="form.visibility===2" label="限定部门"><el-input v-model="form.deptIds" placeholder="部门ID, 逗号分隔" /></el-form-item>
              <el-row :gutter="16">
                <el-col :span="8"><el-form-item label="需登录"><el-switch v-model="form.loginRequired" /></el-form-item></el-col>
                <el-col :span="8"><el-form-item label="多次作答"><el-switch v-model="form.allowMultiBool" :active-value="1" :inactive-value="0" /></el-form-item></el-col>
                <el-col :span="8"><el-form-item label="匿名收集"><el-switch v-model="form.anonymousBool" :active-value="1" :inactive-value="0" /></el-form-item></el-col>
              </el-row>
              <el-row :gutter="16">
                <el-col :span="8"><el-form-item label="显示结果"><el-switch v-model="form.showResultBool" :active-value="1" :inactive-value="0" /></el-form-item></el-col>
              </el-row>
            </el-form>
          </div>
          <div class="setting-group">
            <div class="group-title">时间与配额</div>
            <el-form label-width="120px">
              <el-row :gutter="16">
                <el-col :span="12"><el-form-item label="开始时间"><el-date-picker v-model="form.startDate" type="datetime" placeholder="不限" value-format="x" style="width:100%" /></el-form-item></el-col>
                <el-col :span="12"><el-form-item label="结束时间"><el-date-picker v-model="form.endDate" type="datetime" placeholder="不限" value-format="x" style="width:100%" /></el-form-item></el-col>
              </el-row>
              <el-form-item label="最大答卷数"><el-input-number v-model="form.maxResponse" :min="0" style="width:200px" /><span class="hint">0=不限</span></el-form-item>
            </el-form>
          </div>
        </div>
      </div>

      <!-- 投放视图 -->
      <div v-show="activeView==='share'" class="setting-wrapper">
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

    <el-dialog v-model="exportDialogVisible" title="Schema JSON" width="700px">
      <el-input v-model="exportedJson" type="textarea" :rows="20" readonly style="font-family:monospace" />
      <template #footer>
        <el-button @click="copyJson">复制</el-button>
        <el-button type="primary" @click="exportDialogVisible=false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useRoute, useRouter } from 'vue-router'
import { adminApi } from '../../api'
import DraggableList from './formkit/DraggableList.vue'
import OptionsEditor from './formkit/OptionsEditor.vue'
import ValidateEditor from './formkit/ValidateEditor.vue'
import CalcEditor from './formkit/CalcEditor.vue'
import LogicEditor from './formkit/LogicEditor.vue'
import QuestionIcon from './formkit/QuestionIcon.vue'

const route = useRoute()
const router = useRouter()

const form = reactive<any>({
  id: 0, title: '', description: '', category: '', tags: '', mode: 'survey',
  visibility: 0, allowMultiBool: 0, anonymousBool: 0, showResultBool: 0,
  startDate: null, endDate: null, maxResponse: 0, statusBool: 1, deptIds: '',
  questionNumber: true, progressBar: false, autoSave: false,
  onePageOneQuestion: false, answerSheetVisible: true, copyEnabled: true,
  password: '', triggerType: 'onBlur', loginRequired: false,
  enableUpdate: false, transcriptVisible: true, rankVisible: false,
  redirectUrl: '', endContent: '',
  examRankingEnabled: false, exerciseMode: false, randomOrder: false,
  minSubmitMinutes: 0, maxSubmitMinutes: 0
})

const activeView = ref('edit')
const activeCategory = ref('')
const saving = ref(false)
const collapseActive = ref('')

const types = ref<any[]>([])
const categories = computed(() => {
  const m = new Map<string, string>()
  types.value.forEach(t => m.set(t.category, t.category))
  const defined = [
    { name: 'base', label: '基础' }, { name: 'select', label: '选择' },
    { name: 'media', label: '媒体' }, { name: 'layout', label: '布局' }, { name: 'advanced', label: '高级' }
  ]
  return defined.filter(c => m.has(c.name))
})
const typesByCategory = computed(() => {
  const m: Record<string, any[]> = {}
  for (const t of types.value) {
    if (!m[t.category]) m[t.category] = []
    m[t.category].push(t)
  }
  return m
})
const groupedBySub = computed(() => {
  const cat = activeCategory.value
  if (!cat || !typesByCategory.value[cat]) return {}
  const items = typesByCategory.value[cat]
  // 按子分类分组（后端可能返回 subCategory，暂无则全放一起）
  return { [cat]: items }
})

interface Question { id: string; type: string; title: string; description?: string; required: boolean; placeholder?: string; props?: any; validate?: any[]; logic?: any[]; calcValue?: any; readOnly?: boolean; dataType?: string; examScore?: number; examCorrectAnswer?: string; examAnalysis?: string; [key: string]: any }
const questions = ref<Question[]>([])
const selected = ref<Question | null>(null)
let idCounter = 0

function findById(id: string) { return questions.value.find(q => q.id === id) || null }
function genId() { idCounter++; return 'q' + idCounter }

function addQuestion(t: any) {
  const q: Question = { id: genId(), type: t.type, title: t.displayName, required: false, readOnly: false, dataType: '', placeholder: '', props: t.defaultProps ? JSON.parse(JSON.stringify(t.defaultProps)) : {} }
  if (t.type === 'judge') {
    if (!q.props) q.props = {}
    q.props.options = [{ label: '对', value: 'true' }, { label: '错', value: 'false' }]
  }
  if (t.type === 'select' || t.type === 'radio' || t.type === 'checkbox' || t.type === 'picker' || t.type === 'matrixRadio') {
    if (!q.props) q.props = {}
    if (!q.props.options) q.props.options = [{ value: '选项1', label: '选项1' }, { value: '选项2', label: '选项2' }]
  }
  questions.value.push(q)
  selected.value = findById(q.id)
}

function selectQuestion(id: string) { selected.value = findById(id) }
function onQuestionsUpdate(newQ: Question[]) {
  questions.value = newQ
  if (selected.value) {
    selected.value = newQ.find(q => q.id === selected.value!.id) || null
  }
}
async function removeSelected() {
  if (!selected.value) return
  try {
    await ElMessageBox.confirm('确定删除此题？', '确认', { confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning' })
  } catch { return }
  const sel = selected.value
  questions.value = questions.value.filter(q => q.id !== sel.id)
  selected.value = questions.value.length ? questions.value[questions.value.length - 1] : null
}
async function clearAll() {
  if (!questions.value.length) return
  try {
    await ElMessageBox.confirm(`确定清空全部 ${questions.value.length} 题？`, '确认', { confirmButtonText: '清空', cancelButtonText: '取消', type: 'warning' })
  } catch { return }
  questions.value = []; selected.value = null; idCounter = 0
}

function hasOptions(q: Question) { return ['select','radio','checkbox','picker','matrixRadio','matrixCheckbox','judge'].includes(q.type) }
function hasDataType(q: Question) { return ['input','textarea','number'].includes(q.type) }

function onOptionsUpdate(v: any) { if (selected.value && selected.value.props) selected.value.props.options = v }
function onValidateUpdate(v: any) { if (selected.value) selected.value.validate = v }
function onCalcUpdate(v: any) { if (selected.value) selected.value.calcValue = v }
function onLogicUpdate(v: any) { if (selected.value) selected.value.logic = v }

const envFromAnswers = computed(() => {
  const env: Record<string, any> = {}
  for (const q of questions.value) env[q.id] = undefined
  return env
})

const publicUrl = computed(() => form.id ? `${window.location.origin}/#/pages/survey/fill?id=${form.id}` : '')
const qrUrl = ref('')
const embedCode = computed(() => form.id ? `<iframe src="${publicUrl.value}" width="100%" height="600" frameborder="0"></iframe>` : '')

const exportDialogVisible = ref(false)
const exportedJson = ref('')

function copyLink() { navigator.clipboard.writeText(publicUrl.value); ElMessage.success('已复制') }
function genQR() { qrUrl.value = `https://api.qrserver.com/v1/create-qr-code/?size=200x200&data=${encodeURIComponent(publicUrl.value)}` }
function copyJson() { navigator.clipboard.writeText(exportedJson.value); ElMessage.success('已复制') }

async function load() {
  const id = Number(route.query.id || 0)
  if (!id) return
  try {
    const res: any = await adminApi.surveyDetail(id)
    const sv = res.data.survey
    const base: any = {
      id: sv.id, title: sv.title, description: sv.description,
      category: sv.category, tags: sv.tags, visibility: sv.visibility,
      allowMultiBool: sv.allowMulti, anonymousBool: sv.anonymous, showResultBool: sv.showResult,
      startDate: sv.startTime || null, endDate: sv.endTime || null,
      maxResponse: sv.maxResponse, statusBool: sv.status, deptIds: sv.deptIds,
      mode: sv.mode || 'survey'
    }
    if (sv.settings) {
      try {
        const s = JSON.parse(sv.settings)
        Object.assign(base, {
          questionNumber: s.questionNumber ?? true, progressBar: s.progressBar ?? false,
          autoSave: s.autoSave ?? false, password: s.password || '',
          loginRequired: s.loginRequired ?? false,
          onePageOneQuestion: s.onePageOneQuestion ?? false,
          answerSheetVisible: s.answerSheetVisible ?? true,
          copyEnabled: s.copyEnabled ?? true, triggerType: s.triggerType || 'onBlur',
          enableUpdate: s.enableUpdate ?? false, transcriptVisible: s.transcriptVisible ?? true,
          rankVisible: s.rankVisible ?? false, redirectUrl: s.redirectUrl || '',
          endContent: s.endContent || '',
          examRankingEnabled: s.examRankingEnabled ?? false,
          exerciseMode: s.exerciseMode ?? false, randomOrder: s.randomOrder ?? false,
          minSubmitMinutes: s.minSubmitMinutes || 0, maxSubmitMinutes: s.maxSubmitMinutes || 0
        })
      } catch {}
    } else {
      Object.assign(base, {
        questionNumber: true, progressBar: false, autoSave: false, password: '',
        loginRequired: false, onePageOneQuestion: false, answerSheetVisible: true,
        copyEnabled: true, triggerType: 'onBlur', enableUpdate: false,
        transcriptVisible: true, rankVisible: false, redirectUrl: '', endContent: '',
        examRankingEnabled: false, exerciseMode: false, randomOrder: false,
        minSubmitMinutes: 0, maxSubmitMinutes: 0
      })
    }
    Object.assign(form, base)
    const rawSchema = res.data.schema
    if (rawSchema) {
      try {
        const sch = JSON.parse(rawSchema)
        if (sch.questions) {
          questions.value = sch.questions
          idCounter = questions.value.length
          selected.value = questions.value.length > 0 ? questions.value[0] : null
        }
      } catch {}
    }
  } catch { ElMessage.error('加载失败') }
}

async function save() {
  if (!form.title) { ElMessage.warning('请填写标题'); return }
  saving.value = true
  try {
    const schema = JSON.stringify({ version: '2.0', questions: questions.value, setting: {} })
    const settings = JSON.stringify({
      questionNumber: form.questionNumber, progressBar: form.progressBar,
      autoSave: form.autoSave, password: form.password,
      loginRequired: form.loginRequired, onePageOneQuestion: form.onePageOneQuestion,
      answerSheetVisible: form.answerSheetVisible, copyEnabled: form.copyEnabled,
      triggerType: form.triggerType, enableUpdate: form.enableUpdate,
      transcriptVisible: form.transcriptVisible, rankVisible: form.rankVisible,
      redirectUrl: form.redirectUrl, endContent: form.endContent,
      examRankingEnabled: form.examRankingEnabled, exerciseMode: form.exerciseMode,
      randomOrder: form.randomOrder, minSubmitMinutes: form.minSubmitMinutes,
      maxSubmitMinutes: form.maxSubmitMinutes
    })
    const payload: any = {
      title: form.title, desc: form.description, category: form.category, tags: form.tags,
      visibility: form.visibility, allowMulti: form.allowMultiBool,
      anonymous: form.anonymousBool, showResult: form.showResultBool,
      startTime: form.startDate || 0, endTime: form.endDate || 0,
      maxResponse: form.maxResponse, status: form.statusBool,
      schema, deptIds: form.deptIds, mode: form.mode, settings
    }
    if (form.id) {
      payload.id = form.id
      await adminApi.surveyEdit(payload)
      ElMessage.success('已更新')
    } else {
      const r: any = await adminApi.surveyInsert(payload)
      form.id = r.data?.id || r.id
      ElMessage.success('已创建')
      router.replace({ query: { id: String(form.id) } })
    }
  } catch { ElMessage.error('保存失败') }
  finally { saving.value = false }
}

function goBack() { router.push('/survey') }

onMounted(async () => {
  try {
    const res: any = await adminApi.formkitTypes()
    types.value = res.data || []
    const cats = categories.value
    if (cats.length) activeCategory.value = cats[0].name
  } catch {}
  load()
})
</script>

<style scoped>
/* ======= 主布局 ======= */
.survey-main { display:flex; height:calc(100vh - 56px); background:#fff; }

/* ======= 左侧导航 ======= */
.survey-main-navigator {
  display:flex; flex:0 0 54px; flex-direction:column; justify-content:space-between;
  border-right:1px solid #e8e8e8; background:#fafafa; padding:8px 0; align-items:center;
}
.nav-actions { display:flex; flex-direction:column; align-items:center; gap:4px; }
.nav-bottom { display:flex; flex-direction:column; align-items:center; }
.nav-divider { width:20px; height:1px; background:#e0e0e0; margin:6px 0; }
.nav-btn {
  width:36px; height:36px; border-radius:8px; color:#999; font-size:13px;
  display:flex; align-items:center; justify-content:center;
}
.nav-btn:hover { background:#f0f0f0; color:#666; }
.nav-btn.active { background:#fff; color:#fb454c; box-shadow:0 1px 4px rgba(0,0,0,0.08); }

/* ======= 主内容 ======= */
.survey-main-content { position:relative; flex:1 1 auto; overflow:hidden; }

/* ======= 编辑器 ======= */
.survey-editor { display:flex; height:100%; font-family:"PingFang SC",-apple-system,"Helvetica Neue",Helvetica,BlinkMacSystemFont,"Microsoft YaHei",tahoma,Arial,"Open Sans","Hiragino Sans GB","Heiti SC","WenQuanYi Micro Hei",sans-serif; }

/* ======= 左侧题型面板 ======= */
.survey-sidebar-panel { position:relative; z-index:2; display:flex; flex-grow:0; flex-shrink:0; user-select:none; }
.survey-sidebar-panel-tabs {
  display:flex; flex-direction:column; font-size:12px;
  background:#f5f5f5; border-right:1px solid #e8e8e8;
}
.survey-sidebar-panel-tabs-pane {
  display:flex; flex-direction:column; align-items:center; justify-content:center;
  width:40px; height:40px; color:#999; cursor:pointer; border-radius:0;
  margin:2px; transition:all 0.15s;
}
.survey-sidebar-panel-tabs-pane:hover { color:#666; }
.survey-sidebar-panel-tabs-pane.active { color:#fb454c; background:#fff; border-radius:6px; }
.survey-sidebar-panel-tabs-content {
  display:flex; flex-direction:column; width:240px; height:100%;
  border-right:1px solid #e8e8e8; overflow-y:auto;
}

/* 题型分类标签（卡片式） */
.question-panel { padding:8px; }
.type-tabs-card { --el-tabs-header-height:28px; }
.type-tabs-card :deep(.el-tabs__header) { margin:0 0 8px; }
.type-tabs-card :deep(.el-tabs__nav) { display:flex; justify-content:center; }
.type-tabs-card :deep(.el-tabs__item) {
  padding:0 10px; font-size:12px; line-height:26px; height:26px;
  border:1px solid #e8e8e8; border-radius:4px; margin:0 2px; color:#666;
}
.type-tabs-card :deep(.el-tabs__item.is-active) {
  color:#fff; background:#fb454c; border-color:#fb454c; font-weight:500;
}
.type-tabs-card :deep(.el-tabs__active-bar) { display:none; }
.type-tabs-card :deep(.el-tabs__nav-wrap::after) { display:none; }
.type-tabs-card :deep(.el-tabs__header) { border-bottom:none; }

.question-type { padding:4px 0; }
.menu-group { margin:0 0 12px; }
.menu-group-title {
  font-size:12px; font-weight:500; color:#888; margin-bottom:6px; padding:0 4px;
}
.menu-group-item {
  display:flex; align-items:center; gap:6px; padding:6px 8px;
  border-radius:6px; cursor:pointer; transition:all 0.12s; line-height:20px;
}
.menu-group-item:hover { background:#fff5f5; color:#fb454c; }
.itemIcon { flex-shrink:0; font-size:16px; width:20px; text-align:center; color:#999; }
.menu-group-item:hover .itemIcon { color:#fb454c; }
.item-label { font-size:13px; flex:1; }
.item-type { font-size:10px; color:#bbb; }

/* ======= 中间画布 ======= */
.survey-main-panel {
  display:flex; flex:1 1; flex-direction:column; overflow:hidden;
  background-color:#f7f8fa; padding:0;
}
.survey-main-panel-toolbar {
  display:flex; align-items:center; justify-content:space-between;
  padding:8px 16px; background:#fff; border-bottom:1px solid #e8e8e8; gap:12px;
}
.title-input { font-size:16px; font-weight:600; max-width:360px; flex:1; }
.title-input :deep(.el-input__wrapper) { box-shadow:none!important; padding-left:0; }
.title-input :deep(.el-input__inner) { font-size:16px; font-weight:600; }
.toolbar-actions { display:flex; align-items:center; gap:8px; flex-shrink:0; }

.survey-main-panel-content { flex:1; overflow-y:auto; padding:20px 40px; }
.editor-wrapper { max-width:720px; margin:0 auto; }
.editor { background:#fff; border-radius:12px; box-shadow:0 2px 12px rgba(0,0,0,0.06); overflow:hidden; }

/* 问卷头 */
.header {
  padding:24px 28px 12px; text-align:center; border-bottom:1px solid #f0f0f0;
}
.prefix { font-size:13px; color:#999; margin-bottom:4px; }
.header-content .title { font-size:22px; font-weight:600; color:#303133; line-height:1.4; }

.questions-area { padding:16px 24px; }
.questions-area :deep(.draggable-list) { gap:8px; }

.footer {
  padding:24px 28px; text-align:center; border-top:1px solid #f0f0f0;
}
.footer-inner { font-size:13px; color:#999; }

/* ======= 右侧属性面板 ======= */
.survey-setting-panel { width:265px; flex-shrink:0; border-left:1px solid #e8e8e8; background:#fafafa; overflow-y:auto; }
.props-panel { padding:12px; }
.props-panel h3 { font-size:14px; font-weight:500; color:#fb454c; margin:0 0 8px; padding-bottom:8px; border-bottom:2px solid #fb454c; }
.props-panel :deep(.el-form-item) { margin-bottom:8px; }
.props-panel :deep(.el-form-item__label) { font-size:12px; color:#666; padding-bottom:2px; font-weight:500; line-height:1.2; }
.props-panel :deep(.el-divider) { margin:8px 0; }
.props-panel :deep(.el-collapse-item__header) { font-size:12px; font-weight:500; }

/* ======= 设置/投放视图 ======= */
.setting-wrapper { height:100%; overflow-y:auto; background:#f5f6f8; }
.setting-scroll { max-width:860px; margin:0 auto; padding:20px 24px; }
.setting-group { margin-bottom:20px; padding:16px 20px; background:#fff; border-radius:8px; box-shadow:0 1px 4px rgba(0,0,0,0.04); }
.group-title { font-size:13px; font-weight:500; color:#888; margin-bottom:12px; letter-spacing:0.5px; }
.hint { font-size:12px; color:#aaa; margin-left:8px; }
.save-tip { padding:60px 0; text-align:center; color:#aaa; font-size:14px; }
.share-card { max-width:600px; }
.share-title { font-size:13px; font-weight:500; color:#555; margin:16px 0 8px; }
.share-row { margin-bottom:8px; }
.qr-box { margin-top:12px; display:inline-block; padding:12px; background:#fff; border:1px solid #f0f0f0; border-radius:12px; }
</style>
