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
          <div class="survey-sidebar-panel-tabs-content" v-show="middleTab!=='logic'">
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
                      <div v-for="(child, ci) in outlineRoot.children" :key="child.q.id" class="tree-child" :class="{ active: child.q.id === selected?.id }" @click="selectQuestion(child.q.id)">
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
                    <div v-for="(item, i) in allBgResources" :key="i" class="appearance-thumb" :class="{ active: isActiveImg('backgroundImages', item) }" @click="applyImage('backgroundImages', item)">
                      <img :src="(item.domain||'')+item.url" />
                      <div class="appearance-overlay"><span>点击应用</span></div>
                      <button class="appearance-remove" @click.stop="removeResource('backgroundImages', item, i)">✕</button>
                    </div>
                    <el-upload :action="`/admin/survey/resource_upload`" :show-file-list="false" :on-success="handleBgSuccess" :on-error="()=>ElMessage.error('上传失败')" :headers="uploadHeaders" accept="image/*" :data="{ surveyId: form.id, resType: 'bg' }" :before-upload="checkSaved">
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
                    <el-upload :action="`/admin/survey/resource_upload`" :show-file-list="false" :on-success="handleHeaderSuccess" :on-error="()=>ElMessage.error('上传失败')" :headers="uploadHeaders" accept="image/*" :data="{ surveyId: form.id, resType: 'header' }" :before-upload="checkSaved">
                      <div class="appearance-add">+</div>
                    </el-upload>
                  </div>
                  <el-empty v-if="!allHeaderResources.length" description="暂无页眉图" :image-size="30" />
                </el-tab-pane>
              </el-tabs>
            </div>
          </div>
        </div>

        <!-- 逻辑（全宽，不放在侧栏内） -->
        <div v-if="middleTab==='logic'" class="logic-panel logic-full-panel">
          <div class="logic-toolbar">
            <h4 style="margin:0;font-size:14px">自定义逻辑</h4>
            <div style="display:flex;gap:8px">
              <el-button size="small" type="primary" @click="showAddRule = true">+ 增加规则</el-button>
              <el-button size="small" @click="saveLogicRules">保存全部规则</el-button>
            </div>
          </div>
          <!-- 使用说明 -->
          <div style="font-size:12px;color:#606266;background:#fefced;border:1px solid #faf0c7;border-radius:4px;padding:8px 12px;margin-bottom:12px;line-height:1.8">
            <strong>使用说明：</strong>每一条规则由「条件 + 动作 + 目标」组成。例如：<code>如果回答了Q1第1选项，则显示Q2</code>。
            条件可组合多个（且/或/非），动作可选显示隐藏、必填、跳转等。点击下方「+ 增加规则」添加。
          </div>
          <div class="logic-body">
            <!-- 规则列表 -->
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
            <!-- 右侧示例 -->
            <div class="logic-sidebar">
              <div style="font-size:12px;font-weight:600;color:#606266;margin-bottom:6px">📖 案例参考</div>
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
                  <!-- 选项选择（仅选择题） -->
                  <el-select v-model="cond.optionIdx" placeholder="选择选项" style="width:120px" v-if="hasOptionsByIndex(cond.questionIdx)" clearable @change="cond.operator='optSelected'">
                    <el-option v-for="(o, oi) in getOptionsByIndex(cond.questionIdx)" :key="oi" :label="o.label" :value="oi" />
                  </el-select>
                  <!-- 判断条件（所有题型） -->
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

            <!-- 目标选择 (SHOW/HIDE/REQUIRED/CHECK 需要目标题目) -->
            <el-form-item v-if="['show','hide','required'].includes(ruleForm.action)" label="目标题目">
              <el-select v-model="ruleForm.targetQuestionIdx" placeholder="选择题目" style="width:100%">
                <el-option v-for="(q, qi) in questions" :key="qi" :label="`Q${qi+1}: ${q.title?.slice(0,30)}`" :value="qi" />
              </el-select>
            </el-form-item>

            <!-- CHECK 需要目标题目+选项 -->
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

            <!-- BRANCH -->
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

            <!-- ASSIGNMENT / VALIDATE / REPLACE 需要公式输入 -->
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

            <!-- END 不需要额外设置 -->
          </el-form>
          <template #footer>
            <el-button @click="showAddRule=false">取消</el-button>
            <el-button type="primary" @click="confirmRule">确定</el-button>
          </template>
        </el-dialog>

        <!-- 中: 画布 -->
        <div v-show="middleTab!=='logic'" class="survey-main-panel" @click.self="deselectQuestion">
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
                  <draggable-list :questions="questions" @update:questions="onQuestionsUpdate" @select="selectQuestion" :selected-id="selected?.id??null" editing @remove="removeQuestionById" @select-option="selectOption" />
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

        <!-- 右: 属性面板 -->
        <div class="survey-setting-panel" v-show="middleTab!=='logic'" @click.self="deselectQuestion">
          <div v-if="selected" class="props-panel">
            <!-- 矩阵行设置模式 -->
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

                <!-- 选择题选项属性 -->
                <template v-if="hasOptions(selected)">
                  <el-form-item label="选项配额"><el-input-number v-model="selected.props.options[selectedOptIdx].quota" :min="0" style="width:100%" placeholder="0=不限制" /></el-form-item>
                  <div class="setting-row"><span class="setting-label">显示隐藏公式</span><el-button text size="small" @click="openFormulaDialog('optShow')">点击设置</el-button></div>
                  <div class="setting-row"><span class="setting-label">自动勾选公式</span><el-button text size="small" @click="openFormulaDialog('optAutoCheck')">点击设置</el-button></div>
                </template>

                <!-- 输入类题型子字段属性 -->
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

            <!-- 评分/NPS 设置（独立面板，点击评分输入时弹出） -->
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

              <!-- ==== 通用基础字段 ==== -->
              <el-form-item label="ID"><el-input v-model="selected.id" :disabled="!!selected._existing" /></el-form-item>

              <el-row :gutter="8" v-if="!isPureLayout(selected.type)&&selected.type!=='questionSet'">
                <el-col :span="12"><el-form-item label="必填"><el-switch v-model="selected.required" /></el-form-item></el-col>
                <el-col :span="12"><el-form-item label="只读"><el-switch v-model="selected.readOnly" /></el-form-item></el-col>
              </el-row>
              <el-form-item v-if="!isPureLayout(selected.type)&&selected.type!=='questionSet'" label="说明">
                <el-switch v-model="selected.showDescription" />
              </el-form-item>

              <!-- ==== 公式链接 (仅非布局题) ==== -->
              <template v-if="!isPureLayout(selected.type)">
                <div class="setting-row">
                  <span class="setting-label">结束公式</span><el-button text size="small" @click="openFormulaDialog('endFormula')">点击设置</el-button>
                </div>
                <div class="setting-row">
                  <span class="setting-label">跳转公式</span><el-button text size="small" @click="openFormulaDialog('jumpFormula')">点击设置</el-button>
                </div>
                <el-row :gutter="8">
                  <el-col :span="12"><el-form-item label="默认隐藏"><el-switch v-model="selected.defaultHidden" /></el-form-item></el-col>
                  <el-col :span="12" v-if="selected.type==='user'||selected.type==='dept'"><el-form-item label="多选"><el-switch v-model="selected.multiple" /></el-form-item></el-col>
                  <el-col :span="12" v-if="hasOptions(selected) && !isMatrixAll(selected.type)"><el-form-item label="选项布局"><el-input-number v-model="selected.optionLayout" :min="1" :max="6" :step="1" style="width:100%" /></el-form-item></el-col>
                </el-row>
                <div class="setting-row">
                  <span class="setting-label">必填公式</span><el-button text size="small" @click="openFormulaDialog('requiredFormula')">点击设置</el-button>
                </div>
                <div class="setting-row">
                  <span class="setting-label">文本替换</span><el-button text size="small" @click="openFormulaDialog('textReplace')">点击设置</el-button>
                </div>
                <div class="setting-row" v-if="['input','textarea','number','date','time','dateRange','calculate'].includes(selected.type)||isPersonal(selected.type)">
                  <span class="setting-label">计算公式</span><el-button text size="small" @click="openFormulaDialog('calculate')">点击设置</el-button>
                </div>
                <el-form-item v-if="isInput(selected.type)||isPersonal(selected.type)||selected.type==='signature'||selected.type==='scanCode'" label="占位提示">
                  <el-input v-model="selected.placeholder" />
                </el-form-item>
                <el-form-item v-if="selected.type==='richText'" label="占位提示">
                  <el-input v-model="selected.placeholder" placeholder="输入富文本编辑器占位提示" />
                </el-form-item>
              </template>

              <!-- 纯布局题 -->
              <template v-else-if="selected.type==='description'||selected.type==='questionSet'">
                <div class="setting-row">
                  <span class="setting-label">文本替换</span><el-button text size="small" @click="openFormulaDialog('textReplace')">点击设置</el-button>
                </div>
                <el-form-item v-if="selected.type==='description'" label="文本替换规则">
                  <el-input v-model="selected.props.replaceTextRule" type="textarea" :rows="2" placeholder="@公式" />
                </el-form-item>
              </template>

              <!-- ==== 表格自增列编辑 ==== -->
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

              <!-- ==== 多项/横向填空子字段 ==== -->
              <div v-if="(selected.type==='multiInput'||selected.type==='hInput')&&selected.props" class="props-options-section">
                <h4 style="font-size:12px;color:#888;margin:0 0 6px">子字段列表</h4>
                <div v-for="(f, fi) in (selected.props.fields||[])" :key="fi" class="setting-opt-row">
                  <el-input v-model="f.label" size="small" placeholder="字段名" style="flex:1" />
                  <el-input v-model="f.placeholder" size="small" placeholder="提示语" style="width:110px" />
                  <el-select v-model="f.dataType" size="small" style="width:80px" placeholder="类型">
                    <el-option label="文本" value="" />
                    <el-option label="数字" value="number" />
                    <el-option label="手机" value="mobile" />
                    <el-option label="邮箱" value="email" />
                  </el-select>
                  <el-button text size="small" type="danger" @click="selected.props.fields.splice(fi,1)">×</el-button>
                </div>
                <el-button size="small" @click="selected.props.fields.push({label:'填空'+(selected.props.fields.length+1),placeholder:''})" plain>+ 添加字段</el-button>
              </div>

              <!-- ==== 数字范围 (number 类型) ==== -->
              <div v-if="selected.type==='number'&&selected.props" class="props-options-section">
                <h4 style="font-size:12px;color:#888;margin:0 0 6px">数值范围</h4>
                <el-row :gutter="8">
                  <el-col :span="12"><el-form-item label="最小值"><el-input-number v-model="selected.props.min" :step="1" style="width:100%" /></el-form-item></el-col>
                  <el-col :span="12"><el-form-item label="最大值"><el-input-number v-model="selected.props.max" :step="1" style="width:100%" /></el-form-item></el-col>
                </el-row>
                <el-form-item label="小数位数"><el-input-number v-model="selected.props.decimalPlaces" :min="0" :max="6" style="width:100%" /></el-form-item>
              </div>

              <!-- ==== 分页设置 ==== -->
              <div v-if="selected.type==='pagination'" class="props-options-section">
                <h4 style="font-size:12px;color:#888;margin:0 0 6px">分页</h4>
                <el-row :gutter="8">
                  <el-col :span="12"><el-form-item label="当前页"><el-input-number v-model="selected.props.currentPage" :min="1" style="width:100%" /></el-form-item></el-col>
                  <el-col :span="12"><el-form-item label="总页数"><el-input-number v-model="selected.props.totalPage" :min="1" style="width:100%" /></el-form-item></el-col>
                </el-row>
              </div>

              <!-- ==== 成员/部门列表 ==== -->
              <div v-if="(selected.type==='user'||selected.type==='dept')&&selected.props" class="props-options-section">
                <div style="display:flex;align-items:center;gap:4px;margin-bottom:6px">
                  <h4 style="font-size:12px;color:#888;margin:0">{{ selected.type==='user'?'成员':'部门' }}列表</h4>
                  <el-tooltip :content="importFormatHint" placement="right">
                    <el-button text size="small" style="font-size:11px;color:#999;padding:0 4px">格式</el-button>
                  </el-tooltip>
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

              <!-- ==== 签名考试属性 ==== -->
              <el-form-item v-if="selected.type==='signature'" label="考试答案模式">
                <el-select v-model="selected.examAnswerMode" clearable placeholder="不限制" style="width:100%">
                  <el-option label="不限制" value="" />
                  <el-option label="无" value="none" />
                  <el-option label="文字" value="text" />
                  <el-option label="图片" value="image" />
                </el-select>
              </el-form-item>

              <!-- ==== 数据类型 (输入类题型) ==== -->
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

              <!-- ==== 高级设置 ==== -->
              <el-collapse v-if="!isPureLayout(selected.type)&&form.mode==='exam'" v-model="collapseActive" accordion>
                <el-collapse-item title="考试属性" name="exam">
                  <el-form-item label="分值"><el-input-number v-model="selected.examScore" :min="0" :step="0.5" style="width:100%" /></el-form-item>
                  <el-form-item v-if="selected.examScore" label="正确答案"><el-input v-model="selected.examCorrectAnswer" placeholder="填写正确答案" /></el-form-item>
                  <el-form-item v-if="selected.examScore" label="答案解析"><el-input v-model="selected.examAnalysis" type="textarea" :rows="2" placeholder="选填" /></el-form-item>
                </el-collapse-item>
              </el-collapse>
            </el-form>
            <div class="props-footer">
              <el-button type="danger" text style="width:100%" @click="removeSelected">删除此题</el-button>
            </div>
            </template>
          </div>
          <div v-else-if="showSurveySettings" class="props-panel">
            <h3>问卷设置</h3>
            <div class="setting-row"><span class="setting-label">标题</span><el-input v-model="form.title" placeholder="问卷标题" style="flex:1" /></div>
            <div class="setting-row" style="align-items:flex-start"><span class="setting-label" style="margin-top:4px">描述</span><el-input v-model="form.description" type="textarea" :autosize="{ minRows: 3, maxRows: 8 }" placeholder="问卷描述" style="flex:1" /></div>
            <div class="setting-row"><span class="setting-label">模式</span><span>{{ form.mode === 'exam' ? '考试' : '问卷' }}</span></div>
            <el-divider />
            <div class="group-title" style="font-size:13px;font-weight:500;margin-bottom:8px">已应用图片</div>
            <div v-for="(item, i) in form.backgroundImages" :key="'bg-'+i" class="applied-image-row">
              <img :src="typeof item === 'string' ? item : item.url" class="applied-image-thumb" />
              <span class="applied-image-label">背景图</span>
              <button class="applied-image-remove" @click="removeResource('backgroundImages', item, 0)">✕</button>
            </div>
            <div v-for="(item, i) in form.headerImages" :key="'hd-'+i" class="applied-image-row">
              <img :src="typeof item === 'string' ? item : item.url" class="applied-image-thumb" />
              <span class="applied-image-label">页眉图</span>
              <button class="applied-image-remove" @click="removeResource('headerImages', item, 0)">✕</button>
            </div>
            <el-empty v-if="!form.backgroundImages.length && !form.headerImages.length" description="暂无已应用图片" :image-size="30" />
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
            <div v-if="!form.id" class="save-tip">请先保存并发布问卷</div>
            <div v-else style="padding:12px">
              <el-row :gutter="12" style="margin-bottom:12px">
                <el-col :span="6"><el-statistic title="总提交数" :value="stats.totalCount" /></el-col>
                <el-col :span="6"><el-statistic title="今日提交" :value="stats.todayCount" /></el-col>
                <el-col :span="6"><el-statistic title="平均完成时间" :value="stats.avgTime" suffix="秒" /></el-col>
                <el-col :span="6"><el-statistic title="完成率" :value="stats.completionRate" suffix="%" /></el-col>
              </el-row>
              <el-table :data="dummyData" border size="small" style="width:100%" max-height="400px">
                <el-table-column type="index" label="#" width="50" />
                <el-table-column prop="submitter" label="提交人" width="120" />
                <el-table-column prop="submitTime" label="提交时间" width="160" />
                <el-table-column prop="status" label="状态" width="80" />
                <el-table-column label="操作" width="120">
                  <template #default>
                    <el-button text size="small">查看</el-button>
                    <el-button text size="small" type="danger">删除</el-button>
                  </template>
                </el-table-column>
              </el-table>
              <el-empty v-if="!dummyData.length" description="暂无提交数据" :image-size="50" />
            </div>
          </div>
        </div>
      </div>

      <!-- 报表视图 -->
      <div v-show="activeView==='report'" class="setting-wrapper">
        <div class="setting-scroll">
          <div class="setting-group">
            <div class="group-title">报表</div>
            <div v-if="!form.id" class="save-tip">请先保存并发布问卷</div>
            <div v-else style="padding:12px">
              <div v-for="(q, qi) in questions" :key="q.id" style="margin-bottom:16px;border:1px solid #eee;border-radius:6px;padding:12px">
                <div style="font-size:13px;font-weight:600;margin-bottom:8px">{{ qi+1 }}. {{ q.title }}</div>
                <div v-if="hasOptions(q)" style="display:flex;flex-direction:column;gap:4px">
                  <div v-for="o in (q.props?.options||[])" :key="o.value" style="display:flex;align-items:center;gap:8px">
                    <span style="font-size:12px;width:80px;flex-shrink:0">{{ o.label }}</span>
                    <el-progress :percentage="Math.floor(Math.random()*100)" :stroke-width="12" style="flex:1" />
                    <span style="font-size:11px;color:#999;width:40px;text-align:right">{{ Math.floor(Math.random()*50) }}</span>
                  </div>
                </div>
                <div v-else style="color:#999;font-size:12px">暂无统计</div>
              </div>
              <el-empty v-if="!questions.length" description="暂无题目统计" :image-size="50" />
            </div>
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
  </div>

      <!-- 公式编辑器对话框 -->
      <el-dialog v-model="formulaVisible" :title="formulaTitle" :key="formulaKey" width="720px" :close-on-click-modal="false">
        <div style="display:flex;gap:12px;min-height:360px">
          <div style="flex:1">
            <div style="font-size:12px;color:#888;margin-bottom:4px">{{ formulaType==='calculate'?'计算结果将自动填充到当前题目':'公式表达式' }}</div>
            <el-input v-model="formulaText" type="textarea" :rows="12" :placeholder="formulaPlaceholder" />
            <!-- 使用说明 -->
            <div style="font-size:12px;color:#606266;background:#f5f7fa;border-radius:4px;padding:8px 10px;margin-top:8px">
              <div style="font-weight:600;margin-bottom:4px">使用说明</div>
              <ul style="margin:0;padding-left:16px;line-height:1.8">
                <li>运算符、括号、逗号请使用<strong>英文输入法</strong></li>
                <li>判断相等用 <code>==</code>，赋值用 <code>=</code></li>
                <li>手动输入的文本用引号包裹，例如 <code>"优秀"</code></li>
                <li>公式结果类型需与目标题型匹配（数字题不能接收文本结果）</li>
                <li>点击右侧<strong style="color:#fb454c">红色高亮标签</strong>可插入题目标签到光标位置</li>
              </ul>
            </div>
            <!-- 函数参考 + 案例 -->
            <div style="font-size:11px;color:#999;margin-top:8px">
              <div style="font-weight:600;color:#606266;margin-bottom:4px">函数参考 &amp; 案例</div>
              <!-- 通用 -->
              <div><code>IF(条件, TRUE, FALSE)</code> — 条件判断</div>
              <div><code>AND(条件1, 条件2)</code> — 与运算</div>
              <div><code>OR(条件1, 条件2)</code> — 或运算</div>
              <div><code>NOT(条件)</code> — 非运算</div>
              <div><code>COUNT(Q1)</code> — 统计多选已选数量</div>
              <div><code>SCORE(Q1)</code> — 读取选项分值</div>
              <div><code>TEXT(Q1)</code> — 读取选项文本</div>
              <div><code>CURRENT_DATE("YYYY/MM/DD")</code> — 当前日期</div>
              <div><code>INCLUDE("学校", Q1)</code> — 判断文本是否包含关键词</div>
              <!-- 计算公式案例 -->
              <div v-if="formulaType==='calculate'" style="margin-top:6px;padding-top:6px;border-top:1px solid #eee">
                <div style="font-weight:600;color:#606266;margin-bottom:4px">计算公式案例</div>
                <div><code>SUM(Q1, Q2)</code> — 两题求和（计算总分）</div>
                <div><code>ROUND(Q1 / POWER(Q2 / 100, 2), 2)</code> — BMI 指数</div>
                <div><code>IFS(Q1&lt;60,"不合格",Q1&lt;80,"合格",TRUE(),"优秀")</code> — 多条件等级</div>
                <div><code>CONCATENATE("你的总分是", Q1, "分")</code> — 拼接结果文案</div>
                <div><code>Q1 + Q2 + Q3</code> — 连加运算</div>
                <div><code>(Q1 + Q2) / 2</code> — 求平均值</div>
              </div>
              <!-- 跳转/结束案例 -->
              <div v-if="formulaType==='endFormula'||formulaType==='jumpFormula'" style="margin-top:6px;padding-top:6px;border-top:1px solid #eee">
                <div style="font-weight:600;color:#606266;margin-bottom:4px">跳转/结束案例</div>
                <div><code>IF Q1A1 THEN BRANCH FROM Q1 TO Q5</code> — 选择选项A跳到Q5</div>
                <div><code>IF Q1A2 THEN BRANCH FROM Q1 TO END</code> — 选择选项B结束问卷</div>
                <div><code>IF AND(Q1A1, Q2&gt;18) THEN BRANCH FROM Q1 TO Q6</code> — 多条件跳转</div>
              </div>
              <!-- 必填公式案例 -->
              <div v-if="formulaType==='requiredFormula'" style="margin-top:6px;padding-top:6px;border-top:1px solid #eee">
                <div style="font-weight:600;color:#606266;margin-bottom:4px">必填公式案例</div>
                <div><code>IF Q1A1 THEN REQUIRED Q2</code> — 选A时Q2必填</div>
                <div><code>IF Q1A1 AND Q1A2 THEN REQUIRED Q3</code> — 同时选A、B时Q3必填</div>
                <div><code>IF NOT(ISBLANK(Q1)) THEN REQUIRED Q2</code> — Q1非空时Q2必填</div>
              </div>
              <!-- 文本替换案例 -->
              <div v-if="formulaType==='textReplace'" style="margin-top:6px;padding-top:6px;border-top:1px solid #eee">
                <div style="font-weight:600;color:#606266;margin-bottom:4px">文本替换案例</div>
                <div><code>REPLACE Q2 WITH CONCATENATE("你好，", Q1)</code> — 替换Q2标题</div>
                <div><code>REPLACE Q3 WITH CONCATENATE("你的得分：", SUM(Q1, Q2))</code> — 替换为计算结果</div>
              </div>
              <!-- 显示隐藏案例 -->
              <div v-if="formulaType==='optShow'" style="margin-top:6px;padding-top:6px;border-top:1px solid #eee">
                <div style="font-weight:600;color:#606266;margin-bottom:4px">显示隐藏案例</div>
                <div><code>IF Q1A1 THEN SHOW</code> — 选A时显示本选项</div>
                <div><code>IF Q1A2 THEN HIDE</code> — 选A时隐藏本选项</div>
              </div>
              <!-- 自动勾选案例 -->
              <div v-if="formulaType==='optAutoCheck'" style="margin-top:6px;padding-top:6px;border-top:1px solid #eee">
                <div style="font-weight:600;color:#606266;margin-bottom:4px">自动勾选案例</div>
                <div><code>IF Q1A1 THEN CHECK Q2A1</code> — 选A时自动勾选Q2的选项A</div>
                <div><code>IF Q1A1 THEN CHECK Q2A1 Q2A2</code> — 选A时自动勾选Q2的A、B选项</div>
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
                    <el-tag v-for="(o, oi) in q.props.options.slice(0,4)" :key="oi" size="small" type="warning" style="cursor:pointer" @click="insertFormulaTag(`Q${questions.indexOf(q)+1}A${oi+1}`)">Q{{ questions.indexOf(q)+1 }}A{{ oi+1 }}</el-tag>
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
</template>

<script setup lang="ts">
import { ref, computed, reactive, onMounted, nextTick, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useRoute, useRouter } from 'vue-router'
import { adminApi } from '../../api'
import { normalizeQuestions, importFromSurveyKing, exportToSurveyKing, isSurveyKingFormat, toSkType, toWcType } from '../../utils/surveyKingBridge'
import DraggableList from './formkit/DraggableList.vue'

import QuestionIcon from './formkit/QuestionIcon.vue'
import QuestionPreview from './formkit/QuestionPreview.vue'

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
  minSubmitMinutes: 0, maxSubmitMinutes: 0,
        backgroundImages: [] as any[], headerImages: [] as any[]
})

const activeView = ref('edit')
const panelMode = ref<'edit' | 'json' | 'preview'>('edit')
const middleTab = ref('item')
const showSurveySettings = ref(false)
const sideSubTab = ref('types')
const activeCategory = ref('')
const appearanceTab = ref('bg')
const allBgResources = ref<any[]>([])
const allHeaderResources = ref<any[]>([])
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
  title: form.title || '未命名问卷',
  isRoot: true,
  children: questions.value
    .filter((q: any) => q.type !== 'divider' && q.type !== 'pagination' && q.type !== 'description' && q.type !== 'questionSet')
    .map((q: any, i: number) => ({ q, index: i + 1 }))
}))

// 题型 tab 滚动
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
const newRowName = ref('')
const newColName = ref('')
const optDrawerIdx = ref(0)
const optDrawerVisible = ref(false)
const batchAddVisible = ref(false)
const batchAddText = ref('')

function typeName(t: string) {
  const map: Record<string, string> = {
    input: '单行文本', text: '文本', textarea: '多行', number: '数字',
    select: '下拉', radio: '单选', checkbox: '多选', picker: '选择器',
    cascade: '级联选择', judge: '判断',
    multiInput: '多项填空', hInput: '横向填空', scanCode: '扫码',
    signature: '签名', file: '上传文件',
    rating: '评分', nps: 'NPS评分',
    phone: '手机', email: '邮箱', idCard: '身份证', password: '密码',
    switch: '开关', location: '地理位置',
    date: '日期', time: '时间', dateRange: '日期范围',
    matrixRadio: '矩阵单选', matrixCheckbox: '矩阵多选',
    matrixFillBlank: '矩阵填空', matrixAuto: '表格自增',
    divider: '分割线', description: '说明',
    questionSet: '问题组', pagination: '分页',
    user: '成员', dept: '部门', richText: '富文本',
    autopop: '自动填充',
  }
  return map[t] || t
}

const formulaVisible = ref(false)
const formulaTitle = ref('')
const formulaText = ref('')
const formulaType = ref('')
const formulaKey = ref(0)
const formulaQuestions = computed(() => questions.value.filter(q => !isPureLayout(q.type)))
const formulaPlaceholder = computed(() => {
  const map: Record<string, string> = {
    calculate: '输入计算公式，如：\nSUM(Q1, Q2)\nIFS(Q1<60,"不合格","合格")\nCONCATENATE("总分：", Q1)',
    endFormula: '输入结束规则，如：\nIF Q1A2 THEN BRANCH FROM Q1 TO END',
    jumpFormula: '输入跳转规则，如：\nIF Q1A1 THEN BRANCH FROM Q1 TO Q5',
    requiredFormula: '输入必填规则，如：\nIF Q1A1 THEN REQUIRED Q2',
    textReplace: '输入文本替换规则，如：\nREPLACE Q2 WITH CONCATENATE("你好，", Q1)',
    optShow: '输入显示/隐藏规则，如：\nIF Q1A1 THEN SHOW\nIF Q1A2 THEN HIDE',
    optAutoCheck: '输入自动勾选规则，如：\nIF Q1A1 THEN CHECK Q2A1',
  }
  return map[formulaType.value] || '输入公式表达式'
})
function openFormulaDialog(type: string) {
  const titles: Record<string, string> = {
    endFormula: '结束公式', jumpFormula: '跳转公式',
    requiredFormula: '必填公式', textReplace: '文本替换',
    optShow: '显示隐藏公式', optAutoCheck: '自动勾选公式',
    calculate: '计算公式',
  }
  if (!selected.value) return
  formulaType.value = type
  formulaTitle.value = titles[type] || '公式编辑'
  // 从 selected 读取已有公式
  const fieldMap: Record<string, string> = {
    endFormula: 'endFormula', jumpFormula: 'jumpFormula',
    requiredFormula: 'requiredFormula', textReplace: 'replaceTextRule',
    optShow: 'optShowFormula', optAutoCheck: 'optAutoCheckFormula',
    calculate: 'calculateFormula',
  }
  const field = fieldMap[type] || ''
  formulaText.value = selected.value[field] || selected.value.props?.[field] || ''
  formulaKey.value++
  formulaVisible.value = true
}
function confirmFormula() {
  if (!selected.value) return
  const fieldMap: Record<string, string> = {
    endFormula: 'endFormula', jumpFormula: 'jumpFormula',
    requiredFormula: 'requiredFormula', textReplace: 'replaceTextRule',
    optShow: 'optShowFormula', optAutoCheck: 'optAutoCheckFormula',
  }
  const field = fieldMap[formulaType.value] || ''
  // 如果是选项相关的公式或 textReplace，存入 props
  if (['optShow','optAutoCheck','textReplace'].includes(formulaType.value)) {
    if (!selected.value.props) selected.value.props = {}
    selected.value.props[field] = formulaText.value
  } else if (formulaType.value === 'calculate') {
    if (!selected.value.props) selected.value.props = {}
    selected.value.props.calculateFormula = formulaText.value
  } else {
    selected.value[field] = formulaText.value
  }
  formulaVisible.value = false
  ElMessage.success('公式已保存')
}
// ===== 逻辑规则（表单编辑） =====
interface LogicCondition { questionIdx?: number; optionIdx?: number; operator?: string; compareValue?: string }
interface LogicRuleItem { id: string; conditionType: string; conditions: LogicCondition[]; action: string; targetQuestionIdx?: number; targetOptionIdxs?: number[]; branchFromIdx?: number; branchToIdx?: number; branchToEnd?: boolean; formula?: string }
const logicRuleList = ref<LogicRuleItem[]>([])
const showAddRule = ref(false)
const editingRuleIdx = ref(-1)
const defaultRuleForm = (): LogicRuleItem => ({ id: '', conditionType: 'simple', conditions: [{ questionIdx: undefined, optionIdx: undefined, operator: undefined, compareValue: undefined }], action: 'show', targetQuestionIdx: undefined, targetOptionIdxs: [], branchFromIdx: undefined, branchToIdx: undefined, branchToEnd: false, formula: '' })
const ruleForm = ref<LogicRuleItem>(defaultRuleForm())

// 规则变更时自动保存到当前题目
watch(logicRuleList, () => {
  syncLogicRules()
}, { deep: true })

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
  // Validate targets
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
  if (!form.id) { ElMessage.warning('请先保存问卷'); return }
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
    backgroundImages: form.backgroundImages, headerImages: form.headerImages
  })
  const payload: any = {
    id: form.id,
    title: form.title, desc: form.description, category: form.category, tags: form.tags,
    visibility: form.visibility, allowMulti: form.allowMultiBool,
    anonymous: form.anonymousBool, showResult: form.showResultBool,
    startTime: form.startDate || 0, endTime: form.endDate || 0,
    maxResponse: form.maxResponse, status: form.statusBool,
    schema, deptIds: form.deptIds, mode: form.mode, settings
  }
  try {
    await adminApi.surveyEdit(payload)
    ElMessage.success('逻辑规则已保存')
  } catch { ElMessage.error('保存失败') }
}
function parseDSLToRules(dsl: string): LogicRuleItem[] {
  const lines = dsl.split('\n').map(l => l.trim()).filter(Boolean)
  const rules: LogicRuleItem[] = []
  for (const line of lines) {
    const rule: LogicRuleItem = { id: 'rule_' + Date.now() + '_' + Math.random().toString(36).slice(2,6), conditionType: 'simple', conditions: [{ questionIdx: undefined }], action: 'show' }
    // Try to parse IF ... THEN ... pattern
    const ifMatch = line.match(/^IF\s+(.+?)\s+THEN\s+(.+)$/i)
    if (ifMatch) {
      const condStr = ifMatch[1].trim()
      const actionStr = ifMatch[2].trim()
      // Parse conditions
      const andMatch = condStr.match(/^AND\((.+)\)$/i)
      const orMatch = condStr.match(/^OR\((.+)\)$/i)
      const notMatch = condStr.match(/^NOT\((.+)\)$/i)
      if (andMatch || orMatch) {
        rule.conditionType = andMatch ? 'and' : 'or'
        const inner = (andMatch||orMatch)![1]
        rule.conditions = inner.split(',').map(s => parseConditionTag(s.trim()))
      } else if (notMatch) {
        rule.conditionType = 'not'
        rule.conditions = [parseConditionTag(notMatch[1].trim())]
      } else {
        rule.conditionType = 'simple'
        rule.conditions = [parseConditionTag(condStr)]
      }
      // Parse action
      parseActionStr(rule, actionStr)
    } else {
      // No IF - direct action (ASSIGNMENT, VALIDATE, REPLACE)
      rule.conditionType = 'none'
      rule.conditions = []
      parseActionStr(rule, line)
    }
    rules.push(rule)
  }
  return rules
}

function parseConditionTag(s: string): LogicCondition {
  const cond: LogicCondition = {}
  const qMatch = s.match(/Q(\d+)/)
  if (qMatch) {
    cond.questionIdx = parseInt(qMatch[1]) - 1
    const aMatch = s.match(/A(\d+)/)
    if (aMatch) cond.optionIdx = parseInt(aMatch[1]) - 1
  }
  // Handle operators
  if (s.includes('ISBLANK')) {
    cond.operator = s.startsWith('NOT') ? 'filled' : 'empty'
  } else if (s.includes('>')) {
    cond.operator = 'gt'
    const m = s.match(/>(\S+)/)
    if (m) cond.compareValue = m[1]
  } else if (s.includes('<')) {
    cond.operator = 'lt'
    const m = s.match(/<(\S+)/)
    if (m) cond.compareValue = m[1]
  } else if (s.includes('==')) {
    cond.operator = 'eq'
    const m = s.match(/==(\S+)/)
    if (m) cond.compareValue = m[1]
  }
  return cond
}

function parseActionStr(rule: LogicRuleItem, s: string) {
  const branchMatch = s.match(/^BRANCH\s+FROM\s+(Q\d+)\s+TO\s+(Q\d+|END)$/i)
  const checkMatch = s.match(/^CHECK\s+(.+)$/i)
  const assignMatch = s.match(/^ASSIGNMENT\s+(Q\d+)\s+WITH\s+(.+)$/i)
  const validMatch = s.match(/^VALIDATE\s+(Q\d+)\s+WITH\s+(.+)$/i)
  const replaceMatch = s.match(/^REPLACE\s+(Q\d+)\s+WITH\s+(.+)$/i)

  if (branchMatch) {
    rule.action = (branchMatch[2].toUpperCase() === 'END') ? 'end' : 'branch'
    rule.branchFromIdx = parseInt(branchMatch[1].slice(1)) - 1
    rule.branchToEnd = branchMatch[2].toUpperCase() === 'END'
    if (!rule.branchToEnd) rule.branchToIdx = parseInt(branchMatch[2].slice(1)) - 1
  } else if (checkMatch) {
    rule.action = 'check'
    const parts = checkMatch[1].trim().split(/\s+/)
    if (parts.length) {
      const first = parts[0].match(/Q(\d+)/)
      if (first) { rule.targetQuestionIdx = parseInt(first[1]) - 1; rule.targetOptionIdxs = parts.map(p => { const m = p.match(/A(\d+)/); return m ? parseInt(m[1])-1 : -1 }).filter(x => x>=0) }
    }
  } else if (assignMatch) {
    rule.action = 'assignment'
    rule.targetQuestionIdx = parseInt(assignMatch[1].slice(1)) - 1
    rule.formula = assignMatch[2].trim()
  } else if (validMatch) {
    rule.action = 'validate'
    rule.targetQuestionIdx = parseInt(validMatch[1].slice(1)) - 1
    rule.formula = validMatch[2].trim()
  } else if (replaceMatch) {
    rule.action = 'replace'
    rule.targetQuestionIdx = parseInt(replaceMatch[1].slice(1)) - 1
    rule.formula = replaceMatch[2].trim()
  } else {
    // SHOW / HIDE / REQUIRED
    const m = s.match(/^(SHOW|HIDE|REQUIRED)\s+(Q\d+)$/i)
    if (m) {
      rule.action = m[1].toLowerCase()
      rule.targetQuestionIdx = parseInt(m[2].slice(1)) - 1
    }
  }
}const stats = reactive({ totalCount: 0, todayCount: 0, avgTime: 0, completionRate: 0 })
const dummyData = ref<any[]>([])

function insertFormulaTag(tag: string) {
  const ta = document.querySelector('.el-dialog__body textarea') as HTMLTextAreaElement
  if (ta) {
    const start = ta.selectionStart
    const end = ta.selectionEnd
    const text = formulaText.value
    formulaText.value = text.slice(0, start) + tag + text.slice(end)
    nextTick(() => { ta.selectionStart = ta.selectionEnd = start + tag.length; ta.focus() })
  } else {
    formulaText.value += tag
  }
}

function addOption() {
  if (!selected.value?.props) return
  if (!Array.isArray(selected.value.props.options)) selected.value.props.options = []
  const n = selected.value.props.options.length + 1
  const letter = n <= 26 ? String.fromCharCode(64 + n) : `opt${n}`
  selected.value.props.options.push({ label: `选项${n}`, value: letter })
}
function removeOption(idx: number) {
  if (!selected.value?.props?.options) return
  selected.value.props.options.splice(idx, 1)
}
function confirmBatchAdd() {
  if (!selected.value?.props) return
  const lines = batchAddText.value.split('\n').map(s => s.trim()).filter(Boolean)
  for (const line of lines) {
    const parts = line.split('|')
    const label = parts[0].trim()
    const value = parts[1]?.trim() || label
    if (!Array.isArray(selected.value.props.options)) selected.value.props.options = []
    selected.value.props.options.push({ label, value })
  }
  batchAddText.value = ''
  batchAddVisible.value = false
}

function addMatrixRow() {
  if (!selected.value?.props) return
  if (!Array.isArray(selected.value.props.rows)) selected.value.props.rows = []
  const name = newRowName.value.trim() || `行${selected.value.props.rows.length + 1}`
  selected.value.props.rows.push({ title: name, id: genId(), width: 150 })
  newRowName.value = ''
}
function removeMatrixRow(idx: number) {
  selected.value?.props?.rows?.splice(idx, 1)
}
function addMatrixCol() {
  if (!selected.value?.props) return
  if (!Array.isArray(selected.value.props.columns)) selected.value.props.columns = []
  const name = newColName.value.trim() || `列${selected.value.props.columns.length + 1}`
  selected.value.props.columns.push({ title: name, id: genId(), width: 150 })
  newColName.value = ''
}
function removeMatrixCol(idx: number) {
  selected.value?.props?.columns?.splice(idx, 1)
}
function addAutoCol() {
  if (!selected.value?.props) return
  if (!Array.isArray(selected.value.props.columns)) selected.value.props.columns = []
  selected.value.props.columns.push({ label: '字段' + (selected.value.props.columns.length + 1), id: genId(), width: 150, type: 'input' })
}
function removeAutoCol(idx: number) {
  selected.value?.props?.columns?.splice(idx, 1)
}
function addUserOpt() {
  if (!selected.value?.props) return
  if (!Array.isArray(selected.value.props.options)) selected.value.props.options = []
  const n = selected.value.props.options.length + 1
  if (selected.value.type === 'user') {
    selected.value.props.options.push({ label: `成员${n}`, value: `member${n}`, deptName: '', deptId: '', parentDeptId: '' })
  } else {
    selected.value.props.options.push({ label: `部门${n}`, value: `dept${n}`, parentId: '' })
  }
}
function removeUserOpt(idx: number) {
  selected.value?.props?.options?.splice(idx, 1)
}
function importUserOpt() {
  if (!selected.value) return
  const q = selected.value
  const input = document.createElement('input')
  input.type = 'file'
  input.accept = '.json'
  input.onchange = (e: any) => {
    const file = e.target?.files?.[0]
    if (!file) return
    const reader = new FileReader()
    reader.onload = (ev) => {
      try {
        const data = JSON.parse(ev.target?.result as string)
        if (!Array.isArray(data)) { ElMessage.error('JSON 格式错误，应为数组'); return }
        if (!Array.isArray(q.props?.options)) q.props.options = []
        const isUser = q.type === 'user'
        for (const item of data) {
          if (!item.label && !item.value) continue
          const opt: any = { label: item.label || item.name || '', value: item.value || item.id || '' }
          if (isUser) { opt.deptName = item.deptName || item.department || ''; opt.deptId = item.deptId || item.departmentId || ''; opt.parentDeptId = item.parentDeptId || '' }
          else { opt.parentId = item.parentId || '' }
          q.props.options.push(opt)
        }
        ElMessage.success(`成功导入 ${data.filter((d: any) => d.label || d.value).length} 条`)
      } catch { ElMessage.error('JSON 解析失败') }
    }
    reader.readAsText(file)
  }
  input.click()
}

const types = ref<any[]>([])
const categoryDefs = [
  { name: 'select', label: '选择题', types: ['radio','checkbox','select','cascade','picker','judge','file'] },
  { name: 'fill', label: '填空题', types: ['input','textarea','number','multiInput','hInput','signature','scanCode'] },
  { name: 'rating', label: '打分题', types: ['rating','nps'] },
  { name: 'matrix', label: '矩阵题', types: ['matrixRadio','matrixCheckbox','matrixFillBlank','matrixAuto'] },
  { name: 'layout', label: '辅助布局', types: ['divider','description','questionSet','pagination'] },
  { name: 'advanced', label: '高级题型', types: ['user','dept','richText','autopop'] },
  { name: 'personal', label: '个人信息', types: ['phone','email','idCard','password','date','time','dateRange','switch','location'] },
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

interface Question { id: string; type: string; title: string; description?: string; required: boolean; placeholder?: string; props?: any; validate?: any[]; logic?: any[]; calcValue?: any; readOnly?: boolean; dataType?: string; examScore?: number; examCorrectAnswer?: string; examAnalysis?: string; mediaType?: string; mediaUrl?: string; mediaWidth?: string; mediaAlign?: string; [key: string]: any }
const questions = ref<Question[]>([])
const selected = ref<Question | null>(null)

let idCounter = 0

function genId() { idCounter++; return 'q' + idCounter }

function addQuestion(t: any) {
  const q: Question = { id: genId(), type: t.type, title: t.displayName, required: false, readOnly: false, dataType: '', placeholder: '', mediaType: '', mediaUrl: '', mediaWidth: '', mediaAlign: 'center', props: t.defaultProps ? JSON.parse(JSON.stringify(t.defaultProps)) : {} }
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
    q.props.rows = [{ title: '行1', id: genId() }, { title: '行2', id: genId() }, { title: '行3', id: genId() }]
    q.props.columns = [{ title: '列A', id: genId(), width: 150 }, { title: '列B', id: genId(), width: 150 }]
  }
  if (t.type === 'matrixAuto') {
    q.props.columns = [{ label: '姓名', id: genId(), width: 150, type: 'input' }, { label: '数值', id: genId(), width: 150, type: 'number' }]
    q.props.minRows = 0
    q.props.maxRows = 0
  }
  if (t.type === 'questionSet') {
    q.props.children = []
  }
  if (t.type === 'number') {
    q.props.min = undefined
    q.props.max = undefined
    q.props.decimalPlaces = 0
  }
  if (t.type === 'rating' || t.type === 'nps') {
    q.props.maxRating = t.type === 'nps' ? 10 : 5
    q.props.icon = 'star'
  }
  if (t.type === 'multiInput' || t.type === 'hInput') {
    q.props.fields = [{ label: '填空1', placeholder: '' }, { label: '填空2', placeholder: '' }]
  }
  if (['input','textarea'].includes(t.type)) {
    q.props.options = [{ label: '', value: 'A', calculate: '', dataType: '', placeholder: '' }]
    q.props.calculateFormula = ''
  }
  if (t.type === 'user' || t.type === 'dept') {
    q.props.options = t.type === 'user'
      ? [{ label: '成员A', value: 'memberA', deptName: '', deptId: '', parentDeptId: '' }, { label: '成员B', value: 'memberB', deptName: '', deptId: '', parentDeptId: '' }]
      : [{ label: '部门A', value: 'deptA', parentId: '' }, { label: '部门B', value: 'deptB', parentId: '' }]
  }
  if (t.type === 'number') {
    q.props.calculateFormula = ''
  }
  if (t.type === 'date' || t.type === 'time' || t.type === 'dateRange') {
    q.props.calculateFormula = ''
  }
  if (t.type === 'description') {
    q.props.replaceTextRule = ''
  }
  if (t.type === 'pagination') {
    q.props.currentPage = 1
    q.props.totalPage = 1
  }
  if (t.type === 'signature') {
    q.examAnswerMode = ''
  }
  if (t.type === 'file') {
    q.fileTypes = ['image']
    q.fileExtensions = ''
    q.maxFileSize = 10
    q.maxFileCount = 1
  }
  questions.value.push(q)
  selected.value = q
}

const selectedOptIdx = ref(-1)
function selectQuestion(id: string) {
  selected.value = questions.value.find(q => q.id === id) || null
  selectedOptIdx.value = -1
  showSurveySettings.value = false
}
function selectOption(qId: string, optIdx: number) {
  selected.value = questions.value.find(q => q.id === qId) || null
  selectedOptIdx.value = optIdx
  showSurveySettings.value = false
  if (selected.value && (selected.value.type === 'user' || selected.value.type === 'dept')) {
    if (!selected.value.props) selected.value.props = {}
    if (!Array.isArray(selected.value.props.options)) selected.value.props.options = []
    while (selected.value.props.options.length <= optIdx) {
      selected.value.props.options.push(selected.value.type === 'user'
        ? { label: '', value: '', deptName: '', deptId: '', parentDeptId: '' }
        : { label: '', value: '', parentId: '' })
    }
  }
}
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

function hasOptions(q: Question) { return ['select','radio','checkbox','picker','matrixRadio','matrixCheckbox','judge','cascade'].includes(q.type) }
function hasDataType(q: Question) { return ['input','textarea','number','multiInput','hInput','phone','email','idCard','password','date','time','dateRange','switch','location'].includes(q.type) }
const LAYOUT_PURE = ['divider','description','pagination']
const LAYOUT_ALL = ['divider','description','pagination','questionSet']
const CHOICE_TYPES = ['select','radio','checkbox','picker','cascade','judge']
const INPUT_TYPES = ['input','textarea','number','multiInput','hInput']
const MATRIX_QR = ['matrixRadio','matrixCheckbox','matrixFillBlank']
const MATRIX_ALL = ['matrixRadio','matrixCheckbox','matrixFillBlank','matrixAuto']
const importFormatHint = computed(() => {
  if (selected.value?.type === 'user') return 'JSON 格式: [{"label":"姓名","value":"ID","deptName":"部门","deptId":"部门ID","parentDeptId":"父级部门ID"}]'
  if (selected.value?.type === 'dept') return 'JSON 格式: [{"label":"部门名称","value":"部门ID","parentId":"父级部门ID"}]'
  return ''
})
const PERSONAL_TYPES = ['phone','email','idCard','password','date','time','dateRange','switch','location']
function isPureLayout(t: string) { return LAYOUT_PURE.includes(t) }
function isLayoutAll(t: string) { return LAYOUT_ALL.includes(t) }
function isChoice(t: string) { return CHOICE_TYPES.includes(t) }
function isInput(t: string) { return INPUT_TYPES.includes(t) }
function isMatrixQR(t: string) { return MATRIX_QR.includes(t) }
function isMatrixAll(t: string) { return MATRIX_ALL.includes(t) }
function isPersonal(t: string) { return PERSONAL_TYPES.includes(t) }

function removeQuestionById(id: string) {
  questions.value = questions.value.filter(q => q.id !== id)
  if (selected.value?.id === id) selected.value = null
}

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

const publicUrl = computed(() => form.id ? `${window.location.origin}/#/pages/survey/fill?id=${form.id}` : '')
const qrUrl = ref('')
const embedCode = computed(() => form.id ? `<iframe src="${publicUrl.value}" width="100%" height="600" frameborder="0"></iframe>` : '')

const exportedJson = ref('')

function copyLink() { navigator.clipboard.writeText(publicUrl.value); ElMessage.success('已复制') }
function genQR() { qrUrl.value = `https://api.qrserver.com/v1/create-qr-code/?size=200x200&data=${encodeURIComponent(publicUrl.value)}` }
function copyJson() { navigator.clipboard.writeText(exportedJson.value); ElMessage.success('已复制') }

function exportSurveyKingJson() {
  const skQuestions = exportToSurveyKing(questions.value)
  exportedJson.value = JSON.stringify({ version: '2.0', questions: skQuestions, setting: {} }, null, 2)
  ElMessage.success('已转换为 SurveyKing JSON')
}
async function loadSurveyKingJson() {
  try {
    const text = prompt('请粘贴 SurveyKing JSON：')
    if (!text) return
    const parsed = JSON.parse(text)
    const raw = parsed.questions || parsed
    const converted = importFromSurveyKing(raw)
    if (!converted.length) { ElMessage.warning('未识别到有效题目'); return }
    questions.value = converted
    idCounter = questions.value.length
    selected.value = converted[0] || null
    ElMessage.success(`已导入 ${converted.length} 题`)
  } catch { ElMessage.error('JSON 解析失败') }
}
function downloadJson() {
  const blob = new Blob([exportedJson.value], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url; a.download = 'survey.json'; a.click()
  URL.revokeObjectURL(url)
}

function openQuestionBank() {
  ElMessage.info('题库功能待实现')
}

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
  if (!form.id) { ElMessage.warning('请先保存问卷'); return false }
  return true
}

async function loadResources() {
  if (!form.id) return
  try {
    const res: any = await adminApi.surveyResourceList({ surveyId: form.id })
    const list: any[] = res.data || []
    if (!Array.isArray(list)) return
    allBgResources.value = list.filter((r: any) => r.type === 'bg')
    allHeaderResources.value = list.filter((r: any) => r.type === 'header')
  } catch {}
}
function deselectQuestion() {
  selected.value = null
  showSurveySettings.value = false
}
function openSurveySettings() {
  selected.value = null
  showSurveySettings.value = true
}
watch(middleTab, (v) => { if (v === 'appearance') loadResources() })
watch(sideSubTab, (v) => { if (v === 'bank') loadBank() })

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
          minSubmitMinutes: s.minSubmitMinutes || 0, maxSubmitMinutes: s.maxSubmitMinutes || 0,
          backgroundImages: s.backgroundImages || [], headerImages: s.headerImages || []
        })
      } catch {}
    } else {
      Object.assign(base, {
        questionNumber: true, progressBar: false, autoSave: false, password: '',
        loginRequired: false, onePageOneQuestion: false, answerSheetVisible: true,
        copyEnabled: true, triggerType: 'onBlur', enableUpdate: false,
        transcriptVisible: true, rankVisible: false, redirectUrl: '', endContent: '',
        examRankingEnabled: false, exerciseMode: false, randomOrder: false,
        minSubmitMinutes: 0, maxSubmitMinutes: 0,
  backgroundImages: [] as string[], headerImages: [] as string[]
      })
    }
    Object.assign(form, base)
    loadResources()
    const rawSchema = res.data.schema
    if (rawSchema) {
      try {
        const sch = JSON.parse(rawSchema)
        if (sch.questions) {
          questions.value = normalizeQuestions(sch.questions)
          idCounter = questions.value.length
        }
        if (sch.setting?.logicRules) {
          logicRuleList.value = parseDSLToRules(sch.setting.logicRules)
        }
      } catch {}
    }
  } catch { ElMessage.error('加载失败') }
}

async function save() {
  if (!form.title) { ElMessage.warning('请填写标题'); return }
  saving.value = true
  try {
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
      backgroundImages: form.backgroundImages, headerImages: form.headerImages
    })
    const payload: any = {
      title: form.title, description: form.description, category: form.category, tags: form.tags,
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

const FALLBACK_TYPES = [
  // 选择题
  { type:'radio', displayName:'单选题', category:'select' },
  { type:'checkbox', displayName:'多选题', category:'select' },
  { type:'select', displayName:'下拉题', category:'select' },
  { type:'picker', displayName:'选择器', category:'select' },
  { type:'cascade', displayName:'级联选择', category:'select' },
  { type:'judge', displayName:'判断题', category:'select' },
  { type:'file', displayName:'上传文件', category:'select' },
  // 填空题
  { type:'input', displayName:'单行文本', category:'fill' },
  { type:'textarea', displayName:'多行文本', category:'fill' },
  { type:'number', displayName:'数字', category:'fill' },
  { type:'multiInput', displayName:'多项填空', category:'fill' },
  { type:'hInput', displayName:'横向填空', category:'fill' },
  { type:'signature', displayName:'电子签名', category:'fill' },
  { type:'scanCode', displayName:'扫码', category:'fill' },
  // 打分题
  { type:'rating', displayName:'评分', category:'rating' },
  { type:'nps', displayName:'NPS', category:'rating' },
  // 矩阵题
  { type:'matrixRadio', displayName:'矩阵单选', category:'matrix' },
  { type:'matrixCheckbox', displayName:'矩阵多选', category:'matrix' },
  { type:'matrixFillBlank', displayName:'矩阵填空', category:'matrix' },
  { type:'matrixAuto', displayName:'表格自增', category:'matrix' },
  // 辅助布局
  { type:'divider', displayName:'分割线', category:'layout' },
  { type:'description', displayName:'文字描述', category:'layout' },
  { type:'questionSet', displayName:'问题组', category:'layout' },
  { type:'pagination', displayName:'分页', category:'layout' },
  // 高级题型
  { type:'user', displayName:'成员', category:'advanced' },
  { type:'dept', displayName:'部门', category:'advanced' },
  { type:'richText', displayName:'富文本', category:'advanced' },
  { type:'autopop', displayName:'自动填充', category:'advanced' },
  // 个人信息
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
    const cats = categories.value
    if (cats.length) activeCategory.value = cats[0].name
  } catch {
    types.value = FALLBACK_TYPES
  }
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
.appearance-thumb.active { border-color:#fb454c; background:#fff5f5; }
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

/* ======= 逻辑面板（全宽） ======= */
.logic-full-panel {
  display:flex; flex:1; flex-direction:column; overflow:hidden;
  background:#f7f8fa; padding:20px 32px;
}
.logic-toolbar {
  display:flex; align-items:center; justify-content:space-between;
  margin-bottom:12px; flex-shrink:0;
}
.logic-body {
  display:flex; gap:16px; flex:1; overflow:hidden;
}
.logic-editor-area {
  flex:1; display:flex; flex-direction:column; overflow-y:auto;
}
.logic-sidebar {
  width:260px; flex-shrink:0; overflow-y:auto;
  background:#fff; border-radius:6px; padding:12px;
  border:1px solid #e8e8e8; line-height:1.8;
}
.logic-rule-card {
  background:#fff; border:1px solid #e8e8e8; border-radius:6px; margin-bottom:6px; transition:box-shadow .15s;
}
.logic-rule-card:hover { box-shadow:0 1px 4px rgba(0,0,0,0.06); }
.rule-header { display:flex; align-items:center; gap:8px; padding:8px 12px; }
.rule-index { font-size:11px; color:#999; font-weight:600; width:24px; flex-shrink:0; }
.rule-dsl { font-size:12px; font-family:Consolas,Monaco,"Courier New",monospace; flex:1; color:#303133; }
.rule-actions { flex-shrink:0; }

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
.editor-wrapper { max-width:210mm; margin:0 auto; }
.editor { border-radius:12px; box-shadow:0 2px 12px rgba(0,0,0,0.06); overflow:hidden; }
.json-panel, .survey-preview-panel { display:flex; flex-direction:column; height:100%; }
.json-panel { padding:0; }
.json-panel :deep(.el-textarea__inner) { height:100% !important; min-height:500px; font-size:13px; }
.preview-iframe-wrapper { flex:1; }
.preview-iframe { width:100%; height:100%; border-radius:8px; }
.survey-preview-panel { max-width:210mm; margin:0 auto; background-color:transparent; border-radius:12px; box-shadow:0 2px 12px rgba(0,0,0,0.06); overflow-y:auto; }
.survey-preview-header { padding:24px 28px 12px; text-align:center; border-bottom:1px solid #f0f0f0; }
.preview-form-title { font-size:22px; font-weight:600; color:#303133; }
.preview-form-desc { font-size:14px; color:#666; margin-top:6px; white-space:pre-wrap; word-break:break-word; text-align:left; }
.survey-preview-body { flex:1; overflow-y:auto; padding:16px 28px; }
.survey-preview-footer { text-align:center; padding:16px 28px 24px; color:#999; font-size:13px; border-top:1px solid #f0f0f0; }
.preview-question-card { margin-bottom:12px; }

/* 问卷头 */
.header {
  padding:24px 28px 12px; text-align:center; border-bottom:1px solid #f0f0f0;
  position:relative; min-height:60px; transition:background 0.3s;
  background:rgba(255,255,255,0.85); backdrop-filter:blur(4px);
}
.editor-header-img { width:100%; overflow:hidden; cursor:pointer; line-height:0; border-radius:12px 12px 0 0; }
.preview-header-img { width:100%; overflow:hidden; line-height:0; flex-shrink:0; }
.preview-header-img img { width:100%; max-height:240px; object-fit:cover; display:block; }
.editor-header-img img { width:100%; max-height:240px; object-fit:cover; display:block; }
.prefix { font-size:13px; color:#999; margin-bottom:4px; }
.header-content .title { font-size:22px; font-weight:600; color:#303133; line-height:1.4; }
.header-title-input :deep(.el-input__wrapper) { box-shadow:none!important; padding:0; background:transparent; }
.header-title-input :deep(.el-input__inner) { font-size:22px; font-weight:600; color:#303133; text-align:center; border:none; cursor:pointer; }
.header-title-input :deep(.el-input__inner):focus { cursor:text; }
.header-desc { font-size:14px; color:#666; margin-top:6px; white-space:pre-wrap; word-break:break-word; text-align:left; }

.questions-area { padding:16px 24px; }
.questions-area :deep(.draggable-list) { gap:8px; }

.applied-image-row { display:flex; align-items:center; gap:8px; margin-bottom:8px; padding:6px; background:#fff; border-radius:6px; border:1px solid #eee; }
.applied-image-thumb { width:40px; height:40px; border-radius:4px; object-fit:cover; flex-shrink:0; }
.applied-image-label { font-size:12px; color:#666; flex:1; }
.applied-image-remove { width:22px; height:22px; border-radius:50%; border:none; background:#fee; color:#fb454c; cursor:pointer; display:flex; align-items:center; justify-content:center; font-size:12px; flex-shrink:0; }
.applied-image-remove:hover { background:#fb454c; color:#fff; }

.footer {
  padding:24px 28px; text-align:center; border-top:1px solid #f0f0f0;
}
.footer-inner { font-size:13px; color:#999; }

/* ======= 右侧属性面板 ======= */
.survey-setting-panel { width:380px; flex-shrink:0; border-left:1px solid #e8e8e8; background:#fafafa; overflow-y:auto; }
.props-panel { padding:12px; }
.props-panel h3 { font-size:14px; font-weight:500; color:#fb454c; margin:0 0 8px; padding-bottom:8px; border-bottom:2px solid #fb454c; }
.props-panel :deep(.el-form-item) { margin-bottom:8px; }
.props-panel :deep(.el-form-item__label) { font-size:12px; color:#666; padding-bottom:2px; font-weight:500; line-height:1.2; }
.props-panel :deep(.el-divider) { margin:8px 0; }
.props-panel :deep(.el-collapse-item__header) { font-size:12px; font-weight:500; }
.props-footer { padding:8px 12px; border-top:1px solid #e8e8e8; background:#fafafa; position:sticky; bottom:0; z-index:1; }
.setting-row { display:flex; align-items:center; justify-content:space-between; margin-bottom:8px; font-size:12px; color:#666; }
.setting-row .setting-label { font-weight:500; }
.props-options-section { margin-bottom:8px; padding:8px; background:#f5f6f8; border-radius:6px; }
.setting-opt-row { display:flex; align-items:center; gap:6px; margin-bottom:4px; }

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
</style>
