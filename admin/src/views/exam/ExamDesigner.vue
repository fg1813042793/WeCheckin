<template>
  <div class="survey-main">
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
        <button class="nav-btn" @click="goBack" title="返回列表">
          <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2"><path d="M19 12H5M12 19l-7-7 7-7"/></svg>
        </button>
      </div>
    </div>

    <div class="survey-main-content">
      <div v-show="activeView==='edit'" id="editor" class="survey-editor survey-light survey-app pc">
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
                      <question-icon :type="q.type" />
                      <span class="bank-title">{{ q.title || '未命名' }}</span>
                      <span class="bank-type">{{ q.type }}</span>
                      <el-button size="small" text type="primary" @click.stop="addFromBank(q)">+添加</el-button>
                    </div>
                    <el-empty v-if="!bankQuestions.length && !bankLoading" description="题库暂无题目" :image-size="40" />
                  </div>
                </div>
              </el-tab-pane>
              <el-tab-pane label="大纲" name="outline">
                <div class="outline-tree">
                  <div class="tree-root">
                    <span class="tree-root-title">{{ outlineRoot.title }}</span>
                    <span class="tree-root-count">{{ outlineRoot.children.length }}题</span>
                  </div>
                  <div class="tree-children">
                    <div v-for="child in outlineRoot.children" :key="child.q.id" class="tree-child" :class="{ active: child.q.id === selected?.id }" @click="selectQuestion(child.q.id)">
                      <div class="tree-child-body">
                        <span class="tree-index">{{ child.index }}.</span>
                        <question-icon :type="child.q.type" />
                        <span class="tree-title">{{ child.q.title || '未命名' }}</span>
                        <span class="tree-type">{{ child.q.type }}</span>
                      </div>
                    </div>
                  </div>
                  <el-empty v-if="!questions.length" description="暂无题目" :image-size="40" />
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
                    <el-upload :action="`/admin/exam/resource_upload`" :show-file-list="false" :on-success="handleBgSuccess" :on-error="()=>ElMessage.error('上传失败')" :headers="uploadHeaders" accept="image/*" :data="{ examId: form.id, resType: 'bg' }" :before-upload="checkSaved">
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
                    <el-upload :action="`/admin/exam/resource_upload`" :show-file-list="false" :on-success="handleHeaderSuccess" :on-error="()=>ElMessage.error('上传失败')" :headers="uploadHeaders" accept="image/*" :data="{ examId: form.id, resType: 'header' }" :before-upload="checkSaved">
                      <div class="appearance-add">+</div>
                    </el-upload>
                  </div>
                  <el-empty v-if="!allHeaderResources.length" description="暂无页眉图" :image-size="30" />
                </el-tab-pane>
              </el-tabs>
            </div>
          </div>
        </div>

        <!-- 逻辑（全宽） -->
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
                  <span class="rule-index">#{{ ri+1 }}</span>
                  <span class="rule-dsl">{{ renderRuleDSL(rule) }}</span>
                  <div class="rule-actions">
                    <el-button text size="small" @click="editRule(ri)">编辑</el-button>
                    <el-button text size="small" type="danger" @click="removeRule(ri)">删除</el-button>
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
            <div class="toolbar-actions" style="margin-left:auto">
              <el-button type="primary" :loading="saving" size="small" @click="save">
                <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" style="margin-right:4px"><path d="M19 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11l5 5v11a2 2 0 0 1-2 2z"/><polyline points="17 21 17 13 7 13 7 21"/><polyline points="7 3 7 8 15 8"/></svg>保存
              </el-button>
            </div>
          </div>
          <div class="survey-main-panel-content">
            <div class="editor-wrapper">
              <div class="editor">
                <div class="header">
                  <div class="header-content">
                    <el-input v-model="form.title" placeholder="考试标题" class="header-title-input" maxlength="100" />
                  </div>
                </div>
                <div class="questions-area">
                  <draggable-list :questions="questions" @update:questions="onQuestionsUpdate" @select="selectQuestion" :selected-id="selected?.id??null" editing @remove="removeQuestionById" @select-option="selectOption" />
                </div>
                <div class="footer">
                  <div class="footer-inner">感谢您的参与！</div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div v-show="middleTab!=='logic'" class="survey-setting-panel" @click.self="deselectQuestion">
          <div v-if="selected" class="props-panel">
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
                    <el-button text size="small" type="danger" @click="removeAutoCol(ci)">×</el-button>
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
                    <el-button text size="small" type="primary" @click="selectOption(selected.id, i)" style="font-size:11px">编辑</el-button>
                    <el-button text size="small" type="danger" @click="removeUserOpt(i)">×</el-button>
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

                <!-- 选项列表 -->
                <div v-if="hasOptions(selected)&&selected.props" class="props-options-section">
                  <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:6px">
                    <span style="font-size:12px;color:#888">选项列表</span>
                    <el-button size="small" @click="addOption" plain>+ 添加</el-button>
                  </div>
                  <div v-for="(o, oi) in (selected.props.options||[])" :key="oi" class="setting-opt-row">
                    <el-input v-model="o.label" size="small" placeholder="选项名称" style="flex:1" />
                    <el-button text size="small" type="primary" @click="selectOption(selected.id, oi)" style="font-size:11px">编辑</el-button>
                    <el-button text size="small" type="danger" @click="removeOption(oi)">×</el-button>
                  </div>
                </div>

                <el-divider border-style="dashed" style="margin:12px 0" />
                <el-collapse v-model="collapseActive" accordion>
                  <el-collapse-item title="考试属性" name="exam">
                    <el-form-item label="分值"><el-input-number v-model="selected.examScore" :min="0" :step="0.5" style="width:100%" /></el-form-item>
                    <el-form-item v-if="selected.examScore" label="正确答案"><el-input v-model="selected.examCorrectAnswer" placeholder="填写正确答案" /></el-form-item>
                    <el-form-item v-if="selected.examScore" label="答案解析"><el-input v-model="selected.examAnalysis" type="textarea" :rows="2" placeholder="选填" /></el-form-item>
                  </el-collapse-item>
                  <el-collapse-item title="校验规则" name="validate">
                    <ValidateEditor :model-value="selected.validate||[]" @update:model-value="onValidateUpdate" />
                  </el-collapse-item>
                  <el-collapse-item title="计算表达式" name="calc">
                    <CalcEditor :model-value="selected.calcValue" :env="envFromAnswers" @update:model-value="onCalcUpdate" />
                  </el-collapse-item>
                  <el-collapse-item title="显示逻辑" name="logic">
                    <LogicEditor :model-value="selected.logic||[]" :all-questions="questions" @update:model-value="onLogicUpdate" />
                  </el-collapse-item>
                </el-collapse>
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
              <el-form-item v-for="(cond, ci) in ruleForm.conditions" :key="ci" :label="ci===0?'条件':'条件 '+(ci+1)">
                <div style="display:flex;gap:6px;align-items:center;flex-wrap:wrap">
                  <el-select v-model="cond.questionIdx" placeholder="选择题目" style="width:160px" @change="cond.optionIdx=undefined;cond.operator=undefined">
                    <el-option v-for="(q, qi) in questions" :key="qi" :label="`Q${qi+1}: ${q.title?.slice(0,20)}`" :value="qi" />
                  </el-select>
                  <el-select v-model="cond.optionIdx" placeholder="选择选项" style="width:120px" v-if="hasOptionsByIndex(cond.questionIdx)" clearable @change="cond.operator='optSelected'">
                    <el-option v-for="(o, oi) in getOptionsByIndex(cond.questionIdx)" :key="oi" :label="o.label" :value="oi" />
                  </el-select>
                  <el-select v-model="cond.operator" placeholder="判断条件" style="width:120px" v-if="cond.optionIdx===undefined">
                    <el-option label="已填写" value="filled" />
                    <el-option label="未填写" value="empty" />
                    <el-option label="等于" value="eq" />
                    <el-option label="大于" value="gt" />
                    <el-option label="小于" value="lt" />
                  </el-select>
                  <el-input v-model="cond.compareValue" placeholder="比较值" style="width:100px" v-if="cond.operator==='eq'||cond.operator==='gt'||cond.operator==='lt'" />
                  <el-button v-if="ruleForm.conditions.length>1 && ci===ruleForm.conditions.length-1" text size="small" type="danger" @click="ruleForm.conditions.splice(ci,1)">✕</el-button>
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
                <el-option v-for="(q, qi) in questions" :key="qi" :label="`Q${qi+1}: ${q.title?.slice(0,30)}`" :value="qi" />
              </el-select>
            </el-form-item>
            <template v-if="ruleForm.action==='check'">
              <el-form-item label="目标题目">
                <el-select v-model="ruleForm.targetQuestionIdx" placeholder="选择题目" style="width:100%" @change="ruleForm.targetOptionIdxs=[]">
                  <el-option v-for="(q, qi) in questions" :key="qi" :label="`Q${qi+1}: ${q.title?.slice(0,30)}`" :value="qi" />
                </el-select>
              </el-form-item>
              <el-form-item v-if="ruleForm.targetQuestionIdx!==undefined" label="目标选项（可多选）">
                <el-select v-model="ruleForm.targetOptionIdxs" multiple placeholder="选择选项" style="width:100%">
                  <el-option v-for="(o, oi) in getOptionsByIndex(ruleForm.targetQuestionIdx)" :key="oi" :label="o.label" :value="oi" />
                </el-select>
              </el-form-item>
            </template>
            <template v-if="ruleForm.action==='branch'">
              <el-form-item label="从题目">
                <el-select v-model="ruleForm.branchFromIdx" placeholder="选择题目" style="width:100%" :key="'branchFrom'+showAddRule">
                  <el-option v-for="(q, qi) in questions" :key="qi" :label="`Q${qi+1}: ${q.title?.slice(0,30)}`" :value="qi" />
                </el-select>
              </el-form-item>
              <el-form-item label="跳到">
                <div style="display:flex;gap:8px;align-items:center">
                  <el-select v-model="ruleForm.branchToIdx" placeholder="选择目标题目" style="flex:1;min-width:280px" :disabled="ruleForm.branchToEnd" :key="'branchTo'+showAddRule">
                    <el-option v-for="(q, qi) in questions" :key="qi" :label="`Q${qi+1}: ${q.title?.slice(0,30)}`" :value="qi" />
                  </el-select>
                  <el-checkbox v-model="ruleForm.branchToEnd" style="white-space:nowrap">结束问卷</el-checkbox>
                </div>
              </el-form-item>
            </template>
            <el-form-item v-if="ruleForm.action==='assignment'||ruleForm.action==='validate'||ruleForm.action==='replace'" label="目标题目">
              <el-select v-model="ruleForm.targetQuestionIdx" placeholder="选择题目" style="width:100%">
                <el-option v-for="(q, qi) in questions" :key="qi" :label="`Q${qi+1}: ${q.title?.slice(0,30)}`" :value="qi" />
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
        <div class="setting-scroll">
          <div class="setting-group">
            <div class="group-title">显示设置</div>
            <el-form label-position="top">
              <div class="settings-grid">
                <el-form-item label="显示分数"><el-switch v-model="form.showScore" :active-value="1" :inactive-value="0" /></el-form-item>
                <el-form-item label="显示题号"><el-switch v-model="form.questionNumber" /></el-form-item>
                <el-form-item label="显示进度条"><el-switch v-model="form.progressBar" /></el-form-item>
                <el-form-item label="自动暂存"><el-switch v-model="form.autoSave" /></el-form-item>
                <el-form-item label="显示排名"><el-switch v-model="form.examRankingEnabled" /></el-form-item>
                <el-form-item label="练习模式"><el-switch v-model="form.exerciseMode" /></el-form-item>
                <el-form-item label="随机顺序"><el-switch v-model="form.randomOrder" /></el-form-item>
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
            <div class="group-title">考试设置</div>
            <el-form label-position="top">
              <div class="settings-grid">
                <el-form-item label="答题时长(分钟)"><el-input-number v-model="form.duration" :min="1" style="width:100%" /></el-form-item>
                <el-form-item label="最短交卷(分钟)"><el-input-number v-model="form.minSubmitMinutes" :min="0" style="width:100%" /></el-form-item>
                <el-form-item label="最长交卷(分钟)"><el-input-number v-model="form.maxSubmitMinutes" :min="0" style="width:100%" /></el-form-item>
                <el-form-item label="填写密码"><el-input v-model="form.password" placeholder="留空不设密码" /></el-form-item>
                <el-form-item label="每台设备答题次数"><el-input-number v-model="form.deviceLimit" :min="0" style="width:100%" /><span style="font-size:11px;color:#999;margin-left:4px">0=不限</span></el-form-item>
                <el-form-item label="每个IP答题次数"><el-input-number v-model="form.ipLimit" :min="0" style="width:100%" /><span style="font-size:11px;color:#999;margin-left:4px">0=不限</span></el-form-item>
                <el-form-item label="每个账号答题次数"><el-input-number v-model="form.userLimit" :min="0" style="width:100%" /><span style="font-size:11px;color:#999;margin-left:4px">0=不限</span></el-form-item>
                <el-form-item label="开始时间"><el-date-picker v-model="form.startDate" type="datetime" placeholder="不限" value-format="x" style="width:100%" /></el-form-item>
                <el-form-item label="结束时间"><el-date-picker v-model="form.endDate" type="datetime" placeholder="不限" value-format="x" style="width:100%" /></el-form-item>
              </div>
            </el-form>
          </div>
          <div class="setting-group">
            <div class="group-title">提交后设置</div>
            <el-form label-position="top">
              <div class="settings-grid">
                <el-form-item label="允许改答案"><el-switch v-model="form.enableUpdate" /></el-form-item>
                <el-form-item label="显示成绩单"><el-switch v-model="form.transcriptVisible" /></el-form-item>
                <el-form-item label="显示排行榜"><el-switch v-model="form.rankVisible" /></el-form-item>
                <el-form-item label="最大答题次数"><el-input-number v-model="form.maxAttempts" :min="1" :max="99" style="width:100%" /></el-form-item>
                <el-form-item label="最大答卷数"><el-input-number v-model="form.maxResponse" :min="0" style="width:100%" /><span style="font-size:11px;color:#999;margin-left:4px">0=不限</span></el-form-item>
                <el-form-item label="结束语"><el-input v-model="form.endContent" type="textarea" :rows="3" placeholder="提交后显示的 HTML 内容" /></el-form-item>
              </div>
            </el-form>
          </div>
          <div class="setting-group">
            <div class="group-title">访问权限</div>
            <el-form label-position="top">
              <div class="settings-grid">
                <el-form-item class="settings-full" label="可见性">
                  <el-radio-group v-model="form.visibility" style="display:flex;flex-direction:column;gap:8px;align-items:flex-start">
                    <el-radio :value="0" border>公开链接</el-radio>
                    <el-radio :value="1" border>登录可见</el-radio>
                    <el-radio :value="2" border>部门限定</el-radio>
                  </el-radio-group>
                </el-form-item>
                <el-form-item v-if="form.visibility===2" label="限定部门"><el-input v-model="form.deptIds" placeholder="部门ID, 逗号分隔" /></el-form-item>
                <el-form-item label="需登录"><el-switch v-model="form.loginRequired" /></el-form-item>
                <el-form-item label="多次作答"><el-switch v-model="form.allowMultiBool" :active-value="1" :inactive-value="0" /></el-form-item>
                <el-form-item label="匿名收集"><el-switch v-model="form.anonymousBool" :active-value="1" :inactive-value="0" /></el-form-item>
                <el-form-item label="显示结果"><el-switch v-model="form.showResultBool" :active-value="1" :inactive-value="0" /></el-form-item>
              </div>
            </el-form>
          </div>
        </div>
      </div>

      <div v-show="activeView==='overview'" class="overview-wrapper">
        <div style="padding:24px">
          <div class="setting-group">
            <div style="font-size:14px;font-weight:600;margin-bottom:12px">考试概览</div>
            <div style="display:flex;gap:24px;flex-wrap:wrap">
              <div><span style="color:#888">考试名称：</span>{{ form.title || '未命名' }}</div>
              <div><span style="color:#888">题目数量：</span>{{ questions.length }}题</div>
              <div><span style="color:#888">答题时长：</span>{{ form.duration }}分钟</div>
              <div><span style="color:#888">最大次数：</span>{{ form.maxAttempts }}次</div>
              <div><span style="color:#888">状态：</span>{{ form.statusBool ? '启用' : '停用' }}</div>
            </div>
          </div>
          <div class="setting-group" style="margin-top:16px">
            <div class="overview-card-title">基础信息</div>
            <div style="margin-top:8px">
              <el-form :model="form" label-position="top">
                <div class="settings-grid">
                  <el-form-item label="描述"><el-input v-model="form.description" type="textarea" :rows="3" /></el-form-item>
                  <el-form-item label="分类"><el-input v-model="form.category" /></el-form-item>
                  <el-form-item label="标签"><el-input v-model="form.tags" placeholder="逗号分隔" /></el-form-item>
                </div>
              </el-form>
            </div>
          </div>
          <div class="setting-group" style="margin-top:16px">
            <div class="overview-card-title">访问链接</div>
            <div style="margin-top:8px">
              <el-input v-if="form.id" v-model="publicUrl" readonly size="small">
                <template #append><el-button @click="copyLink" size="small">复制</el-button></template>
              </el-input>
              <span v-else style="color:#999;font-size:13px">请先保存考试</span>
            </div>
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
</template>

<script setup lang="ts">
import { ref, computed, reactive, watch, onMounted, nextTick } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useRoute, useRouter } from 'vue-router'
import { adminApi } from '../../api'
import DraggableList from './formkit/DraggableList.vue'
import ValidateEditor from './formkit/ValidateEditor.vue'
import CalcEditor from './formkit/CalcEditor.vue'
import LogicEditor from './formkit/LogicEditor.vue'
import QuestionIcon from './formkit/QuestionIcon.vue'

const route = useRoute()
const router = useRouter()

const form = reactive<any>({
  id: 0, title: '', description: '', category: '', tags: '',
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
  deviceLimit: 0, ipLimit: 0, userLimit: 0,
  backgroundImages: [] as string[], headerImages: [] as string[]
})

const activeView = ref('edit')
const middleTab = ref('item')
const sideSubTab = ref('types')
const activeCategory = ref('')
const saving = ref(false)
const bankKeyword = ref('')
const bankQuestions = ref<any[]>([])
const bankLoading = ref(false)
let bankTimer: any = null

const appearanceTab = ref('bg')
const allBgResources = ref<any[]>([])
const allHeaderResources = ref<any[]>([])

const token = localStorage.getItem('admin_token') || ''
const uploadHeaders = { Authorization: token }
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
interface LogicRuleItem { id: string; conditionType: string; conditions: LogicCondition[]; action: string; targetQuestionIdx?: number; targetOptionIdxs?: number[]; branchFromIdx?: number; branchToIdx?: number; branchToEnd?: boolean; formula?: string }
const logicRuleList = ref<LogicRuleItem[]>([])
const showAddRule = ref(false)
const editingRuleIdx = ref(-1)
const defaultRuleForm = (): LogicRuleItem => ({ id: '', conditionType: 'simple', conditions: [{ questionIdx: undefined, optionIdx: undefined, operator: undefined, compareValue: undefined }], action: 'show', targetQuestionIdx: undefined, targetOptionIdxs: [], branchFromIdx: undefined, branchToIdx: undefined, branchToEnd: false, formula: '' })
const ruleForm = ref<LogicRuleItem>(defaultRuleForm())

watch(logicRuleList, () => { syncLogicRules() }, { deep: true })

function syncLogicRules() {
  if (!selected.value) return
  if (!selected.value.props) selected.value.props = {}
  selected.value.props.logicRules = generateLogicDSL()
}

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

function generateLogicDSL(): string {
  return logicRuleList.value.map(r => renderRuleDSL(r)).join('\n')
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
  rf.id = 'rule_' + Date.now() + '_' + Math.random().toString(36).slice(2,6)
  if (editingRuleIdx.value>=0) {
    logicRuleList.value[editingRuleIdx.value] = { ...rf }
  } else {
    logicRuleList.value.push({ ...rf })
  }
  syncLogicRules()
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
  syncLogicRules()
}

async function saveLogicRules() {
  if (!form.id) { ElMessage.warning('请先保存考试'); return }
  syncLogicRules()
  const dsl = generateLogicDSL()
  const schema = JSON.stringify({ version: '2.0', questions: questions.value, setting: { logicRules: dsl } })
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
const collapseActive = ref('')

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

interface Question { id: string; type: string; title: string; required: boolean; placeholder?: string; props?: any; validate?: any[]; logic?: any[]; calcValue?: any; readOnly?: boolean; showDescription?: boolean; defaultHidden?: boolean; optionLayout?: number; multiple?: boolean; examScore?: number; examCorrectAnswer?: string; examAnalysis?: string; fileTypes?: string[]; fileExtensions?: string; maxFileSize?: number; maxFileCount?: number; examAnswerMode?: string; dataType?: string; _existing?: boolean; [key: string]: any }
const questions = ref<Question[]>([])
const selected = ref<Question | null>(null)
let idCounter = 0
function genId() { idCounter++; return 'q' + idCounter }
function addQuestion(t: any) {
  const q: Question = { id: genId(), type: t.type, title: t.displayName, required: false, readOnly: false, placeholder: '', props: t.defaultProps ? JSON.parse(JSON.stringify(t.defaultProps)) : {} }
  if (!q.props) q.props = {}
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
function selectQuestion(id: string) { selected.value = questions.value.find(q => q.id === id) || null; selectedOptIdx.value = -1 }
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
  selected.value.props.options.push({ value: '选项' + (selected.value.props.options.length + 1), label: '选项' + (selected.value.props.options.length + 1) })
}
function removeOption(idx: number) { if (selected.value?.props?.options) selected.value.props.options.splice(idx, 1) }
function onValidateUpdate(v: any) { if (selected.value) selected.value.validate = v }
function removeQuestionById(id: string) { questions.value = questions.value.filter(q => q.id !== id); if (selected.value?.id === id) selected.value = null }
function onCalcUpdate(v: any) { if (selected.value) selected.value.calcValue = v }
function onLogicUpdate(v: any) { if (selected.value) selected.value.logic = v }

const envFromAnswers = computed(() => {
  const env: Record<string, any> = {}; for (const q of questions.value) env[q.id] = undefined; return env
})

const publicUrl = computed(() => form.id ? `${window.location.origin}/#/ef/${form.id}` : '')
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
      duration: sv.duration ?? 60, maxAttempts: sv.maxAttempts ?? 1, showScore: sv.showScore ?? 1
    })
    if (sv.settings) {
      try {
        const s = JSON.parse(sv.settings)
        Object.assign(form, {
          questionNumber: s.questionNumber ?? true, progressBar: s.progressBar ?? false,
          autoSave: s.autoSave ?? false, password: s.password || '',
          loginRequired: s.loginRequired ?? false,
          onePageOneQuestion: s.onePageOneQuestion ?? false,
          answerSheetVisible: s.answerSheetVisible ?? true,
          copyEnabled: s.copyEnabled ?? true, triggerType: s.triggerType || 'onBlur',
          enableUpdate: s.enableUpdate ?? false, transcriptVisible: s.transcriptVisible ?? true,
          rankVisible: s.rankVisible ?? false, redirectUrl: s.redirectUrl || '', endContent: s.endContent || '',
          examRankingEnabled: s.examRankingEnabled ?? false,
          exerciseMode: s.exerciseMode ?? false, randomOrder: s.randomOrder ?? false,
          minSubmitMinutes: s.minSubmitMinutes || 0, maxSubmitMinutes: s.maxSubmitMinutes || 0,
          deviceLimit: s.deviceLimit || 0, ipLimit: s.ipLimit || 0, userLimit: s.userLimit || 0,
          backgroundImages: s.backgroundImages || [], headerImages: s.headerImages || []
        })
      } catch {}
    }
    const rawSchema = res.data?.schema || ''
    if (rawSchema) {
      try {
        const sch = JSON.parse(rawSchema)
        if (sch.questions) { questions.value = sch.questions; idCounter = questions.value.length; selected.value = questions.value[0] || null }
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
    deviceLimit: form.deviceLimit, ipLimit: form.ipLimit, userLimit: form.userLimit,
    backgroundImages: form.backgroundImages, headerImages: form.headerImages
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
    ElMessage.success(form.id ? '已更新' : '已创建')
  } catch { ElMessage.error('保存失败') }
  finally { saving.value = false }
}

watch(middleTab, (v) => { if (v === 'appearance') loadResources() })
watch(sideSubTab, (v) => { if (v === 'bank') loadBank() })

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
  await nextTick(); updateTabScroll(); load()
})
</script>

<style scoped>
.survey-main { display:flex; height:calc(100vh - 56px); background:#fff; }
.survey-main-navigator { display:flex; flex:0 0 54px; flex-direction:column; justify-content:space-between; border-right:1px solid #e8e8e8; background:#fafafa; padding:8px 0; align-items:center; }
.nav-actions { display:flex; flex-direction:column; align-items:center; gap:4px; }
.nav-btn { width:36px; height:36px; min-width:36px; padding:0; margin:0; border:none; border-radius:8px; color:#999; font-size:13px; cursor:pointer; display:flex; align-items:center; justify-content:center; background:transparent; outline:none; }
.nav-btn:hover { background:#f0f0f0; color:#666; }
.nav-btn.active { background:#fff; color:#fb454c; box-shadow:0 1px 4px rgba(0,0,0,0.08); }
.survey-main-content { flex:1; display:flex; flex-direction:column; overflow:hidden; }
.survey-editor { display:flex; height:100%; }
.survey-sidebar-panel { width:280px; flex-shrink:0; border-right:1px solid #e8e8e8; background:#fafafa; display:flex; flex-direction:row; overflow:hidden; }
.survey-sidebar-panel-tabs { flex:0 0 48px; display:flex; flex-direction:column; border-right:1px solid #e8e8e8; background:#f5f5f5; padding:4px 0; }
.survey-sidebar-panel-tabs-pane { display:flex; flex-direction:column; align-items:center; justify-content:center; gap:2px; padding:10px 2px; cursor:pointer; color:#999; font-size:10px; }
.survey-sidebar-panel-tabs-pane:hover { color:#fb454c; }
.survey-sidebar-panel-tabs-pane.active { color:#fb454c; background:#fff; }
.survey-sidebar-panel-tabs-pane .tab-label { color:inherit; line-height:1.2; }
.survey-sidebar-panel-tabs-content { flex:1; display:flex; flex-direction:column; overflow:hidden; }
.side-sub-tabs { display:flex; flex-direction:column; flex:1; overflow:hidden; }
.side-sub-tabs :deep(.el-tabs__header) { margin:0; padding:0 8px; background:#fafafa; flex-shrink:0; }
.side-sub-tabs :deep(.el-tabs__active-bar) { background:#fb454c; }
.side-sub-tabs :deep(.el-tabs__content) { flex:1; overflow:hidden; }
.side-sub-tabs :deep(.el-tab-pane) { height:100%; overflow-y:auto; }
.question-panel { padding:8px; }
.type-tabs-bar { display:flex; align-items:center; gap:4px; margin-bottom:8px; }
.tab-scroll-btn { width:20px; height:24px; border:1px solid #e8e8e8; border-radius:4px; background:#fff; cursor:pointer; font-size:12px; color:#666; flex-shrink:0; display:flex; align-items:center; justify-content:center; }
.tab-scroll-btn:disabled { opacity:0.3; cursor:default; }
.tab-scroll-viewport { flex:1; overflow:hidden; }
.tab-scroll-track { display:flex; gap:4px; transition:transform 0.2s; }
.type-tab-btn { padding:2px 10px; border:1px solid #e8e8e8; border-radius:4px; background:#fff; cursor:pointer; font-size:12px; color:#666; white-space:nowrap; }
.type-tab-btn.active { color:#fff; background:#fb454c; border-color:#fb454c; }
.type-tab-btn:hover { border-color:#fb454c; color:#fb454c; }
.menu-group { margin:0; }
.menu-group-title { font-size:11px; color:#bbb; padding:4px 0; }
.menu-group-item { display:flex; align-items:center; gap:6px; padding:6px 8px; cursor:pointer; border-radius:4px; font-size:12px; }
.menu-group-item:hover { background:#fff5f5; color:#fb454c; }
.itemIcon { flex-shrink:0; font-size:16px; width:20px; text-align:center; color:#999; }
.menu-group-item:hover .itemIcon { color:#fb454c; }
.item-label { font-size:13px; flex:1; }
.item-type { font-size:10px; color:#bbb; }
.bank-panel { padding:8px; display:flex; flex-direction:column; height:100%; }
.bank-search { margin-bottom:8px; }
.bank-list { flex:1; overflow-y:auto; }
.bank-item { display:flex; align-items:center; gap:6px; padding:6px 8px; cursor:pointer; border-radius:4px; font-size:12px; }
.bank-item:hover { background:#fff5f5; color:#fb454c; }
.bank-icon { flex-shrink:0; font-size:14px; width:18px; text-align:center; color:#999; }
.bank-item:hover .bank-icon { color:#fb454c; }
.bank-index { font-size:12px; color:#bbb; min-width:20px; }
.bank-title { font-size:13px; flex:1; overflow:hidden; text-overflow:ellipsis; white-space:nowrap; }
.bank-type { font-size:10px; color:#bbb; }
.outline-tree { padding:8px; }
.tree-root { display:flex; align-items:center; gap:6px; padding:8px; font-size:13px; font-weight:600; }
.tree-children { padding-left:12px; }
.tree-child { display:flex; align-items:center; padding:6px 8px; cursor:pointer; border-radius:4px; font-size:12px; }
.tree-child:hover { background:#fff5f5; }
.tree-child.active { background:#fff0f0; color:#fb454c; }
.tree-child-body { display:flex; align-items:center; gap:4px; flex:1; overflow:hidden; }
.tree-index { color:#999; flex-shrink:0; }
.tree-title { flex:1; overflow:hidden; text-overflow:ellipsis; white-space:nowrap; }
.tree-type { font-size:10px; color:#bbb; padding:1px 5px; border-radius:3px; background:#f5f5f5; }
.survey-main-panel { flex:1; display:flex; flex-direction:column; overflow:hidden; background-color:#f7f8fa; }
.survey-main-panel-toolbar { display:flex; align-items:center; justify-content:space-between; padding:8px 16px; background:#fff; border-bottom:1px solid #e8e8e8; gap:12px; }
.toolbar-actions { display:flex; align-items:center; gap:8px; flex-shrink:0; }
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
.props-panel { padding:12px; }
.props-panel h3 { font-size:14px; font-weight:500; color:#fb454c; margin:0 0 8px; padding-bottom:8px; border-bottom:2px solid #fb454c; }
.props-panel :deep(.el-form-item) { margin-bottom:8px; }
.props-panel :deep(.el-form-item__label) { font-size:12px; color:#666; padding-bottom:2px; font-weight:500; line-height:1.2; }
.props-panel :deep(.el-divider) { margin:8px 0; }
.props-panel :deep(.el-collapse-item__header) { font-size:12px; font-weight:500; }
.props-options-section { margin-bottom:8px; padding:8px; background:#f5f6f8; border-radius:6px; }
.setting-opt-row { display:flex; align-items:center; gap:6px; margin-bottom:4px; }
.setting-wrapper { height:100%; overflow-y:auto; background:#f5f6f8; }
.setting-scroll { display:grid; grid-template-columns:repeat(auto-fill, minmax(420px, 1fr)); gap:20px; padding:20px 24px; align-items:start; }
.setting-group { padding:16px 20px; background:#fff; border-radius:8px; box-shadow:0 1px 4px rgba(0,0,0,0.04); }
.group-title { font-size:13px; font-weight:500; color:#888; margin-bottom:12px; letter-spacing:0.5px; }
.overview-wrapper { height:100%; overflow-y:auto; background:#f5f6f8; padding:20px 24px; }
.overview-card-title { font-size:14px; font-weight:600; color:#303133; margin-bottom:16px; }

/* 外观面板 */
.appearance-panel { padding:12px; }
.appearance-panel :deep(.el-tabs__header) { margin:0 0 8px; }
.appearance-panel :deep(.el-tabs__item) { font-size:12px; padding:0 12px; }
.appearance-grid { display:flex; flex-wrap:wrap; gap:8px; }
.appearance-add, .appearance-thumb { width:72px; height:72px; border-radius:6px; overflow:hidden; flex-shrink:0; }
.appearance-add { display:flex; align-items:center; justify-content:center; border:2px dashed #e8e8e8; background:#fff; cursor:pointer; font-size:24px; color:#ccc; transition:all .15s; }
.appearance-add:hover { border-color:#fb454c; color:#fb454c; }
.appearance-thumb { position:relative; cursor:pointer; border:2px solid transparent; transition:border-color .15s; }
.appearance-thumb:hover { border-color:#fb454c; }
.appearance-thumb.active { border-color:#fb454c; background:#fff5f5; }
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
:deep(.el-table .cell) { overflow:hidden; text-overflow:ellipsis; white-space:nowrap; }
:deep(.el-table th.el-table__cell > .cell) { overflow:hidden; text-overflow:ellipsis; white-space:nowrap; }
</style>
