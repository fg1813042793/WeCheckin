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
        <div class="survey-sidebar-panel" :class="{ compact: middleTab==='logic' }">
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
          <div v-show="middleTab !== 'logic'" class="survey-sidebar-panel-tabs-content">
            <el-tabs v-if="middleTab==='item'" v-model="sideSubTab" class="side-sub-tabs">
              <el-tab-pane label="题型" name="types">
                <div class="question-panel">
                  <div class="type-tabs-bar" ref="tabsBarRef">
                    <button class="tab-scroll-btn tab-scroll-prev" @click="scrollTabs(-1)" :disabled="atTabStart">‹</button>
                    <div class="tab-scroll-viewport" ref="tabViewportRef">
                      <div class="tab-scroll-track" ref="tabTrackRef">
                        <button v-for="cat in categories" :key="cat.name" class="type-tab-btn" :class="{ active: activeCategory===cat.name }" @click="activeCategory=cat.name">{{ cat.label }}</button>
                      </div>
                    </div>
                    <button class="tab-scroll-btn tab-scroll-next" @click="scrollTabs(1)" :disabled="atTabEnd">›</button>
                  </div>
                  <div class="question-type">
                    <dl class="menu-group" v-for="(items, group) in groupedBySub" :key="group">
                      <dd v-for="t in items" :key="t.type" class="menu-group-item" @click="addQuestion(t)">
                        <span class="itemIcon"><question-icon :type="t.type" /></span>
                        <span class="item-label">{{ t.displayName }}</span>
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
                    <template v-if="bankTree.length">
                      <div v-for="cat in bankTree" :key="cat.key" class="bank-cat">
                        <div class="bank-cat-title" @click="toggleBankExpand(`cat:${cat.expandKey}`)">
                          <span class="bank-arrow" :class="{ expanded: cat._expanded }"></span>
                          <span class="bank-cat-name">{{ cat.label }}</span>
                          <span class="bank-count">{{ cat.count }} 题</span>
                        </div>
                        <div v-show="cat._expanded" class="bank-cat-body">
                          <div v-for="grp in cat.children" :key="grp.key" class="bank-type-group">
                            <div class="bank-type-title" @click="toggleBankExpand(`type:${cat.expandKey}|${grp.expandKey}`)">
                              <span class="bank-arrow" :class="{ expanded: grp._expanded }"></span>
                              <question-icon :type="grp.label" class="bank-icon" />
                              <span class="bank-type-name">{{ typeName(grp.label) }}</span>
                              <span class="bank-count">{{ grp.children.length }} 题</span>
                            </div>
                            <div v-show="grp._expanded" class="bank-type-body">
                              <div v-for="q in grp.children" :key="q.id" class="bank-item" @click="addFromBank(q)">
                                <span class="bank-item-main">
                                  <span class="bank-title" :title="bankQuestionTitle(q)">{{ bankQuestionTitle(q) }}</span>
                                  <span class="bank-meta">#{{ q.id }}</span>
                                </span>
                                <span class="bank-type">{{ typeName(q.type) }}</span>
                              </div>
                            </div>
                          </div>
                        </div>
                      </div>
                    </template>
                    <el-empty v-if="!bankTree.length && !bankLoading" description="题库暂无题目" :image-size="40" />
                    <div v-if="bankLoadingVisible" class="bank-loading">加载中...</div>
                  </div>
                </div>
              </el-tab-pane>
              <el-tab-pane label="大纲" name="outline">
                <div class="outline-tree">
                  <div class="tree-root">
                    <div class="tree-root-main">
                      <span class="tree-root-icon">📋</span>
                      <span class="tree-root-title">{{ outlineRoot.title }}</span>
                    </div>
                    <span class="tree-root-count">{{ outlineRoot.children.length }}题</span>
                  </div>
                  <div v-if="outlineRoot.children.length" class="tree-children">
                    <div v-for="child in outlineRoot.children" :key="child.q.id" class="tree-child" :class="{ active: child.q.id === selected?.id }" @click="selectQuestion(child.q.id, true)">
                      <span v-if="child.q.type !== 'description'" class="tree-index">{{ child.index }}.</span>
                      <div class="tree-child-body">
                        <question-icon :type="child.q.type" class="tree-icon" />
                        <span class="tree-title" :title="outlineQuestionTitle(child.q)">{{ outlineQuestionTitle(child.q) }}</span>
                        <span v-if="child.q.required" class="tree-required">必</span>
                        <span class="tree-type">{{ typeName(child.q.type) }}</span>
                      </div>
                    </div>
                  </div>
                  <el-empty v-else class="outline-empty" description="暂无题目" :image-size="40" />
                </div>
              </el-tab-pane>
            </el-tabs>
            <div v-if="middleTab==='appearance'" class="appearance-panel">
              <el-tabs v-model="appearanceTab" class="appearance-tabs">
                <el-tab-pane label="背景图" name="bg">
                  <div class="appearance-grid">
                    <div v-for="(item, i) in allBgResources" :key="i" class="appearance-thumb" :class="{ active: isActiveImg('backgroundImages', item) }" @click="applyImage('backgroundImages', item)">
                      <img :src="(item.domain||'')+item.url" />
                      <div class="appearance-overlay"><span>点击应用</span></div>
                      <button class="appearance-remove" @click.stop="removeResource('backgroundImages', item, i)">✕</button>
                    </div>
                    <el-upload :action="`/api/v2/admin/exam-resources`" :show-file-list="false" :on-success="handleBgSuccess" :on-error="()=>ElMessage.error('上传失败')" :headers="uploadHeaders" accept="image/*" :data="{ examId: form.id, resType: 'bg' }" :before-upload="checkSaved">
                      <div class="appearance-add">+</div>
                    </el-upload>
                  </div>
                  <el-empty v-if="!allBgResources.length" description="暂无背景图" :image-size="30" />
                </el-tab-pane>
                <el-tab-pane label="页眉图" name="header">
                  <div class="appearance-grid">
                    <div v-for="(item, i) in allHeaderResources" :key="i" class="appearance-thumb" :class="{ active: isActiveImg('headerImages', item) }" @click="applyImage('headerImages', item)">
                      <img :src="(item.domain||'')+item.url" />
                      <div class="appearance-overlay"><span>点击应用</span></div>
                      <button class="appearance-remove" @click.stop="removeResource('headerImages', item, i)">✕</button>
                    </div>
                    <el-upload :action="`/api/v2/admin/exam-resources`" :show-file-list="false" :on-success="handleHeaderSuccess" :on-error="()=>ElMessage.error('上传失败')" :headers="uploadHeaders" accept="image/*" :data="{ examId: form.id, resType: 'header' }" :before-upload="checkSaved">
                      <div class="appearance-add">+</div>
                    </el-upload>
                  </div>
                  <el-empty v-if="!allHeaderResources.length" description="暂无页眉图" :image-size="30" />
                </el-tab-pane>
              </el-tabs>
            </div>
          </div>
        </div>

        <!-- 逻辑 -->
        <div v-if="middleTab==='logic'" class="logic-panel logic-full-panel">
          <div class="logic-toolbar">
            <h4 style="margin:0;font-size:14px">自定义逻辑</h4>
            <div style="display:flex;gap:8px">
              <el-button size="small" type="primary" @click="showAddRule = true">+ 增加规则</el-button>
              <el-button size="small" @click="saveLogicRules">保存全部规则</el-button>
            </div>
          </div>
          <div style="font-size:12px;color:#606266;background:#fefced;border:1px solid #faf0c7;border-radius:4px;padding:8px 12px;margin-bottom:12px;line-height:1.8">
            <strong>使用说明：</strong>每一条规则由「条件 + 动作 + 目标」组成。例如：<code>如果回答了Q1第1选项，则显示Q2</code>。
            条件可组合多个（且/或/非），动作可选显示隐藏、必填、跳转等。点击下方「+ 增加规则」添加。
          </div>
          <div class="logic-body">
            <div class="logic-editor-area">
              <div style="font-size:12px;color:#888;margin-bottom:8px">规则列表（{{ logicRuleList.length }}条）</div>
              <div v-if="logicRuleList.length===0" style="color:#999;font-size:13px;padding:40px 0;text-align:center">暂无规则，点击上方「+ 增加规则」添加</div>
              <div v-for="(rule, ri) in logicRuleList" :key="rule.id" class="logic-rule-card">
                <div class="rule-header">
                  <span class="rule-index">#{{ toNumericIndex(ri) + 1 }}</span>
                  <span class="rule-dsl">{{ renderRuleDSL(rule) }}</span>
                  <div class="rule-actions">
                    <el-button text size="small" @click="editRule(toNumericIndex(ri))">编辑</el-button>
                    <el-button text size="small" type="danger" @click="removeRule(toNumericIndex(ri))">删除</el-button>
                  </div>
                </div>
              </div>
            </div>
            <div class="logic-sidebar">
              <div style="font-size:12px;font-weight:600;color:#606266;margin-bottom:6px">案例参考</div>
              <div style="font-size:12px;line-height:2.2">
                <div><code>IF Q1A1 THEN SHOW Q2</code> — 选A时显示Q2</div>
                <div><code>IF Q1A2 THEN HIDE Q3</code> — 选B时隐藏Q3</div>
                <div><code>IF Q1A1 THEN REQUIRED Q2</code> — 选A时Q2必填</div>
                <div><code>IF Q1A1 THEN BRANCH FROM Q1 TO END</code> — 选A结束问卷</div>
                <div><code>IF Q1A1 THEN BRANCH FROM Q1 TO Q5</code> — 选A跳到Q5</div>
                <div><code>IF AND(Q1A1,Q2>18) THEN BRANCH FROM Q2 TO Q6</code> — 多条件跳转</div>
                <div><code>IF Q1A1 THEN CHECK Q2A1</code> — 选A时勾选Q2的选项A</div>
                <div><code>ASSIGNMENT Q3 WITH SUM(Q1,Q2)</code> — Q3赋值为Q1+Q2</div>
                <div><code>VALIDATE Q2 WITH IF(Q2>100,"超出","")</code> — 校验Q2不超100</div>
                <div><code>REPLACE Q2 WITH "你好"</code> — 替换Q2标题</div>
              </div>
            </div>
          </div>
        </div>

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
                  <el-button text size="small" class="toolbar-btn" :class="{ active: panelMode==='json' }" @click="panelMode='json'; exportedJson = JSON.stringify({ version: '2.0', questions }, null, 2)">
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

        <!-- 添加/编辑规则对话框 -->
        <el-dialog v-model="showAddRule" :title="editingRuleIdx>=0?'编辑规则':'增加规则'" width="560px" :close-on-click-modal="false">
          <el-form label-position="top" size="small">
            <el-form-item label="条件类型">
              <el-select v-model="ruleForm.conditionType" style="width:100%">
                <el-option label="简单条件 (单个)" value="simple" />
                <el-option label="且 (AND)" value="and" />
                <el-option label="或 (OR)" value="or" />
                <el-option label="非 (NOT)" value="not" />
                <el-option label="无条件 (赋值/校验/替换)" value="none" />
              </el-select>
            </el-form-item>
            <template v-if="ruleForm.conditionType!=='none'">
              <el-form-item v-for="(cond, ci) in ruleForm.conditions" :key="ci" :label="toNumericIndex(ci) === 0 ? '条件' : '条件 ' + (toNumericIndex(ci) + 1)">
                <div style="display:flex;gap:6px;align-items:center;flex-wrap:wrap">
                  <el-select v-model="cond.questionIdx" placeholder="选择题目" style="width:160px" @change="cond.optionIdx=undefined;cond.operator=undefined">
                    <el-option v-for="(q, qi) in questions" :key="qi" :label="`Q${toNumericIndex(qi) + 1}: ${q.title?.slice(0,20)}`" :value="toNumericIndex(qi)" />
                  </el-select>
                  <el-select v-model="cond.optionIdx" placeholder="选择选项" style="width:120px" v-if="hasOptionsByIndex(cond.questionIdx)" clearable @change="cond.operator='optSelected'">
                    <el-option v-for="(o, oi) in getOptionsByIndex(cond.questionIdx)" :key="oi" :label="o.label" :value="toNumericIndex(oi)" />
                  </el-select>
                  <el-select v-model="cond.operator" placeholder="判断条件" style="width:120px" v-if="cond.optionIdx===undefined">
                    <el-option label="已填写" value="filled" />
                    <el-option label="未填写" value="empty" />
                    <el-option label="等于" value="eq" />
                    <el-option label="大于" value="gt" />
                    <el-option label="小于" value="lt" />
                  </el-select>
                  <el-input v-model="cond.compareValue" placeholder="比较值" style="width:100px" v-if="cond.operator==='eq'||cond.operator==='gt'||cond.operator==='lt'" />
                  <el-button v-if="ruleForm.conditions.length>1 && toNumericIndex(ci) === ruleForm.conditions.length - 1" text size="small" type="danger" @click="ruleForm.conditions.splice(toNumericIndex(ci), 1)">✕</el-button>
                </div>
              </el-form-item>
              <el-button v-if="ruleForm.conditionType==='and'||ruleForm.conditionType==='or'" text size="small" @click="ruleForm.conditions.push({questionIdx:undefined,optionIdx:undefined})">+ 添加条件</el-button>
            </template>
            <el-form-item label="动作">
              <el-select v-model="ruleForm.action" style="width:100%">
                <el-option label="显示 (SHOW)" value="show" />
                <el-option label="隐藏 (HIDE)" value="hide" />
                <el-option label="必填 (REQUIRED)" value="required" />
                <el-option label="跳转 (BRANCH)" value="branch" />
                <el-option label="勾选 (CHECK)" value="check" />
                <el-option label="赋值 (ASSIGNMENT)" value="assignment" />
                <el-option label="校验 (VALIDATE)" value="validate" />
                <el-option label="文本替换 (REPLACE)" value="replace" />
                <el-option label="结束问卷 (END)" value="end" />
              </el-select>
            </el-form-item>
            <el-form-item v-if="['show','hide','required'].includes(ruleForm.action)" label="目标题目">
              <el-select v-model="ruleForm.targetQuestionIdx" placeholder="选择题目" style="width:100%">
                <el-option v-for="(q, qi) in questions" :key="qi" :label="`Q${toNumericIndex(qi) + 1}: ${q.title?.slice(0,30)}`" :value="toNumericIndex(qi)" />
              </el-select>
            </el-form-item>
            <template v-if="ruleForm.action==='check'">
              <el-form-item label="目标题目">
                <el-select v-model="ruleForm.targetQuestionIdx" placeholder="选择题目" style="width:100%" @change="ruleForm.targetOptionIdxs=[]">
                  <el-option v-for="(q, qi) in questions" :key="qi" :label="`Q${toNumericIndex(qi) + 1}: ${q.title?.slice(0,30)}`" :value="toNumericIndex(qi)" />
                </el-select>
              </el-form-item>
              <el-form-item v-if="ruleForm.targetQuestionIdx!==undefined" label="目标选项（可多选）">
                <el-select v-model="ruleForm.targetOptionIdxs" multiple placeholder="选择选项" style="width:100%">
                  <el-option v-for="(o, oi) in getOptionsByIndex(ruleForm.targetQuestionIdx)" :key="oi" :label="o.label" :value="toNumericIndex(oi)" />
                </el-select>
              </el-form-item>
            </template>
            <template v-if="ruleForm.action==='branch'">
              <el-form-item label="从题目">
                <el-select v-model="ruleForm.branchFromIdx" placeholder="选择题目" style="width:100%" :key="'branchFrom'+showAddRule">
                  <el-option v-for="(q, qi) in questions" :key="qi" :label="`Q${toNumericIndex(qi) + 1}: ${q.title?.slice(0,30)}`" :value="toNumericIndex(qi)" />
                </el-select>
              </el-form-item>
              <el-form-item label="跳到">
                <div style="display:flex;gap:8px;align-items:center">
                  <el-select v-model="ruleForm.branchToIdx" placeholder="选择目标题目" style="flex:1;min-width:280px" :disabled="ruleForm.branchToEnd" :key="'branchTo'+showAddRule">
                    <el-option v-for="(q, qi) in questions" :key="qi" :label="`Q${toNumericIndex(qi) + 1}: ${q.title?.slice(0,30)}`" :value="toNumericIndex(qi)" />
                  </el-select>
                  <el-checkbox v-model="ruleForm.branchToEnd" style="white-space:nowrap">结束问卷</el-checkbox>
                </div>
              </el-form-item>
            </template>
            <el-form-item v-if="ruleForm.action==='assignment'||ruleForm.action==='validate'||ruleForm.action==='replace'" label="目标题目">
              <el-select v-model="ruleForm.targetQuestionIdx" placeholder="选择题目" style="width:100%">
                <el-option v-for="(q, qi) in questions" :key="qi" :label="`Q${toNumericIndex(qi) + 1}: ${q.title?.slice(0,30)}`" :value="toNumericIndex(qi)" />
              </el-select>
            </el-form-item>
            <el-form-item v-if="ruleForm.action==='assignment'" label="赋值公式">
              <el-input v-model="ruleForm.formula" placeholder="如 SUM(Q1, Q2)" />
            </el-form-item>
            <el-form-item v-if="ruleForm.action==='validate'" label="校验公式（返回空字符串表示通过）">
              <el-input v-model="ruleForm.formula" placeholder='如 IF(Q2>100,"不能超过100","")' />
            </el-form-item>
            <el-form-item v-if="ruleForm.action==='replace'" label="替换文本/公式">
              <el-input v-model="ruleForm.formula" placeholder='如 CONCATENATE("你好，",Q1)' />
            </el-form-item>
          </el-form>
          <template #footer>
            <el-button @click="showAddRule=false">取消</el-button>
            <el-button type="primary" @click="confirmRule">确定</el-button>
          </template>
        </el-dialog>
      </div>

      <div v-show="activeView==='setting'" class="setting-wrapper">
        <div class="setting-header">
          <div>
            <div class="setting-page-title">考试配置</div>
            <div class="setting-page-desc">集中管理考试展示、交卷规则、访问限制、投放链接和协作人员</div>
          </div>
          <div class="setting-header-actions">
            <el-tag size="small" :type="form.id ? 'success' : 'info'">{{ form.id ? '配置可保存' : '保存后生成链接' }}</el-tag>
            <el-button type="primary" size="small" :loading="saving" @click="save">保存配置</el-button>
          </div>
        </div>
        <div class="setting-scroll">
          <div class="setting-group">
            <div class="group-header">
              <div>
                <div class="group-title">展示与答题体验</div>
                <div class="group-desc">控制考生端页面结构、校验时机和答题辅助能力</div>
              </div>
            </div>
            <el-form label-position="top">
              <div class="settings-grid">
                <el-form-item label="显示题号"><el-switch v-model="form.questionNumber" /></el-form-item>
                <el-form-item label="显示进度条"><el-switch v-model="form.progressBar" /></el-form-item>
                <el-form-item label="自动暂存"><el-switch v-model="form.autoSave" /></el-form-item>
                <el-form-item label="显示分数"><el-switch v-model="form.showScore" :active-value="1" :inactive-value="0" /></el-form-item>
                <el-form-item label="一页一题"><el-switch v-model="form.onePageOneQuestion" /></el-form-item>
                <el-form-item label="显示答题卡"><el-switch v-model="form.answerSheetVisible" /></el-form-item>
                <el-form-item label="显示答案"><el-switch v-model="form.answerVisible" /></el-form-item>
                <el-form-item label="默认语言">
                  <el-select v-model="form.language" style="width:100%">
                    <el-option value="zh" label="中文" />
                    <el-option value="en" label="英文" />
                  </el-select>
                </el-form-item>
                <el-form-item label="校验触发">
                  <el-select v-model="form.triggerType">
                    <el-option value="onInput" label="输入时" />
                    <el-option value="onBlur" label="失焦时" />
                    <el-option value="onSubmit" label="提交时" />
                  </el-select>
                </el-form-item>
              </div>
            </el-form>
          </div>
          <div class="setting-group">
            <div class="group-header">
              <div>
                <div class="group-title">考试规则</div>
                <div class="group-desc">设置交卷时间、练习模式、题目顺序和考试次数</div>
              </div>
            </div>
            <el-form label-position="top">
              <div class="settings-grid">
                <el-form-item label="最短交卷时间">
                  <div class="setting-control-stack">
                    <el-input-number v-model="form.minSubmitMinutes" :min="0" style="width:100%" />
                    <span class="setting-help">0 表示不限制最短作答时间</span>
                  </div>
                </el-form-item>
                <el-form-item label="最长交卷时间">
                  <div class="setting-control-stack">
                    <el-input-number v-model="form.maxSubmitMinutes" :min="0" style="width:100%" />
                    <span class="setting-help">单位：分钟，0 表示不限时</span>
                  </div>
                </el-form-item>
                <el-form-item label="每人考试次数">
                  <div class="setting-control-stack">
                    <el-input-number v-model="form.maxAttempts" :min="0" style="width:100%" />
                    <span class="setting-help">0 表示不限次数</span>
                  </div>
                </el-form-item>
                <el-form-item label="练习模式"><el-switch v-model="form.exerciseMode" /></el-form-item>
                <el-form-item label="随机顺序"><el-switch v-model="form.randomOrder" /></el-form-item>
                <el-form-item label="开始时间"><el-date-picker v-model="form.startDate" type="datetime" placeholder="不限" value-format="x" style="width:100%" /></el-form-item>
                <el-form-item label="结束时间"><el-date-picker v-model="form.endDate" type="datetime" placeholder="不限" value-format="x" style="width:100%" /></el-form-item>
              </div>
            </el-form>
          </div>
          <div class="setting-group">
            <div class="group-header">
              <div>
                <div class="group-title">访问与回收</div>
                <div class="group-desc">控制谁可以参加考试，以及答卷、设备和 IP 限制</div>
              </div>
            </div>
            <el-form label-position="top">
              <div class="settings-grid">
                <el-form-item class="settings-full" label="可见性">
                  <el-radio-group v-model="form.visibility" class="setting-radio-group">
                    <el-radio :value="0" border>公开链接</el-radio>
                    <el-radio :value="1" border>登录可见</el-radio>
                    <el-radio :value="2" border>部门限定</el-radio>
                  </el-radio-group>
                </el-form-item>
                <el-form-item v-if="form.visibility===2" class="settings-full" label="限定部门">
                  <div class="setting-tree-box">
                    <el-tree
                      ref="deptTreeRef"
                      :data="deptTreeOptions"
                      :props="{ label: 'name', children: 'children' }"
                      node-key="id"
                      show-checkbox
                      default-expand-all
                      @check-change="onDeptCheckChange"
                    />
                  </div>
                </el-form-item>
                <el-form-item label="最大答卷数">
                  <div class="setting-control-stack">
                    <el-input-number v-model="form.maxResponse" :min="0" style="width:100%" />
                    <span class="setting-help">0 表示不限总答卷数</span>
                  </div>
                </el-form-item>
                <el-form-item label="填写密码"><el-input v-model="form.password" placeholder="留空不设密码" /></el-form-item>
                <el-form-item label="每台设备答题次数">
                  <div class="setting-control-stack">
                    <el-input-number v-model="form.deviceLimit" :min="0" style="width:100%" />
                    <span class="setting-help">0 表示不限</span>
                  </div>
                </el-form-item>
                <el-form-item label="每个 IP 答题次数">
                  <div class="setting-control-stack">
                    <el-input-number v-model="form.ipLimit" :min="0" style="width:100%" />
                    <span class="setting-help">0 表示不限</span>
                  </div>
                </el-form-item>
              </div>
            </el-form>
          </div>
          <div class="setting-group">
            <div class="group-header">
              <div>
                <div class="group-title">投放与结果</div>
                <div class="group-desc">管理考后展示、填写模板、公开链接和二维码</div>
              </div>
            </div>
            <el-form label-position="top">
              <div class="settings-grid">
                <el-form-item label="显示排名"><el-switch v-model="form.examRankingEnabled" /></el-form-item>
                <el-form-item label="显示成绩单"><el-switch v-model="form.transcriptVisible" /></el-form-item>
                <el-form-item label="可查看正确答案和解析"><el-switch v-model="form.showAnalysis" /></el-form-item>
                <el-form-item label="考试状态"><el-switch v-model="form.statusBool" :active-value="1" :inactive-value="0" active-text="已开启" inactive-text="已停用" /></el-form-item>
                <el-form-item class="settings-full" label="填写模版">
                  <el-select v-model="form.fillTemplate" style="width:100%">
                    <el-option v-for="(label, val) in fillTemplates" :key="val" :value="val" :label="label" />
                  </el-select>
                </el-form-item>
                <el-form-item class="settings-full" label="问卷链接">
                  <div class="share-link-row">
                    <el-input v-if="form.id" :model-value="publicUrl" readonly size="small" style="flex:1">
                      <template #append><el-button @click="copyLink" size="small">复制</el-button></template>
                    </el-input>
                    <span v-else class="setting-empty-tip">请先保存考试</span>
                  </div>
                </el-form-item>
                <el-form-item class="settings-full" label="二维码">
                  <div v-if="form.id" class="qr-preview">
                    <img :src="`https://api.qrserver.com/v1/create-qr-code/?size=120x120&data=${encodeURIComponent(publicUrl)}`" alt="二维码" />
                    <div class="qr-preview-meta">
                      <strong>扫码进入考试</strong>
                      <span>二维码内容随考试链接自动生成</span>
                    </div>
                  </div>
                  <span v-else class="setting-empty-tip">请先保存考试</span>
                </el-form-item>
                <el-form-item class="settings-full" label="提交后处理">
                  <el-radio-group v-model="completionAction" class="completion-radio-group">
                    <el-radio-button value="default">默认完成页</el-radio-button>
                    <el-radio-button value="content">自定义页面</el-radio-button>
                    <el-radio-button value="redirect">跳转链接</el-radio-button>
                  </el-radio-group>
                </el-form-item>
                <el-form-item v-if="completionAction === 'content'" label="提交完成页内容">
                  <el-input v-model="form.endContent" type="textarea" :rows="3" placeholder="提交后显示的 HTML 内容" />
                </el-form-item>
                <el-form-item v-if="completionAction === 'redirect'" label="提交后跳转链接">
                  <el-input v-model="form.redirectUrl" placeholder="https://example.com 或站内路径" />
                  <div v-if="form.redirectUrl && !isValidRedirectUrl(form.redirectUrl)" class="setting-help danger">请输入 http(s) 链接或以 / 开头的站内路径</div>
                </el-form-item>
              </div>
            </el-form>
          </div>
          <div class="setting-group">
            <div class="group-header">
              <div>
                <div class="group-title">协作管理员</div>
                <div class="group-desc">查看创建者，并指定可以共同维护考试的管理员</div>
              </div>
            </div>
            <el-form label-position="top">
              <el-form-item label="考试创建者">
                <el-input :model-value="creatorName" disabled placeholder="加载中..." />
              </el-form-item>
              <el-form-item class="settings-full" label="协作管理员">
                <div class="setting-tree-box large">
                  <el-tree
                    ref="adminTreeRef"
                    :data="adminTreeData"
                    :props="{ label: 'label', children: 'children' }"
                    node-key="id"
                    show-checkbox
                    :default-checked-keys="collaboratorCheckedKeys"
                    @check="onAdminTreeCheck"
                  />
                </div>
              </el-form-item>
            </el-form>
          </div>
        </div>
      </div>

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

  <!-- 通用公式对话框 -->
  <el-dialog v-model="formulaVisible" :title="formulaTitle" :key="formulaKey" width="720px" :close-on-click-modal="false">
    <div style="display:flex;gap:12px;min-height:360px">
      <div style="flex:1">
        <div style="font-size:12px;color:#888;margin-bottom:4px">{{ formulaType==='calculate'?'计算结果将自动填充到当前题目':'公式表达式' }}</div>
        <el-input v-model="formulaText" type="textarea" :rows="12" :placeholder="formulaPlaceholder" />
        <div style="font-size:12px;color:#606266;background:#f5f7fa;border-radius:4px;padding:8px 10px;margin-top:8px">
          <div style="font-weight:600;margin-bottom:4px">使用说明</div>
          <ul style="margin:0;padding-left:16px;line-height:1.8">
            <li>运算符、括号、逗号请使用<strong>英文输入法</strong></li>
            <li>判断相等用 <code>==</code>，赋值用 <code>=</code></li>
            <li>手动输入的文本用引号包裹，例如 <code>"优秀"</code></li>
            <li>公式结果类型需与目标题型匹配（数字题不能接收文本结果）</li>
            <li>点击右侧<strong style="color:var(--designer-accent,#0f766e)">高亮标签</strong>可插入题目标签到光标位置</li>
          </ul>
        </div>
        <div style="font-size:11px;color:#999;margin-top:8px">
          <div style="font-weight:600;color:#606266;margin-bottom:4px">函数参考 &amp; 案例</div>
          <div><code>IF(条件, TRUE, FALSE)</code> — 条件判断</div>
          <div><code>AND(条件1, 条件2)</code> — 与运算</div>
          <div><code>OR(条件1, 条件2)</code> — 或运算</div>
          <div><code>NOT(条件)</code> — 非运算</div>
          <div><code>COUNT(Q1)</code> — 统计多选已选数量</div>
          <div><code>SCORE(Q1)</code> — 读取选项分值</div>
          <div><code>TEXT(Q1)</code> — 读取选项文本</div>
          <div><code>CURRENT_DATE("YYYY/MM/DD")</code> — 当前日期</div>
          <div><code>INCLUDE("学校", Q1)</code> — 判断文本是否包含关键词</div>
          <div v-if="formulaType==='calculate'" style="margin-top:6px;padding-top:6px;border-top:1px solid #eee">
            <div style="font-weight:600;color:#606266;margin-bottom:4px">计算公式案例</div>
            <div><code>SUM(Q1, Q2)</code> — 两题求和（计算总分）</div>
            <div><code>ROUND(Q1 / POWER(Q2 / 100, 2), 2)</code> — BMI 指数</div>
            <div><code>IFS(Q1&lt;60,"不合格",Q1&lt;80,"合格",TRUE(),"优秀")</code> — 多条件等级</div>
            <div><code>CONCATENATE("你的总分是", Q1, "分")</code> — 拼接结果文案</div>
            <div><code>Q1 + Q2 + Q3</code> — 连加运算</div>
            <div><code>(Q1 + Q2) / 2</code> — 求平均值</div>
          </div>
          <div v-if="formulaType==='endFormula'||formulaType==='jumpFormula'" style="margin-top:6px;padding-top:6px;border-top:1px solid #eee">
            <div style="font-weight:600;color:#606266;margin-bottom:4px">跳转/结束案例</div>
            <div><code>IF Q1A1 THEN BRANCH FROM Q1 TO Q5</code> — 选择选项A跳到Q5</div>
            <div><code>IF Q1A2 THEN BRANCH FROM Q1 TO END</code> — 选择选项B结束问卷</div>
            <div><code>IF AND(Q1A1, Q2&gt;18) THEN BRANCH FROM Q1 TO Q6</code> — 多条件跳转</div>
          </div>
          <div v-if="formulaType==='textReplace'" style="margin-top:6px;padding-top:6px;border-top:1px solid #eee">
            <div style="font-weight:600;color:#606266;margin-bottom:4px">文本替换案例</div>
            <div><code>REPLACE Q2 WITH CONCATENATE("你好，", Q1)</code> — 替换Q2标题</div>
            <div><code>REPLACE Q3 WITH CONCATENATE("你的得分：", SUM(Q1, Q2))</code> — 替换为计算结果</div>
          </div>
        </div>
      </div>
      <div style="width:200px;flex-shrink:0">
        <div style="font-size:12px;color:#888;margin-bottom:4px">题目标签</div>
        <div style="border:1px solid #eee;border-radius:4px;padding:8px;max-height:320px;overflow-y:auto">
          <div v-for="q in formulaQuestions" :key="q.id" style="margin-bottom:6px">
            <div style="font-size:11px;color:#999;margin-bottom:2px">{{ q.title?.slice(0,20) }}</div>
            <div style="display:flex;flex-wrap:wrap;gap:4px">
              <el-tag size="small" type="danger" style="cursor:pointer" @click="insertFormulaTag(`Q${questions.indexOf(q)+1}`)">Q{{ questions.indexOf(q)+1 }}</el-tag>
              <template v-if="q.props?.options?.length">
                <el-tag v-for="(o, oi) in q.props.options.slice(0,4)" :key="oi" size="small" type="warning" style="cursor:pointer" @click="insertFormulaTag(`Q${questions.indexOf(q)+1}A${toNumericIndex(oi) + 1}`)">Q{{ questions.indexOf(q)+1 }}A{{ toNumericIndex(oi) + 1 }}</el-tag>
              </template>
            </div>
          </div>
          <el-empty v-if="!questions.length" description="暂无题目" :image-size="30" />
        </div>
      </div>
    </div>
    <template #footer>
      <el-button @click="formulaVisible=false">取消</el-button>
      <el-button type="primary" @click="confirmFormula">确定</el-button>
    </template>
  </el-dialog>

  <!-- 文本导入弹窗 -->
  <el-dialog v-model="textImport.visible" title="文本导入" width="600px" :close-on-click-modal="false" @closed="textImport.text=''">
    <div style="margin-bottom:8px;display:flex;gap:8px;align-items:center">
      <el-radio-group v-model="textImport.mode" size="small">
        <el-radio-button value="paste">粘贴文本</el-radio-button>
        <el-radio-button value="file">上传文件</el-radio-button>
      </el-radio-group>
      <el-tag type="info" style="font-size:11px">每行一个题目，选项用 A. B. C. 开头</el-tag>
    </div>
    <div v-if="textImport.mode==='paste'">
      <el-input v-model="textImport.text" type="textarea" :rows="12" placeholder="粘贴题目文本&#10;&#10;格式示例：&#10;1. 您最喜欢的颜色？&#10;A. 红色&#10;B. 蓝色&#10;C. 绿色&#10;&#10;2. 您的姓名&#10;&#10;3. [多选题] 您的爱好&#10;A. 阅读&#10;B. 运动&#10;C. 音乐" />
    </div>
    <div v-else>
      <el-upload drag :auto-upload="false" :show-file-list="false" :on-change="onTextFileChange" accept=".txt,.docx,.md">
        <el-icon style="font-size:40px;margin-bottom:8px"><svg viewBox="0 0 24 24" width="40" height="40" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="12" y1="18" x2="12" y2="12"/><line x1="9" y1="15" x2="15" y2="15"/></svg></el-icon>
        <div style="font-size:13px;color:#666">拖拽或点击选择 .txt / .docx / .md 文件</div>
      </el-upload>
      <div v-if="textImport.text" style="margin-top:8px;font-size:12px;color:#999">已解析 {{ textImport.text.split('\n').filter((l: string)=>l.trim()).length }} 行</div>
    </div>
    <div v-if="textImport.preview.length" style="margin-top:12px;border:1px solid #e8e8e8;border-radius:6px;padding:8px;max-height:200px;overflow-y:auto">
      <div style="font-size:12px;color:#666;margin-bottom:4px">预览（{{ textImport.preview.length }} 题）：</div>
      <div v-if="textImport.surveyTitle" style="font-size:13px;font-weight:bold;padding:2px 0 6px;border-bottom:1px dashed #eee;margin-bottom:4px">{{ textImport.surveyTitle }}</div>
      <div v-for="(q, i) in textImport.preview" :key="i" style="font-size:12px;padding:2px 0;display:flex;gap:6px">
        <el-tag size="small" style="flex-shrink:0">{{ typeName(q.type) }}</el-tag>
        <span style="color:#333;overflow:hidden;text-overflow:ellipsis;white-space:nowrap">{{ stripHtmlTag(q.title) }}</span>
      </div>
    </div>
    <template #footer>
      <el-button @click="textImport.visible=false">取消</el-button>
      <el-button type="primary" :disabled="!textImport.parsed.length" :loading="textImport.importing" @click="confirmTextImport">导入 {{ textImport.parsed.length }} 题</el-button>
    </template>
  </el-dialog>

  <!-- 上传题库弹窗 -->
  <el-dialog v-model="bankDialog.visible" title="上传到题库" width="420px" :close-on-click-modal="false">
    <el-form label-position="top" size="small">
      <el-form-item label="题库分类">
        <el-select v-model="bankDialog.category" placeholder="选择或输入分类" filterable allow-create clearable style="width:100%">
          <el-option v-for="cat in bankCategories" :key="cat" :label="cat" :value="cat" />
        </el-select>
      </el-form-item>
      <el-form-item label="标签">
        <el-input v-model="bankDialog.tags" placeholder="用逗号分隔" />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="bankDialog.visible=false">取消</el-button>
      <el-button type="primary" :loading="bankDialog.saving" @click="confirmUploadBank">上传</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, reactive, watch, onMounted, onBeforeUnmount, nextTick } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useRoute, useRouter } from 'vue-router'
import { adminApi } from '../../api'
import DraggableList from './formkit/DraggableList.vue'
import QuestionIcon from './formkit/QuestionIcon.vue'

const route = useRoute()
const router = useRouter()

function toNumericIndex(index: string | number): number {
  return Number(index)
}

const form = reactive<any>({
  id: 0, title: '', description: '', category: '', tags: '',
  visibility: 0, allowMultiBool: 0, anonymousBool: 0, showResultBool: 0,
  startDate: null, endDate: null, maxResponse: 0, statusBool: 1, deptIds: '',
  questionNumber: true, progressBar: false, autoSave: false,
  onePageOneQuestion: false, answerSheetVisible: true, answerVisible: true, copyEnabled: true, language: 'zh',
  password: '', triggerType: 'onBlur', loginRequired: false,
  transcriptVisible: true, rankVisible: false, showAnalysis: true,
  redirectUrl: '', endContent: '',
  examRankingEnabled: false, exerciseMode: false, randomOrder: false,
  minSubmitMinutes: 0, maxSubmitMinutes: 0,
  duration: 60, maxAttempts: 1, showScore: 1,
  deviceLimit: 0, ipLimit: 0, userLimit: 0,
  createBy: 0, collaborators: '',
  backgroundImages: [] as string[], headerImages: [] as string[],
  fillTemplate: 'ef'
})

const activeView = ref('edit')
const middleTab = ref('item')
const sideSubTab = ref('types')
const activeCategory = ref('')
const saving = ref(false)
const panelMode = ref<'edit' | 'json' | 'preview'>('edit')
const exportedJson = ref('')
const completionAction = ref<'default' | 'content' | 'redirect'>('default')

function syncCompletionActionFromForm() {
  if (String(form.redirectUrl || '').trim()) {
    completionAction.value = 'redirect'
  } else if (String(form.endContent || '').trim()) {
    completionAction.value = 'content'
  } else {
    completionAction.value = 'default'
  }
}

watch(completionAction, (value, oldValue) => {
  if (value === 'default') {
    form.redirectUrl = ''
    form.endContent = ''
  } else if (value === 'content' && oldValue !== 'content') {
    form.redirectUrl = ''
  } else if (value === 'redirect' && oldValue !== 'redirect') {
    form.endContent = ''
  }
})

function isValidRedirectUrl(raw: string) {
  if (!raw) return true
  if (raw.startsWith('/')) return true
  try {
    const url = new URL(raw)
    return url.protocol === 'http:' || url.protocol === 'https:'
  } catch {
    return false
  }
}

function normalizeCompletionSettings() {
  if (completionAction.value === 'default') {
    form.redirectUrl = ''
    form.endContent = ''
  } else if (completionAction.value === 'content') {
    form.redirectUrl = ''
  } else {
    form.redirectUrl = String(form.redirectUrl || '').trim()
    form.endContent = ''
  }
}

const textImport = reactive({
  visible: false,
  mode: 'paste' as 'paste' | 'file',
  text: '',
  parsed: [] as any[],
  preview: [] as any[],
  importing: false,
  surveyTitle: ''
})

function stripHtmlTag(html: string) {
  return html.replace(/<[^>]*>/g, '')
}

function parseTextToQuestions(text: string) {
  const lines = text.split('\n')
  const questions: any[] = []
  let current: any = null
  let surveyTitle = ''
  let lineIdx = 0
  const trimmed = lines.map(l => l.trim()).filter(l => l.length > 0)
  if (trimmed.length >= 2 && /^[=\-~]{3,}$/.test(trimmed[1])) {
    surveyTitle = trimmed[0]
    const nonEmptyIdx = lines.reduce<number[]>((acc, l, i) => { if (l.trim()) acc.push(i); return acc }, [])
    if (nonEmptyIdx.length >= 2 && nonEmptyIdx[0] + 1 === nonEmptyIdx[1] && /^[=\-~]{3,}$/.test(lines[nonEmptyIdx[1]].trim())) {
      lineIdx = 2
    }
  }
  const typeMap: Record<string, string> = {
    '单选': 'radio', '单选置': 'radio', '单选题': 'radio',
    '多选': 'checkbox', '多选题': 'checkbox', '多选择': 'checkbox',
    '下拉': 'select', '下拉选择': 'select', '下拉框': 'select',
    '选择器': 'picker', 'picker': 'picker',
    '级联选择': 'cascade', '级联': 'cascade', 'cascade': 'cascade',
    '判断': 'judge', '判断题': 'judge',
    '上传文件': 'file', '文件上传': 'file', '图片上传': 'file', '文件': 'file',
    '单行文本': 'input', '文本': 'input',
    '多行': 'textarea', '多行文本': 'textarea', 'textarea': 'textarea',
    '数字': 'number',
    '多项填空': 'multiInput', 'multiInput': 'multiInput',
    '横向填空': 'hInput', 'hInput': 'hInput',
    '签名': 'signature', 'signature': 'signature',
    '扫码': 'scanCode', 'scanCode': 'scanCode',
    '评分': 'rating', 'rating': 'rating',
    'NPS评分': 'nps', 'nps': 'nps', 'NPS': 'nps',
    '手机': 'phone', '手机号': 'phone', 'phone': 'phone',
    '邮箱': 'email', 'email': 'email',
    '身份证': 'idCard', '身份证号': 'idCard', 'idCard': 'idCard',
    '密码': 'password', 'password': 'password',
    '姓名': 'name', 'name': 'name',
    '学号': 'studentId', 'studentId': 'studentId',
    '工号': 'employeeId', 'employeeId': 'employeeId',
    '班级': 'class', 'class': 'class',
    '开关': 'switch', 'switch': 'switch',
    '地理位置': 'location', '位置': 'location', '打卡': 'location', 'location': 'location',
    '日期': 'date', 'date': 'date',
    '时间': 'time', 'time': 'time',
    '日期范围': 'dateRange', 'dateRange': 'dateRange',
    '矩阵单选': 'matrixRadio', 'matrixRadio': 'matrixRadio',
    '矩阵多选': 'matrixCheckbox', 'matrixCheckbox': 'matrixCheckbox',
    '矩阵填空': 'matrixFillBlank', 'matrixFillBlank': 'matrixFillBlank',
    '表格自增': 'matrixAuto', '矩阵自增': 'matrixAuto', 'matrixAuto': 'matrixAuto',
    '成员': 'user', '人员': 'user', 'user': 'user',
    '部门': 'dept', 'dept': 'dept',
    '富文本': 'richText', 'richText': 'richText',
    '自动填充': 'autopop', 'autopop': 'autopop',
  }
  for (const raw of lines) {
    if (lineIdx > 0) { lineIdx--; continue }
    const line = raw.trim()
    if (!line) {
      if (current && current.title) { questions.push(current); current = null }
      continue
    }
    const optMatch = line.match(/^([A-Za-z])[.、）)]\s*(.*)/)
    if (optMatch) {
      if (!current) current = { type: 'radio', title: '', options: [], rows: [], columns: [], fields: [] }
      if (!current.options) current.options = []
      const label = stripHtmlTag(optMatch[2])
      if (label) {
        current.options.push({ label, value: optMatch[1] })
        if (current.type === 'input') current.type = 'radio'
      }
      continue
    }
    const rowMatch = line.match(/^行[:：]\s*(.+)/)
    const colMatch = line.match(/^列[:：]\s*(.+)/)
    const fieldMatch = line.match(/^字段[:：]\s*(.+)/)
    if (rowMatch || colMatch || fieldMatch) {
      if (!current) current = { type: 'input', title: '', options: [], rows: [], columns: [], fields: [] }
      if (!current.rows) current.rows = []
      if (!current.columns) current.columns = []
      if (!current.fields) current.fields = []
      if (rowMatch) current.rows = rowMatch[1].split('/').map((s: string) => ({ title: s.trim() }))
      if (colMatch) current.columns = colMatch[1].split('/').map((s: string) => ({ title: s.trim() }))
      if (fieldMatch) current.fields = fieldMatch[1].split('/').map((s: string) => ({ label: s.trim(), placeholder: '' }))
      continue
    }
    if (current && current.title) { questions.push(current) }
    const typeMatch = line.match(/^\d*[.、）)]?\s*\[(.+?)\]\s*(.*)/)
    if (typeMatch) {
      const typeName2 = typeMatch[1].trim()
      const rawTitle = typeMatch[2]
      const mappedType = typeMap[typeName2] || 'input'
      let qTitle = stripHtmlTag(rawTitle)
      if (!qTitle) qTitle = stripHtmlTag(line.replace(/^\d*[.、）)]?\s*/, '').replace(/\[.*?\]/, '').trim())
      current = { type: mappedType, title: qTitle, options: [], rows: [], columns: [], fields: [] }
      if (mappedType === 'judge') current.options = [{ label: '对', value: 'true' }, { label: '错', value: 'false' }]
    } else {
      const title = stripHtmlTag(line.replace(/^\d+[.、）)]?\s*/, ''))
      current = { type: 'input', title, options: [], rows: [], columns: [], fields: [] }
    }
  }
  if (current && current.title) questions.push(current)
  questions.forEach((q: any) => {
    const needsOptions = ['radio', 'checkbox', 'select', 'picker', 'cascade', 'user', 'dept']
    if (q.options && q.options.length > 0) {
      if (q.type === 'input') q.type = 'radio'
    } else if (needsOptions.includes(q.type)) {
      q.type = 'input'
      q.options = []
    }
  })
  return { title: surveyTitle, questions }
}

watch(() => textImport.text, (val) => {
  if (textImport.mode === 'paste' && val) {
    const result = parseTextToQuestions(val)
    textImport.surveyTitle = result.title
    textImport.parsed = result.questions
    textImport.preview = textImport.parsed.slice(0, 20)
  } else if (!val) {
    textImport.parsed = []
    textImport.preview = []
    textImport.surveyTitle = ''
  }
})
watch(() => textImport.mode, () => {
  textImport.parsed = []
  textImport.preview = []
})

async function onTextFileChange(file: any) {
  const f = file.raw || file
  if (!f) return
  if (f.name.endsWith('.docx')) {
    try {
      const lib = 'mamm' + 'oth'
      const mammothMod = await import(/* @vite-ignore */ lib)
      const m = (mammothMod.default || mammothMod) as any
      const buf = await f.arrayBuffer()
      const result = await m.extractRawText({ arrayBuffer: buf })
      textImport.text = result.value
    } catch {
      ElMessage.error('Word 文件解析失败，请先安装 mammoth: npm install mammoth')
      return
    }
  } else {
    textImport.text = await f.text()
  }
  const result = parseTextToQuestions(textImport.text)
  textImport.surveyTitle = result.title
  textImport.parsed = result.questions
  textImport.preview = textImport.parsed.slice(0, 20)
}

function confirmTextImport() {
  if (!textImport.parsed.length) return
  textImport.importing = true
  try {
    if (textImport.surveyTitle) form.title = textImport.surveyTitle
    textImport.parsed.forEach((q: any) => {
      const newQ: any = {
        id: genId(),
        type: q.type,
        title: q.title,
        required: false,
        readOnly: false,
        dataType: '',
        placeholder: '',
        mediaType: '',
        mediaUrl: '',
        mediaWidth: '',
        mediaAlign: 'center',
        showDescription: true,
        props: {}
      }
      if (['radio', 'checkbox', 'select', 'picker', 'cascade'].includes(q.type) && q.options?.length) {
        newQ.props.options = q.options
      }
      if (q.type === 'judge') {
        newQ.props.options = [{ label: '对', value: 'true' }, { label: '错', value: 'false' }]
      }
      if (['matrixRadio', 'matrixCheckbox', 'matrixFillBlank'].includes(q.type)) {
        if (q.rows?.length) newQ.props.rows = q.rows
        if (q.columns?.length) newQ.props.columns = q.columns
      }
      if (q.type === 'matrixAuto') {
        if (q.columns?.length) newQ.props.columns = q.columns
      }
      if (q.type === 'multiInput' || q.type === 'hInput') {
        if (q.fields?.length) newQ.props.fields = q.fields
      }
      questions.value.push(newQ)
      selected.value = newQ
    })
    ElMessage.success(`成功导入 ${textImport.parsed.length} 题`)
    textImport.visible = false
  } finally {
    textImport.importing = false
  }
}

function openTextImport() {
  textImport.visible = true
  textImport.text = ''
  textImport.parsed = []
  textImport.preview = []
  textImport.mode = 'paste'
  textImport.surveyTitle = ''
}

function exportExam() {
  const text = surveyToText()
  const blob = new Blob([text], { type: 'text/plain;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `${form.title || '试卷'}.txt`
  a.click()
  URL.revokeObjectURL(url)
  ElMessage.success('已导出')
}

function surveyToText(): string {
  const lines: string[] = []
  const title = form.title || '未命名试卷'
  lines.push(title)
  lines.push('='.repeat(title.length))
  lines.push('')
  let idx = 0
  questions.value.forEach((q: any) => {
    if (['description', 'divider', 'pagination', 'questionSet'].includes(q.type)) {
      if (q.type === 'description') lines.push('--- ' + stripHtmlTag(q.title) + ' ---')
      return
    }
    idx++
    const qTitle = stripHtmlTag(q.title || '未命名')
    const typeLabel = typeName(q.type)
    const required = q.required ? '（必填）' : ''
    lines.push(`${idx}. [${typeLabel}] ${qTitle}${required}`)
    if (q.props?.options?.length) {
      q.props.options.forEach((o: any, oi: number) => {
        const prefix = String.fromCharCode(65 + oi)
        lines.push(`   ${prefix}. ${o.label}`)
      })
    }
    if (q.props?.fields?.length) {
      lines.push(`   字段: ${q.props.fields.map((f: any) => f.label).join(' / ')}`)
    }
    if (q.props?.rows?.length) {
      lines.push(`   行: ${q.props.rows.map((r: any) => r.title).join(' / ')}`)
    }
    if (q.props?.columns?.length) {
      lines.push(`   列: ${q.props.columns.map((c: any) => c.title || c.label).join(' / ')}`)
    }
    lines.push('')
  })
  return lines.join('\n')
}

const bankKeyword = ref('')
const bankQuestions = ref<any[]>([])
const bankLoading = ref(false)
const bankLoadingVisible = ref(false)
let bankTimer: any = null
let bankLoadingTimer: ReturnType<typeof setTimeout> | null = null
let bankLoadSeq = 0
const bankDialog = reactive({ visible: false, qid: '', category: '', tags: '', saving: false })
const bankCategories = ref<string[]>([])

async function onUploadBank(id: string) {
  const q = questions.value.find(x => x.id === id)
  if (!q) return
  bankDialog.qid = id
  bankDialog.category = ''
  bankDialog.tags = ''
  bankDialog.visible = true
  try {
    const res: any = await adminApi.examQuestionBankCategories()
    bankCategories.value = res.data || []
  } catch {}
}

async function confirmUploadBank() {
  const q = questions.value.find((x: any) => x.id === bankDialog.qid)
  if (!q) return
  bankDialog.saving = true
  try {
    const { id, ...rest } = q
    const stripHtml = (html: string) => html.replace(/<[^>]*>/g, '')
    const titlePlain = stripHtml(String(rest.title || '')).slice(0, 50)
    const payload: any = {
      title: titlePlain,
      type: rest.type,
      schema: JSON.stringify(rest),
      category: bankDialog.category || '',
      tags: bankDialog.tags || ''
    }
    await adminApi.examQuestionBankInsert(payload)
    ElMessage.success('已上传到题库')
    bankDialog.visible = false
    loadBank()
  } catch (e: any) {
    ElMessage.error(e?.msg || '上传失败')
  } finally {
    bankDialog.saving = false
  }
}

const bankExpanded = ref<Record<string, boolean>>({})
function toggleBankExpand(key: string) {
  bankExpanded.value[key] = !bankExpanded.value[key]
}

const bankTree = computed(() => {
  const map: Record<string, Record<string, any[]>> = {}
  bankQuestions.value.forEach((q: any) => {
    const cat = q.category || ''
    const type = q.type || ''
    if (!map[cat]) map[cat] = {}
    if (!map[cat][type]) map[cat][type] = []
    map[cat][type].push(q)
  })
  return Object.entries(map).map(([cat, types]) => {
    const children = Object.entries(types).map(([type, items]) => ({
      key: type || '__unknown_type__',
      expandKey: type,
      label: type,
      _expanded: bankExpanded.value[`type:${cat}|${type}`] ?? false,
      children: items
    }))
    return {
      key: cat || '__uncategorized__',
      expandKey: cat,
      label: cat || '未分类',
      count: children.reduce((sum, grp) => sum + grp.children.length, 0),
      _expanded: bankExpanded.value[`cat:${cat}`] ?? false,
      children
    }
  })
})

const appearanceTab = ref('bg')
const allBgResources = ref<any[]>([])
const allHeaderResources = ref<any[]>([])

const token = localStorage.getItem('admin_token') || ''
const uploadHeaders = { Authorization: token }

const creatorName = ref('')
const adminTreeRef = ref<any>(null)
const adminTreeData = ref<any[]>([])
const deptTreeOptions = ref<any[]>([])
const deptTreeRef = ref<any>(null)
const collaboratorCheckedKeys = ref<string[]>([])
const adminMap = ref<Record<number, string>>({})
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

function onAdminTreeCheck() {
  form.collaborators = adminTreeRef.value
    ?.getCheckedKeys(true)
    ?.filter((k: string) => k.startsWith('admin-'))
    ?.map((k: string) => k.replace('admin-', ''))
    ?.join(',') || ''
}

async function loadAdminTree() {
  try {
    const [deptRes, mgrRes] = await Promise.all([
      adminApi.deptTree(),
      adminApi.mgrList({ page: 1, pageSize: 9999 })
    ])
    const depts: any[] = deptRes.data || []
    const mgrs: any[] = mgrRes.data?.list || []
    const map: Record<number, string> = {}
    mgrs.forEach((m: any) => { map[m.id] = m.name })
    adminMap.value = map
    const creator = mgrs.find((m: any) => m.id === form.createBy)
    creatorName.value = creator ? creator.name : (form.createBy ? '未知用户' : '未指定')
    adminTreeData.value = buildAdminTree(depts, mgrs)
    collaboratorCheckedKeys.value = form.collaborators
      ? form.collaborators.split(',').map((id: string) => `admin-${id}`)
      : []
  } catch {
    creatorName.value = '加载失败'
    adminTreeData.value = []
  }
}

async function loadDeptTree() {
  try {
    const res: any = await adminApi.deptTree()
    deptTreeOptions.value = res.data || []
    await nextTick()
    if (form.visibility === 2 && form.deptIds) {
      const keys = form.deptIds.split(',').map((s: string) => parseInt(s.trim())).filter((n: number) => !isNaN(n))
      deptTreeRef.value?.setCheckedKeys(keys)
    }
  } catch { deptTreeOptions.value = [] }
}

function onDeptCheckChange() {
  if (!deptTreeRef.value) return
  form.deptIds = deptTreeRef.value.getCheckedKeys(false).join(',')
}

function buildAdminTree(depts: any[], mgrs: any[]): any[] {
  return depts.map((d: any) => {
    const deptAdmins = mgrs.filter((m: any) => (m.deptIds || []).includes(d.id))
    const adminNodes = deptAdmins.map((m: any) => ({
      id: `admin-${m.id}`, label: m.name, type: 'admin'
    }))
    const children = d.children?.length ? buildAdminTree(d.children, mgrs) : []
    if (adminNodes.length) children.push(...adminNodes)
    return {
      id: `dept-${d.id}`, label: d.name, type: 'dept', children
    }
  }).filter((d: any) => d.children?.length)
}

function handleBgSuccess(res: any) {
  if (!res.data?.url) { ElMessage.error('上传失败：响应数据异常'); return }
  ElMessage.success('已上传到资源库，点击图片应用')
  loadResources()
}
function handleHeaderSuccess(res: any) {
  if (!res.data?.url) { ElMessage.error('上传失败：响应数据异常'); return }
  ElMessage.success('已上传到资源库，点击图片应用')
  loadResources()
}
function isActiveImg(field: string, item: any): boolean {
  const active = (form as any)[field]
  if (!active || !active.length) return false
  const cur = active[0]
  return cur && typeof cur === 'object' && cur.id ? cur.id === item.id : cur === (item.domain || '') + item.url
}
function applyImage(field: string, item: any) {
  const url = (item.domain || '') + item.url
  const active = (form as any)[field]
  if (active && active.length) {
    const cur = active[0]
    const same = cur && typeof cur === 'object' && cur.id ? cur.id === item.id : cur === url
    if (same) return
  }
  (form as any)[field] = [{ id: item.id, url }]
  save()
}
async function removeResource(field: string, item: any, idx: number) {
  if (item.id) {
    try { await adminApi.examResourceDelete({ id: item.id }) } catch {}
  }
  const active = (form as any)[field]
  if (active && active.length) {
    const cur = active[0]
    const url = (item.domain || '') + item.url
    const same = cur && typeof cur === 'object' && cur.id ? cur.id === item.id : cur === url
    if (same) { (form as any)[field] = []; await save() }
  }
  loadResources()
}
function checkSaved() {
  if (!form.id) { ElMessage.warning('请先保存考试'); return false }
  return true
}
async function loadResources() {
  if (!form.id) return
  try {
    const res: any = await adminApi.examResourceList({ examId: form.id })
    const list: any[] = res.data || []
    if (!Array.isArray(list)) return
    allBgResources.value = list.filter((r: any) => r.type === 'bg')
    allHeaderResources.value = list.filter((r: any) => r.type === 'header')
  } catch {}
}

// ===== 逻辑规则 =====
interface LogicCondition { questionIdx?: number; optionIdx?: number; operator?: string; compareValue?: string }
interface LogicRuleItem { id: string; conditionType: string; conditions: LogicCondition[]; action: string; scope: 'frontend' | 'backend'; targetQuestionIdx?: number; targetOptionIdxs?: number[]; branchFromIdx?: number; branchToIdx?: number; branchToEnd?: boolean; formula?: string }
const logicRuleList = ref<LogicRuleItem[]>([])
const showAddRule = ref(false)
const editingRuleIdx = ref(-1)
const defaultRuleForm = (): LogicRuleItem => ({ id: '', conditionType: 'simple', conditions: [{ questionIdx: undefined, optionIdx: undefined, operator: undefined, compareValue: undefined }], action: 'show', scope: 'frontend', targetQuestionIdx: undefined, targetOptionIdxs: [], branchFromIdx: undefined, branchToIdx: undefined, branchToEnd: false, formula: '' })
const ruleForm = ref<LogicRuleItem>(defaultRuleForm())

const choiceTypes = ['select','radio','checkbox','picker','cascade','judge','multiInput','hInput']
function hasOptionsByIndex(idx?: number) { if (idx===undefined) return false; const q = questions.value[idx]; return q && choiceTypes.includes(q.type) && !!q.props?.options?.length }
function getOptionsByIndex(idx?: number) { if (idx===undefined) return []; const q = questions.value[idx]; if (!q || !choiceTypes.includes(q.type)) return []; return q.props?.options || [] }

function conditionToDSL(cond: LogicCondition, qi: number): string {
  const qTag = `Q${cond.questionIdx!+1}`
  if (cond.optionIdx!==undefined) return qTag + `A${cond.optionIdx+1}`
  if (cond.operator==='filled') return `NOT(ISBLANK(${qTag}))`
  if (cond.operator==='empty') return `ISBLANK(${qTag})`
  if (cond.operator) return `${qTag}${cond.operator==='eq'?'==':cond.operator}${cond.compareValue||''}`
  return qTag
}

function renderRuleDSL(rule: LogicRuleItem): string {
  let condStr = ''
  if (rule.conditionType==='none') {
    condStr = ''
  } else if (rule.conditions.length===1) {
    condStr = conditionToDSL(rule.conditions[0], 0)
  } else {
    const parts = rule.conditions.map((c, i) => conditionToDSL(c, i))
    condStr = `${rule.conditionType.toUpperCase()}(${parts.join(', ')})`
  }
  let actionStr = ''
  const actionMap: Record<string, string> = { show: 'SHOW', hide: 'HIDE', required: 'REQUIRED', check: 'CHECK', branch: 'BRANCH', assignment: 'ASSIGNMENT', validate: 'VALIDATE', replace: 'REPLACE', end: 'BRANCH' }
  const a = actionMap[rule.action] || rule.action.toUpperCase()
  if (rule.action==='branch') {
    const fromTag = rule.branchFromIdx!==undefined ? `Q${rule.branchFromIdx+1}` : '?'
    const toTag = rule.branchToEnd ? 'END' : (rule.branchToIdx!==undefined ? `Q${rule.branchToIdx+1}` : '?')
    actionStr = `BRANCH FROM ${fromTag} TO ${toTag}`
  } else if (rule.action==='check') {
    const tTag = rule.targetQuestionIdx!==undefined ? `Q${rule.targetQuestionIdx+1}` : '?'
    const opts = (rule.targetOptionIdxs||[]).map(oi => `${tTag}A${oi+1}`).join(' ')
    actionStr = `CHECK ${opts}`
  } else if (rule.action==='assignment') {
    const tTag = rule.targetQuestionIdx!==undefined ? `Q${rule.targetQuestionIdx+1}` : '?'
    actionStr = `ASSIGNMENT ${tTag} WITH ${rule.formula||'?'}`
  } else if (rule.action==='validate') {
    const tTag = rule.targetQuestionIdx!==undefined ? `Q${rule.targetQuestionIdx+1}` : '?'
    actionStr = `VALIDATE ${tTag} WITH ${rule.formula||'?'}`
  } else if (rule.action==='replace') {
    const tTag = rule.targetQuestionIdx!==undefined ? `Q${rule.targetQuestionIdx+1}` : '?'
    actionStr = `REPLACE ${tTag} WITH ${rule.formula||'?'}`
  } else if (rule.action==='end') {
    const fromTag = rule.branchFromIdx!==undefined ? `Q${rule.branchFromIdx+1}` : '?'
    actionStr = `BRANCH FROM ${fromTag} TO END`
  } else {
    const tTag = rule.targetQuestionIdx!==undefined ? `Q${rule.targetQuestionIdx+1}` : '?'
    actionStr = `${a} ${tTag}`
  }
  return rule.conditionType==='none' ? actionStr : `IF ${condStr} THEN ${actionStr}`
}

function confirmRule() {
  const rf = ruleForm.value
  if (rf.conditionType!=='none') {
    for (const c of rf.conditions) {
      if (c.questionIdx===undefined) { ElMessage.warning('请选择条件题目'); return }
    }
  }
  if (!rf.action) { ElMessage.warning('请选择动作'); return }
  if (['show','hide','required'].includes(rf.action) && rf.targetQuestionIdx===undefined) { ElMessage.warning('请选择目标题目'); return }
  if (rf.action==='check' && rf.targetQuestionIdx===undefined) { ElMessage.warning('请选择目标题目'); return }
  if (rf.action==='branch' && rf.branchFromIdx===undefined) { ElMessage.warning('请选择跳转来源题目'); return }
  if (rf.action==='branch' && !rf.branchToEnd && rf.branchToIdx===undefined) { ElMessage.warning('请选择跳转目标题目或勾选结束问卷'); return }
  if (['assignment','validate','replace'].includes(rf.action) && rf.targetQuestionIdx===undefined) { ElMessage.warning('请选择目标题目'); return }
  rf.scope = rf.action === 'postStat' ? 'backend' : 'frontend'
  rf.id = 'rule_' + Date.now() + '_' + Math.random().toString(36).slice(2,6)
  if (editingRuleIdx.value>=0) {
    logicRuleList.value[editingRuleIdx.value] = { ...rf }
  } else {
    logicRuleList.value.push({ ...rf })
  }
  showAddRule.value = false
  ruleForm.value = defaultRuleForm()
  editingRuleIdx.value = -1
}

function editRule(idx: number) {
  editingRuleIdx.value = idx
  ruleForm.value = JSON.parse(JSON.stringify(logicRuleList.value[idx]))
  showAddRule.value = true
}

function removeRule(idx: number) {
  logicRuleList.value.splice(idx, 1)
}

async function saveLogicRules() {
  if (!form.id) { ElMessage.warning('请先保存考试'); return }
  normalizeCompletionSettings()
  if (form.redirectUrl && !isValidRedirectUrl(form.redirectUrl)) {
    ElMessage.warning('请输入有效的跳转链接')
    return
  }
  const schema = JSON.stringify({ version: '2.0', questions: questions.value })
  const settings = JSON.stringify({
    questionNumber: form.questionNumber, progressBar: form.progressBar,
    autoSave: form.autoSave, password: form.password,
    loginRequired: form.loginRequired, onePageOneQuestion: form.onePageOneQuestion,
      answerSheetVisible: form.answerSheetVisible, answerVisible: form.answerVisible, copyEnabled: form.copyEnabled, language: form.language,
      triggerType: form.triggerType,
      transcriptVisible: form.transcriptVisible, showAnalysis: form.showAnalysis, rankVisible: form.rankVisible,
    redirectUrl: form.redirectUrl, endContent: form.endContent,
    examRankingEnabled: form.examRankingEnabled, exerciseMode: form.exerciseMode,
    randomOrder: form.randomOrder, minSubmitMinutes: form.minSubmitMinutes,
    maxSubmitMinutes: form.maxSubmitMinutes,
    deviceLimit: form.deviceLimit, ipLimit: form.ipLimit, userLimit: form.userLimit,
    backgroundImages: form.backgroundImages, headerImages: form.headerImages
  })
  const payload: any = {
    id: form.id, title: form.title, schema, settings
  }
  try {
    await adminApi.examSave(payload)
    ElMessage.success('逻辑规则已保存')
  } catch { ElMessage.error('保存失败') }
}

async function loadBank() {
  clearTimeout(bankTimer)
  if (bankLoadingTimer) {
    clearTimeout(bankLoadingTimer)
    bankLoadingTimer = null
  }
  bankLoadingVisible.value = false
  const seq = ++bankLoadSeq
  if (!form.id) {
    bankLoading.value = false
    bankQuestions.value = []
    return
  }
  bankTimer = setTimeout(async () => {
    bankLoading.value = true
    bankLoadingTimer = setTimeout(() => {
      if (bankLoading.value && seq === bankLoadSeq) {
        bankLoadingVisible.value = true
      }
    }, 450)
    try {
      const res: any = await adminApi.examQuestionBankList({ keyword: bankKeyword.value, page: 1, pageSize: 100 })
      if (seq === bankLoadSeq) bankQuestions.value = res.data?.list || []
    } catch {
      if (seq === bankLoadSeq) bankQuestions.value = []
    }
    finally {
      if (seq === bankLoadSeq) {
        bankLoading.value = false
        bankLoadingVisible.value = false
        if (bankLoadingTimer) {
          clearTimeout(bankLoadingTimer)
          bankLoadingTimer = null
        }
      }
    }
  }, 300)
}
function addFromBank(q: any) {
  const newQ: any = { id: ++idCounter, type: q.type, title: q.title || '未命名' }
  try { newQ.props = JSON.parse(q.schema) } catch {}
  questions.value.push(newQ)
  selected.value = newQ
}

interface OutlineNode { key: string; title: string; children: { q: any; index: number }[] }
const outlineRoot = computed<OutlineNode>(() => ({
  key: 'root', title: form.title || '未命名考试',
  children: questions.value.filter((q: any) => q.type !== 'divider').map((q: any, i: number) => ({ q, index: i + 1 }))
}))

function plainQuestionTitle(q: any) {
  const d = document.createElement('div')
  d.innerHTML = String(q?.title || '')
  const title = (d.textContent || d.innerText || '').split('\n')[0].replace(/\s+/g, ' ').trim()
  return title || '未命名'
}

function outlineQuestionTitle(q: any) {
  return plainQuestionTitle(q)
}

function bankQuestionTitle(q: any) {
  return plainQuestionTitle(q)
}
const tabsBarRef = ref<HTMLElement | null>(null)
const tabViewportRef = ref<HTMLElement | null>(null)
const tabTrackRef = ref<HTMLElement | null>(null)
const scrollPos = ref(0)
const atTabStart = computed(() => scrollPos.value <= 0)
const atTabEnd = ref(false)
function scrollTabs(dir: number) {
  const vp = tabViewportRef.value; const tr = tabTrackRef.value
  if (!vp || !tr) return
  const step = vp.clientWidth * 0.6
  let newPos = scrollPos.value + dir * step
  newPos = Math.max(0, Math.min(newPos, tr.scrollWidth - vp.clientWidth))
  scrollPos.value = newPos; tr.style.transform = `translateX(${-newPos}px)`
  atTabEnd.value = newPos >= tr.scrollWidth - vp.clientWidth - 1
}
function updateTabScroll() {
  const vp = tabViewportRef.value; const tr = tabTrackRef.value
  if (!vp || !tr) return; scrollPos.value = 0; tr.style.transform = 'translateX(0)'
  atTabEnd.value = tr.scrollWidth <= vp.clientWidth
}
const types = ref<any[]>([])
const FALLBACK_TYPES = [
  { type:'radio', displayName:'单选题', category:'select' },
  { type:'checkbox', displayName:'多选题', category:'select' },
  { type:'select', displayName:'下拉题', category:'select' },
  { type:'picker', displayName:'选择器', category:'select' },
  { type:'cascade', displayName:'级联选择', category:'select' },
  { type:'judge', displayName:'判断题', category:'select' },
  { type:'file', displayName:'上传文件', category:'select' },
  { type:'input', displayName:'单行文本', category:'fill' },
  { type:'textarea', displayName:'多行文本', category:'fill' },
  { type:'number', displayName:'数字', category:'fill' },
  { type:'multiInput', displayName:'多项填空', category:'fill' },
  { type:'hInput', displayName:'横向填空', category:'fill' },
  { type:'signature', displayName:'电子签名', category:'fill' },
  { type:'scanCode', displayName:'扫码', category:'fill' },
  { type:'rating', displayName:'评分', category:'rating' },
  { type:'nps', displayName:'NPS', category:'rating' },
  { type:'matrixRadio', displayName:'矩阵单选', category:'matrix' },
  { type:'matrixCheckbox', displayName:'矩阵多选', category:'matrix' },
  { type:'matrixFillBlank', displayName:'矩阵填空', category:'matrix' },
  { type:'matrixAuto', displayName:'表格自增', category:'matrix' },
  { type:'divider', displayName:'分割线', category:'layout' },
  { type:'description', displayName:'文字描述', category:'layout' },
  { type:'questionSet', displayName:'问题组', category:'layout' },
  { type:'pagination', displayName:'分页', category:'layout' },
  { type:'user', displayName:'成员', category:'advanced' },
  { type:'dept', displayName:'部门', category:'advanced' },
  { type:'richText', displayName:'富文本', category:'advanced' },
  { type:'autopop', displayName:'自动填充', category:'advanced' },
  { type:'name', displayName:'姓名', category:'personal' },
  { type:'studentId', displayName:'学号', category:'personal' },
  { type:'employeeId', displayName:'工号', category:'personal' },
  { type:'class', displayName:'班级', category:'personal' },
  { type:'phone', displayName:'手机', category:'personal' },
  { type:'email', displayName:'邮箱', category:'personal' },
  { type:'idCard', displayName:'身份证', category:'personal' },
  { type:'password', displayName:'密码', category:'personal' },
  { type:'date', displayName:'日期', category:'personal' },
  { type:'time', displayName:'时间', category:'personal' },
  { type:'dateRange', displayName:'日期范围', category:'personal' },
  { type:'switch', displayName:'开关', category:'personal' },
  { type:'location', displayName:'地理位置', category:'personal' },
]
const categoryDefs = [
  { name: 'select', label: '选择题', types: ['radio','checkbox','select','cascade','picker','judge','file'] },
  { name: 'fill', label: '填空题', types: ['input','textarea','number','multiInput','hInput','signature','scanCode'] },
  { name: 'rating', label: '打分题', types: ['rating','nps'] },
  { name: 'matrix', label: '矩阵题', types: ['matrixRadio','matrixCheckbox','matrixFillBlank','matrixAuto'] },
  { name: 'layout', label: '辅助布局', types: ['divider','description','questionSet','pagination'] },
  { name: 'advanced', label: '高级题型', types: ['user','dept','richText','autopop'] },
  { name: 'personal', label: '个人信息', types: ['name','studentId','employeeId','class','phone','email','idCard','password','date','time','dateRange','switch','location'] },
]
const categories = computed(() => categoryDefs.filter(c => types.value.some(t => c.types.includes(t.type))))
const typesByCategory = computed(() => {
  const m: Record<string, any[]> = {}
  for (const cat of categoryDefs) {
    const items = types.value.filter(t => cat.types.includes(t.type))
    if (items.length) m[cat.name] = items
  }
  return m
})
const groupedBySub = computed(() => {
  const cat = activeCategory.value
  if (!cat || !typesByCategory.value[cat]?.length) return {}
  return { [cat]: typesByCategory.value[cat] }
})

function typeName(t: string) {
  const map: Record<string, string> = {
    input:'单行文本',textarea:'多行文本',number:'数字',select:'下拉',radio:'单选',checkbox:'多选',
    picker:'选择器',cascade:'级联',judge:'判断',multiInput:'多项填空',hInput:'横向填空',
    scanCode:'扫码',signature:'签名',file:'上传',rating:'评分',nps:'NPS',
    phone:'手机',email:'邮箱',idCard:'身份证',password:'密码',switch:'开关',
    name:'姓名',studentId:'学号',employeeId:'工号',class:'班级',
    location:'地理位置',date:'日期',time:'时间',dateRange:'日期范围',
    matrixRadio:'矩阵单选',matrixCheckbox:'矩阵多选',matrixFillBlank:'矩阵填空',matrixAuto:'表格自增',
    divider:'分割线',description:'说明',questionSet:'问题组',pagination:'分页',user:'成员',dept:'部门',richText:'富文本'
  }
  return map[t] || t
}

interface Question { id: string; type: string; title: string; required: boolean; placeholder?: string; props?: any; validate?: any[]; logic?: any[]; calcValue?: any; readOnly?: boolean; showDescription?: boolean; defaultHidden?: boolean; optionLayout?: number; multiple?: boolean; examScoreMode?: string; examScore?: number; examCorrectAnswer?: string; examAnalysis?: string; fileTypes?: string[]; fileExtensions?: string; maxFileSize?: number; maxFileCount?: number; examAnswerMode?: string; dataType?: string; _existing?: boolean; [key: string]: any }
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
const formulaVisible = ref(false)
const formulaTitle = ref('')
const formulaText = ref('')
const formulaType = ref('')
const formulaKey = ref(0)
const formulaQuestions = computed(() => questions.value.filter((q: any) => !isPureLayout(q.type)))
const formulaPlaceholder = computed(() => {
  const map: Record<string, string> = {
    endFormula: 'IF Q1A1 THEN END\nIF AND(Q1A1, Q2>10) THEN END',
    jumpFormula: 'IF Q1A1 THEN BRANCH FROM Q1 TO Q5\nIF Q2>100 THEN BRANCH FROM Q2 TO Q8',
    textReplace: 'REPLACE Q2 WITH CONCATENATE("你好", Q1)',
    calculate: 'SUM(Q1, Q2)\nQ1 + Q2\nIF(Q1>60, "及格", "不及格")'
  }
  return map[formulaType.value] || '输入公式表达式'
})
function openFormulaDialog(type: string) {
  if (!selected.value) return
  const titles: Record<string, string> = {
    endFormula: '结束公式', jumpFormula: '跳转公式',
    textReplace: '文本替换', calculate: '计算公式', validate: '校验规则'
  }
  const fieldMap: Record<string, string> = {
    endFormula: 'endFormula', jumpFormula: 'jumpFormula',
    textReplace: 'replaceTextRule', calculate: 'calculateFormula', validate: 'validateRule'
  }
  const field = fieldMap[type] || ''
  formulaType.value = type
  formulaTitle.value = titles[type] || '公式编辑'
  formulaText.value = selected.value[field] || selected.value.props?.[field] || ''
  formulaKey.value++
  formulaVisible.value = true
}
function confirmFormula() {
  if (!selected.value) return
  const fieldMap: Record<string, string> = {
    endFormula: 'endFormula', jumpFormula: 'jumpFormula',
    textReplace: 'replaceTextRule', calculate: 'calculateFormula', validate: 'validateRule'
  }
  const field = fieldMap[formulaType.value] || ''
  if (formulaType.value === 'calculate' || formulaType.value === 'textReplace') {
    if (!selected.value.props) selected.value.props = {}
    selected.value.props[field] = formulaText.value
  } else {
    selected.value[field] = formulaText.value
  }
  formulaVisible.value = false
}
function insertFormulaTag(tag: string) {
  const ta = document.querySelector('.el-dialog .el-textarea textarea') as HTMLTextAreaElement
  if (ta) {
    const start = ta.selectionStart, end = ta.selectionEnd
    formulaText.value = formulaText.value.slice(0, start) + tag + formulaText.value.slice(end)
    nextTick(() => { ta.selectionStart = ta.selectionEnd = start + tag.length; ta.focus() })
  } else {
    formulaText.value += tag
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

const fillTemplates: Record<string, string> = { ef: '默认模版', ef1: '新样式模版' }
const publicUrl = computed(() => form.id ? `${window.location.origin}/${form.fillTemplate || 'ef'}/${form.id}` : '')
function copyLink() { navigator.clipboard.writeText(publicUrl.value); ElMessage.success('已复制') }

function goBack() { router.push('/exam/list') }
function goResponses() {
  if (form.id) router.push({ path: '/exam/responses', query: { examId: String(form.id), title: form.title } })
}
function goStatistic() {
  if (form.id) router.push({ path: '/exam/statistic', query: { examId: String(form.id), title: form.title } })
}

async function load() {
  const id = Number(route.query.id || 0)
  if (!id) return
  try {
    const res: any = await adminApi.examDetail(id)
    const sv = res.data?.survey
    if (!sv) { ElMessage.error('加载失败'); return }
    Object.assign(form, {
      id: sv.id, title: sv.title, description: sv.description || '',
      category: sv.category || '', tags: sv.tags || '',
      visibility: sv.visibility ?? 0, allowMultiBool: sv.allowMulti ?? 0,
      anonymousBool: sv.anonymous ?? 0, showResultBool: sv.showResult ?? 0,
      startDate: sv.startTime || null, endDate: sv.endTime || null,
      maxResponse: sv.maxResponse ?? 0, statusBool: sv.status ?? 1, deptIds: sv.deptIds || '',
      duration: sv.duration ?? 60, maxAttempts: sv.maxAttempts ?? 1, showScore: sv.showScore ?? 1,
      createBy: sv.createBy || 0
    })
    if (sv.settings) {
      try {
        const s = JSON.parse(sv.settings)
        Object.assign(form, {
          questionNumber: s.questionNumber ?? true, progressBar: s.progressBar ?? false,
          autoSave: s.autoSave ?? false, password: s.password || '',
          loginRequired: s.loginRequired ?? false,
          onePageOneQuestion: s.onePageOneQuestion ?? false,
          answerSheetVisible: s.answerSheetVisible ?? true, answerVisible: s.answerVisible ?? true,
          copyEnabled: s.copyEnabled ?? true, language: s.language || 'zh', triggerType: s.triggerType || 'onBlur',
          transcriptVisible: s.transcriptVisible ?? true, showAnalysis: s.showAnalysis ?? true,
          rankVisible: s.rankVisible ?? false, redirectUrl: s.redirectUrl || '', endContent: s.endContent || '',
          examRankingEnabled: s.examRankingEnabled ?? false,
          exerciseMode: s.exerciseMode ?? false, randomOrder: s.randomOrder ?? false,
          minSubmitMinutes: s.minSubmitMinutes || 0, maxSubmitMinutes: s.maxSubmitMinutes || 0,
          deviceLimit: s.deviceLimit || 0, ipLimit: s.ipLimit || 0, userLimit: s.userLimit || 0,
          backgroundImages: s.backgroundImages || [], headerImages: s.headerImages || [],
          collaborators: s.collaborators || '',
          fillTemplate: s.fillTemplate || 'ef'
        })
      } catch {}
    }
    syncCompletionActionFromForm()
    const rawSchema = res.data?.schema || ''
    if (rawSchema) {
      try {
        const sch = JSON.parse(rawSchema)
        if (sch.questions) { questions.value = sch.questions; idCounter = questions.value.length; selected.value = questions.value[0] || null }
      } catch {}
    }
    loadResources()
    await loadAdminTree()
    await loadDeptTree()
  } catch { ElMessage.error('加载失败') }
}

function downloadJson() {
  const blob = new Blob([exportedJson.value], { type: 'application/json;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `${form.title || 'exam'}.json`
  a.click()
  URL.revokeObjectURL(url)
}
function loadJson() {
  const input = document.createElement('input')
  input.type = 'file'
  input.accept = '.json'
  input.onchange = async (e: any) => {
    const file = e.target.files[0]
    if (!file) return
    try {
      const text = await file.text()
      const data = JSON.parse(text)
      if (data.questions) questions.value = data.questions
      exportedJson.value = text
      ElMessage.success('已导入')
    } catch { ElMessage.error('JSON 解析失败') }
  }
  input.click()
}

let autoSaveTimer: ReturnType<typeof setTimeout> | null = null
let designerLoaded = false
const designerAutoSaveSnapshot = computed(() => JSON.stringify({
  title: form.title,
  description: form.description,
  category: form.category,
  tags: form.tags,
  visibility: form.visibility,
  allowMultiBool: form.allowMultiBool,
  anonymousBool: form.anonymousBool,
  showResultBool: form.showResultBool,
  startDate: form.startDate,
  endDate: form.endDate,
  maxResponse: form.maxResponse,
  statusBool: form.statusBool,
  deptIds: form.deptIds,
  duration: form.duration,
  maxAttempts: form.maxAttempts,
  showScore: form.showScore,
  collaborators: form.collaborators,
  questionNumber: form.questionNumber,
  progressBar: form.progressBar,
  autoSave: form.autoSave,
  password: form.password,
  loginRequired: form.loginRequired,
  onePageOneQuestion: form.onePageOneQuestion,
  answerSheetVisible: form.answerSheetVisible,
  answerVisible: form.answerVisible,
  copyEnabled: form.copyEnabled,
  language: form.language,
  triggerType: form.triggerType,
  transcriptVisible: form.transcriptVisible,
  showAnalysis: form.showAnalysis,
  rankVisible: form.rankVisible,
  redirectUrl: form.redirectUrl,
  endContent: form.endContent,
  examRankingEnabled: form.examRankingEnabled,
  exerciseMode: form.exerciseMode,
  randomOrder: form.randomOrder,
  minSubmitMinutes: form.minSubmitMinutes,
  maxSubmitMinutes: form.maxSubmitMinutes,
  deviceLimit: form.deviceLimit,
  ipLimit: form.ipLimit,
  userLimit: form.userLimit,
  backgroundImages: form.backgroundImages,
  headerImages: form.headerImages,
  fillTemplate: form.fillTemplate
}))

async function save(showMessage: boolean | Event = true) {
  const shouldShowMessage = typeof showMessage === 'boolean' ? showMessage : true
  if (!form.title) {
    if (shouldShowMessage) ElMessage.warning('请填写标题')
    return
  }
  normalizeCompletionSettings()
  if (form.redirectUrl && !isValidRedirectUrl(form.redirectUrl)) {
    if (shouldShowMessage) ElMessage.warning('请输入有效的跳转链接')
    return
  }
  if (autoSaveTimer) {
    clearTimeout(autoSaveTimer)
    autoSaveTimer = null
  }
  saving.value = true
  try {
    const schema = JSON.stringify({ version: '2.0', questions: questions.value, setting: {} })
    const settings = JSON.stringify({
      collaborators: form.collaborators,
      questionNumber: form.questionNumber, progressBar: form.progressBar,
      autoSave: form.autoSave, password: form.password,
      loginRequired: form.loginRequired, onePageOneQuestion: form.onePageOneQuestion,
    answerSheetVisible: form.answerSheetVisible, answerVisible: form.answerVisible, copyEnabled: form.copyEnabled, language: form.language,
      triggerType: form.triggerType,
      transcriptVisible: form.transcriptVisible, showAnalysis: form.showAnalysis, rankVisible: form.rankVisible,
    redirectUrl: form.redirectUrl, endContent: form.endContent,
    examRankingEnabled: form.examRankingEnabled, exerciseMode: form.exerciseMode,
    randomOrder: form.randomOrder, minSubmitMinutes: form.minSubmitMinutes,
    maxSubmitMinutes: form.maxSubmitMinutes,
    deviceLimit: form.deviceLimit, ipLimit: form.ipLimit, userLimit: form.userLimit,
    backgroundImages: form.backgroundImages, headerImages: form.headerImages,
    fillTemplate: form.fillTemplate
  })
  const payload: any = {
    title: form.title, description: form.description, category: form.category, tags: form.tags,
    visibility: form.visibility, allowMulti: form.allowMultiBool,
    anonymous: form.anonymousBool, showResult: form.showResultBool,
    startTime: form.startDate || 0, endTime: form.endDate || 0,
    maxResponse: form.maxResponse, status: form.statusBool, mode: 'exam',
    deptIds: form.deptIds, schema, settings,
    duration: form.duration, maxAttempts: form.maxAttempts, showScore: form.showScore
  }
  if (form.id) payload.id = form.id
  const r: any = await adminApi.examSave(payload)
  if (!form.id) { form.id = r.id || r.data?.id; router.replace({ query: { id: String(form.id) } }) }
  if (shouldShowMessage) ElMessage.success('已保存')
} catch {
  if (shouldShowMessage) ElMessage.error('保存失败')
}
finally { saving.value = false }
}

watch(middleTab, (v) => { if (v === 'appearance') loadResources() })
watch(sideSubTab, (v) => { if (v === 'bank') loadBank() })
watch(() => form.visibility, (v) => {
  if (v === 2 && deptTreeOptions.value.length) {
    nextTick(() => {
      if (form.deptIds) {
        const keys = form.deptIds.split(',').map((s: string) => parseInt(s.trim())).filter((n: number) => !isNaN(n))
        deptTreeRef.value?.setCheckedKeys(keys)
      }
    })
  }
})

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

watch(designerAutoSaveSnapshot, () => {
  if (!form.id || !designerLoaded) return
  if (autoSaveTimer) clearTimeout(autoSaveTimer)
  autoSaveTimer = setTimeout(() => { save(false) }, 500)
})

onBeforeUnmount(() => {
  if (autoSaveTimer) clearTimeout(autoSaveTimer)
})

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
  const cats = categories.value; if (cats.length) activeCategory.value = cats[0].name
  await nextTick(); updateTabScroll(); await load(); await nextTick(); designerLoaded = true
})
</script>

<style scoped>
.survey-main {
  display:flex; height:calc(100vh - 56px); background:#fff;
  --designer-accent:#0f766e;
  --designer-accent-soft:#ecfdf5;
  --designer-accent-border:#99f6e4;
  --designer-accent-hover:#ccfbf1;
  --designer-border:#e5e7eb;
  --designer-bg:#f8fafc;
}
.survey-main-navigator { display:flex; flex:0 0 54px; flex-direction:column; justify-content:space-between; border-right:1px solid #e8e8e8; background:#fafafa; padding:8px 0; align-items:center; }
.nav-actions { display:flex; flex-direction:column; align-items:center; gap:4px; }
.nav-btn { width:36px; height:36px; min-width:36px; padding:0; margin:0; border:none; border-radius:8px; color:#999; font-size:13px; cursor:pointer; display:flex; align-items:center; justify-content:center; background:transparent; outline:none; }
.nav-btn:hover { background:#f0f0f0; color:#666; }
.nav-btn.active { background:var(--designer-accent-soft); color:var(--designer-accent); box-shadow:0 6px 14px rgba(15,118,110,0.12); }
.survey-main-content { flex:1; display:flex; flex-direction:column; overflow:hidden; }
.survey-editor { display:flex; height:100%; }
.survey-sidebar-panel { width:280px; flex-shrink:0; border-right:1px solid #e8e8e8; background:#fafafa; display:flex; flex-direction:row; overflow:hidden; }
.survey-sidebar-panel.compact { width:48px; background:#f5f5f5; }
.survey-sidebar-panel.compact .survey-sidebar-panel-tabs { border-right:none; }
.survey-sidebar-panel-tabs { flex:0 0 48px; display:flex; flex-direction:column; border-right:1px solid #e8e8e8; background:#f5f5f5; padding:4px 0; }
.survey-sidebar-panel-tabs-pane { display:flex; flex-direction:column; align-items:center; justify-content:center; gap:2px; padding:10px 2px; cursor:pointer; color:#999; font-size:10px; }
.survey-sidebar-panel-tabs-pane:hover { color:var(--designer-accent); }
.survey-sidebar-panel-tabs-pane.active { color:var(--designer-accent); background:#fff; }
.survey-sidebar-panel-tabs-pane .tab-label { color:inherit; line-height:1.2; }
.survey-sidebar-panel-tabs-content { flex:1; display:flex; flex-direction:column; overflow:hidden; }
.side-sub-tabs { display:flex; flex-direction:column; flex:1; overflow:hidden; }
.side-sub-tabs :deep(.el-tabs__header) { margin:0; padding:0 8px; background:#fafafa; flex-shrink:0; }
.side-sub-tabs :deep(.el-tabs__active-bar) { background:var(--designer-accent); }
.side-sub-tabs :deep(.el-tabs__content) { flex:1; overflow:hidden; }
.side-sub-tabs :deep(.el-tab-pane) { height:100%; overflow-y:auto; }
.question-panel { padding:8px; }
.type-tabs-bar { display:flex; align-items:center; gap:4px; margin-bottom:8px; }
.tab-scroll-btn { width:20px; height:24px; border:1px solid #e8e8e8; border-radius:4px; background:#fff; cursor:pointer; font-size:12px; color:#666; flex-shrink:0; display:flex; align-items:center; justify-content:center; }
.tab-scroll-btn:disabled { opacity:0.3; cursor:default; }
.tab-scroll-viewport { flex:1; overflow:hidden; }
.tab-scroll-track { display:flex; gap:4px; transition:transform 0.2s; }
.type-tab-btn { padding:2px 10px; border:1px solid #e8e8e8; border-radius:4px; background:#fff; cursor:pointer; font-size:12px; color:#666; white-space:nowrap; }
.type-tab-btn.active { color:#fff; background:var(--designer-accent); border-color:var(--designer-accent); }
.type-tab-btn:hover { border-color:var(--designer-accent-border); color:var(--designer-accent); background:var(--designer-accent-soft); }
.question-type { padding:0 2px 8px; }
.menu-group { margin:0; padding:0; }
.menu-group-item {
  display:flex; align-items:center; justify-content:center; gap:7px; width:calc(100% - 12px);
  min-height:32px; padding:5px 10px; margin:0 auto 6px; box-sizing:border-box; line-height:20px;
  cursor:pointer; border-radius:8px; border:1px solid #e5e7eb; background:#fff; font-size:12px;
  transition:background 0.12s, border-color 0.12s, color 0.12s, box-shadow 0.12s;
}
.menu-group-item:hover { background:var(--designer-accent-soft); color:var(--designer-accent); border-color:var(--designer-accent-border); box-shadow:0 6px 14px rgba(15,118,110,0.08); }
.itemIcon {
  flex-shrink:0; width:20px; height:20px; display:inline-flex; align-items:center; justify-content:center;
  color:#999; font-size:15px; line-height:1;
}
.itemIcon :deep(svg) { display:block; width:1em; height:1em; }
.menu-group-item:hover .itemIcon { color:var(--designer-accent); }
.item-label { flex:0 1 auto; min-width:0; overflow:hidden; text-overflow:ellipsis; white-space:nowrap; font-size:13px; }
.outline-tree { height:100%; min-height:0; padding:6px 7px 10px; display:flex; flex-direction:column; background:#fff; box-sizing:border-box; }
.tree-root {
  display:flex; align-items:center; justify-content:space-between; gap:8px; flex-shrink:0;
  padding:9px 8px 10px; border-bottom:1px solid var(--designer-border); margin-bottom:6px; background:#fff;
}
.tree-root-main { display:flex; align-items:center; gap:8px; min-width:0; }
.tree-root-icon { font-size:16px; flex-shrink:0; }
.tree-root-title { flex:1; min-width:0; overflow:hidden; text-overflow:ellipsis; white-space:nowrap; font-weight:650; font-size:13px; color:#1f2937; }
.tree-root-count {
  flex-shrink:0; font-size:11px; color:var(--designer-accent); font-weight:600;
  background:var(--designer-accent-soft); border:1px solid var(--designer-accent-border);
  border-radius:999px; padding:1px 8px; line-height:18px;
}
.tree-children {
  position:relative; flex:1; min-height:0; overflow-y:auto; padding:4px 2px 2px 18px;
  display:flex; flex-direction:column; gap:3px;
}
.tree-children::before {
  content:''; position:absolute; left:9px; top:10px; bottom:10px;
  border-left:1px dashed #cbd5e1; pointer-events:none;
}
.tree-child {
  display:flex; align-items:center; gap:7px; cursor:pointer; border-radius:8px; min-height:34px;
  padding:5px 6px; border:1px solid transparent; position:relative;
  transition:background 0.12s, border-color 0.12s, box-shadow 0.12s;
}
.tree-child::after {
  content:''; position:absolute; left:-9px; top:50%; width:9px;
  border-top:1px dashed #cbd5e1; transform:translateY(-50%); pointer-events:none;
}
.tree-child:hover { background:#f0fdfa; border-color:var(--designer-accent-border); }
.tree-child.active { background:var(--designer-accent-soft); border-color:var(--designer-accent-border); box-shadow:none; }
.tree-child.active::before {
  content:''; position:absolute; left:0; top:7px; bottom:7px; width:3px;
  border-radius:999px; background:var(--designer-accent);
}
.tree-child-body {
  display:flex; align-items:center; gap:6px; flex:1; min-width:0;
}
.tree-child.active .tree-child-body { font-weight:500; }
.tree-index {
  display:flex; align-items:center; justify-content:flex-end; flex:0 0 26px; height:22px;
  color:#98a2b3; font-size:12px; font-weight:600; line-height:1;
}
.tree-child.active .tree-index { color:var(--designer-accent); }
.tree-icon { flex-shrink:0; font-size:13px; width:16px; text-align:center; color:#98a2b3; }
.tree-child:hover .tree-icon,
.tree-child.active .tree-icon { color:var(--designer-accent); }
.tree-title {
  flex:1; min-width:0; overflow:hidden; text-overflow:ellipsis; white-space:nowrap;
  color:#344054; font-size:13px; line-height:20px;
}
.tree-child.active .tree-title { color:#1f2937; font-weight:650; }
.tree-type {
  max-width:64px; overflow:hidden; text-overflow:ellipsis; white-space:nowrap;
  font-size:10px; color:#667085; padding:1px 6px; border-radius:999px; background:#f2f4f7; flex-shrink:0; line-height:16px;
}
.tree-required {
  display:flex; align-items:center; justify-content:center; flex:0 0 16px; height:16px;
  font-size:10px; color:#dc6803; border-radius:999px; background:#fff7ed;
}
.tree-child:hover .tree-type,
.tree-child.active .tree-type { background:var(--designer-accent-soft); color:var(--designer-accent); }
.outline-empty { flex:1; display:flex; align-items:center; justify-content:center; }
.survey-main-panel { flex:1; display:flex; flex-direction:column; overflow:hidden; background-color:#f7f8fa; }
.survey-main-panel-toolbar { display:flex; align-items:center; justify-content:space-between; padding:8px 16px; background:#fff; border-bottom:1px solid #e8e8e8; gap:12px; }
.toolbar-left { display:flex; align-items:center; gap:8px; }
.toolbar-right { display:flex; align-items:center; gap:8px; }
.toolbar-btn-group { display:inline-flex; align-items:center; background:#f5f6f8; border-radius:6px; padding:2px; gap:0; }
.toolbar-btn-group .el-button { border:none; }
.toolbar-btn.active { background:#fff; border-radius:4px; box-shadow:0 1px 2px rgba(0,0,0,0.08); }
.toolbar-divider { width:1px; height:20px; background:#e0e0e0; margin:0 4px; }
.toolbar-score { font-size:13px; color:#606266; white-space:nowrap; margin-right:4px; }
.survey-main-panel-content { flex:1; overflow-y:auto; padding:20px 40px; }
.editor-wrapper { max-width:210mm; margin:0 auto; padding:20px 0; }
.editor { background:#fff; border-radius:12px; box-shadow:0 2px 12px rgba(0,0,0,0.06); overflow:hidden; }
.header { padding:24px 28px 12px; text-align:center; border-bottom:1px solid #f0f0f0; position:relative; min-height:60px; background:rgba(255,255,255,0.85); backdrop-filter:blur(4px); }
.header-title-input :deep(.el-input__wrapper) { box-shadow:none!important; padding:0; background:transparent; }
.header-title-input :deep(.el-input__inner) { font-size:22px; font-weight:600; color:#303133; text-align:center; border:none; cursor:pointer; }
.header-title-input :deep(.el-input__inner):focus { cursor:text; }
.questions-area { padding:16px 24px; }
.questions-area :deep(.draggable-list) { gap:8px; }
.footer { padding:24px 28px; text-align:center; color:#999; font-size:13px; border-top:1px solid #f0f0f0; }
.survey-setting-panel { width:380px; flex-shrink:0; border-left:1px solid #e8e8e8; background:#fafafa; overflow-y:auto; }
.json-panel :deep(.el-textarea__inner) { height:100% !important; min-height:500px; font-size:13px; }
.survey-preview-panel { max-width:210mm; margin:0 auto; padding:20px 40px; overflow-y:auto; }
.props-panel { padding:12px; }
.props-panel h3 { font-size:14px; font-weight:500; color:var(--designer-accent); margin:0 0 8px; padding-bottom:8px; border-bottom:2px solid var(--designer-accent); }
.props-panel :deep(.el-form-item) { margin-bottom:8px; }
.props-panel :deep(.el-form-item__label) { font-size:12px; color:#666; padding-bottom:2px; font-weight:500; line-height:1.2; }
.props-panel :deep(.el-divider) { margin:8px 0; }
.props-panel :deep(.el-collapse-item__header) { font-size:12px; font-weight:500; }
.exam-scoring-section { padding:0 2px; }
.exam-scoring-section .section-title { font-size:13px; font-weight:500; color:#303133; margin-bottom:8px; }
.exam-formula-rows { padding:0 2px; border-top:1px dashed #e0e0e0; margin-top:8px; padding-top:8px; }
.exam-formula-rows .setting-row { display:flex; align-items:center; justify-content:space-between; margin-bottom:6px; font-size:12px; color:#666; }
.exam-formula-rows .setting-row .setting-label { font-weight:500; }
.props-options-section { margin-bottom:8px; padding:8px; background:#f5f6f8; border-radius:6px; }
.setting-opt-row { display:flex; align-items:center; gap:6px; margin-bottom:4px; }
.setting-wrapper { height:100%; overflow-y:auto; background:var(--designer-bg); }
.setting-header {
  position:sticky; top:0; z-index:5; display:flex; align-items:center; justify-content:space-between;
  gap:16px; padding:18px 24px; background:rgba(248,250,252,0.94);
  border-bottom:1px solid var(--designer-border); backdrop-filter:saturate(120%) blur(10px);
}
.setting-page-title { font-size:18px; font-weight:700; color:#1f2937; line-height:1.3; }
.setting-page-desc { margin-top:3px; font-size:12px; color:#667085; line-height:1.5; }
.setting-header-actions { display:flex; align-items:center; justify-content:flex-end; gap:10px; flex-shrink:0; }
.setting-scroll { display:grid; grid-template-columns:repeat(auto-fill, minmax(360px, 1fr)); gap:18px; padding:24px; align-items:start; }
.setting-group {
  padding:18px 20px; background:#fff; border-radius:10px;
  border:1px solid var(--designer-border); box-shadow:0 8px 18px rgba(15,23,42,0.04);
}
.group-header { display:flex; align-items:flex-start; justify-content:space-between; gap:12px; margin-bottom:14px; }
.group-title { font-size:14px; font-weight:650; color:#344054; margin-bottom:3px; letter-spacing:0; }
.group-desc { color:#98a2b3; font-size:12px; line-height:1.5; }
/* 外观面板 */
.appearance-panel { padding:12px; }
.appearance-panel :deep(.el-tabs__header) { margin:0 0 8px; }
.appearance-panel :deep(.el-tabs__item) { font-size:12px; padding:0 12px; }
.appearance-grid { display:flex; flex-wrap:wrap; gap:8px; }
.appearance-add, .appearance-thumb { width:72px; height:72px; border-radius:6px; overflow:hidden; flex-shrink:0; }
.appearance-add { display:flex; align-items:center; justify-content:center; border:2px dashed #e8e8e8; background:#fff; cursor:pointer; font-size:24px; color:#ccc; transition:all .15s; }
.appearance-add:hover { border-color:var(--designer-accent); color:var(--designer-accent); }
.appearance-thumb { position:relative; cursor:pointer; border:2px solid transparent; transition:border-color .15s; }
.appearance-thumb:hover { border-color:var(--designer-accent); }
.appearance-thumb.active { border-color:var(--designer-accent); background:var(--designer-accent-soft); }
.appearance-thumb img { display:block; width:100%; height:100%; object-fit:cover; }
.appearance-overlay { position:absolute; inset:0; display:flex; align-items:center; justify-content:center; background:rgba(0,0,0,0.4); opacity:0; transition:opacity .15s; font-size:11px; color:#fff; }
.appearance-thumb:hover .appearance-overlay { opacity:1; }
.appearance-remove { position:absolute; top:2px; right:2px; width:18px; height:18px; border-radius:50%; border:none; background:rgba(0,0,0,0.5); color:#fff; font-size:10px; line-height:1; cursor:pointer; display:flex; align-items:center; justify-content:center; }

/* 逻辑面板 */
.logic-full-panel { display:flex; flex:1; flex-direction:column; overflow:hidden; background:#f7f8fa; padding:20px 32px; }
.logic-toolbar { display:flex; align-items:center; justify-content:space-between; margin-bottom:12px; flex-shrink:0; }
.logic-body { display:flex; gap:16px; flex:1; overflow:hidden; }
.logic-editor-area { flex:1; display:flex; flex-direction:column; overflow-y:auto; }
.logic-sidebar { width:260px; flex-shrink:0; overflow-y:auto; background:#fff; border-radius:6px; padding:12px; border:1px solid #e8e8e8; line-height:1.8; }
.logic-rule-card { background:#fff; border:1px solid #e8e8e8; border-radius:6px; margin-bottom:6px; transition:box-shadow .15s; }
.logic-rule-card:hover { box-shadow:0 1px 4px rgba(0,0,0,0.06); }
.rule-header { display:flex; align-items:center; gap:8px; padding:8px 12px; }
.rule-index { font-size:11px; color:#999; font-weight:600; width:24px; flex-shrink:0; }
.rule-dsl { font-size:12px; font-family:Consolas,Monaco,"Courier New",monospace; flex:1; color:#303133; }
.rule-actions { flex-shrink:0; }

.settings-grid { display:flex; flex-direction:column; gap:0; }
.settings-grid .el-form-item { width:100%; min-width:0; box-sizing:border-box; }
.settings-grid .el-form-item .el-input,
.settings-grid .el-form-item .el-select,
.settings-grid .el-form-item .el-input-number,
.settings-grid .el-form-item .el-date-editor { width:100% !important; }
.settings-grid .el-form-item .el-form-item__content { min-width:0; overflow:hidden; }
.setting-group { overflow:hidden; }
.settings-full { width:100%; }
.setting-control-stack { display:flex; flex-direction:column; gap:4px; width:100%; min-width:0; }
.setting-help { width:100%; margin-top:4px; font-size:12px; color:#98a2b3; line-height:1.5; }
.setting-help.danger { color:#f04438; }
.setting-empty-tip { color:#98a2b3; font-size:13px; line-height:32px; }
.completion-radio-group { display:flex; width:100%; min-width:0; }
.completion-radio-group :deep(.el-radio-button) { flex:1; min-width:0; }
.completion-radio-group :deep(.el-radio-button__inner) { width:100%; padding-left:8px; padding-right:8px; }
.setting-radio-group {
  display:grid; grid-template-columns:repeat(3, minmax(0, 1fr)); gap:8px; width:100%;
}
.setting-radio-group :deep(.el-radio) {
  width:100%; margin-right:0; justify-content:center; background:#fff;
}
.setting-tree-box {
  width:100%; max-height:240px; overflow:auto; border:1px solid #dcdfe6;
  border-radius:8px; padding:8px; box-sizing:border-box; background:#fbfcfe;
}
.setting-tree-box.large { max-height:400px; }
.share-link-row { display:flex; gap:8px; width:100%; min-width:0; }
.qr-preview {
  display:flex; align-items:center; gap:14px; width:100%; padding:12px;
  border:1px solid #e5e7eb; border-radius:10px; background:#fbfefd; box-sizing:border-box;
}
.qr-preview img {
  width:104px; height:104px; flex-shrink:0; border:1px solid #d9f3ee; border-radius:8px; background:#fff;
}
.qr-preview-meta { display:flex; flex-direction:column; gap:4px; min-width:0; }
.qr-preview-meta strong { color:#1f2937; font-size:13px; }
.qr-preview-meta span { color:#98a2b3; font-size:12px; line-height:1.5; }
:deep(.el-table .cell) { overflow:hidden; text-overflow:ellipsis; white-space:nowrap; }
:deep(.el-table th.el-table__cell > .cell) { overflow:hidden; text-overflow:ellipsis; white-space:nowrap; }

/* 题库面板 */
.bank-panel { padding:10px 8px 12px; display:flex; flex-direction:column; height:100%; background:#fff; box-sizing:border-box; }
.bank-search { margin-bottom:10px; flex-shrink:0; }
.bank-search :deep(.el-input__wrapper) { border-radius:8px; box-shadow:0 0 0 1px #e5e7eb inset; }
.bank-search :deep(.el-input__wrapper:hover) { box-shadow:0 0 0 1px var(--designer-accent-border) inset; }
.bank-search :deep(.el-input__wrapper.is-focus) { box-shadow:0 0 0 1px var(--designer-accent) inset; }
.bank-list { flex:1; min-height:0; overflow-y:auto; padding-right:2px; display:flex; flex-direction:column; gap:6px; }
.bank-cat { border:1px solid #eef2f7; border-radius:9px; background:#fff; overflow:hidden; }
.bank-cat-title,
.bank-type-title {
  display:flex; align-items:center; gap:7px; min-width:0; cursor:pointer;
  transition:background 0.12s, color 0.12s, border-color 0.12s;
}
.bank-cat-title { padding:7px 8px; font-size:13px; font-weight:650; color:#1f2937; background:#f8fafc; }
.bank-cat-title:hover { background:var(--designer-accent-soft); color:var(--designer-accent); }
.bank-cat-name,
.bank-type-name { flex:1; min-width:0; overflow:hidden; text-overflow:ellipsis; white-space:nowrap; }
.bank-arrow {
  width:16px; height:16px; flex-shrink:0; border-radius:5px; color:#94a3b8;
  display:flex; align-items:center; justify-content:center;
}
.bank-arrow::before {
  content:''; width:5px; height:5px; border:solid currentColor; border-width:0 1.5px 1.5px 0;
  transform:rotate(-45deg); transition:transform 0.12s; margin-left:-2px;
}
.bank-arrow.expanded::before { transform:rotate(45deg); margin-left:0; margin-top:-2px; }
.bank-count {
  flex-shrink:0; font-size:11px; color:var(--designer-accent); font-weight:600;
  background:var(--designer-accent-soft); border:1px solid var(--designer-accent-border);
  border-radius:999px; padding:1px 7px; line-height:18px;
}
.bank-cat-body { padding:4px 6px 6px; }
.bank-type-group { margin-bottom:3px; }
.bank-type-group:last-child { margin-bottom:0; }
.bank-type-title { padding:5px 6px; font-size:12px; font-weight:600; color:#475569; border-radius:7px; }
.bank-type-title:hover { background:#f8fafc; color:var(--designer-accent); }
.bank-type-title .bank-count { color:#64748b; background:#f8fafc; border-color:#e2e8f0; }
.bank-type-body {
  display:flex; flex-direction:column; gap:2px; padding:1px 0 1px 28px;
}
.bank-item {
  display:flex; align-items:center; gap:6px; min-height:30px; padding:4px 6px;
  border:1px solid transparent; border-radius:7px; background:transparent; cursor:pointer;
  transition:background 0.12s, border-color 0.12s, color 0.12s;
}
.bank-item:hover {
  background:#f0fdfa; border-color:var(--designer-accent-border);
}
.bank-icon {
  flex-shrink:0; width:16px; height:16px; border-radius:5px; background:#f1f5f9;
  color:#64748b; display:flex; align-items:center; justify-content:center; font-size:10px;
}
.bank-type-title:hover .bank-icon { background:var(--designer-accent-soft); color:var(--designer-accent); }
.bank-item-main { flex:1; min-width:0; display:flex; align-items:center; gap:6px; }
.bank-title { flex:1; min-width:0; font-size:13px; color:#1f2937; line-height:20px; overflow:hidden; text-overflow:ellipsis; white-space:nowrap; }
.bank-meta { flex-shrink:0; max-width:42px; font-size:11px; line-height:16px; color:#94a3b8; overflow:hidden; text-overflow:ellipsis; white-space:nowrap; }
.bank-type {
  flex-shrink:0; max-width:54px; overflow:hidden; text-overflow:ellipsis; white-space:nowrap;
  font-size:11px; color:var(--designer-accent); background:var(--designer-accent-soft);
  border:1px solid var(--designer-accent-border); border-radius:999px; padding:1px 6px; line-height:16px;
}
.bank-loading {
  display:flex; align-items:center; justify-content:center; gap:8px; padding:18px 0;
  color:#64748b; font-size:12px;
}
.bank-loading::before {
  content:''; width:12px; height:12px; border:2px solid var(--designer-accent-hover); border-top-color:var(--designer-accent);
  border-radius:50%; animation:bank-spin 0.8s linear infinite;
}
.bank-list :deep(.el-empty) { padding:34px 0; }
@keyframes bank-spin { to { transform:rotate(360deg); } }
</style>
<style>
.correct-answer-popper .el-select-dropdown__item { height:auto; min-height:34px; line-height:1.4; padding-top:6px; padding-bottom:6px; white-space:normal; }
</style>
