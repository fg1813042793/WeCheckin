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
            <div class="survey-sidebar-panel-tabs-pane" :class="{ active: middleTab==='result' }" title="统计" @click="middleTab='result'">
              <svg viewBox="0 0 1024 1024" width="20" height="20" fill="currentColor"><path d="M128 128h768v85.333333H213.333333v682.666667H128V128z m170.666667 170.666667h85.333333v426.666666h-85.333333V298.666667z m170.666666 128h85.333334v298.666666h-85.333334V426.666667z m170.666667-42.666667h85.333333v341.333333h-85.333333V384z"/></svg>
              <span class="tab-label">统计</span>
              <span v-if="form.resultConfig.viewMode === 'perResponse'" class="tab-badge">逐份</span>
            </div>
          </div>
          <div class="survey-sidebar-panel-tabs-content" v-show="middleTab!=='logic' && middleTab!=='result'">
            <!-- 题目 -->
            <template v-if="middleTab==='item'">
              <el-tabs v-model="sideSubTab" class="side-sub-tabs">
                <el-tab-pane label="题型" name="types">
                  <SurveyDesignerQuestionPalette :types="types" @add="addQuestion" />
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
                  <SurveyDesignerOutline :title="form.title" :questions="questions" :selected-id="selected?.id" :invalid-ids="invalidQuestionIds" :type-name="typeName" @select="selectQuestion($event, true)" />
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
                    <el-upload :action="`/api/v2/admin/survey-resources`" :show-file-list="false" :on-success="handleBgSuccess" :on-error="()=>ElMessage.error('上传失败')" :headers="uploadHeaders" accept="image/*" :data="{ surveyId: form.id, resType: 'bg' }" :before-upload="checkSaved">
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
                    <el-upload :action="`/api/v2/admin/survey-resources`" :show-file-list="false" :on-success="handleHeaderSuccess" :on-error="()=>ElMessage.error('上传失败')" :headers="uploadHeaders" accept="image/*" :data="{ surveyId: form.id, resType: 'header' }" :before-upload="checkSaved">
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
              <el-button size="small" type="primary" @click="addNewRule">+ 增加规则</el-button>
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
                  <span class="rule-index">#{{ toNumericIndex(ri) + 1 }}</span>
                  <span class="rule-dsl">{{ renderRuleDSL(rule) }}</span>
                  <div class="rule-actions">
                    <el-button text size="small" @click="editRule(toNumericIndex(ri))">编辑</el-button>
                    <el-button text size="small" type="danger" @click="removeRule(toNumericIndex(ri))">删除</el-button>
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
                <div><code>POST_STAT VALUE CHANNEL(webhook) WEBHOOK(dingtalk:https://oapi.dingtalk.com/robot/send?access_token=xxx) NOTIFY_ADMIN USERS(123,456)</code> — 交卷后按选项值统计，钉钉群通知+站内通知管理员和指定成员</div>
              </div>
            </div>
          </div>
        </div>

        <!-- 统计配置（全宽） -->
        <div v-if="middleTab==='result'" class="logic-panel logic-full-panel">
          <div class="logic-toolbar">
            <h4 style="margin:0;font-size:14px">问卷统计配置</h4>
          </div>
          <div style="font-size:12px;color:#606266;background:#e8f4fd;border:1px solid #b3d8f0;border-radius:4px;padding:8px 12px;margin-bottom:12px;line-height:1.8">
            配置问卷提交后的结果统计方式，包括统计字段、图表展示、数据导出等。
          </div>
          <div class="logic-body">
            <div class="logic-editor-area">
              <el-form label-position="top" size="small">
                <el-form-item label="统计方式">
                  <el-radio-group v-model="form.resultConfig.statType">
                    <el-radio value="value">按选项值统计</el-radio>
                    <el-radio value="count">按选项计数统计</el-radio>
                    <el-radio value="score">按分值统计（仅考试模式）</el-radio>
                  </el-radio-group>
                  <div style="font-size:11px;color:#999;margin-top:4px">
                    <template v-if="form.resultConfig.statType==='value'">统计时使用选项设置的「选项值」字段进行计算</template>
                    <template v-else-if="form.resultConfig.statType==='count'">统计每个选项被选择的次数</template>
                    <template v-else>基于题目设置的「分值」计算总分/平均分</template>
                  </div>
                </el-form-item>
                <el-divider />
                <el-form-item label="查看模式">
                  <el-radio-group v-model="form.resultConfig.viewMode">
                    <el-radio value="aggregate">总览统计（所有答卷汇总）</el-radio>
                    <el-radio value="perResponse">逐份统计（每份答卷单独查看）</el-radio>
                  </el-radio-group>
                </el-form-item>
                <el-form-item v-if="form.resultConfig.viewMode === 'aggregate'" label="统计范围">
                  <el-checkbox-group v-model="form.resultConfig.scope">
                    <el-checkbox value="all">全部题目</el-checkbox>
                    <el-checkbox value="choice">仅选择题</el-checkbox>
                    <el-checkbox value="score">仅评分题</el-checkbox>
                  </el-checkbox-group>
                </el-form-item>
                <el-divider />
                <el-form-item label="结果显示">
                  <div style="display:flex;flex-direction:column;gap:10px">
                    <el-switch v-model="form.resultConfig.showChart" active-text="显示图表" inactive-text="隐藏图表" />
                    <el-switch v-model="form.resultConfig.showDetail" active-text="显示详细数据" inactive-text="隐藏详细数据" />
                  </div>
                </el-form-item>
                <el-form-item label="导出时选项字段">
                  <el-radio-group v-model="form.resultConfig.exportField">
                    <el-radio value="value">选项值 (value)</el-radio>
                    <el-radio value="label">选项文本 (label)</el-radio>
                    <el-radio value="both">值+文本 (value / label)</el-radio>
                  </el-radio-group>
                </el-form-item>
              </el-form>
            </div>
            <div class="logic-sidebar">
              <div style="font-size:12px;font-weight:600;color:#606266;margin-bottom:6px">配置说明</div>
              <div style="font-size:12px;line-height:2.2">
                <div><strong>按选项值统计：</strong>适用于选项有数值含义的题目（如评分、分级），使用选项的 value 字段参与计算。</div>
                <div><strong>按选项计数统计：</strong>统计每个选项被选中次数，适用于普通选择题。</div>
                <div><strong>按分值统计：</strong>基于题目设置的分值（考试模式）计算总分和平均分。</div>
                <div><strong>总览统计：</strong>所有答卷合并，以图表/表格展示每题的整体分布。</div>
                <div><strong>逐份统计：</strong>每份答卷独立展示，便于查看每个提交者的具体作答。</div>
                <div><strong>导出字段：</strong>控制 CSV/Excel 导出时使用选项值还是选项文本。</div>
              </div>
            </div>
          </div>
          <div style="padding:12px;border-top:1px solid #eee;display:flex;gap:8px;justify-content:center">
            <el-button type="primary" size="small" :loading="saving" @click="save">保存配置</el-button>
            <el-button size="small" plain @click="goToStatReport">查看统计报表</el-button>
          </div>
        </div>

        <!-- 添加/编辑规则对话框 -->
        <el-dialog v-model="showAddRule" :title="editingRuleIdx>=0?'编辑规则':'增加规则'" width="680px" :close-on-click-modal="false">
          <el-form label-position="top" size="small">
            <el-form-item label="条件类型">
              <el-select v-model="ruleForm.conditionType" style="width:100%" :disabled="ruleForm.action==='postStat'">
                <el-option label="简单条件 (单个)" value="simple" />
                <el-option label="且 (AND)" value="and" />
                <el-option label="或 (OR)" value="or" />
                <el-option label="非 (NOT)" value="not" />
                <el-option label="无条件 (赋值/校验/替换/交卷后统计)" value="none" />
              </el-select>
            </el-form-item>
            <template v-if="ruleForm.conditionType!=='none'">
              <el-form-item v-for="(cond, ci) in ruleForm.conditions" :key="ci" :label="toNumericIndex(ci) === 0 ? '条件' : '条件 ' + (toNumericIndex(ci) + 1)">
                <div style="display:flex;gap:6px;align-items:center;flex-wrap:wrap">
                  <el-select v-model="cond.questionIdx" placeholder="选择题目" style="width:160px" @change="cond.optionIdx=undefined;cond.operator=undefined">
                    <el-option v-for="(q, qi) in questions" :key="qi" :label="`Q${toNumericIndex(qi) + 1}: ${q.title?.slice(0,20)}`" :value="toNumericIndex(qi)" />
                  </el-select>
                  <!-- 选项选择（仅选择题） -->
                  <el-select v-model="cond.optionIdx" placeholder="选择选项" style="width:120px" v-if="hasOptionsByIndex(cond.questionIdx)" clearable @change="cond.operator='optSelected'">
                    <el-option v-for="(o, oi) in getOptionsByIndex(cond.questionIdx)" :key="oi" :label="o.label" :value="toNumericIndex(oi)" />
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
                <el-option label="交卷后统计 (POST_STAT)" value="postStat" />
              </el-select>
            </el-form-item>

            <!-- 目标选择 (SHOW/HIDE/REQUIRED/CHECK 需要目标题目) -->
            <el-form-item v-if="['show','hide','required'].includes(ruleForm.action)" label="目标题目">
              <el-select v-model="ruleForm.targetQuestionIdx" placeholder="选择题目" style="width:100%">
                <el-option v-for="(q, qi) in questions" :key="qi" :label="`Q${toNumericIndex(qi) + 1}: ${q.title?.slice(0,30)}`" :value="toNumericIndex(qi)" />
              </el-select>
            </el-form-item>

            <!-- CHECK 需要目标题目+选项 -->
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

            <!-- BRANCH -->
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

            <!-- END -->
            <template v-if="ruleForm.action==='end'">
              <el-form-item label="从题目">
                <el-select v-model="ruleForm.branchFromIdx" placeholder="选择触发结束的题目" style="width:100%" :key="'endFrom'+showAddRule">
                  <el-option v-for="(q, qi) in questions" :key="qi" :label="`Q${toNumericIndex(qi) + 1}: ${q.title?.slice(0,30)}`" :value="toNumericIndex(qi)" />
                </el-select>
              </el-form-item>
            </template>

            <!-- ASSIGNMENT / VALIDATE / REPLACE 需要公式输入 -->
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

            <!-- POST_STAT 交卷后统计 -->
            <template v-if="ruleForm.action==='postStat'">
              <el-form-item label="统计字段">
                <el-radio-group v-model="ruleForm.statField">
                  <el-radio value="label">选项标签 (label)</el-radio>
                  <el-radio value="value">选项值 (value)</el-radio>
                </el-radio-group>
                <div style="font-size:11px;color:#999;margin-top:4px">选择统计时使用的字段，与「统计配置」中的统计方式联动</div>
              </el-form-item>
              <el-form-item label="统计范围">
                <el-radio-group v-model="ruleForm.statScope">
                  <el-radio value="all">全部统计（统计所有答卷）</el-radio>
                  <el-radio value="single">单卷统计（仅统计当前答卷）</el-radio>
                </el-radio-group>
              </el-form-item>
              <el-divider />
              <el-form-item label="通知方式">
                <el-radio-group v-model="ruleForm.notifyChannel">
                  <el-radio value="internal">站内通知</el-radio>
                  <el-radio value="webhook">Webhook</el-radio>
                  <el-radio value="both">全部</el-radio>
                </el-radio-group>
              </el-form-item>
              <template v-if="ruleForm.notifyChannel==='webhook'||ruleForm.notifyChannel==='both'">
                <el-form-item label="Webhook 类型">
                  <el-select v-model="ruleForm.webhookType" style="width:100%">
                    <el-option label="钉钉" value="dingtalk" />
                    <el-option label="企业微信" value="wecom" />
                    <el-option label="飞书" value="lark" />
                    <el-option label="自定义" value="custom" />
                  </el-select>
                </el-form-item>
                <el-form-item label="Webhook URL">
                  <el-input v-model="ruleForm.webhookUrl" :placeholder="webhookPlaceholder" clearable />
                  <div style="font-size:11px;color:#999;margin-top:4px">{{ webhookHint }}</div>
                </el-form-item>
              </template>
              <el-divider />
              <el-form-item label="通知消息模板">
                <div style="display:flex;gap:8px;margin-bottom:4px;align-items:center">
                  <el-select v-model="templatePreset" size="small" style="width:140px" placeholder="模版预设" @change="applyTemplatePreset" clearable>
                    <el-option-group label="内置模版">
                      <el-option v-for="p in allPresets.filter(x=>x.group==='builtin')" :key="p.value" :label="p.label" :value="p.value" />
                    </el-option-group>
                    <el-option-group label="自定义模版" v-if="customPresets.length">
                      <el-option v-for="p in allPresets.filter(x=>x.group==='custom')" :key="p.value" :label="p.label" :value="p.value">
                        <span>{{ p.label }}</span>
                        <span style="float:right;color:#999;font-size:11px;cursor:pointer" @click.stop="deleteCustomPreset(p.label)">删除</span>
                      </el-option>
                    </el-option-group>
                  </el-select>
                  <el-button text size="small" @click="showSavePresetDialog = true; newPresetName = ''">另存为</el-button>
                  <el-button text size="small" @click="resetTemplate" :disabled="!ruleForm.messageTemplate">恢复默认</el-button>
                </div>
                <el-input v-model="ruleForm.messageTemplate" type="textarea" :rows="5" placeholder="📋 问卷「{title}」收到新答卷&#10;提交人：{submitter}　时间：{date}&#10;共 {total} 份提交&#10;&#10;{result}" />
                <div style="font-size:11px;color:#999;margin-top:4px;line-height:1.8">
                  可用变量：
                  <code>{title}</code> 问卷标题 ·
                  <code>{questionCount}</code> 题目数 ·
                  <code>{total}</code> 总答卷数 ·
                  <code>{submitter}</code> 提交人 ·
                  <code>{date}</code> 提交时间 ·
                  <code>{result}</code> 统计结果（自动生成表格）
                  <br>支持插入 <code>\n</code> 换行，留空使用默认模版
                </div>
              </el-form-item>
              <el-divider />
              <el-form-item label="通知对象">
                <el-checkbox v-model="ruleForm.notifyAdmin" style="margin-bottom:8px;display:block">通知管理员</el-checkbox>
                <el-input v-model="ruleForm.notifyUserIds" placeholder="指定成员 ID（多个用逗号隔开）" clearable />
                <div style="font-size:11px;color:#999;margin-top:4px">通知管理员和指定成员为选填，仅用于站内通知</div>
              </el-form-item>
            </template>

            <!-- END 不需要额外设置 -->
          </el-form>
          <template #footer>
            <el-button @click="showAddRule=false">取消</el-button>
            <el-button type="primary" @click="confirmRule">确定</el-button>
          </template>
        </el-dialog>

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

        <!-- 右: 属性面板 -->
        <div class="survey-setting-panel" v-show="middleTab!=='logic' && middleTab!=='result'" @click.self="deselectQuestion">
          <div v-if="selected" class="props-panel">
            <!-- 矩阵行设置模式 -->
            <!-- 文件上传设置 -->
            <template v-if="selectedOptIdx>=0 && selected.type==='file'">
              <SurveyDesignerPanelTitle title="文件上传设置" :context="selectedQuestionContext" />
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
              <SurveyDesignerPanelTitle title="矩阵行设置" :context="selectedQuestionContext" />
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
              <SurveyDesignerPanelTitle :title="`${selected.type==='user'?'成员':'部门'}设置`" :context="selectedQuestionContext" />
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
              <SurveyDesignerPanelTitle :title="`选项设置 - ${stripHtml(selected.props.options[selectedOptIdx].label || '')}`" :context="selectedQuestionContext" />
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

            <!-- 多项/横向填空子字段设置 -->
            <template v-else-if="selectedOptIdx>=0 && (selected.type==='multiInput'||selected.type==='hInput') && selected.props?.fields?.[selectedOptIdx]">
              <SurveyDesignerPanelTitle :title="`字段设置 - ${selected.props.fields[selectedOptIdx].label}`" :context="selectedQuestionContext" />
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

            <!-- 评分/NPS 设置（独立面板，点击评分输入时弹出） -->
            <template v-else-if="selectedOptIdx>=0 && (selected.type==='rating'||selected.type==='nps')">
            <SurveyDesignerPanelTitle :title="`${selected.type==='rating'?'评分':'NPS'}设置`" :context="selectedQuestionContext" />
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
            <SurveyDesignerPanelTitle :title="`${typeName(selected.type)}设置`" :context="selectedQuestionContext" />
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
                  <el-button text size="small" type="danger" @click="removeAutoCol(toNumericIndex(ci))">×</el-button>
                </div>
                <el-button size="small" @click="addAutoCol" plain>+ 添加列</el-button>
                <el-row :gutter="8" style="margin-top:8px">
                  <el-col :span="12"><el-form-item label="最少行数"><el-input-number v-model="selected.props.minRows" :min="0" style="width:100%" /></el-form-item></el-col>
                  <el-col :span="12"><el-form-item label="最多行数"><el-input-number v-model="selected.props.maxRows" :min="0" style="width:100%" /></el-form-item></el-col>
                </el-row>
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
                  <el-button text size="small" type="primary" @click="selectOption(selected.id, toNumericIndex(i))" style="font-size:11px">编辑</el-button>
                  <el-button text size="small" type="danger" @click="removeUserOpt(toNumericIndex(i))">×</el-button>
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
            <SurveyDesignerPanelTitle title="问卷设置" :context="form.title || '未命名问卷'" />
            <div class="setting-row"><span class="setting-label">标题</span><el-input v-model="form.title" placeholder="问卷标题" style="flex:1" /></div>
            <div class="setting-row" style="align-items:flex-start"><span class="setting-label" style="margin-top:4px">描述</span><el-input v-model="form.description" type="textarea" :autosize="{ minRows: 3, maxRows: 8 }" placeholder="问卷描述" style="flex:1" /></div>
            <div class="setting-row"><span class="setting-label">模式</span><span>{{ form.mode === 'exam' ? '考试' : '问卷' }}</span></div>
            <el-divider />
            <div class="group-title" style="font-size:13px;font-weight:500;margin-bottom:8px">已应用图片</div>
            <div v-for="(item, i) in form.backgroundImages" :key="'bg-'+i" class="applied-image-row">
              <img :src="typeof item === 'string' ? item : item.url" class="applied-image-thumb" />
              <span class="applied-image-label">背景图</span>
              <button class="applied-image-remove" @click="removeAppliedImage('backgroundImages')">✕</button>
            </div>
            <div v-for="(item, i) in form.headerImages" :key="'hd-'+i" class="applied-image-row">
              <img :src="typeof item === 'string' ? item : item.url" class="applied-image-thumb" />
              <span class="applied-image-label">页眉图</span>
              <button class="applied-image-remove" @click="removeAppliedImage('headerImages')">✕</button>
            </div>
            <el-empty v-if="!form.backgroundImages.length && !form.headerImages.length" description="暂无已应用图片" :image-size="30" />
          </div>
          <div v-else class="props-panel props-empty-panel">
            <el-empty description="未选择题目" :image-size="60">
              <el-button size="small" type="primary" @click="openSurveySettings">问卷设置</el-button>
            </el-empty>
          </div>
        </div>
      </div>

      <!-- 设置视图 -->
      <div v-show="activeView==='setting'" class="setting-wrapper">
        <div class="setting-header">
          <div>
            <div class="setting-page-title">配置中心</div>
            <div class="setting-page-desc">管理问卷展示、回收规则、投放链接和协作人员</div>
          </div>
          <div class="setting-header-actions">
            <el-tag size="small" :type="saveStatusTagType">{{ saveStatusLabel }}</el-tag>
            <el-button type="primary" size="small" :loading="saving" @click="save">保存配置</el-button>
          </div>
        </div>
        <div class="setting-scroll">
          <!-- 问卷显示设置 -->
          <div class="setting-group">
            <div class="group-title">问卷显示设置</div>
            <el-form label-position="top">
              <div class="settings-grid">
                <el-form-item label="填写端自动暂存"><el-switch v-model="form.autoSave" /></el-form-item>
                <el-form-item label="显示题目序号"><el-switch v-model="form.questionNumber" /></el-form-item>
                <el-form-item label="显示进度条"><el-switch v-model="form.progressBar" /></el-form-item>
                <!--<el-form-item label="设置问卷默认答案"><el-switch v-model="form.defaultAnswer" /></el-form-item>-->
                <el-form-item label="一页一题"><el-switch v-model="form.onePageOneQuestion" /></el-form-item>
                <el-form-item label="显示答题卡"><el-switch v-model="form.answerSheetVisible" /></el-form-item>
                <!--<el-form-item label="允许复制题目"><el-switch v-model="form.copyEnabled" /></el-form-item>-->
              </div>
            </el-form>
          </div>
          <!-- 回收设置 -->
          <div class="setting-group">
            <div class="group-title">回收设置</div>
            <el-form label-position="top">
              <div class="settings-grid">
                <el-form-item class="settings-full" label="可见性">
                  <el-radio-group v-model="form.visibility" style="display:flex;flex-direction:column;gap:8px;align-items:flex-start">
                    <el-radio :value="0" border>公开链接</el-radio>
                    <el-radio :value="1" border>登录可见</el-radio>
                    <el-radio :value="2" border>部门限定</el-radio>
                  </el-radio-group>
                </el-form-item>
                <el-form-item v-if="form.visibility===2" class="settings-full" label="限定部门" :class="{ 'is-error': selectedDeptCount === 0 }">
                  <div class="setting-tree-box" :class="{ invalid: selectedDeptCount === 0 }">
                    <el-tree
                      ref="deptTreeRef"
                      :data="deptTreeData"
                      :props="{ label: 'name', children: 'children' }"
                      node-key="id"
                      show-checkbox
                      check-strictly
                      :default-checked-keys="deptCheckedKeys"
                      @check="onDeptCheck"
                    />
                  </div>
                  <div class="setting-help" :class="{ danger: selectedDeptCount === 0 }">
                    {{ selectedDeptCount ? `已选择 ${selectedDeptCount} 个部门` : '部门限定需要至少选择一个部门' }}
                  </div>
                </el-form-item>
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
                <!--<el-form-item label="每个账号答题次数"><el-input-number v-model="form.userLimit" :min="0" style="width:100%" /><span style="font-size:11px;color:#999;margin-left:4px">0=不限</span></el-form-item>-->
                <el-form-item label="允许多次填写"><el-switch v-model="form.allowMultiBool" :active-value="1" :inactive-value="0" active-text="允许" inactive-text="仅一次" /></el-form-item>
                <el-form-item label="开始时间"><el-date-picker v-model="form.startDate" type="datetime" placeholder="不限" value-format="x" style="width:100%" /></el-form-item>
                <el-form-item label="结束时间">
                  <el-date-picker v-model="form.endDate" type="datetime" placeholder="不限" value-format="x" style="width:100%" />
                  <div v-if="hasInvalidTimeRange" class="setting-help danger">结束时间不能早于开始时间</div>
                </el-form-item>
                <el-form-item label="回收上限">
                  <div class="setting-control-stack">
                    <el-input-number v-model="form.maxResponse" :min="0" style="width:100%" />
                    <span class="setting-help">0 表示不限</span>
                  </div>
                </el-form-item>
                <el-form-item label="作答时间（分钟）">
                  <div class="setting-control-stack">
                    <el-input-number v-model="form.timeLimit" :min="0" style="width:100%" />
                    <span class="setting-help">0 表示不限</span>
                  </div>
                </el-form-item>
              </div>
            </el-form>
          </div>
          <div v-if="form.mode==='exam'" class="setting-group">
            <div class="group-title">考试设置</div>
            <el-form label-position="top">
              <div class="settings-grid">
                <el-form-item label="显示排名"><el-switch v-model="form.examRankingEnabled" /></el-form-item>
                <el-form-item label="练习模式"><el-switch v-model="form.exerciseMode" /></el-form-item>
                <el-form-item label="随机顺序"><el-switch v-model="form.randomOrder" /></el-form-item>
                <el-form-item label="最短交卷(分钟)"><el-input-number v-model="form.minSubmitMinutes" :min="0" style="width:100%" /></el-form-item>
                <el-form-item label="最长交卷(分钟)"><el-input-number v-model="form.maxSubmitMinutes" :min="0" style="width:100%" /></el-form-item>
              </div>
            </el-form>
          </div>
          <!-- 投放与分享 -->
          <div class="setting-group">
            <div class="group-title">投放与分享</div>
            <el-form label-position="top">
              <el-form-item class="settings-full" label="提交后处理">
                <el-radio-group v-model="completionAction" class="completion-radio-group">
                  <el-radio-button value="default">默认完成页</el-radio-button>
                  <el-radio-button value="content">自定义页面</el-radio-button>
                  <el-radio-button value="redirect">跳转链接</el-radio-button>
                </el-radio-group>
              </el-form-item>
              <div class="settings-grid">
                <el-form-item v-if="completionAction === 'content'" label="自定义完成页"><el-input v-model="form.endContent" type="textarea" :rows="3" placeholder="答题完成后显示的 HTML 内容" /></el-form-item>
                <el-form-item v-if="completionAction === 'redirect'" label="跳转链接">
                  <el-input v-model="form.redirectUrl" placeholder="https://example.com 或站内路径" />
                  <div v-if="form.redirectUrl && !isValidRedirectUrl(form.redirectUrl)" class="setting-help danger">请输入 http(s) 链接或以 / 开头的站内路径</div>
                </el-form-item>
              </div>
              <el-form-item class="settings-full" label="填写模版">
                <el-select v-model="form.fillTemplate" style="width:100%">
                  <el-option v-for="(label, val) in fillTemplates" :key="val" :value="val" :label="label" />
                </el-select>
              </el-form-item>
              <el-form-item class="settings-full" label="问卷链接">
                <div style="display:flex;flex-direction:column;gap:8px;width:100%">
                  <el-input v-if="form.id" v-model="publicUrl" readonly><template #append><el-button @click="copyLink">复制</el-button></template></el-input>
                  <span v-else style="color:#999;">请先保存问卷</span>
                  <div v-if="form.id" style="display:flex;align-items:center;gap:12px">
                    <el-image v-if="qrUrl" :src="qrUrl" style="width:100px;height:100px;border-radius:4px;border:1px solid #eee" />
                    <div style="display:flex;gap:6px;align-items:center">
                      <el-button size="small" @click="genQR">生成二维码</el-button>
                      <el-button v-if="qrUrl" size="small" @click="downloadQR">下载二维码</el-button>
                    </div>
                  </div>
                </div>
              </el-form-item>
              <div class="settings-grid">
                <el-form-item label="问卷状态">
                  <el-switch v-model="form.statusBool" :active-value="1" :inactive-value="0" active-text="已开启" inactive-text="已停用" />
                </el-form-item>
              </div>
            </el-form>
          </div>
          <!-- 协作管理员 -->
          <div class="setting-group">
            <div class="group-title">协作管理员</div>
            <el-form label-position="top">
              <div class="settings-grid">
                <el-form-item label="问卷创建者"><el-input :model-value="creatorName" disabled placeholder="加载中..." /></el-form-item>
              </div>
              <el-form-item class="settings-full" label="协作管理员">
                <div style="width:100%;max-height:400px;overflow:auto;border:1px solid #dcdfe6;border-radius:4px;padding:8px">
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

      <!-- 概览视图 -->
      <!--<div v-show="activeView==='overview'" class="overview-wrapper">
        <div class="overview-scroll">
          <!-- 日收集问卷图表 -->
          <!--<div class="overview-card">
            <div class="overview-card-title">📊 日收集问卷</div>
            <div class="chart-placeholder">
              <div class="chart-bars">
                <div v-for="i in 7" :key="i" class="chart-bar-group">
                  <div class="chart-bar" :style="{ height: (20 + Math.random()*60) + 'px' }"></div>
                  <div class="chart-label">{{ ['一','二','三','四','五','六','日'][i-1] }}</div>
                </div>
              </div>
            </div>
          </div>-->

          <!-- 系统设置 -->
          <!--<div class="overview-card">
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
      </div>-->

      <!-- 数据视图 -->
      <div v-show="activeView==='data'" class="setting-wrapper">
        <div class="setting-scroll" style="display:block;padding:20px 24px">
          <div class="setting-group">
            <div class="group-title">数据统计</div>
            <div v-if="!form.id" class="save-tip">请先保存并发布问卷</div>
            <div v-else style="padding:12px">
              <div style="display:flex;gap:8px;margin-bottom:12px;flex-wrap:wrap;align-items:center">
                <el-button size="small" type="primary" @click="addResponse">添加数据</el-button>
                <el-button size="small" @click="editResponse">编辑数据</el-button>
                <el-button size="small" type="danger" @click="batchDeleteResponses">删除数据</el-button>
                <el-button size="small" @click="exportResponses">导出数据</el-button>
                <el-button size="small" @click="importResponses">导入数据</el-button>
                <el-button size="small" @click="showTrash = true">回收站</el-button>
                <el-button size="small" style="margin-left:auto" @click="refreshResponses" title="刷新数据">
                  <el-icon><Refresh /></el-icon>
                </el-button>
              </div>
              <div style="display:flex;gap:8px;margin-bottom:12px;align-items:center">
                <el-input v-model="responseKeyword" placeholder="搜索提交人..." size="small" clearable style="width:200px" @input="onSearch" />
              </div>
              <!--<el-row :gutter="12" style="margin-bottom:12px">
                <el-col :span="6"><el-statistic title="总提交数" :value="stats.totalCount" /></el-col>
                <el-col :span="6"><el-statistic title="今日提交" :value="stats.todayCount" /></el-col>
                <el-col :span="6"><el-statistic title="平均完成时间" :value="stats.avgTime" suffix="秒" /></el-col>
                <el-col :span="6"><el-statistic title="完成率" :value="stats.completionRate" suffix="%" /></el-col>
              </el-row>-->
              <el-table ref="dataTableRef" :data="responseRows" border size="small" style="width:100%" max-height="500px" @selection-change="onSelectionChange">
                <el-table-column type="selection" width="40" fixed />
                <el-table-column v-for="col in dataColumns" :key="col.key" :label="col.label" :width="col.width" :min-width="col.minWidth" :show-overflow-tooltip="col.showOverflowTooltip" :fixed="col.fixed">
                  <template #default="{ row }">
                    <span v-if="col.prop">{{ row[col.prop] ?? '-' }}</span>
                    <span v-else-if="col.key === 'nickname'">{{ row.nickname || '匿名' }}</span>
                    <span v-else-if="col.key === 'startTime'">{{ formatTime(row.startTime) }}</span>
                    <span v-else-if="col.key === 'submitTime'">{{ formatTime(row.submitTime) }}</span>
                    <span v-else-if="col.key === 'duration'">{{ row.duration ? row.duration + 's' : '-' }}</span>
                    <span v-else-if="col.key === 'ip'">{{ row.ip || '-' }}</span>
                    <span v-else-if="col.key === 'browser'">{{ row.browser || '-' }}</span>
                    <span v-else-if="col.key === 'deviceType'">{{ row.deviceType || '-' }}</span>
                    <span v-else-if="col.key === 'platformType'">{{ row.platformType || '-' }}</span>
                    <span v-else-if="col.key === 'isAutoSubmit'">
                      <el-tag v-if="row.isAutoSubmit === 1" size="small" type="warning">自动</el-tag>
                      <span v-else>-</span>
                    </span>
                    <span v-else-if="col.key === 'answer-'+col.qId">{{ formatAnswer(row.answers?.[col.qId], col.qType) }}</span>
                    <span v-else-if="col.key === 'actions'">
                      <el-button text size="small" type="danger" @click="deleteResponse(row.id)">删除</el-button>
                    </span>
                  </template>
                </el-table-column>
              </el-table>
              <div v-if="!responseRows.length" style="padding:40px 0;text-align:center;color:#aaa;font-size:14px">暂无提交数据</div>
              <div style="margin-top:12px;display:flex;justify-content:flex-start">
                 <el-pagination
                   size="small"
                  layout="total, sizes, prev, pager, next"
                  :total="responseTotal"
                  :page-size="responsePageSize"
                  :current-page="responsePage"
                  :page-sizes="[10, 20, 50, 100]"
                  @current-change="onPageChange"
                  @size-change="onPageSizeChange"
                />
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 报表视图 -->
      <div v-show="activeView==='report'" class="setting-wrapper">
        <div style="padding:20px 24px">
          <div class="setting-group">
            <div class="group-title">报表</div>
            <div v-if="!form.id" class="save-tip">请先保存并发布问卷</div>
            <div v-else-if="reportLoading" style="padding:40px;text-align:center;color:#999">加载中...</div>
            <div v-else-if="!reportData?.fieldStats?.length" style="padding:12px">
              <el-empty description="暂无数据" :image-size="50" />
            </div>
            <div v-else style="padding:12px;max-width:600px;margin:0 auto">
              <div v-for="(fs, qi) in (reportData.fieldStats||[])" :key="fs?.questionId||qi" style="margin-bottom:16px;border:1px solid #eee;border-radius:6px;padding:12px">
                <template v-if="fs && fs.totalCount > 0 && fs.dist">
                <div style="display:flex;align-items:center;gap:8px;margin-bottom:8px">
                  <span style="font-size:13px;font-weight:600;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;flex:1">{{ toNumericIndex(qi) + 1 }}. {{ firstLine(stripHtml(fs.title)) }}</span>
                  <span class="chart-type-btns">
                    <button :class="{ active: (chartTypes[fs.questionId]||'bar')==='bar' }" @click="chartTypes[fs.questionId]='bar'" title="条形图">≡</button>
                    <button :class="{ active: (chartTypes[fs.questionId]||'bar')==='column' }" @click="chartTypes[fs.questionId]='column'" title="柱形图">▯</button>
                    <button :class="{ active: (chartTypes[fs.questionId]||'bar')==='pie' }" @click="chartTypes[fs.questionId]='pie'" title="扇形图">◯</button>
                  </span>
                </div>
                  <div v-if="(chartTypes[fs.questionId]||'bar')==='bar'" style="display:flex;flex-direction:column;gap:4px">
                    <div v-for="(cnt, label) in (fs.dist||{})" :key="label" style="display:flex;align-items:center;gap:8px">
                      <span style="font-size:12px;width:100px;flex-shrink:0;overflow:hidden;text-overflow:ellipsis;white-space:nowrap">{{ label || '(空)' }}</span>
                      <div v-if="fs.totalCount > 0" style="flex:1;height:12px;background:#e8e8e8;border-radius:6px;overflow:hidden">
                        <div :style="{ width: Math.max(Math.round((cnt??0)/fs.totalCount*100),0) + '%', height: '100%', background: '#409eff', borderRadius: '6px', transition: 'width .3s' }"></div>
                      </div>
                      <span style="font-size:11px;color:#999;width:60px;text-align:right">{{ cnt }} / {{ fs.totalCount }}</span>
                    </div>
                  </div>
                  <div v-else-if="(chartTypes[fs.questionId]||'bar')==='column'" style="display:flex;align-items:flex-end;justify-content:center;gap:12px;padding:16px 0;min-height:160px">
                    <div v-for="(cnt, label) in fs.dist" :key="label" style="display:flex;flex-direction:column;align-items:center;gap:4px;flex:1;max-width:80px">
                      <span style="font-size:11px;color:#999">{{ cnt }}</span>
                      <div :style="{ height: Math.max(cnt/fs.totalCount*140, 4) + 'px', width: '100%', background: '#2563eb', borderRadius: '4px 4px 0 0', transition: 'height .3s' }"></div>
                      <span style="font-size:11px;color:#666;text-align:center;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;max-width:100%">{{ label || '(空)' }}</span>
                    </div>
                  </div>
                  <div v-else-if="(chartTypes[fs.questionId]||'bar')==='pie'" style="display:flex;justify-content:center;padding:12px 0">
                    <svg :viewBox="'0 0 160 160'" style="width:160px;height:160px;transform:rotate(-90deg)">
                      <circle cx="80" cy="80" r="70" fill="none" stroke="#f0f0f0" stroke-width="30" />
                      <template v-for="(cnt, label) in (Object.fromEntries(Object.entries(fs.dist||{})))" :key="label">
                        <circle cx="80" cy="80" r="70" fill="none" :stroke="PIE_COLORS[Object.keys(fs.dist||{}).indexOf(String(label)) % PIE_COLORS.length]" stroke-width="30"
                          :stroke-dasharray="pieDasharray(Number(cnt), fs.totalCount)"
                          stroke-dashoffset="0" style="transition: stroke-dasharray .3s" />
                      </template>
                    </svg>
                    <div style="display:flex;flex-direction:column;gap:4px;margin-left:16px;justify-content:center;font-size:12px">
                      <div v-for="(cnt, label) in fs.dist" :key="label" style="display:flex;align-items:center;gap:6px">
                        <span :style="{ width:10, height:10, borderRadius:'50%', background: PIE_COLORS[Object.keys(fs.dist).indexOf(String(label)) % PIE_COLORS.length], display:'inline-block' }"></span>
                        <span style="color:#666;max-width:80px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap">{{ label || '(空)' }}</span>
                        <span style="color:#999">{{ Math.round(Number(cnt)/fs.totalCount*100) }}%</span>
                      </div>
                    </div>
                  </div>
                </template>
                <template v-else-if="fs && fs.tableData?.length">
                  <div style="font-size:13px;font-weight:600;margin-bottom:6px">{{ toNumericIndex(qi) + 1 }}. {{ firstLine(stripHtml(fs.title)) }}</div>
                  <div style="max-height:200px;overflow:auto">
                    <el-table :data="fs.tableData" border size="small" style="width:100%">
                      <el-table-column type="index" label="#" width="40" fixed />
                      <el-table-column v-for="(col, ci) in (fs.tableCols||['回答'])" :key="ci" :label="col" min-width="100">
                        <template #default="{ row }">{{ stripHtml(row[ci] ?? '') }}</template>
                      </el-table-column>
                    </el-table>
                  </div>
                </template>
                <div v-else-if="fs.numericStat" style="font-size:12px;color:#606266;line-height:2">
                  总和: {{ fs.numericStat.sum }} | 平均: {{ fs.numericStat.avg.toFixed(2) }} | 最小: {{ fs.numericStat.min }} | 最大: {{ fs.numericStat.max }}
                </div>
                <div v-else style="color:#999;font-size:12px">暂无统计</div>
              </div>
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
                <li>点击右侧<strong style="color:#2563eb">蓝色高亮标签</strong>可插入题目标签到光标位置</li>
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
                  <el-tag size="small" type="primary" style="cursor:pointer" @click="insertFormulaTag(`Q${questions.indexOf(q)+1}`)">Q{{ questions.indexOf(q)+1 }}</el-tag>
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

      <!-- 保存自定义模版弹窗 -->
      <el-dialog v-model="showSavePresetDialog" title="保存自定义模版" width="360px" :close-on-click-modal="false">
        <el-form label-position="top" size="small">
          <el-form-item label="模版名称">
            <el-input v-model="newPresetName" placeholder="输入模版名称" maxlength="50" />
          </el-form-item>
        </el-form>
        <template #footer>
          <el-button @click="showSavePresetDialog=false">取消</el-button>
          <el-button type="primary" @click="confirmSavePreset">保存</el-button>
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
            <div style="font-size:13px;color:#666">拖拽或点击选择 .txt / .docx 文件</div>
          </el-upload>
          <div v-if="textImport.text" style="margin-top:8px;font-size:12px;color:#999">已解析 {{ textImport.text.split('\n').filter(l=>l.trim()).length }} 行</div>
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
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { onBeforeRouteLeave, useRoute, useRouter } from 'vue-router'
import { adminApi } from '../../api'
import request, { showRequestError } from '../../utils/request'
import { normalizeQuestions, importFromSurveyKing, exportToSurveyKing, isSurveyKingFormat, toSkType, toWcType } from '../../utils/surveyKingBridge'
import DraggableList from './formkit/DraggableList.vue'

import QuestionIcon from './formkit/QuestionIcon.vue'
import QuestionPreview from './formkit/QuestionPreview.vue'
import SurveyDesignerIssues from './components/SurveyDesignerIssues.vue'
import SurveyDesignerOutline from './components/SurveyDesignerOutline.vue'
import SurveyDesignerPanelTitle from './components/SurveyDesignerPanelTitle.vue'
import SurveyDesignerQuestionPalette from './components/SurveyDesignerQuestionPalette.vue'
import { FALLBACK_TYPES, getSurveyQuestionTypeName as typeName } from './survey-designer-question-types'
import { parseSurveyText, serializeSurveyText, stripSurveyMarkup as stripHtmlTag, type ParsedSurveyTextQuestion } from './survey-designer-text-transfer'
import { collectSurveyDesignerIssues, isValidRedirectUrl, type SurveyDesignerIssue } from './survey-designer-validation'
import { useSurveyDesignerHistory } from './composables/useSurveyDesignerHistory'
import './survey-designer-responsive.css'

const route = useRoute()
const router = useRouter()

function toNumericIndex(index: string | number): number {
  return Number(index)
}

const form = reactive<any>({
  id: 0, title: '', description: '', category: '', tags: '', mode: 'survey',
  visibility: 0, allowMultiBool: 0, anonymousBool: 0, showResultBool: 0,
  startDate: null, endDate: null, maxResponse: 0, statusBool: 1, deptIds: '',
  questionNumber: true, progressBar: false, autoSave: false,
  onePageOneQuestion: false, answerSheetVisible: true, copyEnabled: true,
  password: '', triggerType: 'onBlur', loginRequired: false,
  redirectUrl: '', endContent: '',
  examRankingEnabled: false, exerciseMode: false, randomOrder: false,
  minSubmitMinutes: 0, maxSubmitMinutes: 0,
  backgroundImages: [] as any[], headerImages: [] as any[],
  defaultAnswer: false, defaultLang: 'zh-CN',
  deviceLimit: 0, ipLimit: 0, userLimit: 0,
  publicQuery: false, showAnswerAnalysis: false,
  createBy: 0, collaborators: '', timeLimit: 0,
  fillTemplate: 'sf1',
  resultConfig: {
    statType: 'value',
    viewMode: 'aggregate',
    scope: ['all'],
    showChart: true,
    showDetail: true,
    exportField: 'value',
  },
})

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

const hasInvalidTimeRange = computed(() => {
  const start = Number(form.startDate || 0)
  const end = Number(form.endDate || 0)
  return start > 0 && end > 0 && end < start
})

const activeView = ref('edit')
const panelMode = ref<'edit' | 'json' | 'preview'>('edit')
const panelModeLabel = computed(() => ({ edit: '编辑模式', json: 'JSON模式', preview: '预览模式' }[panelMode.value]))
const middleTab = ref('item')
const showSurveySettings = ref(false)
const reportData = ref<any>(null)
const reportLoading = ref(false)
const chartTypes = ref<Record<string,string>>({})
const sideSubTab = ref('types')
const appearanceTab = ref('bg')
const allBgResources = ref<any[]>([])
const allHeaderResources = ref<any[]>([])
const saving = ref(false)
const bankKeyword = ref('')
const bankQuestions = ref<any[]>([])
const bankLoading = ref(false)
const bankLoadingVisible = ref(false)
let bankTimer: any = null
let bankLoadingTimer: ReturnType<typeof setTimeout> | null = null
let bankLoadSeq = 0

const bankCategories = ref<string[]>([])

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

const bankExpanded = ref<Record<string, boolean>>({})
function toggleBankExpand(key: string) {
  bankExpanded.value[key] = !bankExpanded.value[key]
}

const bankDialog = reactive({
  visible: false,
  qid: '',
  category: '',
  tags: '',
  saving: false
})

const textImport = reactive({
  visible: false,
  mode: 'paste' as 'paste' | 'file',
  text: '',
  parsed: [] as ParsedSurveyTextQuestion[],
  preview: [] as ParsedSurveyTextQuestion[],
  importing: false,
  surveyTitle: ''
})

watch(() => textImport.text, (val) => {
  if (textImport.mode === 'paste' && val) {
    const result = parseSurveyText(val)
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
  const result = parseSurveyText(textImport.text)
  textImport.surveyTitle = result.title
  textImport.parsed = result.questions
  textImport.preview = textImport.parsed.slice(0, 20)
}

function confirmTextImport() {
  if (!textImport.parsed.length) return
  textImport.importing = true
  try {
    if (textImport.surveyTitle) {
      form.title = textImport.surveyTitle
    }
    textImport.parsed.forEach(q => {
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

function exportSurvey() {
  const text = serializeSurveyText({ title: form.title, description: form.description, questions: questions.value })
  const blob = new Blob([text], { type: 'text/plain;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `${form.title || '问卷'}.txt`
  a.click()
  URL.revokeObjectURL(url)
  ElMessage.success('已导出')
}

const deptTreeRef = ref<any>(null)
const deptTreeData = ref<any[]>([])
const deptCheckedKeys = ref<number[]>([])
const selectedDeptCount = computed(() => parseDeptIds(form.deptIds).length)

function parseDeptIds(value: string) {
  return String(value || '').split(',').map(id => Number(id.trim())).filter(Boolean)
}

function onDeptCheck() {
  form.deptIds = deptTreeRef.value?.getCheckedKeys()?.join(',') || ''
}

async function loadDeptTree() {
  if (form.visibility !== 2) return
  try {
    const res: any = await adminApi.deptTree()
    deptTreeData.value = res.data || []
  } catch { deptTreeData.value = [] }
}

const creatorName = ref('')
const adminTreeRef = ref<any>(null)
const adminTreeData = ref<any[]>([])
const collaboratorCheckedKeys = ref<string[]>([])
const adminMap = ref<Record<number, string>>({})

function onAdminTreeCheck() {
  form.collaborators = adminTreeRef.value?.getCheckedKeys(true)?.filter((k: string) => k.startsWith('admin-'))?.map((k: string) => k.replace('admin-', ''))?.join(',') || ''
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
    collaboratorCheckedKeys.value = form.collaborators ? form.collaborators.split(',').map((id: string) => `admin-${id}`) : []
  } catch {
    creatorName.value = '加载失败'
    adminTreeData.value = []
  }
}

function buildAdminTree(depts: any[], mgrs: any[]): any[] {
  return depts.map((d: any) => {
    const deptAdmins = mgrs.filter((m: any) => (m.deptIds || []).includes(d.id))
    const adminNodes = deptAdmins.map((m: any) => ({ id: `admin-${m.id}`, label: m.name, type: 'admin' }))
    const children = d.children?.length ? buildAdminTree(d.children, mgrs) : []
    if (adminNodes.length) children.push(...adminNodes)
    return { id: `dept-${d.id}`, label: d.name, type: 'dept', children }
  }).filter((d: any) => d.children?.length)
}

watch(() => form.visibility, (v) => {
  if (v === 2) { loadDeptTree() }
})
watch(() => form.id, (v) => { if (v) { loadReport() } else { reportData.value = null } })
async function loadBank() {
  clearTimeout(bankTimer)
  if (bankLoadingTimer) {
    clearTimeout(bankLoadingTimer)
    bankLoadingTimer = null
  }
  bankLoadingVisible.value = false
  const seq = ++bankLoadSeq
  bankTimer = setTimeout(async () => {
    bankLoading.value = true
    bankLoadingTimer = setTimeout(() => {
      if (bankLoading.value && seq === bankLoadSeq) {
        bankLoadingVisible.value = true
      }
    }, 450)
    try {
      const res: any = await adminApi.surveyQuestionBankList({ keyword: bankKeyword.value, page: 1, pageSize: 100 })
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
function clonePlain<T>(value: T): T {
  return JSON.parse(JSON.stringify(value ?? {}))
}

function getBankQuestionSchema(item: any) {
  let parsed: any = {}
  try {
    parsed = typeof item.schema === 'string' ? JSON.parse(item.schema) : (item.schema || {})
  } catch {
    parsed = {}
  }
  const raw = Array.isArray(parsed?.questions) ? parsed.questions[0] : parsed
  if (!raw || typeof raw !== 'object' || Array.isArray(raw)) return {}
  return normalizeQuestions([raw])[0] || raw
}

function addFromBank(q: any) {
  const source: any = clonePlain(getBankQuestionSchema(q))
  const newQ: Question = {
    ...source,
    id: genId(),
    type: source.type || q.type || 'input',
    title: source.title || q.title || '未命名',
    required: source.required ?? false,
    readOnly: source.readOnly ?? false,
    dataType: source.dataType ?? '',
    placeholder: source.placeholder ?? '',
    mediaType: source.mediaType ?? '',
    mediaUrl: source.mediaUrl ?? '',
    mediaWidth: source.mediaWidth ?? '',
    mediaAlign: source.mediaAlign ?? 'center',
    showDescription: source.showDescription ?? !['description', 'questionSet', 'pagination', 'divider'].includes(source.type || q.type || 'input'),
    props: source.props ? clonePlain(source.props) : {}
  }
  delete (newQ as any).schema
  questions.value.push(newQ)
  selected.value = newQ
  selectedOptIdx.value = -1
}

async function onUploadBank(id: string) {
  const q = questions.value.find(x => x.id === id)
  if (!q) return
  bankDialog.qid = id
  bankDialog.category = ''
  bankDialog.tags = ''
  bankDialog.visible = true
  // 获取题库已有分类
  try {
    const res: any = await adminApi.surveyQuestionBankCategories()
    bankCategories.value = res.data || []
  } catch {}
}

async function confirmUploadBank() {
  const q = questions.value.find(x => x.id === bankDialog.qid)
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
    await adminApi.formkitSaveToBank(payload)
    ElMessage.success('已上传到题库')
    bankDialog.visible = false
    loadBank()
  } catch (error) {
    showRequestError(error, '上传失败')
  } finally {
    bankDialog.saving = false
  }
}

function bankQuestionTitle(q: any) {
  const title = firstLine(stripHtml(String(q?.title || ''))).replace(/\s+/g, ' ').trim()
  return title || '未命名'
}

const collapseActive = ref('')
const mediaCollapseActive = ref('')
const newRowName = ref('')
const newColName = ref('')
const optDrawerIdx = ref(0)
const optDrawerVisible = ref(false)
const batchAddVisible = ref(false)
const batchAddText = ref('')

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
interface LogicRuleItem { id: string; conditionType: string; conditions: LogicCondition[]; action: string; scope: 'frontend' | 'backend'; targetQuestionIdx?: number; targetOptionIdxs?: number[]; branchFromIdx?: number; branchToIdx?: number; branchToEnd?: boolean; formula?: string; statField?: string; statScope?: string; notifyAdmin?: boolean; notifyUserIds?: string; notifyChannel?: string; webhookType?: string; webhookUrl?: string; messageTemplate?: string }
const logicRuleList = ref<LogicRuleItem[]>([])
const showAddRule = ref(false)
const editingRuleIdx = ref(-1)
const defaultRuleForm = (): LogicRuleItem => ({ id: '', conditionType: 'simple', conditions: [{ questionIdx: undefined, optionIdx: undefined, operator: undefined, compareValue: undefined }], action: 'show', scope: 'frontend', targetQuestionIdx: undefined, targetOptionIdxs: [], branchFromIdx: undefined, branchToIdx: undefined, branchToEnd: false, formula: '', statField: 'label', statScope: 'all', notifyAdmin: false, notifyUserIds: '', notifyChannel: 'internal', webhookType: 'dingtalk', webhookUrl: '', messageTemplate: '' })
const ruleForm = ref<LogicRuleItem>(defaultRuleForm())

const webhookPlaceholder = computed(() => {
  const t = ruleForm.value.webhookType || 'dingtalk'
  const map: Record<string, string> = {
    dingtalk: 'https://oapi.dingtalk.com/robot/send?access_token=xxx',
    wecom: 'https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=xxx',
    lark: 'https://open.feishu.cn/open-apis/bot/v2/hook/xxx',
    custom: 'https://your-domain.com/webhook/xxx'
  }
  return map[t] || ''
})
const webhookHint = computed(() => {
  const t = ruleForm.value.webhookType || 'dingtalk'
  const map: Record<string, string> = {
    dingtalk: '在钉钉群中添加机器人获取 Webhook 地址',
    wecom: '在企业微信群中添加机器人获取 Webhook 地址',
    lark: '在飞书群中添加机器人获取 Webhook 地址',
    custom: '通用 Webhook，POST JSON 格式消息体'
  }
  return map[t] || ''
})

const templatePreset = ref('')
const customPresets = ref<{ label: string; value: string }[]>([])
async function loadCustomPresets() {
  try {
    const res = await adminApi.surveyTemplatePresetsGet()
    if (res.code === 0 && Array.isArray(res.data)) {
      customPresets.value = res.data
    }
  } catch {}
}
loadCustomPresets()

const defaultBuiltinPresets: Record<string, string> = {
  '简洁': '📋 问卷「{title}」收到新答卷\n新提交人：{submitter}\n时间：{date}\n共 {total} 份提交\n\n{result}',
  '详细': '📊 问卷统计报告\n━━━━━━━━━━━━━━━━━━\n📌 问卷：{title}\n👤 新提交人：{submitter}\n🕐 时间：{date}\n📈 总提交数：{total}\n\n{result}\n━━━━━━━━━━━━━━━━━━',
  '仅统计结果': '{result}'
}
const builtinPresets = ref<Record<string, string>>({...defaultBuiltinPresets})
async function loadBuiltinPresets() {
  try {
    const res = await request.get<string | { label: string; value: string }[]>('/api/v2/home/setup', { params: { key: 'BUILTIN_TEMPLATE_PRESETS' } })
    if (res.code === 0 && res.data) {
      const parsed = typeof res.data === 'string' ? JSON.parse(res.data) : res.data
      if (Array.isArray(parsed) && parsed.length) {
        const obj: Record<string, string> = {}
        for (const p of parsed) {
          obj[p.label] = p.value
        }
        builtinPresets.value = obj
      }
    }
  } catch {}
}
loadBuiltinPresets()

const allPresets = computed(() => {
  const items: { label: string; value: string; group: string }[] = []
  for (const key of Object.keys(builtinPresets.value)) {
    items.push({ label: key, value: key, group: 'builtin' })
  }
  for (const c of customPresets.value) {
    items.push({ label: c.label, value: 'custom:' + c.label, group: 'custom' })
  }
  return items
})

const showSavePresetDialog = ref(false)
const newPresetName = ref('')

async function saveCustomPresets() {
  try {
    await adminApi.surveyTemplatePresetsSave({ presets: customPresets.value })
  } catch {}
}
function confirmSavePreset() {
  const name = newPresetName.value.trim()
  if (!name) { ElMessage.warning('请输入模版名称'); return }
  if (!ruleForm.value.messageTemplate) { ElMessage.warning('请先填写模版内容'); return }
  const existing = customPresets.value.findIndex(c => c.label === name)
  const entry = { label: name, value: ruleForm.value.messageTemplate }
  if (existing >= 0) {
    customPresets.value[existing] = entry
  } else {
    customPresets.value.push(entry)
  }
  saveCustomPresets()
  showSavePresetDialog.value = false
  ElMessage.success('已保存自定义模版')
}
function deleteCustomPreset(label: string) {
  customPresets.value = customPresets.value.filter(c => c.label !== label)
  saveCustomPresets()
}

function applyTemplatePreset(val: string) {
  if (!val) return
  if (val.startsWith('custom:')) {
    const label = val.slice(7)
    const c = customPresets.value.find(p => p.label === label)
    if (c) ruleForm.value.messageTemplate = c.value
  } else if (builtinPresets.value[val]) {
    ruleForm.value.messageTemplate = builtinPresets.value[val]
  }
}
function resetTemplate() {
  ruleForm.value.messageTemplate = ''
}

function buildSettingsPayload() {
  return {
    questionNumber: form.questionNumber,
    progressBar: form.progressBar,
    autoSave: form.autoSave,
    password: form.password,
    loginRequired: form.loginRequired,
    onePageOneQuestion: form.onePageOneQuestion,
    answerSheetVisible: form.answerSheetVisible,
    copyEnabled: form.copyEnabled,
    triggerType: form.triggerType,
    redirectUrl: form.redirectUrl,
    endContent: form.endContent,
    examRankingEnabled: form.examRankingEnabled,
    exerciseMode: form.exerciseMode,
    randomOrder: form.randomOrder,
    minSubmitMinutes: form.minSubmitMinutes,
    maxSubmitMinutes: form.maxSubmitMinutes,
    backgroundImages: form.backgroundImages,
    headerImages: form.headerImages,
    defaultAnswer: form.defaultAnswer,
    defaultLang: form.defaultLang,
    deviceLimit: form.deviceLimit,
    ipLimit: form.ipLimit,
    userLimit: form.userLimit,
    publicQuery: form.publicQuery,
    showAnswerAnalysis: form.showAnswerAnalysis,
    collaborators: form.collaborators,
    timeLimit: form.timeLimit,
    fillTemplate: form.fillTemplate,
    resultConfig: form.resultConfig,
    logicRules: JSON.stringify(logicRuleList.value)
  }
}

function buildSurveyPayload(schema: string, settings: string) {
  return {
    title: form.title,
    description: form.description,
    category: form.category,
    tags: form.tags,
    visibility: form.visibility,
    allowMulti: form.allowMultiBool,
    anonymous: form.anonymousBool,
    showResult: form.showResultBool,
    startTime: form.startDate || 0,
    endTime: form.endDate || 0,
    maxResponse: form.maxResponse,
    status: form.statusBool,
    schema,
    deptIds: form.deptIds,
    mode: form.mode,
    settings
  }
}

function normalizeCompletionSettings() {
  form.title = String(form.title || '').trim()
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

function validateSurveySettings(showMessage = true) {
  const blockingCodes = new Set(['survey_title_required', 'survey_department_required', 'survey_time_range_invalid', 'survey_redirect_url_invalid'])
  const issue = designerIssues.value.find(item => blockingCodes.has(item.code))
  if (issue && showMessage) { ElMessage.warning(issue.message); locateDesignerIssue(issue) }
  return !issue
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
  const actionMap: Record<string, string> = { show: 'SHOW', hide: 'HIDE', required: 'REQUIRED', check: 'CHECK', branch: 'BRANCH', assignment: 'ASSIGNMENT', validate: 'VALIDATE', replace: 'REPLACE', end: 'BRANCH', postStat: 'POST_STAT' }
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
  } else if (rule.action==='postStat') {
    const field = rule.statField === 'value' ? 'VALUE' : 'LABEL'
    const channel = rule.notifyChannel || 'internal'
    const parts = [`POST_STAT ${field} CHANNEL(${channel})`]
    if (rule.webhookUrl) parts.push(`WEBHOOK(${rule.webhookType || 'dingtalk'}:${rule.webhookUrl})`)
    if (rule.notifyAdmin) parts.push('NOTIFY_ADMIN')
    if (rule.notifyUserIds) parts.push(`USERS(${rule.notifyUserIds})`)
    if (rule.messageTemplate) parts.push(`TEMPLATE("${rule.messageTemplate.replace(/"/g, '\\"')}")`)
    actionStr = parts.join(' ')
  } else {
    const tTag = rule.targetQuestionIdx!==undefined ? `Q${rule.targetQuestionIdx+1}` : '?'
    actionStr = `${a} ${tTag}`
  }

  return rule.conditionType==='none' ? actionStr : `IF ${condStr} THEN ${actionStr}`
}

async function confirmRule() {
  const rf = ruleForm.value
  if (rf.conditionType!=='none') {
    for (const c of rf.conditions) {
      if (c.questionIdx===undefined) { ElMessage.warning('请选择条件题目'); return }
    }
  }
  if (!rf.action) { ElMessage.warning('请选择动作'); return }
  if (rf.action==='postStat') {
    if (!rf.statField) { ElMessage.warning('请选择统计字段'); return }
    if (!rf.notifyChannel) { ElMessage.warning('请选择通知方式'); return }
    if ((rf.notifyChannel==='webhook'||rf.notifyChannel==='both') && !rf.webhookUrl) { ElMessage.warning('请填写 Webhook URL'); return }
    if ((rf.notifyChannel==='webhook'||rf.notifyChannel==='both') && !rf.webhookType) { ElMessage.warning('请选择 Webhook 类型'); return }
  }
  // Validate targets
  if (['show','hide','required'].includes(rf.action) && rf.targetQuestionIdx===undefined) { ElMessage.warning('请选择目标题目'); return }
  if (rf.action==='check' && rf.targetQuestionIdx===undefined) { ElMessage.warning('请选择目标题目'); return }
  if (rf.action==='branch' && rf.branchFromIdx===undefined) { ElMessage.warning('请选择跳转来源题目'); return }
  if (rf.action==='branch' && !rf.branchToEnd && rf.branchToIdx===undefined) { ElMessage.warning('请选择跳转目标题目或勾选结束问卷'); return }
  if (rf.action==='end' && rf.branchFromIdx===undefined) { ElMessage.warning('请选择结束来源题目'); return }
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
  if (form.id) await saveLogicRules(false)
}

function addNewRule() {
  editingRuleIdx.value = -1
  ruleForm.value = defaultRuleForm()
  showAddRule.value = true
}

function editRule(idx: number) {
  editingRuleIdx.value = idx
  ruleForm.value = JSON.parse(JSON.stringify(logicRuleList.value[idx]))
  showAddRule.value = true
}

function removeRule(idx: number) {
  ElMessageBox.confirm('确定删除该条规则？', '提示', { type: 'warning', confirmButtonText: '删除', cancelButtonText: '取消' }).then(() => {
    logicRuleList.value.splice(idx, 1)
    saveLogicRules(true)
  }).catch(() => {})
}

async function saveLogicRules(showMsg = true) {
  if (!form.id) { if (showMsg) ElMessage.warning('请先保存问卷'); return }
  if (!validateSurveySettings(showMsg)) return
  normalizeCompletionSettings()
  saving.value = true
  const schema = JSON.stringify({ version: '2.0', questions: questions.value })
  const settings = JSON.stringify(buildSettingsPayload())
  const payload: any = { id: form.id, ...buildSurveyPayload(schema, settings) }
  try {
    const r: any = await adminApi.surveyEdit(payload)
    if (!form.id) { form.id = r.id || r.data?.id; router.replace({ query: { id: String(form.id) } }) }
    if (showMsg) ElMessage.success('已保存')
  } catch (error) { if (showMsg) showRequestError(error, '保存失败') }
  finally { saving.value = false }
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
const responseRows = ref<any[]>([])
const responseTotal = ref(0)
const responsePage = ref(1)
const responsePageSize = ref(20)
const responseKeyword = ref('')
const responseSource = ref<any[]>([])
const dataTableRef = ref<any>(null)
const dragColKey = ref('')
const dataBaseOrder = ref<string[]>(['nickname', 'startTime', 'submitTime', 'duration', 'ip', 'browser', 'deviceType', 'platformType', 'isAutoSubmit'])

const dataColumnKeys = computed(() => {
  const keys = [...dataBaseOrder.value]
  questions.value.forEach((q: any) => { if (!keys.includes('answer-'+q.id)) keys.push('answer-'+q.id) })
  keys.push('actions')
  return keys
})

const dataColumns = computed(() => {
  const cols: any[] = []
  for (const key of dataColumnKeys.value) {
    if (key.startsWith('answer-')) {
      const qId = key.replace('answer-', '')
      const q = questions.value.find((x: any) => x.id === qId)
      if (!q) continue
      cols.push({ key, label: stripHtml(q.title), minWidth: 120, showOverflowTooltip: true, qId, qType: q.type })
    } else {
      const defs: Record<string, any> = {
        nickname: { key: 'nickname', label: '提交人', width: 90 },
        startTime: { key: 'startTime', label: '开始时间', width: 150 },
        submitTime: { key: 'submitTime', label: '提交时间', width: 150 },
        duration: { key: 'duration', label: '答题时长', width: 80 },
        ip: { key: 'ip', label: 'IP地址', width: 120 },
        browser: { key: 'browser', label: '浏览器', width: 100 },
        deviceType: { key: 'deviceType', label: '设备类型', width: 90 },
        platformType: { key: 'platformType', label: '平台类型', width: 90 },
        isAutoSubmit: { key: 'isAutoSubmit', label: '自动提交', width: 80 },
        actions: { key: 'actions', label: '操作', width: 80, fixed: 'right' },
      }
      if (defs[key]) cols.push(defs[key])
    }
  }
  return cols
})

function setupColDrag() {
  nextTick(() => {
    const wrapper = dataTableRef.value?.$el?.querySelector('.el-table__header-wrapper')
    if (!wrapper) return
    const ths = wrapper.querySelectorAll('th:not(.el-table__cell--selection)')
    ths.forEach((th: any, i: number) => {
      th.draggable = true
      const idx = i
      th.addEventListener('dragstart', () => { dragColKey.value = dataColumnKeys.value[idx] })
      th.addEventListener('dragover', (e: DragEvent) => {
        if (dragColKey.value && dragColKey.value !== dataColumnKeys.value[idx]) {
          e.preventDefault()
          const key = dragColKey.value
          const arr = [...dataBaseOrder.value]
          const from = arr.indexOf(dragColKey.value)
          const toIdx = dataColumnKeys.value.indexOf(dataColumnKeys.value[idx])
          if (from >= 0) { arr.splice(from, 1); arr.splice(Math.min(toIdx, arr.length), 0, key); dataBaseOrder.value = arr }
        }
      })
      th.addEventListener('dragend', () => { dragColKey.value = '' })
    })
  })
}

async function loadResponses(page = 1) {
  if (!form.id) return
  responsePage.value = page
  try {
    const keyword = responseKeyword.value.trim()
    const res: any = await adminApi.surveyResponseList({ surveyId: form.id, page, pageSize: responsePageSize.value, keyword })
    const list: any[] = res.data?.list || []
    responseTotal.value = res.data?.total || 0
    responseSource.value = list.map((r: any) => {
      let answers: any = r.answers || {}
      if (typeof answers === 'string') { try { answers = JSON.parse(answers) } catch { answers = {} } }
      return {
        id: r.id, nickname: r.nickname, submitTime: r.submitTime, startTime: r.startTime,
        duration: r.duration, ip: r.ip,
        browser: r.browser || '-', deviceType: r.deviceType || '-', platformType: r.platformType || '-',
        isAutoSubmit: r.isAutoSubmit,
        answers
      }
    })
    stats.totalCount = responseTotal.value
    responseRows.value = responseSource.value
    setupColDrag()
  } catch { responseRows.value = [] }
}

function onSearch() { responsePage.value = 1; loadResponses(1) }

function onPageChange(page: number) { loadResponses(page) }
function onPageSizeChange(size: number) { responsePageSize.value = size; loadResponses(1) }

function formatTime(ts: number) {
  if (!ts) return '-'
  return new Date(ts).toLocaleString('zh-CN', { hour12: false })
}

function formatDateStr(s: string) {
  if (!s) return '-'
  const d = new Date(s)
  if (isNaN(d.getTime())) return s
  return d.toLocaleDateString('zh-CN', { year: 'numeric', month: '2-digit', day: '2-digit' }).replace(/\//g, '-')
}

function formatTimeStr(s: string) {
  if (!s) return '-'
  const d = new Date(s)
  if (isNaN(d.getTime())) return s
  return d.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit', second: '2-digit', hour12: false })
}

function formatAnswer(v: any, qType: string) {
  if (v == null || v === '') return '-'
  if (qType === 'date') return formatDateStr(v)
  if (qType === 'time') return formatTimeStr(v)
  if (qType === 'dateRange') {
    if (Array.isArray(v)) return v.map((d: any) => formatDateStr(d)).join(' ~ ')
    return formatDateStr(v)
  }
  if (Array.isArray(v)) return v.join(', ')
  if (typeof v === 'object') return JSON.stringify(v)
  return String(v)
}

function stripHtml(html: string) {
  const d = document.createElement('div')
  d.innerHTML = html
  return d.textContent || d.innerText || ''
}
function firstLine(s: string) { return s.split('\n')[0] || s }
const PIE_COLORS = ['#2563eb','#06b6d4','#22c55e','#f59e0b','#8b5cf6','#14b8a6','#64748b','#f97316']
function pieDasharray(cnt: number, total: number) {
  const ratio = cnt / total
  const circumference = 2 * Math.PI * 70
  return `${ratio * circumference} ${circumference - ratio * circumference}`
}

async function deleteResponse(id: number) {
  try {
    await adminApi.surveyResponseDel({ id })
    responseSource.value = responseSource.value.filter((r: any) => r.id !== id)
    responseTotal.value--
    stats.totalCount = responseTotal.value
    responseRows.value = responseSource.value
  } catch (error) { showRequestError(error, '删除失败') }
}

const showTrash = ref(false)
const selectedResponseIds = ref<number[]>([])

function onSelectionChange(rows: any[]) { selectedResponseIds.value = rows.map((r: any) => r.id) }

function addResponse() { ElMessage.info('添加数据功能待实现') }
function editResponse() { ElMessage.info('编辑数据功能待实现') }
function batchDeleteResponses() {
  if (!selectedResponseIds.value.length) { ElMessage.warning('请先选择数据'); return }
  ElMessageBox.confirm(`确认删除 ${selectedResponseIds.value.length} 条数据?`, '提示', { type: 'warning' }).then(async () => {
    for (const id of selectedResponseIds.value) {
      try { await adminApi.surveyResponseDel({ id }) } catch {}
    }
    selectedResponseIds.value = []
    loadResponses(responsePage.value)
    ElMessage.success('已删除')
  }).catch(() => {})
}
function exportResponses() { ElMessage.info('导出数据功能待实现') }
function importResponses() { ElMessage.info('导入数据功能待实现') }
function refreshResponses() { loadResponses(responsePage.value); ElMessage.success('已刷新') }

async function loadReport() {
  if (!form.id) { reportData.value = null; return }
  reportLoading.value = true
  try {
    const res: any = await adminApi.surveyStatistic(form.id)
    reportData.value = res.data || res
  } catch { reportData.value = null }
  finally { reportLoading.value = false }
}

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
interface Question { id: string; type: string; title: string; description?: string; required: boolean; placeholder?: string; props?: any; validate?: any[]; logic?: any[]; calcValue?: any; readOnly?: boolean; dataType?: string; examScore?: number; examCorrectAnswer?: string; examAnalysis?: string; mediaType?: string; mediaUrl?: string; mediaWidth?: string; mediaAlign?: string; [key: string]: any }
const questions = ref<Question[]>([])
const selected = ref<Question | null>(null)
const selectedQuestionContext = computed(() => {
  if (!selected.value) return ''
  const index = questions.value.findIndex(question => question.id === selected.value?.id)
  const title = firstLine(stripHtmlTag(String(selected.value.title || ''))).replace(/\s+/g, ' ').trim() || '未命名'
  return `${index >= 0 ? `第 ${index + 1} 项` : '当前项'} · ${title}`
})

let idCounter = 0

function syncIdCounterFromQuestions(list: Question[] = questions.value) {
  idCounter = list.reduce((max, q) => {
    const match = String(q.id || '').match(/^q(\d+)$/)
    return match ? Math.max(max, Number(match[1])) : max
  }, 0)
}

function genId() {
  let id = ''
  do {
    idCounter++
    id = 'q' + idCounter
  } while (questions.value.some(q => q.id === id))
  return id
}

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
    q.multiple = false
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
function scrollQuestionIntoView(id: string) {
  nextTick(() => {
    const cards = Array.from(document.querySelectorAll<HTMLElement>('[data-survey-question-id]'))
    const target = cards.find(el => el.dataset.surveyQuestionId === id)
    target?.scrollIntoView({ behavior: 'smooth', block: 'center', inline: 'nearest' })
  })
}

function selectQuestion(id: string, shouldScroll = false) {
  selected.value = questions.value.find(q => q.id === id) || null
  selectedOptIdx.value = -1
  showSurveySettings.value = false
  if (shouldScroll && selected.value) {
    panelMode.value = 'edit'
    scrollQuestionIntoView(id)
  }
}
function selectOption(qId: string, optIdx: number) {
  selected.value = questions.value.find(q => q.id === qId) || null
  selectedOptIdx.value = optIdx
  showSurveySettings.value = false
  if (selected.value) {
    if (!selected.value.props) selected.value.props = {}
    if (selected.value.type === 'user' || selected.value.type === 'dept') {
      if (!Array.isArray(selected.value.props.options)) selected.value.props.options = []
      while (selected.value.props.options.length <= optIdx) {
        selected.value.props.options.push(selected.value.type === 'user'
          ? { label: '', value: '', deptName: '', deptId: '', parentDeptId: '' }
          : { label: '', value: '', parentId: '' })
      }
    }
  }
}
function onQuestionsUpdate(newQ: Question[]) {
  replaceQuestions(newQ)
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
  replaceQuestions(questions.value.filter(q => q.id !== sel.id))
  selected.value = questions.value.length ? questions.value[questions.value.length - 1] : null
  selectedOptIdx.value = -1
}
async function clearAll() {
  if (!questions.value.length) return
  try {
    await ElMessageBox.confirm(`确定清空全部 ${questions.value.length} 题？`, '确认', { confirmButtonText: '清空', cancelButtonText: '取消', type: 'warning' })
  } catch { return }
  questions.value = []
  logicRuleList.value = []
  selected.value = null
  selectedOptIdx.value = -1
  idCounter = 0
}

function hasOptions(q: Question) { return ['select','radio','checkbox','picker','matrixRadio','matrixCheckbox','judge','cascade'].includes(q.type) }
function hasDataType(q: Question) { return ['input','textarea','number','name','studentId','employeeId','class','phone','email','idCard','password','date','time','dateRange','switch','location'].includes(q.type) }
const LAYOUT_PURE = ['divider','description','pagination']
const LAYOUT_ALL = ['divider','description','pagination','questionSet']
const answerableQuestionCount = computed(() => questions.value.filter(q => !LAYOUT_ALL.includes(q.type)).length)
const CHOICE_TYPES = ['select','radio','checkbox','picker','cascade','judge']
const INPUT_TYPES = ['input','textarea','number','multiInput','hInput']
const MATRIX_QR = ['matrixRadio','matrixCheckbox','matrixFillBlank']
const MATRIX_ALL = ['matrixRadio','matrixCheckbox','matrixFillBlank','matrixAuto']
const importFormatHint = computed(() => {
  if (selected.value?.type === 'user') return 'JSON 格式: [{"label":"姓名","value":"ID","deptName":"部门","deptId":"部门ID","parentDeptId":"父级部门ID"}]'
  if (selected.value?.type === 'dept') return 'JSON 格式: [{"label":"部门名称","value":"部门ID","parentId":"父级部门ID"}]'
  return ''
})
const PERSONAL_TYPES = ['name','studentId','employeeId','class','phone','email','idCard','password','date','time','dateRange','switch','location']
function isPureLayout(t: string) { return LAYOUT_PURE.includes(t) }
function isLayoutAll(t: string) { return LAYOUT_ALL.includes(t) }
function isChoice(t: string) { return CHOICE_TYPES.includes(t) }
function isInput(t: string) { return INPUT_TYPES.includes(t) }
function isMatrixQR(t: string) { return MATRIX_QR.includes(t) }
function isMatrixAll(t: string) { return MATRIX_ALL.includes(t) }
function isPersonal(t: string) { return PERSONAL_TYPES.includes(t) }

function remapQuestionIndex(idx: number | undefined, oldQuestions: Question[], nextIndexById: Map<string, number>) {
  if (idx === undefined) return undefined
  const oldId = oldQuestions[idx]?.id
  if (!oldId) return undefined
  return nextIndexById.get(oldId)
}

function isLogicRuleUsable(rule: LogicRuleItem) {
  if (rule.conditionType !== 'none' && (rule.conditions || []).some(cond => cond.questionIdx === undefined)) return false
  if (['show','hide','required','check','assignment','validate','replace'].includes(rule.action) && rule.targetQuestionIdx === undefined) return false
  if (['branch','end'].includes(rule.action) && rule.branchFromIdx === undefined) return false
  if (rule.action === 'branch' && !rule.branchToEnd && rule.branchToIdx === undefined) return false
  return true
}

function syncLogicRulesAfterQuestionChange(oldQuestions: Question[], nextQuestions: Question[]) {
  if (!logicRuleList.value.length) return
  const nextIndexById = new Map(nextQuestions.map((q, idx) => [q.id, idx]))
  const remappedRules = logicRuleList.value.map((rule) => {
    const nextRule: LogicRuleItem = {
      ...rule,
      conditions: (rule.conditions || []).map((cond) => ({
        ...cond,
        questionIdx: remapQuestionIndex(cond.questionIdx, oldQuestions, nextIndexById)
      }))
    }
    nextRule.targetQuestionIdx = remapQuestionIndex(rule.targetQuestionIdx, oldQuestions, nextIndexById)
    nextRule.branchFromIdx = remapQuestionIndex(rule.branchFromIdx, oldQuestions, nextIndexById)
    nextRule.branchToIdx = rule.branchToEnd ? undefined : remapQuestionIndex(rule.branchToIdx, oldQuestions, nextIndexById)
    return nextRule
  })
  const nextRules = remappedRules.filter(isLogicRuleUsable)
  const removed = remappedRules.length - nextRules.length
  logicRuleList.value = nextRules
  if (removed > 0) {
    ElMessage.warning(`已移除 ${removed} 条关联失效题目的逻辑规则`)
  }
}

function replaceQuestions(nextQuestions: Question[]) {
  const oldQuestions = [...questions.value]
  questions.value.splice(0, questions.value.length, ...nextQuestions)
  syncLogicRulesAfterQuestionChange(oldQuestions, questions.value)
}

function removeQuestionById(id: string) {
  replaceQuestions(questions.value.filter(q => q.id !== id))
  if (selected.value?.id === id) {
    selected.value = null
    selectedOptIdx.value = -1
  }
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

const fillTemplates: Record<string, string> = { sf: '默认模版', sf1: '新样式模版' }
const publicUrl = computed(() => form.id ? `${window.location.origin}/${form.fillTemplate || 'sf'}/${form.id}` : '')
const qrUrl = computed(() => publicUrl.value ? `https://api.qrserver.com/v1/create-qr-code/?size=200x200&data=${encodeURIComponent(publicUrl.value)}` : '')
const embedCode = computed(() => form.id ? `<iframe src="${publicUrl.value}" width="100%" height="600" frameborder="0"></iframe>` : '')

const exportedJson = ref('')

function genQR() {
  if (!form.id) {
    ElMessage.warning('请先保存问卷')
    return
  }
  ElMessage.success('二维码已生成')
}
function copyLink() { navigator.clipboard.writeText(publicUrl.value); ElMessage.success('已复制') }
function downloadQR() {
  fetch(qrUrl.value)
    .then(r => r.blob())
    .then(blob => {
      const url = URL.createObjectURL(blob)
      const link = document.createElement('a')
      link.href = url
      link.download = `问卷二维码_${form.id}.png`
      link.click()
      URL.revokeObjectURL(url)
    })
    .catch(() => ElMessage.error('下载失败'))
}
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
    replaceQuestions(converted)
    syncIdCounterFromQuestions()
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
  if (item.id) {
    try { await adminApi.surveyResourceDelete({ id: item.id }) } catch {}
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
function removeAppliedImage(field: string) {
  (form as any)[field] = []
  save()
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
watch(() => activeView.value, (v) => { if (v === 'data') loadResponses() })

let autoSaveTimer: ReturnType<typeof setTimeout> | null = null
let dataLoaded = false
let savePending = false
const designerLoading = ref(true)
const designerLoadError = ref(false)
const saveFailed = ref(false)
const savedSnapshot = ref('')
const designerIssues = computed(() => collectSurveyDesignerIssues({ form, questions: questions.value }))
const invalidQuestionIds = computed(() => Array.from(new Set(designerIssues.value.filter(issue => issue.severity === 'error' && issue.questionId).map(issue => issue.questionId!))))
const designerSnapshot = computed(() => JSON.stringify(buildSurveyPayload(
  JSON.stringify({ version: '2.0', questions: questions.value }),
  JSON.stringify(buildSettingsPayload()),
)))
const historySnapshot = computed(() => JSON.stringify({ questions: questions.value, logicRules: logicRuleList.value }))
const hasUnsavedChanges = computed(() => dataLoaded && designerSnapshot.value !== savedSnapshot.value)
const saveStatusLabel = computed(() => saving.value ? '保存中' : saveFailed.value ? '保存失败' : hasUnsavedChanges.value ? '未保存' : form.id ? '已保存' : '尚未保存')
const saveStatusClass = computed(() => saveFailed.value ? 'is-error' : hasUnsavedChanges.value ? 'is-dirty' : saving.value ? 'is-saving' : 'is-saved')
const saveStatusTagType = computed<'success' | 'warning' | 'danger' | 'info'>(() => saveFailed.value ? 'danger' : hasUnsavedChanges.value ? 'warning' : form.id ? 'success' : 'info')

const designerHistory = useSurveyDesignerHistory({
  capture: () => historySnapshot.value,
  restore: (snapshot) => {
    const state = JSON.parse(snapshot) as { questions: Question[]; logicRules: LogicRuleItem[] }
    const selectedId = selected.value?.id
    questions.value.splice(0, questions.value.length, ...normalizeQuestions(state.questions))
    logicRuleList.value = state.logicRules
    syncIdCounterFromQuestions()
    selected.value = questions.value.find(question => question.id === selectedId) || null
  },
})
const { canUndo, canRedo } = designerHistory
watch(historySnapshot, snapshot => designerHistory.record(snapshot))

function scheduleAutoSave() {
  if (!form.id || !dataLoaded) return
  saveFailed.value = false
  if (autoSaveTimer) clearTimeout(autoSaveTimer)
  autoSaveTimer = setTimeout(() => { void save(false) }, 800)
}
watch(designerSnapshot, scheduleAutoSave)

function undoDesignerChange() { designerHistory.undo() }
function redoDesignerChange() { designerHistory.redo() }
function locateDesignerIssue(issue: SurveyDesignerIssue) {
  const question = issue.questionId ? questions.value.find(item => item.id === issue.questionId) : questions.value[issue.questionIndex ?? -1]
  if (question) {
    activeView.value = 'edit'; middleTab.value = 'item'; sideSubTab.value = 'outline'; panelMode.value = 'edit'
    selected.value = question; selectedOptIdx.value = -1; showSurveySettings.value = false
    if (question.id) scrollQuestionIntoView(question.id)
    return
  }
  if (issue.section === 'edit') { activeView.value = 'edit'; panelMode.value = 'edit'; openSurveySettings() }
  else activeView.value = 'setting'
}
function showShortcutHelp() {
  ElMessageBox.alert('Ctrl/Cmd + S：保存问卷\nCtrl/Cmd + Z：撤销\nCtrl/Cmd + Shift + Z 或 Ctrl/Cmd + Y：重做', '快捷键', { confirmButtonText: '知道了' })
}
function isTextEditingTarget(target: EventTarget | null) {
  return target instanceof HTMLElement && Boolean(target.closest('input, textarea, [contenteditable="true"]'))
}
function handleDesignerKeydown(event: KeyboardEvent) {
  if (!event.metaKey && !event.ctrlKey) return
  const key = event.key.toLowerCase()
  if (key === 's') { event.preventDefault(); void save(); return }
  if (isTextEditingTarget(event.target)) return
  if (key === 'z') { event.preventDefault(); event.shiftKey ? redoDesignerChange() : undoDesignerChange() }
  else if (key === 'y') { event.preventDefault(); redoDesignerChange() }
}
function handleBeforeUnload(event: BeforeUnloadEvent) {
  if (!hasUnsavedChanges.value) return
  event.preventDefault()
  event.returnValue = ''
}
onBeforeRouteLeave(async () => {
  if (!hasUnsavedChanges.value) return true
  try {
    await ElMessageBox.confirm('当前问卷还有未保存的修改，离开后将丢失这些内容。', '确认离开', { confirmButtonText: '离开', cancelButtonText: '继续编辑', type: 'warning' })
    return true
  } catch { return false }
})

async function load() {
  const id = Number(route.query.id || 0)
  designerLoading.value = true
  designerLoadError.value = false
  if (!id) {
    dataLoaded = true
    savedSnapshot.value = designerSnapshot.value
    designerHistory.reset(historySnapshot.value)
    designerLoading.value = false
    return
  }
  try {
    const res: any = await adminApi.surveyDetail(id)
    const sv = res.data.survey
    const base: any = {
      id: sv.id, title: sv.title, description: sv.description,
      category: sv.category, tags: sv.tags, visibility: sv.visibility,
      allowMultiBool: sv.allowMulti, anonymousBool: sv.anonymous, showResultBool: sv.showResult,
      startDate: sv.startTime || null, endDate: sv.endTime || null,
      maxResponse: sv.maxResponse, statusBool: sv.status, deptIds: sv.deptIds,
      mode: sv.mode || 'survey',
      createBy: sv.createBy || 0
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
          transcriptVisible: s.transcriptVisible ?? true,
          rankVisible: s.rankVisible ?? false, redirectUrl: s.redirectUrl || '',
          endContent: s.endContent || '',
          examRankingEnabled: s.examRankingEnabled ?? false,
          exerciseMode: s.exerciseMode ?? false, randomOrder: s.randomOrder ?? false,
          minSubmitMinutes: s.minSubmitMinutes || 0, maxSubmitMinutes: s.maxSubmitMinutes || 0,
          backgroundImages: s.backgroundImages || [], headerImages: s.headerImages || [],
          defaultAnswer: s.defaultAnswer ?? false, defaultLang: s.defaultLang || 'zh-CN',
          deviceLimit: s.deviceLimit || 0, ipLimit: s.ipLimit || 0,
          userLimit: s.userLimit || 0,
          publicQuery: s.publicQuery ?? false, showAnswerAnalysis: s.showAnswerAnalysis ?? false,
          collaborators: s.collaborators || '', timeLimit: s.timeLimit || 0,
          fillTemplate: s.fillTemplate || 'sf1',
          resultConfig: Object.assign({ statType: 'value', viewMode: 'aggregate', scope: ['all'], showChart: true, showDetail: true, exportField: 'value' }, s.resultConfig || {}),
          logicRules: s.logicRules || '[]'
        })
        try { logicRuleList.value = JSON.parse(s.logicRules || '[]') } catch { logicRuleList.value = [] }
      } catch {}
    } else {
      Object.assign(base, {
        questionNumber: true, progressBar: false, autoSave: false, password: '',
        loginRequired: false, onePageOneQuestion: false, answerSheetVisible: true,
        copyEnabled: true, triggerType: 'onBlur',
        transcriptVisible: true, rankVisible: false, redirectUrl: '', endContent: '',
        examRankingEnabled: false, exerciseMode: false, randomOrder: false,
        minSubmitMinutes: 0, maxSubmitMinutes: 0,
  backgroundImages: [] as string[], headerImages: [] as string[],
        defaultAnswer: false, defaultLang: 'zh-CN',
        deviceLimit: 0, ipLimit: 0, userLimit: 0,
  publicQuery: false,
        collaborators: '', timeLimit: 0,
        fillTemplate: 'sf1',
        resultConfig: { statType: 'value', viewMode: 'aggregate', scope: ['all'], showChart: true, showDetail: true, exportField: 'value' }
      })
    }
    Object.assign(form, base)
    syncCompletionActionFromForm()
    deptCheckedKeys.value = form.deptIds ? form.deptIds.split(',').map(Number).filter(Boolean) : []
    loadAdminTree()
    loadResources()
    const rawSchema = res.data?.schema || ''
    if (rawSchema) {
      try {
        const sch = JSON.parse(rawSchema)
        if (sch.questions) {
          questions.value = normalizeQuestions(sch.questions)
          syncIdCounterFromQuestions()
        }
      } catch {}
    }
    await nextTick()
    dataLoaded = true
    savedSnapshot.value = designerSnapshot.value
    designerHistory.reset(historySnapshot.value)
    saveFailed.value = false
  } catch (error) {
    designerLoadError.value = true
    showRequestError(error, '加载失败')
  } finally { designerLoading.value = false }
}

function retryLoad() { void load() }

async function save(showMessage: boolean | Event = true) {
  const shouldShowMessage = typeof showMessage === 'boolean' ? showMessage : true
  if (saving.value) { savePending = true; return false }
  if (!validateSurveySettings(shouldShowMessage)) return false
  normalizeCompletionSettings()
  if (autoSaveTimer) clearTimeout(autoSaveTimer)
  autoSaveTimer = null
  const snapshotToSave = designerSnapshot.value
  const creating = !form.id
  saveFailed.value = false
  saving.value = true
  try {
    const schema = JSON.stringify({ version: '2.0', questions: questions.value })
    const settings = JSON.stringify(buildSettingsPayload())
    const payload: any = buildSurveyPayload(schema, settings)
    if (form.id) {
      payload.id = form.id
      await adminApi.surveyEdit(payload)
      if (shouldShowMessage) ElMessage.success('已更新')
    } else {
      const r: any = await adminApi.surveyInsert(payload)
      form.id = r.data?.id || r.id
      if (shouldShowMessage) ElMessage.success('已创建')
      router.replace({ query: { id: String(form.id) } })
    }
    savedSnapshot.value = snapshotToSave
    if (creating) await loadAdminTree()
    return true
  } catch (error) {
    saveFailed.value = true
    if (shouldShowMessage) showRequestError(error, '保存失败')
    return false
  }
  finally {
    const shouldRetry = savePending
    savePending = false
    saving.value = false
    if (shouldRetry && form.id && hasUnsavedChanges.value) queueMicrotask(() => { void save(false) })
  }
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
  window.addEventListener('keydown', handleDesignerKeydown)
  window.addEventListener('beforeunload', handleBeforeUnload)
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

onBeforeUnmount(() => {
  window.removeEventListener('keydown', handleDesignerKeydown)
  window.removeEventListener('beforeunload', handleBeforeUnload)
  if (autoSaveTimer) clearTimeout(autoSaveTimer)
  clearTimeout(bankTimer)
  if (bankLoadingTimer) clearTimeout(bankLoadingTimer)
  designerHistory.dispose()
})
</script>

<style scoped src="./survey-designer.css"></style>
