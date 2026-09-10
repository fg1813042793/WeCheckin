<template>
  <div class="survey-main">
    <div class="survey-main-navigator">
      <div class="nav-actions">
        <button class="nav-btn" :class="{ active: activeView==='edit' }" @click="activeView='edit'" title="编辑">
          <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
        </button>
        <button class="nav-btn" :class="{ active: activeView==='setting' }" @click="activeView='setting'" title="设置">
          <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"/></svg>
        </button>
        <button class="nav-btn" :class="{ active: activeView==='data' }" @click="activeView='data'" title="数据">
          <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2"><ellipse cx="12" cy="5" rx="9" ry="3"/><path d="M21 12c0 1.66-4 3-9 3s-9-1.34-9-3"/><path d="M3 5v14c0 1.66 4 3 9 3s9-1.34 9-3V5"/></svg>
        </button>
        <button class="nav-btn" @click="goBack" title="返回列表">
          <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2"><path d="M19 12H5M12 19l-7-7 7-7"/></svg>
        </button>
      </div>
    </div>

    <div class="survey-main-content">
      <div v-show="activeView==='edit'" id="editor" class="survey-editor survey-light survey-app pc">
        <ExamDesignerSidebar
          ref="sidebarRef"
          v-model="middleTab"
          :form="form"
          :types="types"
          :questions="questions"
          :selected-id="selected?.id || null"
          :type-name="typeName"
          @add="addQuestion"
          @add-bank="addFromBank"
          @select="selectQuestion($event, true)"
          @save="save"
        />

        <ExamDesignerLogicPanel v-if="middleTab === 'logic'" v-model="logicRuleList" :questions="questions" @save="saveLogicRules" />

        <div v-show="middleTab!=='logic'" class="survey-main-panel" @click.self="deselectQuestion">
          <div class="survey-main-panel-toolbar">
            <div class="toolbar-left">
              <span class="toolbar-score">试卷总分：<strong>{{ paperTotalScore }}</strong> 分</span>
              <span class="toolbar-divider" />
              <el-button-group class="toolbar-btn-group">
                <el-tooltip content="文本导入" placement="bottom">
                  <el-button text size="small" class="toolbar-btn" @click="openTextImport">
                    <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
                  </el-button>
                </el-tooltip>
                <el-tooltip content="导出试卷" placement="bottom">
                  <el-button text size="small" class="toolbar-btn" @click="exportExam">
                    <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/></svg>
                  </el-button>
                </el-tooltip>
              </el-button-group>
            </div>
            <div class="toolbar-right">
              <el-button-group class="toolbar-btn-group">
                <el-tooltip content="编辑" placement="bottom">
                  <el-button text size="small" class="toolbar-btn" :class="{ active: panelMode==='edit' }" @click="panelMode='edit'">
                    <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
                  </el-button>
                </el-tooltip>
                <el-tooltip content="JSON" placement="bottom">
                  <el-button text size="small" class="toolbar-btn" :class="{ active: panelMode==='json' }" @click="panelMode='json'; prepareJson()">
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
                <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" style="margin-right:4px"><path d="M19 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11l5 5v11a2 2 0 0 1-2 2z"/><polyline points="17 21 17 13 7 13 7 21"/><polyline points="7 3 7 8 15 8"/></svg>保存
              </el-button>
            </div>
          </div>
          <div v-if="panelMode==='edit'" class="survey-main-panel-content">
            <div class="editor-wrapper">
              <div class="editor">
                <div class="header">
                  <div class="header-content">
                    <el-input v-model="form.title" placeholder="考试标题" class="header-title-input" maxlength="100" />
                  </div>
                </div>
                <div class="questions-area">
                  <draggable-list :questions="questions" @update:questions="onQuestionsUpdate" @select="selectQuestion" :selected-id="selected?.id??null" editing @remove="removeQuestionById" @select-option="selectOption" @upload-bank="onUploadBank" />
                </div>
                <div class="footer">
                  <div class="footer-inner">感谢您的参与！</div>
                </div>
              </div>
            </div>
          </div>
          <div v-else-if="panelMode==='json'" class="json-panel" style="display:flex;flex-direction:column;height:100%">
            <div style="padding:12px;display:flex;gap:8px;flex-shrink:0">
              <el-button size="small" @click="downloadJson">下载 JSON</el-button>
              <el-button size="small" @click="loadJson">导入 JSON</el-button>
            </div>
            <div style="flex:1;overflow:auto;padding:0 12px 12px">
              <el-input v-model="exportedJson" type="textarea" :rows="20" style="font-family:monospace;font-size:12px" />
            </div>
          </div>
          <div v-else-if="panelMode==='preview'" class="survey-preview-panel">
            <div style="padding:12px">
              <div class="header" style="margin-bottom:12px">
                <div class="header-content" style="font-size:18px;font-weight:600;text-align:center;padding:12px">{{ form.title }}</div>
              </div>
              <draggable-list :questions="questions" @select="selectQuestion" :selected-id="selected?.id??null" @upload-bank="onUploadBank" />
            </div>
          </div>
        </div>

        <div v-show="middleTab!=='logic'" class="survey-setting-panel" @click.self="deselectQuestion">
          <div v-if="selected" class="props-panel" :key="selected?.id || 'none'">
            <!-- 文件上传设置 -->
            <template v-if="selectedOptIdx>=0 && selected.type==='file'">
              <h3>文件上传设置</h3>
              <el-form label-position="top" size="small">
                <el-form-item label="允许类型">
                  <el-checkbox-group v-model="selected.fileTypes">
                    <el-checkbox value="image">图片</el-checkbox>
                    <el-checkbox value="video">视频</el-checkbox>
                    <el-checkbox value="audio">音频</el-checkbox>
                    <el-checkbox value="doc">文档</el-checkbox>
                    <el-checkbox value="other">其他</el-checkbox>
                  </el-checkbox-group>
                </el-form-item>
                <el-form-item label="自定义扩展名">
                  <el-input v-model="selected.fileExtensions" placeholder="jpg,png,pdf 逗号分隔" />
                </el-form-item>
                <el-row :gutter="8">
                  <el-col :span="12"><el-form-item label="单文件限制"><el-input-number v-model="selected.maxFileSize" :min="0" style="width:100%" /><span style="font-size:11px;color:#999;margin-left:4px">MB</span></el-form-item></el-col>
                  <el-col :span="12"><el-form-item label="文件数量"><el-input-number v-model="selected.maxFileCount" :min="1" :max="99" style="width:100%" /></el-form-item></el-col>
                </el-row>
                <el-button text size="small" style="margin-top:8px" @click="selectedOptIdx=-1">← 返回题目设置</el-button>
              </el-form>
            </template>

            <!-- 矩阵行设置 -->
            <template v-else-if="selectedOptIdx>=0 && isMatrixAll(selected.type) && selected.props?.rows?.[selectedOptIdx]">
              <h3>矩阵行设置</h3>
              <el-form label-position="top" size="small">
                <el-form-item label="行名">
                  <el-input v-model="selected.props.rows[selectedOptIdx].title" />
                </el-form-item>
                <el-form-item label="行ID">
                  <el-input v-model="selected.props.rows[selectedOptIdx].id" :disabled="true" />
                </el-form-item>
                <el-button text size="small" style="margin-top:8px" @click="selectedOptIdx=-1">← 返回题目设置</el-button>
              </el-form>
            </template>

            <!-- 成员/部门设置 -->
            <template v-else-if="selectedOptIdx>=0 && (selected.type==='user'||selected.type==='dept')">
              <h3>{{ selected.type==='user'?'成员':'部门' }}设置</h3>
              <el-form label-position="top" size="small">
                <el-form-item v-if="selected.type==='user'" label="成员名称">
                  <el-input v-model="selected.props.options[selectedOptIdx].label" placeholder="输入成员名称" />
                </el-form-item>
                <el-form-item v-if="selected.type==='user'" label="成员ID">
                  <el-input v-model="selected.props.options[selectedOptIdx].value" placeholder="输入成员ID" />
                </el-form-item>
                <el-form-item v-if="selected.type==='user'" label="部门名称">
                  <el-input v-model="selected.props.options[selectedOptIdx].deptName" placeholder="输入部门名称" />
                </el-form-item>
                <el-form-item v-if="selected.type==='user'" label="部门ID">
                  <el-input v-model="selected.props.options[selectedOptIdx].deptId" placeholder="输入部门ID" />
                </el-form-item>
                <el-form-item v-if="selected.type==='user'" label="父级部门ID">
                  <el-input v-model="selected.props.options[selectedOptIdx].parentDeptId" placeholder="输入父级部门ID" />
                </el-form-item>
                <el-form-item v-if="selected.type==='dept'" label="部门名称">
                  <el-input v-model="selected.props.options[selectedOptIdx].label" placeholder="输入部门名称" />
                </el-form-item>
                <el-form-item v-if="selected.type==='dept'" label="部门ID">
                  <el-input v-model="selected.props.options[selectedOptIdx].value" placeholder="输入部门ID" />
                </el-form-item>
                <el-form-item v-if="selected.type==='dept'" label="父级部门ID">
                  <el-input v-model="selected.props.options[selectedOptIdx].parentId" placeholder="输入父级部门ID" />
                </el-form-item>
                <el-button text size="small" style="margin-top:8px" @click="selectedOptIdx=-1">← 返回题目设置</el-button>
              </el-form>
            </template>

            <!-- 选项设置模式 -->
            <template v-else-if="selectedOptIdx>=0 && selected.props?.options?.[selectedOptIdx]">
              <h3>选项设置 - {{ selected.props.options[selectedOptIdx].label }}</h3>
              <el-form label-position="top" size="small">
                <el-form-item label="选项值"><el-input v-model="selected.props.options[selectedOptIdx].value" placeholder="默认同选项名称" /></el-form-item>
                <template v-if="hasOptions(selected)">
                  <el-form-item label="选项配额"><el-input-number v-model="selected.props.options[selectedOptIdx].quota" :min="0" style="width:100%" placeholder="0=不限制" /></el-form-item>
                </template>
                <template v-if="isInput(selected.type)">
                  <el-form-item label="必填"><el-switch v-model="selected.props.options[selectedOptIdx].required" /></el-form-item>
                  <el-form-item label="只读"><el-switch v-model="selected.props.options[selectedOptIdx].readOnly" /></el-form-item>
                  <el-form-item label="计算公式"><el-input v-model="selected.props.options[selectedOptIdx].calculate" placeholder="#{qId} 引用其他题目" /></el-form-item>
                  <el-form-item label="内容限制"><el-select v-model="selected.props.options[selectedOptIdx].dataType" placeholder="不限制" clearable style="width:100%"><el-option label="不限制" value="" /><el-option label="数字" value="number" /><el-option label="手机号" value="mobile" /><el-option label="邮箱" value="email" /><el-option label="身份证" value="idCard" /></el-select></el-form-item>
                  <el-form-item v-if="selected.props.options[selectedOptIdx].dataType==='number'" label="小数位数"><el-input-number v-model="selected.props.options[selectedOptIdx].decimalPlaces" :min="0" :max="6" style="width:100%" /></el-form-item>
                  <el-form-item label="最少填写"><el-input-number v-model="selected.props.options[selectedOptIdx].minLength" :min="0" style="width:100%" /></el-form-item>
                  <el-form-item label="最多填写"><el-input-number v-model="selected.props.options[selectedOptIdx].maxLength" :min="0" style="width:100%" /></el-form-item>
                  <el-form-item label="后缀文字"><el-input v-model="selected.props.options[selectedOptIdx].suffix" placeholder="如 cm / kg" /></el-form-item>
                  <el-form-item label="答案唯一"><el-switch v-model="selected.props.options[selectedOptIdx].unique" /></el-form-item>
                  <el-form-item label="提示语"><el-input v-model="selected.props.options[selectedOptIdx].placeholder" /></el-form-item>
                </template>
                <el-button text size="small" style="margin-top:8px" @click="selectedOptIdx=-1">← 返回题目设置</el-button>
              </el-form>
            </template>

            <!-- 多项/横向填空子字段设置 -->
            <template v-else-if="selectedOptIdx>=0 && (selected.type==='multiInput'||selected.type==='hInput') && selected.props?.fields?.[selectedOptIdx]">
              <h3>字段设置 - {{ selected.props.fields[selectedOptIdx].label }}</h3>
              <el-form label-position="top" size="small">
                <el-form-item label="字段名">
                  <el-input v-model="selected.props.fields[selectedOptIdx].label" placeholder="输入字段名" />
                </el-form-item>
                <el-form-item label="占位提示">
                  <el-input v-model="selected.props.fields[selectedOptIdx].placeholder" placeholder="输入占位提示" />
                </el-form-item>
                <el-form-item label="数据类型">
                  <el-select v-model="selected.props.fields[selectedOptIdx].dataType" placeholder="不限制" clearable style="width:100%">
                    <el-option label="不限制" value="" />
                    <el-option label="数字" value="number" />
                    <el-option label="手机号" value="mobile" />
                    <el-option label="邮箱" value="email" />
                    <el-option label="身份证" value="idCard" />
                  </el-select>
                </el-form-item>
                <el-form-item v-if="selected.props.fields[selectedOptIdx].dataType==='number'" label="小数位数">
                  <el-input-number v-model="selected.props.fields[selectedOptIdx].decimalPlaces" :min="0" :max="6" style="width:100%" />
                </el-form-item>
                <el-button text size="small" style="margin-top:8px" @click="selectedOptIdx=-1">← 返回题目设置</el-button>
              </el-form>
            </template>

            <!-- 评分/NPS 设置 -->
            <template v-else-if="selectedOptIdx>=0 && (selected.type==='rating'||selected.type==='nps')">
              <h3>{{ selected.type==='rating'?'评分':'NPS' }}设置</h3>
              <el-form label-position="top" size="small">
                <el-form-item label="最大分值">
                  <el-input-number v-model="selected.props.maxRating" :min="2" :max="10" :step="1" style="width:100%" />
                </el-form-item>
                <el-form-item v-if="selected.type==='rating'" label="图标类型">
                  <el-radio-group v-model="selected.props.icon">
                    <el-radio value="star">星标</el-radio>
                    <el-radio value="heart">爱心</el-radio>
                    <el-radio value="smiley">笑脸</el-radio>
                  </el-radio-group>
                </el-form-item>
              </el-form>
              <el-button text size="small" style="margin-top:8px" @click="selectedOptIdx=-1">← 返回题目设置</el-button>
              <el-button type="danger" text style="margin-top:12px;width:100%" @click="removeSelected">删除此题</el-button>
            </template>

            <!-- fallback: selectedOptIdx 未匹配任何已知模式 -->
            <template v-else-if="selectedOptIdx>=0">
              <el-button text size="small" @click="selectedOptIdx=-1">← 返回题目设置</el-button>
            </template>

            <!-- 题目设置模式 -->
            <template v-else>
              <h3>{{ typeName(selected.type) }}设置</h3>
              <el-form label-position="top" size="small">
                <!-- 通用基础字段 -->
                <el-form-item label="ID"><el-input v-model="selected.id" :disabled="!!selected._existing" /></el-form-item>
                <el-row :gutter="8" v-if="!isPureLayout(selected.type)&&selected.type!=='questionSet'">
                  <el-col :span="12"><el-form-item label="必填"><el-switch v-model="selected.required" /></el-form-item></el-col>
                  <el-col :span="12"><el-form-item label="只读"><el-switch v-model="selected.readOnly" /></el-form-item></el-col>
                </el-row>
                <el-form-item v-if="!isPureLayout(selected.type)&&selected.type!=='questionSet'" label="说明">
                  <el-switch v-model="selected.showDescription" />
                </el-form-item>

                <!-- 非常规布局题属性 -->
                <template v-if="!isPureLayout(selected.type)">
                  <el-row :gutter="8">
                    <el-col :span="12"><el-form-item label="默认隐藏"><el-switch v-model="selected.defaultHidden" /></el-form-item></el-col>
                    <el-col :span="12" v-if="selected.type==='user'||selected.type==='dept'"><el-form-item label="多选"><el-switch v-model="selected.multiple" /></el-form-item></el-col>
                    <el-col :span="12" v-if="hasOptions(selected) && !isMatrixAll(selected.type)"><el-form-item label="选项布局"><el-input-number v-model="selected.optionLayout" :min="1" :max="6" :step="1" style="width:100%" /></el-form-item></el-col>
                  </el-row>
                  <el-form-item v-if="isInput(selected.type)||isPersonal(selected.type)||selected.type==='signature'||selected.type==='scanCode'" label="占位提示">
                    <el-input v-model="selected.placeholder" />
                  </el-form-item>
                  <el-form-item v-if="selected.type==='richText'" label="占位提示">
                    <el-input v-model="selected.placeholder" placeholder="输入富文本编辑器占位提示" />
                  </el-form-item>
                </template>

                <!-- 表格自增列编辑 -->
                <div v-if="selected.type==='matrixAuto'&&selected.props" class="props-options-section">
                  <h4 style="font-size:12px;color:#888;margin:0 0 6px">列定义</h4>
                  <div v-for="(col, ci) in (selected.props.columns||[])" :key="ci" class="setting-opt-row">
                    <el-input v-model="col.label" size="small" placeholder="列名" style="flex:1" />
                    <el-select v-model="col.type" size="small" style="width:90px">
                      <el-option label="文本" value="input" />
                      <el-option label="数字" value="number" />
                      <el-option label="多行" value="textarea" />
                      <el-option label="日期" value="date" />
                    </el-select>
                    <el-button text size="small" type="danger" @click="removeAutoCol(toNumericIndex(ci))">×</el-button>
                  </div>
                  <el-button size="small" @click="addAutoCol" plain>+ 添加列</el-button>
                  <el-row :gutter="8" style="margin-top:8px">
                    <el-col :span="12"><el-form-item label="最少行数"><el-input-number v-model="selected.props.minRows" :min="0" style="width:100%" /></el-form-item></el-col>
                    <el-col :span="12"><el-form-item label="最多行数"><el-input-number v-model="selected.props.maxRows" :min="0" style="width:100%" /></el-form-item></el-col>
                  </el-row>
                </div>

                <!-- 数字范围 -->
                <div v-if="selected.type==='number'&&selected.props" class="props-options-section">
                  <h4 style="font-size:12px;color:#888;margin:0 0 6px">数值范围</h4>
                  <el-row :gutter="8">
                    <el-col :span="12"><el-form-item label="最小值"><el-input-number v-model="selected.props.min" :step="1" style="width:100%" /></el-form-item></el-col>
                    <el-col :span="12"><el-form-item label="最大值"><el-input-number v-model="selected.props.max" :step="1" style="width:100%" /></el-form-item></el-col>
                  </el-row>
                  <el-form-item label="小数位数"><el-input-number v-model="selected.props.decimalPlaces" :min="0" :max="6" style="width:100%" /></el-form-item>
                </div>

                <!-- 分页设置 -->
                <div v-if="selected.type==='pagination'" class="props-options-section">
                  <h4 style="font-size:12px;color:#888;margin:0 0 6px">分页</h4>
                  <el-row :gutter="8">
                    <el-col :span="12"><el-form-item label="当前页"><el-input-number v-model="selected.props.currentPage" :min="1" style="width:100%" /></el-form-item></el-col>
                    <el-col :span="12"><el-form-item label="总页数"><el-input-number v-model="selected.props.totalPage" :min="1" style="width:100%" /></el-form-item></el-col>
                  </el-row>
                </div>

                <!-- 成员/部门列表 -->
                <div v-if="(selected.type==='user'||selected.type==='dept')&&selected.props" class="props-options-section">
                  <div style="display:flex;align-items:center;gap:4px;margin-bottom:6px">
                    <h4 style="font-size:12px;color:#888;margin:0">{{ selected.type==='user'?'成员':'部门' }}列表</h4>
                  </div>
                  <div v-for="(o, i) in (selected.props.options||[])" :key="i" class="setting-opt-row">
                    <span style="font-size:12px;flex:1;overflow:hidden;text-overflow:ellipsis;white-space:nowrap">
                      {{ o.label || (selected.type==='user'?'成员':'部门') }}
                      <span v-if="selected.type==='user' && o.deptName" style="color:#999;margin-left:4px">({{ o.deptName }})</span>
                      <span v-if="selected.type==='user' && o.parentDeptId" style="color:#999;margin-left:4px;font-size:11px">父级:{{ o.parentDeptId }}</span>
                      <span v-if="selected.type==='dept' && o.parentId" style="color:#999;margin-left:4px;font-size:11px">父级:{{ o.parentId }}</span>
                    </span>
                    <el-button text size="small" type="primary" @click="selectOption(selected.id, toNumericIndex(i))" style="font-size:11px">编辑</el-button>
                    <el-button text size="small" type="danger" @click="removeUserOpt(toNumericIndex(i))">×</el-button>
                  </div>
                  <div style="display:flex;gap:4px;margin-top:4px">
                    <el-button size="small" @click="addUserOpt" plain>{{ selected.type==='user'?'+ 添加成员':'+ 添加部门' }}</el-button>
                    <el-button size="small" @click="importUserOpt" plain>导入</el-button>
                  </div>
                </div>

                <!-- 签名考试属性 -->
                <el-form-item v-if="selected.type==='signature'" label="考试答案模式">
                  <el-select v-model="selected.examAnswerMode" clearable placeholder="不限制" style="width:100%">
                    <el-option label="不限制" value="" />
                    <el-option label="无" value="none" />
                    <el-option label="文字" value="text" />
                    <el-option label="图片" value="image" />
                  </el-select>
                </el-form-item>

                <!-- 数据类型 -->
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

                <el-divider border-style="dashed" style="margin:12px 0" />
                <div class="exam-scoring-section">
                  <el-form-item label="计分方式" style="margin-bottom:4px">
                    <el-select v-model="selected.examScoreMode" placeholder="不计分" clearable style="width:100%" teleported="false" popper-style="width:auto;min-width:280px">
                      <el-option value="single" label="此题有唯一答案和分值" />
                      <el-option value="perOption" label="每个选项都有对应的分值" :disabled="!hasOptions(selected)&&!isMatrixAll(selected.type)&&!['multiInput','hInput'].includes(selected.type)" />
                      <el-option value="allCorrect" label="全部答对才得分" :disabled="!hasOptions(selected)&&!isMatrixAll(selected.type)&&!['multiInput','hInput'].includes(selected.type)" />
                      <el-option value="partialCorrect" label="答对几项得几分，答错不得分" :disabled="!hasOptions(selected)&&!isMatrixAll(selected.type)&&!['multiInput','hInput'].includes(selected.type)" />
                    </el-select>
                  </el-form-item>
                  <template v-if="selected.examScoreMode">
                    <div class="props-options-section" style="margin-bottom:8px;padding:8px">
                      <template v-if="selected.examScoreMode==='single'">
                        <div style="font-size:12px;color:#999;margin-bottom:6px">选择/填写正确答案</div>
                        <template v-if="isMatrixAll(selected.type)">
                          <span style="font-size:12px;color:#999">为每道矩阵行统一设置分值</span>
                        </template>
                        <template v-else-if="hasOptions(selected)">
                          <el-select ref="correctSelectRef" v-model="selected.examCorrectAnswer" placeholder="选择正确答案" clearable style="width:100%" popper-style="width:auto;min-width:260px" popper-class="correct-answer-popper" @visible-change="updateCorrectLabel">
                            <el-option v-for="(o, i) in (selected.props?.options||[])" :key="i" :value="o.value" :label="o.label?.replace(/<[^>]*>/g,'') || o.value"><span v-html="o.label||o.value" /></el-option>
                          </el-select>
                        </template>
                        <template v-else-if="isInput(selected.type)">
                          <div style="font-size:12px;color:#999;margin-bottom:4px">为每个输入框设置正确答案</div>
                          <div v-for="(f, i) in (selected.props?.fields||[])" :key="i" class="setting-opt-row" style="gap:4px">
                            <span style="font-size:12px;color:#666;white-space:nowrap;min-width:60px">{{ f.label || '输入框' + (toNumericIndex(i) + 1) }}</span>
                            <el-input v-model="f.examCorrectAnswer" :placeholder="f.examCorrectAnswer?'':(f.label || '输入框' + (toNumericIndex(i) + 1))" size="small" style="flex:1" @change="syncFieldsCorrectAnswer" />
                          </div>
                        </template>
                        <template v-else>
                          <el-input v-model="selected.examCorrectAnswer" placeholder="填写正确答案" size="small" />
                        </template>
                      </template>
                      <template v-else-if="selected.examScoreMode==='perOption'">
                        <template v-if="hasOptions(selected)">
                          <div style="font-size:12px;color:#999;margin-bottom:4px">为每个选项设置分值</div>
                          <div v-for="(o, i) in (selected.props?.options||[])" :key="i" class="setting-opt-row" style="gap:4px">
                            <span style="font-size:12px;color:#666;white-space:nowrap;min-width:40px" v-html="o.label || '选项' + (toNumericIndex(i) + 1)"></span>
                            <el-input-number v-model="o.examScore" :min="0" :step="0.5" size="small" style="flex:1" placeholder="0" controls-position="right" />
                          </div>
                        </template>
                        <template v-else-if="isMatrixAll(selected.type)">
                          <el-form-item label="每行分值" style="margin:0"><el-input-number v-model="selected.examScore" :min="0" :step="0.5" size="small" style="width:100%" /></el-form-item>
                        </template>
                        <template v-else-if="isInput(selected.type)">
                          <div style="font-size:12px;color:#999;margin-bottom:4px">为每个输入框设置分值和正确答案</div>
                          <div v-for="(f, i) in (selected.props?.fields||[])" :key="i" class="setting-opt-row" style="gap:4px">
                            <span style="font-size:12px;color:#666;white-space:nowrap;min-width:60px">{{ f.label || '输入框' + (toNumericIndex(i) + 1) }}</span>
                            <el-input v-model="f.examCorrectAnswer" placeholder="正确答案" size="small" style="flex:1;min-width:60px" />
                            <el-input-number v-model="f.examScore" :min="0" :step="0.5" size="small" style="width:90px" placeholder="分值" controls-position="right" />
                          </div>
                        </template>
                      </template>
                      <template v-else-if="selected.examScoreMode==='allCorrect'">
                        <div style="font-size:12px;color:#999;margin-bottom:4px">勾选所有正确选项</div>
                        <template v-if="hasOptions(selected)">
                          <div v-for="(o, i) in (selected.props?.options||[])" :key="i" style="display:flex;align-items:center;gap:8px;margin-bottom:4px;padding:4px 6px;border-radius:4px;background:#fff;border:1px solid #e8e8e8">
                            <el-checkbox v-model="o.examCorrect" size="small" />
                            <span style="font-size:13px;flex:1" v-html="o.label || `选项${toNumericIndex(i) + 1}`"></span>
                            <el-tag v-if="o.examCorrect" size="small" type="success" effect="dark" style="flex-shrink:0">正确</el-tag>
                          </div>
                        </template>
                        <template v-else-if="isMatrixAll(selected.type)">
                          <el-form-item label="每行分值" style="margin:0"><el-input-number v-model="selected.examScore" :min="0" :step="0.5" size="small" style="width:100%" /></el-form-item>
                        </template>
                        <template v-else-if="isInput(selected.type)">
                          <div style="font-size:12px;color:#999;margin-bottom:4px">勾选所有必须正确的输入框，并设置每项正确答案</div>
                          <div v-for="(f, i) in (selected.props?.fields||[])" :key="i" style="display:flex;align-items:center;gap:8px;margin-bottom:4px;padding:4px 6px;border-radius:4px;background:#fff;border:1px solid #e8e8e8">
                            <el-checkbox v-model="f.examCorrect" size="small" />
                            <span style="font-size:13px;flex:1;white-space:nowrap;min-width:50px">{{ f.label || '输入框' + (toNumericIndex(i) + 1) }}</span>
                            <el-input v-model="f.examCorrectAnswer" placeholder="正确答案" size="small" style="flex:1;min-width:60px" />
                            <el-tag v-if="f.examCorrect" size="small" type="success" effect="dark" style="flex-shrink:0">必对</el-tag>
                          </div>
                        </template>
                        </template>
                        <template v-else-if="selected.examScoreMode==='partialCorrect'">
                        <div style="font-size:12px;color:#999;margin-bottom:4px">勾选答对的选项并设置每题分值</div>
                        <template v-if="hasOptions(selected)">
                          <div v-for="(o, i) in (selected.props?.options||[])" :key="i" style="display:flex;align-items:center;gap:8px;margin-bottom:4px;padding:4px 6px;border-radius:4px;background:#fff;border:1px solid #e8e8e8">
                            <el-checkbox v-model="o.examCorrect" size="small" />
                            <span style="font-size:13px;flex:1" v-html="o.label || `选项${toNumericIndex(i) + 1}`"></span>
                            <el-input-number v-if="o.examCorrect" v-model="o.examScore" :min="0" :step="0.5" size="small" style="width:90px" controls-position="right" />
                            <el-tag v-if="o.examCorrect" size="small" type="success" effect="dark" style="flex-shrink:0">正确</el-tag>
                          </div>
                        </template>
                        <template v-else-if="isMatrixAll(selected.type)">
                          <el-form-item label="每行分值" style="margin:0"><el-input-number v-model="selected.examScore" :min="0" :step="0.5" size="small" style="width:100%" /></el-form-item>
                        </template>
                        <template v-else-if="isInput(selected.type)">
                          <div style="font-size:12px;color:#999;margin-bottom:4px">勾选答对的输入框并设置每题分值</div>
                          <div v-for="(f, i) in (selected.props?.fields||[])" :key="i" style="display:flex;align-items:center;gap:8px;margin-bottom:4px;padding:4px 6px;border-radius:4px;background:#fff;border:1px solid #e8e8e8">
                            <el-checkbox v-model="f.examCorrect" size="small" />
                            <el-input v-model="f.examCorrectAnswer" placeholder="正确答案" size="small" style="flex:1;min-width:60px" />
                            <el-input-number v-if="f.examCorrect" v-model="f.examScore" :min="0" :step="0.5" size="small" style="width:90px" controls-position="right" />
                            <el-tag v-if="f.examCorrect" size="small" type="success" effect="dark" style="flex-shrink:0">正确</el-tag>
                          </div>
                        </template>
                      </template>
                    </div>
                    <el-row :gutter="8">
                      <el-col :span="12"><el-form-item label="此题分值" style="margin-bottom:4px">
                        <el-input-number v-if="selected.examScoreMode==='partialCorrect'" :model-value="partialTotalScore" :min="0" :step="0.5" size="small" style="width:100%" disabled />
                        <el-input-number v-else-if="selected.examScoreMode==='perOption'" :model-value="perOptionTotalScore" :min="0" :step="0.5" size="small" style="width:100%" disabled />
                        <el-input-number v-else v-model="selected.examScore" :min="0" :step="0.5" size="small" style="width:100%" controls-position="right" />
                      </el-form-item></el-col>
                    </el-row>
                    <el-form-item label="答案解析" style="margin-bottom:0">
                      <el-input v-model="selected.examAnalysis" type="textarea" :rows="2" placeholder="选填" size="small" />
                    </el-form-item>
                  </template>
                </div>
                <div class="exam-formula-rows">
                  <div class="setting-row"><span class="setting-label">结束公式</span><el-button text size="small" @click="openFormulaDialog('endFormula')">点击设置</el-button></div>
                  <div class="setting-row"><span class="setting-label">跳转公式</span><el-button text size="small" @click="openFormulaDialog('jumpFormula')">点击设置</el-button></div>
                  <div class="setting-row"><span class="setting-label">文本替换</span><el-button text size="small" @click="openFormulaDialog('textReplace')">点击设置</el-button></div>
                  <div class="setting-row" v-if="isInput(selected.type)||isPersonal(selected.type)"><span class="setting-label">计算公式</span><el-button text size="small" @click="openFormulaDialog('calculate')">点击设置</el-button></div>
                  <div class="setting-row"><span class="setting-label">校验规则</span><el-button text size="small" @click="openFormulaDialog('validate')">点击设置</el-button></div>
                </div>
              </el-form>
              <div class="props-footer">
                <el-button type="danger" text style="width:100%" @click="removeSelected">删除此题</el-button>
              </div>
            </template>
          </div>
          <el-empty v-else description="选择题目编辑属性" :image-size="60" />
        </div>

      </div>

      <ExamDesignerSettingsPanel
        ref="settingsRef"
        v-show="activeView === 'setting'"
        v-model="completionAction"
        v-model:form="form"
        :saving="saving"
        :public-url="publicUrl"
        :fill-templates="fillTemplates"
        @save="save"
        @copy-link="copyLink"
      />

      <div v-show="activeView==='data'" class="setting-wrapper">
        <div class="setting-scroll">
          <div class="setting-group">
            <div class="group-title">数据统计</div>
            <div style="display:flex;gap:16px;padding:12px 0">
              <el-button size="large" @click="goResponses"><el-icon style="margin-right:6px"><Document /></el-icon>查看答卷</el-button>
              <el-button size="large" type="primary" @click="goStatistic"><el-icon style="margin-right:6px"><DataAnalysis /></el-icon>查看统计</el-button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>

  <SurveyDesignerFormulaDialog ref="formulaDialogRef" :questions="questions" @save="saveFormula" />
  <SurveyDesignerBankUploadDialog ref="bankUploadDialogRef" @confirm="confirmUploadBank" />
  <SurveyDesignerTextImportDialog ref="textImportDialogRef" @import="importTextQuestions" />
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, nextTick } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useRoute, useRouter } from 'vue-router'
import { adminApi } from '../../api'
import DraggableList from './formkit/DraggableList.vue'
import ExamDesignerSidebar from './components/ExamDesignerSidebar.vue'
import ExamDesignerLogicPanel from './components/ExamDesignerLogicPanel.vue'
import ExamDesignerSettingsPanel from './components/ExamDesignerSettingsPanel.vue'
import SurveyDesignerFormulaDialog from '../survey/components/SurveyDesignerFormulaDialog.vue'
import SurveyDesignerTextImportDialog from '../survey/components/SurveyDesignerTextImportDialog.vue'
import SurveyDesignerBankUploadDialog from '../survey/components/SurveyDesignerBankUploadDialog.vue'
import { FALLBACK_TYPES, getSurveyQuestionTypeName as typeName } from '../survey/survey-designer-question-types'
import type { SurveyFormulaPayload } from '../survey/survey-designer-types'
import type { ExamDesignerQuestion as Question, ExamLogicRule } from './exam-designer-types'
import { createExamDesignerForm, useExamDesignerDocument } from './composables/useExamDesignerDocument'
import { useExamDesignerTransfer } from './composables/useExamDesignerTransfer'

const route = useRoute()
const router = useRouter()
const toNumericIndex = (index: string | number) => Number(index)

const form = createExamDesignerForm()

const activeView = ref('edit')
const middleTab = ref('item')
const panelMode = ref<'edit' | 'json' | 'preview'>('edit')
const logicRuleList = ref<ExamLogicRule[]>([])

interface SidebarHandle { loadBank: () => void; loadResources: () => Promise<void> }
interface SettingsHandle { loadData: () => Promise<void> }
const sidebarRef = ref<SidebarHandle | null>(null)
const settingsRef = ref<SettingsHandle | null>(null)

const correctSelectRef = ref<any>(null)
const selectedCorrectLabel = computed(() => {
  const current = selected.value
  if (!current?.examCorrectAnswer) return ''
  const opt = (current.props?.options || []).find((o: any) => o.value === current.examCorrectAnswer)
  return opt ? (opt.label || opt.value) : current.examCorrectAnswer
})
function updateCorrectLabel() {
  nextTick(() => {
    if (!correctSelectRef.value?.$el) return
    const wrap = correctSelectRef.value.$el.querySelector('.el-select__wrapper')
    if (!wrap) return
    const el = wrap.querySelector('.el-select__placeholder') || wrap.querySelector('.el-select__selected-item')
    if (el) el.innerHTML = selectedCorrectLabel.value
  })
}

const types = ref<any[]>([])
const questions = ref<Question[]>([])
const selected = ref<Question | null>(null)
let idCounter = 0
function genId() { idCounter++; return 'q' + idCounter }
function addQuestion(t: any) {
  const q: Question = { id: genId(), type: t.type, title: t.displayName, required: false, readOnly: false, placeholder: '', props: t.defaultProps ? JSON.parse(JSON.stringify(t.defaultProps)) : {} }
  if (!q.props) q.props = {}
  if (t.type !== 'description' && t.type !== 'questionSet' && t.type !== 'pagination' && t.type !== 'divider') {
    q.examScoreMode = 'single'
  }
  if (t.type === 'judge') {
    q.props.options = [{ label: '对', value: 'true' }, { label: '错', value: 'false' }]
  }
  if (['select','radio','checkbox','picker','cascade'].includes(t.type)) {
    q.props.options = [
      { label: '选项A', value: 'A' },
      { label: '选项B', value: 'B' },
      { label: '选项C', value: 'C' },
      { label: '选项D', value: 'D' },
    ]
  }
  if (t.type !== 'description' && t.type !== 'questionSet' && t.type !== 'pagination' && t.type !== 'divider') {
    q.showDescription = true
  }
  if (t.type === 'matrixRadio' || t.type === 'matrixCheckbox' || t.type === 'matrixFillBlank') {
    q.props.rows = [{ title: '行1', id: 'r1' }, { title: '行2', id: 'r2' }]
    q.props.columns = [{ title: '列A', id: 'c1' }, { title: '列B', id: 'c2' }]
  }
  if (t.type === 'matrixAuto') {
    q.props.columns = [{ label: '姓名', type: 'input' }, { label: '数值', type: 'number' }]
  }
  if (t.type === 'multiInput' || t.type === 'hInput') {
    q.props.fields = [{ label: '填空1', placeholder: '' }, { label: '填空2', placeholder: '' }]
  }
  if (t.type === 'rating') {
    q.props.maxRating = 5
    q.props.icon = 'star'
  }
  if (t.type === 'nps') {
    q.props.maxRating = 10
  }
  if (t.type === 'file') {
    q.fileTypes = ['image']
    q.maxFileSize = 10
    q.maxFileCount = 1
  }
  if (t.type === 'signature' || t.type === 'scanCode') {
    q.dataType = ''
  }
  questions.value.push(q); selected.value = q
}
function scrollQuestionIntoView(id: string) {
  nextTick(() => {
    const cards = Array.from(document.querySelectorAll<HTMLElement>('[data-exam-question-id]'))
    const target = cards.find(el => el.dataset.examQuestionId === id)
    target?.scrollIntoView({ behavior: 'smooth', block: 'center', inline: 'nearest' })
  })
}

function selectQuestion(id: string, shouldScroll = false) {
  selected.value = questions.value.find(q => q.id === id) || null
  selectedOptIdx.value = -1
  populateFieldsFromCorrectAnswer()
  if (shouldScroll && selected.value) {
    panelMode.value = 'edit'
    scrollQuestionIntoView(id)
  }
}
function deselectQuestion() { selected.value = null; selectedOptIdx.value = -1 }
function onQuestionsUpdate(newQ: Question[]) {
  questions.value.splice(0, questions.value.length, ...newQ)
  if (selected.value) selected.value = questions.value.find(q => q.id === selected.value!.id) || null
}
async function removeSelected() {
  if (!selected.value) return
  try { await ElMessageBox.confirm('确定删除此题？', '确认', { type: 'warning' }) } catch { return }
  const sel = selected.value; questions.value = questions.value.filter(q => q.id !== sel.id)
  selected.value = questions.value.length ? questions.value[questions.value.length - 1] : null
}
function hasOptions(q: Question) { return ['select','radio','checkbox','picker','cascade','judge','matrixRadio','matrixCheckbox'].includes(q.type) }
function isPureLayout(t: string) { return ['divider','description','pagination'].includes(t) }
function isInput(t: string) { return ['input','textarea','number','multiInput','hInput'].includes(t) }
function isMatrixQR(t: string) { return ['matrixRadio','matrixCheckbox','matrixFillBlank'].includes(t) }
function isMatrixAll(t: string) { return ['matrixRadio','matrixCheckbox','matrixFillBlank','matrixAuto'].includes(t) }
function isChoice(t: string) { return ['select','radio','checkbox','picker','cascade','judge'].includes(t) }
function hasDataType(q: Question) { return ['input','textarea','number','name','studentId','employeeId','class','phone','email','idCard','password','date','time','dateRange','switch','location'].includes(q.type) }
function isPersonal(t: string) { return ['name','studentId','employeeId','class','phone','email','idCard','password'].includes(t) }
const selectedOptIdx = ref(-1)
interface FormulaDialogHandle { open: (type: string, initialText?: string) => void }
const formulaDialogRef = ref<FormulaDialogHandle | null>(null)
function openFormulaDialog(type: string) {
  if (!selected.value) return
  const fieldMap: Record<string, string> = {
    endFormula: 'endFormula', jumpFormula: 'jumpFormula',
    textReplace: 'replaceTextRule', calculate: 'calculateFormula', validate: 'validateRule'
  }
  const field = fieldMap[type] || ''
  formulaDialogRef.value?.open(type, selected.value[field] || selected.value.props?.[field] || '')
}
function saveFormula(payload: SurveyFormulaPayload) {
  if (!selected.value) return
  const fieldMap: Record<string, string> = {
    endFormula: 'endFormula', jumpFormula: 'jumpFormula',
    textReplace: 'replaceTextRule', calculate: 'calculateFormula', validate: 'validateRule'
  }
  const field = fieldMap[payload.type] || ''
  if (payload.type === 'calculate' || payload.type === 'textReplace') {
    if (!selected.value.props) selected.value.props = {}
    selected.value.props[field] = payload.text
  } else {
    selected.value[field] = payload.text
  }
}
function selectOption(qId: string, optIdx: number) { selected.value = questions.value.find(q => q.id === qId) || null; selectedOptIdx.value = optIdx }
function addAutoCol() {
  if (!selected.value?.props) return
  if (!selected.value.props.columns) selected.value.props.columns = []
  selected.value.props.columns.push({ label: '列' + (selected.value.props.columns.length + 1), type: 'input' })
}
function removeAutoCol(ci: number) { if (selected.value?.props?.columns) selected.value.props.columns.splice(ci, 1) }
function addUserOpt() {
  if (!selected.value?.props) return
  if (!selected.value.props.options) selected.value.props.options = []
  selected.value.props.options.push({ label: selected.value.type === 'user' ? '成员' + (selected.value.props.options.length + 1) : '部门' + (selected.value.props.options.length + 1), value: '' })
}
function removeUserOpt(idx: number) { if (selected.value?.props?.options) selected.value.props.options.splice(idx, 1) }
function importUserOpt() {
  const t = selected.value?.type
  if (!t) return
  const text = prompt(`请粘贴${t==='user'?'成员':'部门'}数据，格式：\n名称,ID[,部门名称,部门ID] (每行一个)`)
  if (!text) return
  const lines = text.trim().split('\n')
  for (const line of lines) {
    const parts = line.split(',').map(s => s.trim())
    if (!parts[0]) continue
    const item: any = { label: parts[0], value: parts[1] || parts[0] }
    if (t === 'user') {
      if (parts[2]) item.deptName = parts[2]
      if (parts[3]) item.deptId = parts[3]
    } else {
      if (parts[2]) item.parentId = parts[2]
    }
    if (!selected.value?.props?.options) { if (selected.value) selected.value.props = { ...(selected.value.props || {}), options: [] } }
    selected.value?.props?.options?.push(item)
  }
  ElMessage.success(`已导入 ${lines.filter(l=>l.trim()).length} 条`)
}
function addOption() {
  if (!selected.value) return
  if (!selected.value.props) selected.value.props = {}
  if (!selected.value.props.options) selected.value.props.options = []
  const n = selected.value.props.options.length + 1
  const letter = n <= 26 ? String.fromCharCode(64 + n) : `opt${n}`
  selected.value.props.options.push({ label: `选项${letter}`, value: letter })
}
function removeOption(idx: number) { if (selected.value?.props?.options) selected.value.props.options.splice(idx, 1) }
function removeQuestionById(id: string) { questions.value = questions.value.filter(q => q.id !== id); if (selected.value?.id === id) selected.value = null }

const partialTotalScore = computed(() => {
  if (!selected.value) return 0
  const items = ['multiInput','hInput'].includes(selected.value.type) ? (selected.value.props?.fields||[]) : (selected.value.props?.options||[])
  return items.reduce((sum: number, o: any) => sum + (o.examCorrect ? (Number(o.examScore) || 0) : 0), 0)
})
const perOptionTotalScore = computed(() => {
  if (!selected.value) return 0
  const items = ['multiInput','hInput'].includes(selected.value.type) ? (selected.value.props?.fields||[]) : (selected.value.props?.options||[])
  return items.reduce((sum: number, o: any) => sum + (Number(o.examScore) || 0), 0)
})

function questionScore(q: any): number {
  if (!q.examScoreMode) return 0
  if (q.examScoreMode === 'single') return Number(q.examScore) || 0
  if (q.examScoreMode === 'perOption') {
    const items = ['multiInput','hInput'].includes(q.type) ? (q.props?.fields||[]) : (q.props?.options||[])
    return items.reduce((s: number, o: any) => s + (Number(o.examScore) || 0), 0)
  }
  if (q.examScoreMode === 'allCorrect') return Number(q.examScore) || 0
  if (q.examScoreMode === 'partialCorrect') {
    const items = ['multiInput','hInput'].includes(q.type) ? (q.props?.fields||[]) : (q.props?.options||[])
    return items.reduce((s: number, o: any) => s + (o.examCorrect ? (Number(o.examScore) || 0) : 0), 0)
  }
  return 0
}
const paperTotalScore = computed(() => questions.value.reduce((sum, q) => sum + questionScore(q), 0))

const envFromAnswers = computed(() => {
  const env: Record<string, any> = {}; for (const q of questions.value) env[q.id] = undefined; return env
})

const {
  textImportDialogRef, bankUploadDialogRef, exportedJson, fillTemplates,
  publicUrl, importTextQuestions, openTextImport, exportExam, addFromBank,
  onUploadBank, confirmUploadBank, copyLink, prepareJson, downloadJson, loadJson,
} = useExamDesignerTransfer({
  form,
  questions,
  selected,
  generateId: genId,
  reloadBank: () => sidebarRef.value?.loadBank(),
  typeName,
})

const {
  saving, completionAction, load, save, saveLogicRules,
} = useExamDesignerDocument({
  form,
  questions,
  selected,
  route,
  router,
  onQuestionsLoaded: () => { idCounter = questions.value.length },
  reloadResources: () => sidebarRef.value?.loadResources(),
  reloadSettings: () => settingsRef.value?.loadData(),
})

function goBack() { void router.push('/exam/list') }
function goResponses() { if (form.id) void router.push({ path: '/exam/responses', query: { examId: String(form.id), title: form.title } }) }
function goStatistic() { if (form.id) void router.push({ path: '/exam/statistic', query: { examId: String(form.id), title: form.title } }) }

watch([partialTotalScore, perOptionTotalScore], () => {
  if (!selected.value) return
  if (selected.value.examScoreMode === 'partialCorrect') selected.value.examScore = partialTotalScore.value
  else if (selected.value.examScoreMode === 'perOption') selected.value.examScore = perOptionTotalScore.value
})

watch(() => selected.value?.examScoreMode, (cur, prev) => {
  if (!selected.value || !prev) return
  if (prev === 'single') {
    selected.value.examCorrectAnswer = undefined
  }
  if (prev === 'perOption' || prev === 'partialCorrect') {
    selected.value.props?.options?.forEach((o: any) => { o.examScore = undefined })
    selected.value.props?.fields?.forEach((f: any) => { f.examScore = undefined })
  }
  if (prev === 'allCorrect' || prev === 'partialCorrect') {
    selected.value.props?.options?.forEach((o: any) => { o.examCorrect = undefined })
    selected.value.props?.fields?.forEach((f: any) => { f.examCorrect = undefined })
  }
  if (cur === 'single' && ['multiInput','hInput'].includes(selected.value.type)) {
    populateFieldsFromCorrectAnswer()
  }
})

function populateFieldsFromCorrectAnswer() {
  if (!selected.value || !['multiInput','hInput'].includes(selected.value.type)) return
  const correctAnswer = selected.value.examCorrectAnswer
  if (!correctAnswer) return
  const fields = selected.value.props?.fields || []
  const parts = correctAnswer.split(',')
  fields.forEach((f: any, i: number) => {
    if (parts[i] !== undefined) f.examCorrectAnswer = parts[i]
  })
}

function deriveCorrectAnswerFromMode() {
  if (!selected.value) return
  if (!selected.value.examScoreMode || selected.value.examScoreMode === 'single') {
    if (['multiInput','hInput'].includes(selected.value.type)) {
      const fields = selected.value.props?.fields || []
      const vals = fields.map((f: any) => f.examCorrectAnswer || '').filter(Boolean)
      selected.value.examCorrectAnswer = vals.length ? vals.join(',') : undefined
    } else {
      if (selected.value.examCorrectAnswer?.includes(',')) selected.value.examCorrectAnswer = undefined
    }
    return
  }
  const isInputType = ['multiInput','hInput'].includes(selected.value.type)
  const items = isInputType ? (selected.value.props?.fields || []) : (selected.value.props?.options || [])
  let correctVals: string[] = []
  if (selected.value.examScoreMode === 'allCorrect' || selected.value.examScoreMode === 'partialCorrect') {
    correctVals = items.filter((item: any) => item.examCorrect).map((item: any) => item.value || item.examCorrectAnswer || '')
  } else if (selected.value.examScoreMode === 'perOption') {
    correctVals = items.filter((item: any) => Number(item.examScore) > 0).map((item: any) => isInputType ? (item.examCorrectAnswer || '') : (item.value || ''))
  }
  selected.value.examCorrectAnswer = correctVals.length ? correctVals.join(',') : undefined
}

function syncFieldsCorrectAnswer() {
  if (!selected.value) return
  const fields = selected.value.props?.fields || []
  const vals = fields.map((f: any) => f.examCorrectAnswer || '').filter(Boolean)
  selected.value.examCorrectAnswer = vals.length ? vals.join(',') : undefined
}
watch(() => selected.value?.examScoreMode, deriveCorrectAnswerFromMode)
watch(() => selected.value?.props?.options?.map((o: any) => o.examCorrect), deriveCorrectAnswerFromMode, { deep: true })
watch(() => selected.value?.props?.options?.map((o: any) => o.examScore), deriveCorrectAnswerFromMode, { deep: true })
watch(() => selected.value?.props?.fields?.map((f: any) => f.examCorrect), deriveCorrectAnswerFromMode, { deep: true })
watch(() => selected.value?.props?.fields?.map((f: any) => f.examScore), deriveCorrectAnswerFromMode, { deep: true })
watch(() => selected.value?.props?.fields?.map((f: any) => f.examCorrectAnswer), deriveCorrectAnswerFromMode, { deep: true })

watch(() => selected.value?.examCorrectAnswer, updateCorrectLabel, { immediate: true })
watch(() => selected.value?.props?.options, updateCorrectLabel, { deep: true })

onMounted(async () => {
  try {
    const res: any = await adminApi.formkitTypes()
    const apiTypes: any[] = res.data || []
    const apiMap = new Map(apiTypes.map((t:any) => [t.type, t]))
    types.value = FALLBACK_TYPES.map(ft => {
      const api = apiMap.get(ft.type)
      return api ? { ...ft, ...api, category: ft.category } : ft
    })
  } catch {
    types.value = FALLBACK_TYPES
  }
  await nextTick()
  await load()
})
</script>

<style scoped src="./exam-designer.css"></style>
<style>
.correct-answer-popper .el-select-dropdown__item { height:auto; min-height:34px; line-height:1.4; padding-top:6px; padding-bottom:6px; white-space:normal; }
</style>
