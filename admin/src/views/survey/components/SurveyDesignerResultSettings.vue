<template>
  <div class="logic-panel logic-full-panel">
    <div class="logic-toolbar"><h4>问卷统计配置</h4></div>
    <div class="settings-tip">配置问卷提交后的结果统计方式，包括统计字段、图表展示、数据导出等。</div>
    <div class="logic-body">
      <div class="logic-editor-area">
        <el-form label-position="top" size="small">
          <el-form-item label="统计方式">
            <el-radio-group v-model="form.resultConfig.statType">
              <el-radio value="value">按选项值统计</el-radio>
              <el-radio value="count">按选项计数统计</el-radio>
              <el-radio value="score">按分值统计（仅考试模式）</el-radio>
            </el-radio-group>
            <div class="field-help">
              <template v-if="form.resultConfig.statType === 'value'">统计时使用选项设置的「选项值」字段进行计算</template>
              <template v-else-if="form.resultConfig.statType === 'count'">统计每个选项被选择的次数</template>
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
          <el-form-item label="结果显示"><div class="switch-list"><el-switch v-model="form.resultConfig.showChart" active-text="显示图表" inactive-text="隐藏图表" /><el-switch v-model="form.resultConfig.showDetail" active-text="显示详细数据" inactive-text="隐藏详细数据" /></div></el-form-item>
          <el-form-item label="导出时选项字段">
            <el-radio-group v-model="form.resultConfig.exportField"><el-radio value="value">选项值 (value)</el-radio><el-radio value="label">选项文本 (label)</el-radio><el-radio value="both">值+文本 (value / label)</el-radio></el-radio-group>
          </el-form-item>
        </el-form>
      </div>
      <aside class="logic-sidebar">
        <strong>配置说明</strong>
        <div class="description-list">
          <div><b>按选项值统计：</b>适用于选项有数值含义的题目（如评分、分级），使用选项的 value 字段参与计算。</div>
          <div><b>按选项计数统计：</b>统计每个选项被选中次数，适用于普通选择题。</div>
          <div><b>按分值统计：</b>基于题目设置的分值（考试模式）计算总分和平均分。</div>
          <div><b>总览统计：</b>所有答卷合并，以图表/表格展示每题的整体分布。</div>
          <div><b>逐份统计：</b>每份答卷独立展示，便于查看每个提交者的具体作答。</div>
          <div><b>导出字段：</b>控制 CSV/Excel 导出时使用选项值还是选项文本。</div>
        </div>
      </aside>
    </div>
    <div class="panel-actions"><el-button type="primary" size="small" :loading="saving" @click="emit('save')">保存配置</el-button><el-button size="small" plain @click="emit('openReport')">查看统计报表</el-button></div>
  </div>
</template>

<script setup lang="ts">
defineProps<{ form: Record<string, any>; saving: boolean }>()
const emit = defineEmits<{ save: []; openReport: [] }>()
</script>

<style scoped>
.logic-full-panel { display:flex; flex:1; flex-direction:column; overflow:hidden; background:var(--designer-bg); padding:24px 32px; }
.logic-toolbar { display:flex; align-items:center; justify-content:space-between; margin-bottom:14px; flex-shrink:0; padding:12px 14px; border:1px solid var(--designer-border); border-radius:8px; background:#fff; box-shadow:0 8px 18px rgba(15,23,42,.04); }
.logic-toolbar h4 { margin:0; font-size:14px; }
.settings-tip { font-size:12px; color:#606266; background:#e8f4fd; border:1px solid #b3d8f0; border-radius:4px; padding:8px 12px; margin-bottom:12px; line-height:1.8; }
.logic-body { display:flex; gap:16px; flex:1; overflow:hidden; }
.logic-editor-area { flex:1; display:flex; flex-direction:column; overflow-y:auto; min-width:0; }
.logic-sidebar { width:260px; flex-shrink:0; overflow-y:auto; background:#fff; border-radius:8px; padding:14px; border:1px solid var(--designer-border); line-height:1.8; box-shadow:0 8px 18px rgba(15,23,42,.04); font-size:12px; }
.logic-sidebar > strong { display:block; color:#606266; margin-bottom:6px; }
.description-list { line-height:2.2; }
.field-help { font-size:11px; color:#999; margin-top:4px; }
.switch-list { display:flex; flex-direction:column; gap:10px; }
.panel-actions { padding:12px; border-top:1px solid #eee; display:flex; gap:8px; justify-content:center; }
</style>
