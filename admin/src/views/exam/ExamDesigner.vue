<template>
  <div class="survey-main">
    <!-- 最左侧导航: 概述 / 编辑 / 设置 / 数据 / 报表 / 项目 -->
    <div class="survey-main-navigator">
      <div class="nav-actions">
        <button class="nav-btn" :class="{ active: activeView==='overview' }" @click="activeView='overview'" title="概述">
          <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="16" y1="13" x2="8" y2="13"/><line x1="16" y1="17" x2="8" y2="17"/><polyline points="10 9 9 9 8 9"/></svg>
        </button>
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
        <button class="nav-btn" :class="{ active: activeView==='project' }" @click="activeView='project'" title="项目">
          <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2"><path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/></svg>
        </button>
      </div>
    </div>

    <!-- 主内容 -->
    <div class="survey-main-content">
      <!-- 编辑视图 -->
      <div v-show="activeView==='edit'" id="editor" class="survey-editor survey-light survey-app pc">
        <!-- 中间侧边栏: 项目 / 外观 / 逻辑 -->
        <div class="survey-sidebar-panel">
          <div class="survey-sidebar-panel-tabs">
            <div class="survey-sidebar-panel-tabs-pane" :class="{ active: middleTab==='item' }" title="题目" @click="middleTab='item'">
              <svg viewBox="0 0 1024 1024" width="20" height="20" fill="currentColor"><path d="M810.666667 128H213.333333c-46.933333 0-85.333333 38.4-85.333333 85.333333v597.333334c0 46.933333 38.4 85.333333 85.333333 85.333333h597.333334c46.933333 0 85.333333-38.4 85.333333-85.333333V213.333333c0-46.933333-38.4-85.333333-85.333333-85.333333z m-42.666667 554.666667H256v-85.333334h512v85.333334z m0-170.666667H256v-85.333333h512v85.333333z m0-170.666667H256V256h512v85.333333z"/></svg>
              <span class="tab-label">题目</span>
            </div>
            <div class="survey-sidebar-panel-tabs-pane" :class="{ active: middleTab==='appearance' }" title="外观" @click="middleTab='appearance'">
              <svg viewBox="0 0 1024 1024" width="20" height="20" fill="currentColor"><path d="M512 128c-212.053333 0-384 171.946667-384 384s171.946667 384 384 384 384-171.946667 384-384-171.946667-384-384-384z m0 682.666667c-164.693333 0-298.666667-133.973333-298.666667-298.666667s133.973333-298.666667 298.666667-298.666667 298.666667 133.973333 298.666667 298.666667-133.973333 298.666667-298.666667 298.666667zM512 341.333333c-94.293333 0-170.666667 76.373333-170.666667 170.666667s76.373333 170.666667 170.666667 170.666667 170.666667-76.373333 170.666667-170.666667-76.373333-170.666667-170.666667-170.666667z"/></svg>
              <span class="tab-label">外观</span>
            </div>
            <div class="survey-sidebar-panel-tabs-pane" :class="{ active: middleTab==='logic' }" title="逻辑" @click="middleTab='logic'">
              <svg viewBox="0 0 1024 1024" width="20" height="20" fill="currentColor"><path d="M632.888889 56.888889c0-31.288889-25.6-56.888889-56.888889-56.888889H455.111111c-31.288889 0-56.888889 25.6-56.888889 56.888889v120.888889H170.666667c-31.288889 0-56.888889 25.6-56.888889 56.888889v113.777777c0 31.288889 25.6 56.888889 56.888889 56.888889h120.888889v369.777778c0 31.288889 25.6 56.888889 56.888889 56.888889h113.777777c31.288889 0 56.888889-25.6 56.888889-56.888889V462.222222h113.777778c31.288889 0 56.888889-25.6 56.888889-56.888889V291.555556c0-31.288889-25.6-56.888889-56.888889-56.888889H632.888889V56.888889z"/></svg>
              <span class="tab-label">逻辑</span>
            </div>
          </div>
          <div class="survey-sidebar-panel-tabs-content">
            <!-- 题目 -->
            <template v-if="middleTab==='item'">
              <el-tabs v-model="sideSubTab" class="side-sub-tabs">
                <el-tab-pane label="题型" name="types">
                  <div class="question-panel">
                    <div class="type-tabs-bar" ref="tabsBarRef">
                      <button class="tab-scroll-btn tab-scroll-prev" @click="scrollTabs(-1)" :disabled="atTabStart">‹</button>
                      <div class="tab-scroll-viewport" ref="tabViewportRef">
                        <div class="tab-scroll-track" ref="tabTrackRef">
                          <button
                            v-for="cat in categories" :key="cat.name"
                            class="type-tab-btn"
                            :class="{ active: activeCategory===cat.name }"
                            @click="activeCategory=cat.name"
                          >{{ cat.label }}</button>
                        </div>
                      </div>
                      <button class="tab-scroll-btn tab-scroll-next" @click="scrollTabs(1)" :disabled="atTabEnd">›</button>
                    </div>
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
                  </div>
                </el-tab-pane>
                <el-tab-pane label="题库" name="bank">
                  <div class="bank-panel">
                    <div class="bank-search">
                      <el-input v-model="bankKeyword" placeholder="搜索题目..." size="small" clearable @input="loadBank" />
                    </div>
                    <div class="bank-list">
                      <div v-for="q in bankQuestions" :key="q.id" class="bank-item" @dblclick="addFromBank(q)">
                        <question-icon :type="q.type" class="bank-icon" />
                        <span class="bank-title">{{ q.title || '未命名' }}</span>
                        <span class="bank-type">{{ q.type }}</span>
                        <el-button size="small" text type="primary" @click.stop="addFromBank(q)">+添加</el-button>
                      </div>
                      <el-empty v-if="!bankQuestions.length && !bankLoading" description="题库暂无题目" :image-size="40" />
                      <div v-if="bankLoading" class="bank-loading">加载中...</div>
                    </div>
                  </div>
                </el-tab-pane>
                <el-tab-pane label="大纲" name="outline">
                  <div class="outline-tree">
                    <div class="tree-root">
                      <span class="tree-root-icon">📋</span>
                      <span class="tree-root-title">{{ outlineRoot.title }}</span>
                      <span class="tree-root-count">{{ outlineRoot.children.length }}题</span>
                    </div>
                    <div class="tree-root-line"></div>
                    <div class="tree-children">
                      <div v-for="child in outlineRoot.children" :key="child.q.id" class="tree-child" :class="{ active: child.q.id === selected?.id }" @click="selectQuestion(child.q.id)">
                        <div class="tree-connector">
                          <span class="tree-line"></span>
                          <span class="tree-branch"></span>
                        </div>
                        <div class="tree-child-body">
                          <span v-if="child.q.type !== 'description'" class="tree-index">{{ child.index }}.</span>
                          <question-icon :type="child.q.type" class="tree-icon" />
                          <span class="tree-title">{{ child.q.title || '未命名' }}</span>
                          <span class="tree-type">{{ child.q.type }}</span>
                        </div>
                      </div>
                    </div>
                    <el-empty v-if="!questions.length" description="暂无题目" :image-size="40" />
                  </div>
                </el-tab-pane>
              </el-tabs>
            </template>
            <!-- 外观 -->
            <div v-if="middleTab==='appearance'" class="appearance-panel">
              <el-tabs v-model="appearanceTab" class="appearance-tabs">
                <el-tab-pane label="背景图" name="bg">
                  <div class="appearance-grid">
                    <div v-for="(url, i) in form.backgroundImages" :key="i" class="appearance-thumb" @click="applyToPreview(url)">
                      <img :src="url" />
                      <div class="appearance-overlay"><span>点击应用</span></div>
                      <button class="appearance-remove" @click.stop="removeImage('backgroundImages', i)">✕</button>
                    </div>
                    <el-upload :action="`/admin/survey/resource_upload`" :show-file-list="false" :on-success="handleBgSuccess" :on-error="()=>ElMessage.error('上传失败')" :headers="uploadHeaders" accept="image/*" :data="{ surveyId: form.id, resType: 'bg' }" :before-upload="checkSaved">
                      <div class="appearance-add">+</div>
                    </el-upload>
                  </div>
                  <el-empty v-if="!form.backgroundImages.length" description="暂无背景图" :image-size="30" />
                </el-tab-pane>
                <el-tab-pane label="页眉图" name="header">
                  <div class="appearance-grid">
                    <div v-for="(url, i) in form.headerImages" :key="i" class="appearance-thumb" @click="applyToPreview(url)">
                      <img :src="url" />
                      <div class="appearance-overlay"><span>点击应用</span></div>
                      <button class="appearance-remove" @click.stop="removeImage('headerImages', i)">✕</button>
                    </div>
                    <el-upload :action="`/admin/survey/resource_upload`" :show-file-list="false" :on-success="handleHeaderSuccess" :on-error="()=>ElMessage.error('上传失败')" :headers="uploadHeaders" accept="image/*" :data="{ surveyId: form.id, resType: 'header' }" :before-upload="checkSaved">
                      <div class="appearance-add">+</div>
                    </el-upload>
                  </div>
                  <el-empty v-if="!form.headerImages.length" description="暂无页眉图" :image-size="30" />
                </el-tab-pane>
              </el-tabs>
            </div>
            <!-- 逻辑 -->
            <div v-if="middleTab==='logic'" class="logic-panel">
              <p style="color:#909399;padding:20px;text-align:center">逻辑功能待实现</p>
            </div>
          </div>
        </div>

        <!-- 中: 画布 -->
        <div class="survey-main-panel">
          <div class="survey-main-panel-toolbar">
            <div class="toolbar-left">
              <el-button-group class="toolbar-btn-group">
                <el-tooltip content="撤销" placement="bottom">
                  <el-button text size="small" class="toolbar-btn" disabled>
                    <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><polyline points="1 4 1 10 7 10"/><path d="M3.51 15a9 9 0 1 0 2.13-9.36L1 10"/></svg>
                  </el-button>
                </el-tooltip>
                <el-tooltip content="重做" placement="bottom">
                  <el-button text size="small" class="toolbar-btn" disabled>
                    <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><polyline points="23 4 23 10 17 10"/><path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/></svg>
                  </el-button>
                </el-tooltip>
                <el-tooltip content="快捷键" placement="bottom">
                  <el-button text size="small" class="toolbar-btn">
                    <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><rect x="2" y="4" width="20" height="16" rx="2"/><path d="M6 8h.01M10 8h.01M14 8h.01M18 8h.01M8 12h.01M12 12h.01M16 12h.01M6 16h.01M10 16h.01M14 16h.01M18 16h.01"/></svg>
                  </el-button>
                </el-tooltip>
                <el-tooltip content="文本导入" placement="bottom">
                  <el-button text size="small" class="toolbar-btn">
                    <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
                  </el-button>
                </el-tooltip>
              </el-button-group>
            </div>
            <div class="toolbar-right">
              <el-button-group class="toolbar-btn-group">
                <el-tooltip content="编辑" placement="bottom">
                  <el-button text size="small" class="toolbar-btn" :class="{ active: sideSubTab==='types' }" @click="sideSubTab='types'">
                    <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
                  </el-button>
                </el-tooltip>
                <el-tooltip content="JSON" placement="bottom">
                  <el-button text size="small" class="toolbar-btn" @click="exportJSON">
                    <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><polyline points="16 18 22 12 16 6"/><polyline points="8 6 2 12 8 18"/></svg>
                  </el-button>
                </el-tooltip>
                <el-tooltip content="预览" placement="bottom">
                  <el-button text size="small" class="toolbar-btn" @click="preview">
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
          <div class="survey-main-panel-content">
            <div class="editor-wrapper">
              <div class="editor">
                <!-- 问卷头 -->
                <div class="header">
                  <div class="prefix">问卷</div>
                  <div class="header-content">
                    <el-input v-model="form.title" placeholder="问卷标题" class="header-title-input" maxlength="100" />
                  </div>
                </div>

                <!-- 题目列表 -->
                <div class="questions-area">
                  <draggable-list :questions="questions" @update:questions="onQuestionsUpdate" @select="selectQuestion" :selected-id="selected?.id??null" editing @remove="removeQuestionById" />
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
            <h3>题目设置</h3>
            <el-form label-position="top" size="small">
              <el-form-item label="ID"><el-input v-model="selected.id" :disabled="!!selected._existing" /></el-form-item>
              <el-form-item label="占位提示"><el-input v-model="selected.placeholder" /></el-form-item>
              <el-row :gutter="8">
                <el-col :span="12"><el-form-item label="必填"><el-switch v-model="selected.required" /></el-form-item></el-col>
                <el-col :span="12"><el-form-item label="只读"><el-switch v-model="selected.readOnly" /></el-form-item></el-col>
              </el-row>
              <el-divider border-style="dashed" style="margin:12px 0" />
              <el-collapse v-model="mediaCollapseActive" accordion>
                <el-collapse-item title="媒体设置" name="media">
                  <el-form-item label="媒体类型">
                    <el-radio-group v-model="selected.mediaType" @change="onMediaTypeChange">
                      <el-radio value="">无</el-radio>
                      <el-radio value="image">图片</el-radio>
                      <el-radio value="video">视频</el-radio>
                      <el-radio value="audio">音频</el-radio>
                    </el-radio-group>
                  </el-form-item>
                  <el-form-item v-if="selected.mediaType" label="媒体地址">
                    <el-input v-model="selected.mediaUrl" placeholder="输入URL或上传">
                      <template #append>
                        <el-button @click="triggerMediaUpload">上传</el-button>
                      </template>
                    </el-input>
                  </el-form-item>
                  <el-row v-if="selected.mediaType" :gutter="8">
                    <el-col :span="12">
                      <el-form-item label="显示宽度">
                        <el-input v-model="selected.mediaWidth" placeholder="auto">
                          <template #append>px</template>
                        </el-input>
                      </el-form-item>
                    </el-col>
                    <el-col :span="12">
                      <el-form-item label="对齐方式">
                        <el-radio-group v-model="selected.mediaAlign">
                          <el-radio value="left">左</el-radio>
                          <el-radio value="center">中</el-radio>
                          <el-radio value="right">右</el-radio>
                        </el-radio-group>
                      </el-form-item>
                    </el-col>
                  </el-row>
                </el-collapse-item>
              </el-collapse>
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

      <!-- 概览视图 -->
      <div v-show="activeView==='overview'" class="overview-wrapper">
        <div class="overview-scroll">
          <!-- 日收集问卷图表 -->
          <div class="overview-card">
            <div class="overview-card-title">📊 日收集问卷</div>
            <div class="chart-placeholder">
              <div class="chart-bars">
                <div v-for="i in 7" :key="i" class="chart-bar-group">
                  <div class="chart-bar" :style="{ height: (20 + Math.random()*60) + 'px' }"></div>
                  <div class="chart-label">{{ ['一','二','三','四','五','六','日'][i-1] }}</div>
                </div>
              </div>
            </div>
          </div>

          <!-- 系统设置 -->
          <div class="overview-card">
            <div class="overview-card-title">⚙️ 系统设置</div>
            <div class="overview-settings">
              <div class="setting-row">
                <span class="setting-label">访问链接</span>
                <div class="setting-value">
                  <el-input v-if="form.id" v-model="publicUrl" readonly size="small">
                    <template #append><el-button @click="copyLink" size="small">复制</el-button></template>
                  </el-input>
                  <span v-else class="save-tip">请先保存问卷</span>
                </div>
              </div>
              <div class="setting-row">
                <span class="setting-label">二维码</span>
                <div class="setting-value">
                  <el-button size="small" @click="genQR">生成二维码</el-button>
                  <div v-if="qrUrl" class="qr-box" style="margin-top:8px"><el-image :src="qrUrl" style="width:120px;height:120px;border-radius:6px" /></div>
                </div>
              </div>
              <div class="setting-row">
                <span class="setting-label">问卷状态</span>
                <div class="setting-value">
                  <el-switch v-model="form.statusBool" :active-value="1" :inactive-value="0" active-text="已开启" inactive-text="已停用" />
                </div>
              </div>
              <div class="setting-row">
                <span class="setting-label">自动保存</span>
                <div class="setting-value">
                  <el-switch v-model="form.autoSave" active-text="开启" inactive-text="关闭" />
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 数据视图 -->
      <div v-show="activeView==='data'" class="setting-wrapper">
        <div class="setting-scroll">
          <div class="setting-group">
            <div class="group-title">数据统计</div>
            <p style="color:#909399;padding:20px">数据统计功能待实现</p>
          </div>
        </div>
      </div>

      <!-- 报表视图 -->
      <div v-show="activeView==='report'" class="setting-wrapper">
        <div class="setting-scroll">
          <div class="setting-group">
            <div class="group-title">报表</div>
            <p style="color:#909399;padding:20px">报表功能待实现</p>
          </div>
        </div>
      </div>

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
import { ref, computed, reactive, onMounted, nextTick, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useRoute, useRouter } from 'vue-router'
import { adminApi } from '../../api'
import DraggableList from '../survey/formkit/DraggableList.vue'

import ValidateEditor from '../survey/formkit/ValidateEditor.vue'
import CalcEditor from '../survey/formkit/CalcEditor.vue'
import LogicEditor from '../survey/formkit/LogicEditor.vue'
import QuestionIcon from '../survey/formkit/QuestionIcon.vue'

const route = useRoute()
const router = useRouter()

const form = reactive<any>({
  id: 0, title: '', description: '', category: '', tags: '', mode: 'exam',
  visibility: 0, allowMultiBool: 0, anonymousBool: 0, showResultBool: 0,
  startDate: null, endDate: null, maxResponse: 0, statusBool: 1, deptIds: '',
  questionNumber: true, progressBar: false, autoSave: false,
  onePageOneQuestion: false, answerSheetVisible: true, copyEnabled: true,
  password: '', triggerType: 'onBlur', loginRequired: false,
  enableUpdate: false, transcriptVisible: true, rankVisible: false,
  redirectUrl: '', endContent: '',
  examRankingEnabled: false, exerciseMode: false, randomOrder: false,
  minSubmitMinutes: 0, maxSubmitMinutes: 0,
  duration: 60, maxAttempts: 1, showScore: 1,
        backgroundImages: [] as string[], headerImages: [] as string[]
})

const activeView = ref('edit')
const middleTab = ref('item')
const sideSubTab = ref('types')
const activeCategory = ref('')
const appearanceTab = ref('bg')
const saving = ref(false)
const bankKeyword = ref('')
const bankQuestions = ref<any[]>([])
const bankLoading = ref(false)
let bankTimer: any = null
async function loadBank() {
  clearTimeout(bankTimer)
  if (!form.id) return
  bankTimer = setTimeout(async () => {
    bankLoading.value = true
    try {
      const res: any = await adminApi.surveyQuestionBankList({ keyword: bankKeyword.value, page: 1, pageSize: 100 })
      bankQuestions.value = res.list || []
    } catch { bankQuestions.value = [] }
    finally { bankLoading.value = false }
  }, 300)
}
function addFromBank(q: any) {
  let schema = {}
  try { schema = JSON.parse(q.schema) } catch {}
  const newQ: any = {
    id: ++idCounter,
    type: q.type,
    title: q.title || '未命名',
    schema: schema
  }
  questions.value.push(newQ)
  selected.value = newQ
}

interface OutlineNode {
  key: string
  title: string
  isRoot?: boolean
  children: { q: any; index: number }[]
}
const outlineRoot = computed<OutlineNode>(() => ({
  key: 'root',
  title: form.title || '未命名考试',
  isRoot: true,
  children: questions.value
    .filter((q: any) => q.type !== 'divider')
    .map((q: any, i: number) => ({ q, index: i + 1 }))
}))
const tabsBarRef = ref<HTMLElement | null>(null)
const tabViewportRef = ref<HTMLElement | null>(null)
const tabTrackRef = ref<HTMLElement | null>(null)
const scrollPos = ref(0)
const atTabStart = computed(() => scrollPos.value <= 0)
const atTabEnd = ref(false)
function scrollTabs(dir: number) {
  const vp = tabViewportRef.value
  const tr = tabTrackRef.value
  if (!vp || !tr) return
  const step = vp.clientWidth * 0.6
  let newPos = scrollPos.value + dir * step
  newPos = Math.max(0, Math.min(newPos, tr.scrollWidth - vp.clientWidth))
  scrollPos.value = newPos
  tr.style.transform = `translateX(${-newPos}px)`
  atTabEnd.value = newPos >= tr.scrollWidth - vp.clientWidth - 1
}
function updateTabScroll() {
  const vp = tabViewportRef.value
  const tr = tabTrackRef.value
  if (!vp || !tr) return
  scrollPos.value = 0
  tr.style.transform = 'translateX(0)'
  atTabEnd.value = tr.scrollWidth <= vp.clientWidth
}
const collapseActive = ref('')
const mediaCollapseActive = ref('')

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

interface Question { id: string; type: string; title: string; description?: string; required: boolean; placeholder?: string; props?: any; validate?: any[]; logic?: any[]; calcValue?: any; readOnly?: boolean; dataType?: string; examScore?: number; examCorrectAnswer?: string; examAnalysis?: string; mediaType?: string; mediaUrl?: string; mediaWidth?: string; mediaAlign?: string; [key: string]: any }
const questions = ref<Question[]>([])
const selected = ref<Question | null>(null)
let idCounter = 0

function genId() { idCounter++; return 'q' + idCounter }

function addQuestion(t: any) {
  const q: Question = { id: genId(), type: t.type, title: t.displayName, required: false, readOnly: false, dataType: '', placeholder: '', mediaType: '', mediaUrl: '', mediaWidth: '', mediaAlign: 'center', props: t.defaultProps ? JSON.parse(JSON.stringify(t.defaultProps)) : {} }
  if (t.type === 'judge') {
    if (!q.props) q.props = {}
    q.props.options = [{ label: '对', value: 'true' }, { label: '错', value: 'false' }]
  }
  if (t.type === 'select' || t.type === 'radio' || t.type === 'checkbox' || t.type === 'picker' || t.type === 'matrixRadio') {
    if (!q.props) q.props = {}
    if (!q.props.options) q.props.options = [{ value: '选项1', label: '选项1' }, { value: '选项2', label: '选项2' }]
  }
  questions.value.push(q)
  selected.value = q
}

function selectQuestion(id: string) { selected.value = questions.value.find(q => q.id === id) || null }
function onQuestionsUpdate(newQ: Question[]) {
  questions.value.splice(0, questions.value.length, ...newQ)
  if (selected.value) {
    selected.value = questions.value.find(q => q.id === selected.value!.id) || null
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

function onValidateUpdate(v: any) { if (selected.value) selected.value.validate = v }
function removeQuestionById(id: string) {
  questions.value = questions.value.filter(q => q.id !== id)
  if (selected.value?.id === id) selected.value = null
}
function onCalcUpdate(v: any) { if (selected.value) selected.value.calcValue = v }
function onLogicUpdate(v: any) { if (selected.value) selected.value.logic = v }

function onMediaTypeChange(v: string) {
  if (!selected.value) return
  selected.value.mediaType = v
  if (!v) { selected.value.mediaUrl = ''; selected.value.mediaWidth = '' }
}
function triggerMediaUpload() {
  const input = document.createElement('input')
  input.type = 'file'
  input.accept = 'image/*,video/*,audio/*'
  input.onchange = () => {
    const file = input.files?.[0]
    if (!file || !selected.value) return
    const url = URL.createObjectURL(file)
    selected.value.mediaUrl = url
    if (!selected.value.mediaType) {
      if (file.type.startsWith('video/')) selected.value.mediaType = 'video'
      else if (file.type.startsWith('audio/')) selected.value.mediaType = 'audio'
      else selected.value.mediaType = 'image'
    }
  }
  input.click()
}

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

function exportJSON() {
  exportedJson.value = JSON.stringify({ version: '2.0', questions: questions.value, setting: {} }, null, 2)
  exportDialogVisible.value = true
}
function preview() {
  window.open('/survey/preview?id=' + form.id, '_blank')
}
function openQuestionBank() {
  ElMessage.info('题库功能待实现')
}

const token = localStorage.getItem('admin_token') || ''
const uploadHeaders = { Authorization: token }
function handleBgSuccess(res: any) {
  if (!res.data?.url) { ElMessage.error('上传失败：响应数据异常'); return }
  form.backgroundImages.push(res.data.domain + res.data.url)
}
function handleHeaderSuccess(res: any) {
  if (!res.data?.url) { ElMessage.error('上传失败：响应数据异常'); return }
  form.headerImages.push(res.data.domain + res.data.url)
}
function removeImage(field: string, idx: number) { (form as any)[field].splice(idx, 1) }
function applyToPreview(url: string) { ElMessage.success('已应用到问卷') }
function checkSaved() {
  if (!form.id) { ElMessage.warning('请先保存问卷'); return false }
  return true
}

async function loadResources() {
  if (!form.id) return
  try {
    const res: any = await adminApi.surveyResourceList({ surveyId: form.id })
    const list: any[] = res.data || []
    if (!Array.isArray(list)) return
    form.backgroundImages = list.filter((r: any) => r.type === 'bg').map((r: any) => (r.domain || '') + r.url)
    form.headerImages = list.filter((r: any) => r.type === 'header').map((r: any) => (r.domain || '') + r.url)
  } catch {}
}
watch(middleTab, (v) => { if (v === 'appearance') loadResources() })
watch(sideSubTab, (v) => { if (v === 'bank') loadBank() })

async function load() {
  const id = Number(route.query.id || 0)
  if (!id) return
  try {
    const res: any = await adminApi.examDetail(id)
    Object.assign(form, {
      id: res.id, title: res.title, description: res.desc || '',
      startDate: res.startTime || null, endDate: res.endTime || null,
      duration: res.duration, maxAttempts: res.maxAttempts,
      showScore: res.showScore, statusBool: res.status
    })
    if (res.settings) {
      try {
        const s = JSON.parse(res.settings)
        Object.assign(form, {
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
          minSubmitMinutes: s.minSubmitMinutes || 0, maxSubmitMinutes: s.maxSubmitMinutes || 0,
          backgroundImages: s.backgroundImages || [], headerImages: s.headerImages || []
        })
      } catch {}
    }
    if (res.schema) {
      try {
        const sch = JSON.parse(res.schema)
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
      maxSubmitMinutes: form.maxSubmitMinutes,
      backgroundImages: form.backgroundImages, headerImages: form.headerImages
    })
    const payload: any = {
      title: form.title, desc: form.description,
      startTime: form.startDate || 0, endTime: form.endDate || 0,
      duration: form.duration, maxAttempts: form.maxAttempts,
      showScore: form.showScore, status: form.statusBool,
      schema, settings
    }
    if (form.id) payload.id = form.id
    const res: any = await adminApi.examSave(payload)
    if (!form.id && res.id) form.id = res.id
    ElMessage.success('保存成功')
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
  await nextTick()
  updateTabScroll()
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
.nav-btn {
  width:36px; height:36px; min-width:36px; padding:0; margin:0;
  border:none; border-radius:8px; color:#999; font-size:13px; cursor:pointer;
  display:flex; align-items:center; justify-content:center; flex-shrink:0;
  background:transparent; outline:none; line-height:1;
}
.nav-btn:hover { background:#f0f0f0; color:#666; }
.nav-btn.active { background:#fff; color:#fb454c; box-shadow:0 1px 4px rgba(0,0,0,0.08); }
.nav-btn svg { display:block; flex-shrink:0; }

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

/* 题型 / 大纲 tab 切换 */
.side-sub-tabs { display:flex; flex-direction:column; height:100%; }
.side-sub-tabs :deep(.el-tabs__header) { margin:0; padding:0 8px; background:#fafafa; border-bottom:1px solid #e8e8e8; }
.side-sub-tabs :deep(.el-tabs__nav-wrap) { margin-bottom:0; }
.side-sub-tabs :deep(.el-tabs__active-bar) { background:#fb454c; height:2px; }
.side-sub-tabs :deep(.el-tabs__item) {
  padding:0 12px; font-size:13px; height:32px; line-height:32px; color:#666; transition:color 0.15s;
}
.side-sub-tabs :deep(.el-tabs__item:hover) { color:#333; }
.side-sub-tabs :deep(.el-tabs__item.is-active) { color:#fb454c; font-weight:500; }
.side-sub-tabs :deep(.el-tabs__content) { flex:1; overflow-y:auto; }

/* 题型分类标签（自定义按钮滚动） */
.question-panel { padding:8px; }
.type-tabs-bar { display:flex; align-items:stretch; gap:0; margin-bottom:8px; }
.tab-scroll-viewport { flex:1; overflow:hidden; }
.tab-scroll-track { display:flex; gap:4px; transition:transform .2s ease; }
.tab-scroll-btn {
  flex:0 0 20px; border:1px solid #e8e8e8; border-radius:4px; background:#fff;
  color:#999; cursor:pointer; font-size:14px; line-height:1;
  display:flex; align-items:center; justify-content:center; padding:0;
  transition:all .15s;
}
.tab-scroll-btn:hover:not(:disabled) { color:#fb454c; border-color:#fb454c; }
.tab-scroll-btn:disabled { opacity:0.3; cursor:default; }
.type-tab-btn {
  flex:0 0 auto; min-width:0; padding:3px 6px; font-size:12px; line-height:20px; height:26px;
  border:1px solid #e8e8e8; border-radius:4px; color:#666; cursor:pointer; text-align:center;
  background:#fff; transition:all 0.15s; white-space:nowrap;
}
.type-tab-btn:hover { border-color:#fb454c; color:#fb454c; }
.type-tab-btn.active { color:#fff; background:#fb454c; border-color:#fb454c; font-weight:500; }

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

/* 外观面板 */
.appearance-panel { padding:12px; }
.appearance-panel :deep(.el-tabs__header) { margin:0 0 8px; }
.appearance-panel :deep(.el-tabs__item) { font-size:12px; padding:0 12px; }
.appearance-grid { display:flex; flex-wrap:wrap; gap:8px; }
.appearance-add, .appearance-thumb {
  width:72px; height:72px; border-radius:6px; overflow:hidden; flex-shrink:0;
}
.appearance-add {
  display:flex; align-items:center; justify-content:center;
  border:2px dashed #e8e8e8; background:#fff; cursor:pointer;
  font-size:24px; color:#ccc; transition:all .15s;
}
.appearance-add:hover { border-color:#fb454c; color:#fb454c; }
.appearance-thumb {
  position:relative; cursor:pointer;
  border:2px solid transparent; transition:border-color .15s;
}
.appearance-thumb:hover { border-color:#fb454c; }
.appearance-thumb img { display:block; width:100%; height:100%; object-fit:cover; }
.appearance-overlay {
  position:absolute; inset:0; display:flex; align-items:center; justify-content:center;
  background:rgba(0,0,0,0.4); opacity:0; transition:opacity .15s;
  font-size:11px; color:#fff;
}
.appearance-thumb:hover .appearance-overlay { opacity:1; }
.appearance-remove {
  position:absolute; top:2px; right:2px; width:18px; height:18px; border-radius:50%;
  border:none; background:rgba(0,0,0,0.5); color:#fff; font-size:10px; line-height:1;
  cursor:pointer; display:flex; align-items:center; justify-content:center;
}

/* ======= 中间画布 ======= */
.survey-main-panel {
  display:flex; flex:1 1; flex-direction:column; overflow:hidden;
  background-color:#f7f8fa; padding:0;
}
.survey-main-panel-toolbar {
  display:flex; align-items:center; justify-content:space-between;
  padding:8px 16px; background:#fff; border-bottom:1px solid #e8e8e8; gap:12px;
}
.toolbar-actions { display:flex; align-items:center; gap:8px; flex-shrink:0; }
.toolbar-btn-group { display:inline-flex; align-items:center; background:#f5f6f8; border-radius:6px; padding:2px; gap:0; }
.toolbar-btn-group .el-button { border:none; }

.survey-main-panel-content { flex:1; overflow-y:auto; padding:20px 40px; }
.editor-wrapper { max-width:720px; margin:0 auto; }
.editor { background:#fff; border-radius:12px; box-shadow:0 2px 12px rgba(0,0,0,0.06); overflow:hidden; }

/* 问卷头 */
.header {
  padding:24px 28px 12px; text-align:center; border-bottom:1px solid #f0f0f0;
}
.prefix { font-size:13px; color:#999; margin-bottom:4px; }
.header-content .title { font-size:22px; font-weight:600; color:#303133; line-height:1.4; }
.header-title-input :deep(.el-input__wrapper) { box-shadow:none!important; padding:0; background:transparent; }
.header-title-input :deep(.el-input__inner) { font-size:22px; font-weight:600; color:#303133; text-align:center; border:none; }

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

/* ======= 概览视图 ======= */
.overview-wrapper { height:100%; overflow-y:auto; background:#f5f6f8; padding:20px 24px; }
.overview-scroll { max-width:860px; margin:0 auto; }
.overview-card { margin-bottom:20px; padding:20px; background:#fff; border-radius:8px; box-shadow:0 1px 4px rgba(0,0,0,0.04); }
.overview-card-title { font-size:14px; font-weight:600; color:#303133; margin-bottom:16px; }
.chart-placeholder { padding:10px 0; }
.chart-bars { display:flex; align-items:flex-end; justify-content:space-around; height:150px; gap:8px; }
.chart-bar-group { display:flex; flex-direction:column; align-items:center; gap:4px; flex:1; }
.chart-bar { width:100%; max-width:40px; border-radius:4px 4px 0 0; background:linear-gradient(to top,#fb454c,#ff7875); transition:height 0.3s; min-height:8px; }
.chart-label { font-size:11px; color:#999; }
.overview-settings { display:flex; flex-direction:column; gap:14px; }
.setting-row { display:flex; align-items:center; gap:12px; }
.setting-label { font-size:13px; color:#555; min-width:80px; flex-shrink:0; }
.setting-value { flex:1; }

/* 题库面板 */
.bank-panel { padding:8px; display:flex; flex-direction:column; height:100%; }
.bank-search { margin-bottom:8px; }
.bank-list { flex:1; overflow-y:auto; }
.bank-item {
  display:flex; align-items:center; gap:6px; padding:6px 8px;
  border-radius:6px; cursor:pointer; transition:all 0.12s; line-height:20px;
}
.bank-item:hover { background:#fff5f5; color:#fb454c; }
.bank-item.active { background:#fff0f0; color:#fb454c; }
.bank-icon { flex-shrink:0; font-size:14px; width:18px; text-align:center; color:#999; }
.bank-item:hover .bank-icon { color:#fb454c; }
.bank-index { font-size:12px; color:#bbb; min-width:20px; }
.bank-title { font-size:13px; flex:1; overflow:hidden; text-overflow:ellipsis; white-space:nowrap; }
.bank-type { font-size:10px; color:#bbb; }

/* 大纲树 */
.outline-tree { padding:4px 0; }
.tree-root {
  display:flex; align-items:center; gap:8px; padding:8px 10px;
  border-radius:8px; font-weight:600; font-size:14px; color:#333;
  background:#f5f7fa;
}
.tree-root-icon { font-size:16px; }
.tree-root-title { flex:1; overflow:hidden; text-overflow:ellipsis; white-space:nowrap; }
.tree-root-count { font-size:11px; color:#999; font-weight:400; }
.tree-root-line {
  height:12px; margin-left:29px; border-left:2px dashed #d0d5dd;
}
.tree-children { padding-left:20px; }
.tree-child {
  display:flex; align-items:stretch; cursor:pointer; border-radius:6px;
  transition:all 0.12s; min-height:32px; position:relative;
}
.tree-child:hover { background:#fff5f5; }
.tree-child.active { background:#fff0f0; }
.tree-connector {
  position:relative; width:20px; flex-shrink:0;
}
.tree-line {
  position:absolute; left:50%; top:0; bottom:0;
  border-left:2px dashed #d0d5dd; margin-left:-1px;
}
.tree-child:last-child .tree-line { bottom:50%; }
.tree-branch {
  position:absolute; left:50%; top:50%; width:10px;
  border-top:2px dashed #d0d5dd;
}
.tree-child-body {
  display:flex; align-items:center; gap:6px; padding:6px 8px; flex:1; min-width:0;
  border-radius:0 6px 6px 0;
}
.tree-child.active .tree-child-body { font-weight:500; }
.tree-index { font-size:12px; color:#bbb; min-width:22px; text-align:right; flex-shrink:0; }
.tree-icon { flex-shrink:0; font-size:14px; width:18px; text-align:center; color:#999; }
.tree-child:hover .tree-icon,
.tree-child.active .tree-icon { color:#fb454c; }
.tree-title { font-size:13px; flex:1; overflow:hidden; text-overflow:ellipsis; white-space:nowrap; }
.tree-type { font-size:10px; color:#bbb; padding:1px 5px; border-radius:3px; background:#f5f5f5; flex-shrink:0; }
.tree-child:hover .tree-type,
.tree-child.active .tree-type { background:#fee; color:#fb454c; }
.tree-icon { flex-shrink:0; font-size:14px; width:18px; text-align:center; color:#999; }
.tree-child:hover .tree-icon,
.tree-child.active .tree-icon { color:#fb454c; }
.tree-title { font-size:13px; flex:1; overflow:hidden; text-overflow:ellipsis; white-space:nowrap; }
.tree-type { font-size:10px; color:#bbb; padding:1px 5px; border-radius:3px; background:#f5f5f5; flex-shrink:0; }
.tree-child:hover .tree-type,
.tree-child.active .tree-type { background:#fee; color:#fb454c; }
</style>
