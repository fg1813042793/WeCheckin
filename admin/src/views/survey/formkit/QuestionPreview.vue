<template>
  <div class="question-preview" :class="{ editing, 'is-hidden': q.defaultHidden }">
    <!-- 题目标题 -->
    <div v-show="editing" class="edit-title-line">
      <el-tag size="small" :type="tagType(q.type)||undefined" style="flex-shrink:0;margin-top:5px" :class="{'required-tag': q.required}">{{ typeName(q.type) }}</el-tag>
      <div class="title-editor-wrap" @click.stop>
        <QuillEditor v-model:content="q.title" content-type="html"
          :options="titleEditorOptions"
          placeholder="输入标题"
          @ready="onTitleEditorReady" />
      </div>
      <span class="preview-actions">
        <el-tooltip content="逻辑设置" placement="top">
          <el-button text size="small" @click.stop="$emit('open-logic')">
            <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="3"/><path d="M12 1v2M12 21v2M4.22 4.22l1.42 1.42M18.36 18.36l1.42 1.42M1 12h2M21 12h2M4.22 19.78l1.42-1.42M18.36 5.64l1.42-1.42"/></svg>
          </el-button>
        </el-tooltip>
        <el-tooltip content="复制" placement="top">
          <el-button text size="small" @click.stop="$emit('copy')">
            <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><rect x="9" y="9" width="13" height="13" rx="2"/><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/></svg>
          </el-button>
        </el-tooltip>
        <el-tooltip content="删除" placement="top">
          <el-button text size="small" type="danger" @click.stop="$emit('remove')">
            <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
          </el-button>
        </el-tooltip>
        <el-tooltip content="上传题库" placement="top">
          <el-button text size="small" @click.stop="$emit('upload-bank')">
            <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/></svg>
          </el-button>
        </el-tooltip>
      </span>
    </div>
    <div v-show="!editing" class="preview-header">
      <el-tag size="small" :type="tagType(q.type)||undefined" class="preview-type-tag" :class="{'required-tag': q.required}">{{ typeName(q.type) }}</el-tag>
      <div class="preview-title" v-html="q.title"></div>
    </div>

    <!-- 说明 -->
    <div v-if="editing && q.showDescription === true" class="edit-desc-line">
      <div
        ref="descRef"
        class="desc-editable"
        contenteditable
        placeholder="添加说明"
        @blur="onDescBlur"
      >{{ q.description }}</div>
    </div>
    <div v-else-if="q.description && q.showDescription === true" class="preview-desc">{{ q.description }}</div>

    <!-- 媒体 -->
    <div v-if="q.mediaUrl" class="preview-media" :style="{ textAlign: q.mediaAlign || 'center' }">
      <img v-if="q.mediaType==='image'" :src="q.mediaUrl" :style="{ maxWidth: q.mediaWidth ? q.mediaWidth+'px' : '100%', maxHeight:'300px' }" class="media-preview" />
      <video v-else-if="q.mediaType==='video'" :src="q.mediaUrl" controls :style="{ maxWidth: q.mediaWidth ? q.mediaWidth+'px' : '100%', maxHeight:'300px' }" class="media-preview" />
      <audio v-else-if="q.mediaType==='audio'" :src="q.mediaUrl" controls class="media-preview" style="width:100%;max-width:400px" />
    </div>

    <!-- 选项画布编辑 (选择题) -->
      <div v-if="editing && isChoiceType" class="edit-options-area" :style="optionGrid(q)">
      <div v-for="(o, i) in (q.props?.options||[])" :key="o.value || i" class="edit-option-row" @click.stop="emit('select-option', i)">
        <span class="option-icon">{{ fieldIcon }}</span>
        <div class="option-editor-wrap">
          <QuillEditor v-model:content="o.label" content-type="html"
            :options="optionEditorOptions"
            placeholder="输入选项"
            @ready="(quill: any) => onOptionEditorReady(quill, i)"
            @click.stop />
        </div>
        <el-button text size="small" type="danger" class="opt-del-btn" @click="removeOption(i)">
          <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.5"><circle cx="12" cy="12" r="9"/><line x1="8" y1="12" x2="16" y2="12"/></svg>
        </el-button>
        <el-tag v-if="o.value===examCorrectAnswer" size="small" type="success" effect="dark" style="margin-left:4px;">✓</el-tag>
      </div>
    </div>
    <el-button v-if="editing && isChoiceType" text size="small" type="primary" @click="addOption" class="add-option-btn">添加选项</el-button>

    <!-- 矩阵题行/列画布编辑 -->
    <div v-if="editing && isMatrixType" class="edit-options-area" style="border:1px solid #d0d0d0;border-radius:6px;padding:0">
      <div style="overflow-x:auto">
        <div :style="{display:'grid',gridTemplateColumns:matrixGridCols,gap:0,padding:0}">
          <div style="padding:6px 4px;font-size:12px;color:#999;text-align:center;border:1px solid #d0d0d0;background:#f0f2f5;min-width:0"> </div>
          <div style="padding:6px 8px;font-size:12px;color:#999;border:1px solid #d0d0d0;background:#f0f2f5;font-weight:500;border-left:none;min-width:0">行\\列</div>
          <div v-for="(col, ci) in (q.props?.columns||[])" :key="ci" style="padding:4px 6px;border:1px solid #d0d0d0;background:#eef1f6;display:flex;align-items:center;gap:2px;justify-content:center;white-space:nowrap;border-left:none;min-width:0">
            <div style="font-size:12px;color:#666;padding:2px 4px;outline:none;cursor:text;min-width:40px;text-align:center;border:1px solid #d0d0d0;border-radius:3px;background:#fff" contenteditable @blur="e => onMatrixColLabelBlur(ci, e)" @keydown.enter.prevent="(e:any)=>e.target.blur()">{{ typeof col==='string'?col:(col.title||col.label||'') }}</div>
            <el-button text size="small" type="danger" style="font-size:12px;padding:0 2px;min-height:0" @click.stop="removeMatrixCol(ci)">×</el-button>
          </div>
          <div v-if="q.type!=='matrixAuto'" style="border:1px solid #d0d0d0;background:#f0f2f5;border-left:none;min-width:0"></div>
        </div>
        <template v-if="q.type==='matrixAuto' && (!q.props?.rows || !q.props.rows.length)">
          <div class="matrix-edit-row" :style="{display:'grid',gridTemplateColumns:matrixGridCols,gap:0,padding:0,cursor:'default',opacity:0.6}">
            <span style="padding:6px 4px;font-size:11px;color:#999;text-align:center;border:1px solid #e8e8e8;border-top:none;background:#fafafa;min-width:0">1.</span>
            <div style="padding:6px 8px;border:1px solid #e8e8e8;border-top:none;border-left:none;font-size:12px;color:#999;min-width:0">示例行</div>
            <div v-for="col in (q.props?.columns||[])" :key="col.id||col.label||col" style="padding:4px;border:1px solid #e8e8e8;border-top:none;border-left:none;min-width:0"><el-input disabled size="small" :placeholder="col.label||'值'" style="pointer-events:none" /></div>
          </div>
        </template>
        <div v-for="(r, ri) in (q.props?.rows||[])" :key="ri" class="matrix-edit-row" @click.stop="emit('select-option', ri)" :style="{display:'grid',gridTemplateColumns:matrixGridCols,gap:0,padding:0,cursor:'pointer'}">
          <span style="padding:6px 4px;font-size:11px;color:#999;text-align:center;border:1px solid #e8e8e8;border-top:none;background:#fafafa;min-width:0">{{ ri+1 }}.</span>
          <div style="padding:6px 8px;border:1px solid #e8e8e8;border-top:none;border-left:none;font-size:12px;color:#606266;outline:none;min-width:0" class="option-editable" contenteditable @blur="e => onMatrixRowLabelBlur(ri, e)" @keydown.enter.prevent="(e:any)=>e.target.blur()">{{ typeof r==='string'?r:r.title }}</div>
          <template v-if="q.type==='matrixAuto'">
            <div v-for="col in (q.props?.columns||[])" :key="col.id||col.label||col" style="padding:4px;border:1px solid #e8e8e8;border-top:none;border-left:none;min-width:0"><el-input disabled size="small" :placeholder="col.label||'值'" style="pointer-events:none" /></div>
          </template>
          <template v-else-if="q.type!=='matrixFillBlank'">
            <div v-for="col in (q.props?.columns||[])" :key="col.id||col.title||col" class="matrix-col-label" style="padding:6px 8px;border:1px solid #e8e8e8;border-top:none;border-left:none;background:#fafafa;font-size:12px;color:#666;text-align:center;min-width:0">{{ typeof col==='string'?col:(col.title||col.label||'') }}</div>
          </template>
          <template v-else>
            <div v-for="col in (q.props?.columns||[])" :key="col.id||col.title||col" style="padding:4px;border:1px solid #e8e8e8;border-top:none;border-left:none;min-width:0"><el-input disabled size="small" placeholder="填空" style="pointer-events:none" /></div>
          </template>
          <el-button v-if="q.type!=='matrixAuto'" text size="small" type="danger" class="opt-del-btn" @click.stop="removeMatrixRow(ri)" style="padding:4px 6px;border:1px solid #e8e8e8;border-top:none;border-left:none;min-width:0">
            <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.5"><circle cx="12" cy="12" r="9"/><line x1="8" y1="12" x2="16" y2="12"/></svg>
          </el-button>
        </div>
      </div>
      <div class="matrix-edit-add" style="display:flex;align-items:center;gap:8px;padding:6px 8px;border-top:1px solid #e0e0e0">
        <el-button text size="small" type="primary" @click="addMatrixRow">+ 添加行</el-button>
        <template v-if="q.type==='matrixAuto'">
          <el-button text size="small" type="danger" @click="removeMatrixLastRow">
            <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="9"/><line x1="8" y1="12" x2="16" y2="12"/></svg>
            删除行
          </el-button>
        </template>
        <span style="flex:1"></span>
        <el-button text size="small" type="primary" @click="addMatrixCol">+ 添加列</el-button>
      </div>
    </div>

    <!-- 单项填空/多行文本 → 直接渲染输入框 -->
    <div v-if="editing && isSingleInput" class="edit-options-area" style="padding:4px 8px;cursor:pointer" @click.stop="emit('select-option', 0)">
      <div style="pointer-events:none">
        <el-input v-if="q.type==='input'" size="small" :placeholder="(q.props?.options?.[0]?.placeholder)||'请输入'" disabled />
        <el-input v-else size="small" type="textarea" :placeholder="(q.props?.options?.[0]?.placeholder)||'请输入'" :rows="2" disabled />
      </div>
    </div>

    <!-- 评分 / NPS 输入（点击弹出评分设置） -->
    <div v-if="editing && (q.type==='rating'||q.type==='nps')" class="edit-options-area" style="padding:8px;cursor:pointer" @click.stop="emit('select-option', 0)">
      <div style="pointer-events:none;display:flex;align-items:center;gap:8px">
        <span v-if="q.type==='rating'" class="rate-icons-preview" v-for="i in (q.props?.maxRating || 5)" :key="i" v-html="rateIconSvg(q.props?.icon || 'star')"></span>
        <template v-else>
          <span style="font-size:11px;color:#909399">0</span>
          <el-rate :model-value="0" disabled :max="10" />
          <span style="font-size:11px;color:#909399">10</span>
        </template>
      </div>
    </div>

    <!-- 文件上传（画布中显示上传框） -->
    <div v-if="editing && q.type==='file'" class="edit-options-area" style="padding:8px;cursor:pointer" @click.stop="emit('select-option', 0)">
      <div class="file-upload-placeholder" style="border:1px dashed #d9d9d9;border-radius:6px;padding:16px 12px;text-align:center;color:#999;pointer-events:none">
        <svg viewBox="0 0 1024 1024" width="24" height="24" fill="currentColor" style="display:block;margin:0 auto 6px"><path d="M854.6 288.6L639.4 73.4c-6-6-14.1-9.4-22.6-9.4H192c-17.7 0-32 14.3-32 32v832c0 17.7 14.3 32 32 32h640c17.7 0 32-14.3 32-32V311.3c0-8.5-3.4-16.7-9.4-22.7z"/></svg>
        <span>上传文件</span>
        <div style="font-size:11px;margin-top:4px">点击设置上传参数</div>
      </div>
    </div>

    <!-- 成员/部门预览 -->
    <div v-if="editing && (q.type==='user'||q.type==='dept')" class="edit-options-area" style="padding:8px">
      <div style="display:flex;flex-wrap:wrap;gap:4px;cursor:pointer">
        <el-tag v-for="(o, i) in (q.props?.options||[])" :key="i" size="small" effect="plain" style="font-size:12px;cursor:pointer" @click.stop="emit('select-option', i)">{{ o.label || (q.type==='user'?'成员':'部门') }}</el-tag>
      </div>
    </div>

    <!-- 其他填空题子字段预览 -->
    <div v-if="editing && isInputSubFieldType && !isSingleInput && q.type!=='file' && q.type!=='signature' && q.type!=='scanCode'" class="edit-options-area">
      <div v-for="(o, i) in (q.props?.options||[])" :key="i" class="edit-option-row input-field-row" @click.stop="emit('select-option', i)">
        <span class="option-icon">▸</span>
        <span class="input-field-label">{{ o.label || '字段' }}</span>
        <el-input size="small" :placeholder="o.placeholder||'请输入'" disabled style="flex:1;max-width:200px;pointer-events:none" />
        <el-tag v-if="o.calculate" size="small" type="warning" effect="plain" style="margin:0 4px">公式</el-tag>
        <el-tag v-if="o.dataType" size="small" effect="plain" style="margin:0 4px">{{ o.dataType }}</el-tag>
        <el-button text size="small" type="danger" class="opt-del-btn" @click.stop="removeOption(i)">
          <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.5"><circle cx="12" cy="12" r="9"/><line x1="8" y1="12" x2="16" y2="12"/></svg>
        </el-button>
      </div>
    </div>
    <el-button v-if="editing && isInputSubFieldType && !isSingleInput && q.type!=='file' && q.type!=='signature' && q.type!=='scanCode'" text size="small" type="primary" @click="addOption" class="add-option-btn">添加字段</el-button>

    <!-- 子字段画布编辑 (multiInput / hInput) -->
    <div v-if="editing && hasFields" class="edit-options-area">
      <div v-for="(f, i) in (q.props?.fields||[])" :key="i" class="edit-option-row input-field-row" @click.stop="emit('select-option', i)">
        <span class="option-icon">⊞</span>
        <span class="input-field-label">{{ f.label || '字段' }}</span>
        <el-input size="small" :placeholder="f.placeholder||'请输入'" disabled style="flex:1;max-width:200px;pointer-events:none" />
        <el-tag v-if="f.dataType" size="small" effect="plain" style="margin:0 4px">{{ f.dataType }}</el-tag>
        <el-button text size="small" type="danger" class="opt-del-btn" @click.stop="removeField(i)">
          <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="1.5"><circle cx="12" cy="12" r="9"/><line x1="8" y1="12" x2="16" y2="12"/></svg>
        </el-button>
      </div>
    </div>
    <el-button v-if="editing && hasFields" text size="small" type="primary" @click="addField" class="add-option-btn">添加字段</el-button>

    <!-- 个人信息编辑预览 -->
    <div v-if="editing && ['phone','email','idCard','password','date','time','dateRange','switch','location'].includes(q.type)" class="edit-options-area" style="padding:4px 8px;cursor:pointer" @click.stop="emit('select-option', 0)">
      <div style="pointer-events:none">
        <el-input v-if="['phone','email','idCard','password'].includes(q.type)" size="small" :placeholder="q.placeholder||'请输入'" disabled />
        <el-date-picker v-else-if="q.type==='date'" disabled type="date" placeholder="选择日期" style="width:100%" />
        <el-time-picker v-else-if="q.type==='time'" disabled placeholder="选择时间" style="width:100%" />
        <el-switch v-else-if="q.type==='switch'" disabled />
        <div v-else-if="q.type==='dateRange'" style="display:flex;flex-direction:column;gap:4px;width:100%"><el-date-picker disabled type="date" placeholder="开始日期" style="width:100%" /><el-date-picker disabled type="date" placeholder="结束日期" style="width:100%" /></div>
        <div v-else-if="q.type==='location'" style="color:#999;font-size:13px;padding:4px 0"><el-button text size="small" disabled><svg viewBox="0 0 1024 1024" width="16" height="16" fill="currentColor" style="vertical-align:middle;margin-right:4px"><path d="M512 64C367.2 64 248 183.2 248 328c0 163.2 233.6 524.8 252 551.2 3.2 4.8 8 7.2 12 7.2s8.8-2.4 12-7.2C542.4 852.8 776 491.2 776 328 776 183.2 656.8 64 512 64z m0 400c-39.2 0-72-32.8-72-72s32.8-72 72-72 72 32.8 72 72-32.8 72-72 72z"/></svg>选择位置</el-button></div>
      </div>
    </div>

    <!-- 富文本编辑预览（画布中显示模拟编辑器） -->
    <div v-if="editing && q.type==='richText'" class="edit-options-area" style="padding:0;cursor:pointer;overflow:hidden;border:1px solid #dcdfe6;border-radius:4px" @click.stop="emit('select-option', 0)">
      <div style="pointer-events:none;min-height:80px;padding:10px 12px;color:#c0c4cc;font-size:13px">{{ q.placeholder || '输入富文本内容...' }}</div>
    </div>

    <!-- 自动填充编辑预览 -->
    <div v-if="editing && q.type==='autopop'" class="edit-options-area" style="padding:4px 8px;cursor:pointer" @click.stop="emit('select-option', 0)">
      <div style="pointer-events:none">
        <el-input size="small" placeholder="自动填充" disabled />
      </div>
    </div>

    <!-- 扫码编辑预览 -->
    <div v-if="editing && q.type==='scanCode'" class="edit-options-area" style="padding:4px 8px;cursor:pointer" @click.stop="emit('select-option', 0)">
      <div style="pointer-events:none">
        <el-input size="small" placeholder="扫码" disabled>
          <template #prefix><svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="3" width="7" height="7"/><rect x="14" y="3" width="7" height="7"/><rect x="14" y="14" width="7" height="7"/><rect x="3" y="14" width="7" height="7"/></svg></template>
        </el-input>
      </div>
    </div>

    <!-- 签名编辑预览 -->
    <div v-if="editing && q.type==='signature'" class="edit-options-area" style="padding:0;cursor:pointer;overflow:hidden;border:1px solid #dcdfe6;border-radius:4px" @click.stop="emit('select-option', 0)">
      <div style="pointer-events:none;min-height:100px;display:flex;align-items:center;justify-content:center;color:#c0c4cc;font-size:13px">点击设置签名</div>
    </div>

    <div v-if="!editing" class="preview-body">
      <el-input v-if="['input','text'].includes(q.type)" :placeholder="q.placeholder||'请输入'" v-model="val" />
      <div v-else-if="q.type==='multiInput'" style="display:flex;flex-direction:column;gap:4px">
        <el-input v-for="(f, fi) in (q.props?.fields||[])" :key="fi" :placeholder="f.placeholder||'请输入'" v-model="val[fi]" />
      </div>
      <div v-else-if="q.type==='hInput'" style="display:flex;flex-wrap:wrap;gap:4px">
        <el-input v-for="(f, fi) in (q.props?.fields||[])" :key="fi" :placeholder="f.placeholder||'请输入'" v-model="val[fi]" style="flex:1;min-width:120px" />
      </div>
      <el-input v-else-if="q.type==='scanCode'" placeholder="扫码" v-model="val" class="scan-code-input">
        <template #prefix><svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="3" width="7" height="7"/><rect x="14" y="3" width="7" height="7"/><rect x="14" y="14" width="7" height="7"/><rect x="3" y="14" width="7" height="7"/></svg></template>
        <template #suffix>
          <el-button text type="primary" size="small" @click="showScanner = true" style="margin-right:-8px;height:28px">
            <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><path d="M1 7V4a2 2 0 0 1 2-2h3M1 17v3a2 2 0 0 0 2 2h3M23 7V4a2 2 0 0 0-2-2h-3M23 17v3a2 2 0 0 1-2 2h-3"/><rect x="8" y="8" width="8" height="8" rx="1"/></svg>
          </el-button>
        </template>
      </el-input>
      <div v-else-if="q.type==='signature'" class="signature-pad-wrap">
        <canvas ref="sigCanvasRef" class="sig-canvas" @mousedown="onSigMouseDown" @mousemove="onSigMouseMove" @mouseup="onSigMouseUp" @mouseleave="onSigMouseUp" @touchstart.prevent="onSigTouchStart" @touchmove.prevent="onSigTouchMove" @touchend="onSigMouseUp"></canvas>
        <div style="display:flex;gap:8px;margin-top:4px"><el-button size="small" text @click="clearSignature">清除</el-button></div>
      </div>
      <el-input v-else-if="q.type==='textarea'" type="textarea" :placeholder="q.placeholder||'请输入'" :rows="2" v-model="val" />
      <el-input-number v-else-if="q.type==='number'" v-model="val" style="width:100%;--el-input-width:100%" />
      <el-radio-group v-else-if="q.type==='radio'" v-model="val" class="preview-options preview-radio-group" :style="optionGrid(q)">
        <el-radio v-for="o in (q.props?.options||[])" :key="o.value" :value="o.value"><span v-html="o.label"></span></el-radio>
      </el-radio-group>
      <el-checkbox-group v-else-if="q.type==='checkbox'" v-model="val" class="preview-options preview-checkbox-group" :style="optionGrid(q)">
        <el-checkbox v-for="o in (q.props?.options||[])" :key="o.value" :value="o.value"><span v-html="o.label"></span></el-checkbox>
      </el-checkbox-group>
      <el-select v-else-if="q.type==='select'" v-model="val" placeholder="请选择" style="width:100%">
        <el-option v-for="o in (q.props?.options||[])" :key="o.value" :value="o.value" :label="o.label" />
      </el-select>
      <el-select v-else-if="q.type==='picker'" v-model="val" placeholder="请选择" style="width:100%">
        <el-option v-for="o in (q.props?.options||[])" :key="o.value" :value="o.value" :label="o.label" />
      </el-select>
      <el-cascader v-else-if="q.type==='cascade'" v-model="val" placeholder="请选择" style="width:100%" :options="q.props?.options||[]" clearable />
      <el-radio-group v-else-if="q.type==='judge'" v-model="val" class="preview-options preview-radio-group">
        <el-radio value="true">对</el-radio>
        <el-radio value="false">错</el-radio>
      </el-radio-group>
      <div v-else-if="q.type==='rating'" style="padding:4px 0">
        <el-rate v-model="val" :max="q.props?.maxRating || 5" />
      </div>
      <el-date-picker v-else-if="q.type==='date'" v-model="val" type="date" placeholder="选择日期" style="width:100%" />
      <el-time-picker v-else-if="q.type==='time'" v-model="val" placeholder="选择时间" style="width:100%" />
      <el-switch v-else-if="q.type==='switch'" v-model="val" />
      <el-divider v-else-if="q.type==='divider'" style="margin:4px 0" />
      <div v-else-if="q.type==='description'" class="preview-plain">{{ q.description }}</div>
      <div v-else-if="q.type==='file'" style="border:1px dashed #d9d9d9;border-radius:6px;padding:12px;text-align:center;color:#999;margin:4px 0">
        <svg viewBox="0 0 1024 1024" width="22" height="22" fill="currentColor" style="display:block;margin:0 auto 4px"><path d="M854.6 288.6L639.4 73.4c-6-6-14.1-9.4-22.6-9.4H192c-17.7 0-32 14.3-32 32v832c0 17.7 14.3 32 32 32h640c17.7 0 32-14.3 32-32V311.3c0-8.5-3.4-16.7-9.4-22.7z"/></svg>
        <span style="font-size:13px">上传文件</span>
      </div>
      <div v-else-if="q.type==='location'" class="preview-plain"><el-button text @click.prevent><svg viewBox="0 0 1024 1024" width="16" height="16" fill="currentColor" style="vertical-align:middle;margin-right:4px"><path d="M512 64C367.2 64 248 183.2 248 328c0 163.2 233.6 524.8 252 551.2 3.2 4.8 8 7.2 12 7.2s8.8-2.4 12-7.2C542.4 852.8 776 491.2 776 328 776 183.2 656.8 64 512 64z m0 400c-39.2 0-72-32.8-72-72s32.8-72 72-72 72 32.8 72 72-32.8 72-72 72z"/></svg>选择位置</el-button></div>
      <el-input v-else-if="q.type==='phone'" placeholder="手机号" v-model="val" />
      <el-input v-else-if="q.type==='phone'" placeholder="手机号" v-model="val" />
      <el-input v-else-if="q.type==='email'" placeholder="邮箱地址" v-model="val" />
      <el-input v-else-if="q.type==='idCard'" placeholder="身份证号" v-model="val" />
      <el-input v-else-if="q.type==='password'" type="password" placeholder="密码" v-model="val" />
      <div v-else-if="q.type==='dateRange'" style="display:flex;flex-direction:column;gap:4px;width:100%"><el-date-picker v-model="val[0]" type="date" placeholder="开始日期" style="width:100%" /><el-date-picker v-model="val[1]" type="date" placeholder="结束日期" style="width:100%" /></div>
      <div v-else-if="q.type==='matrixRadio'" class="preview-matrix">
        <table><thead><tr><th class="corner">行\\列</th><th v-for="c in (q.props?.columns||[{title:'列A'},{title:'列B'}])" :key="typeof c==='string'?c:(c.title||c.label)">{{ typeof c==='string'?c:(c.title||c.label) }}</th></tr></thead><tbody><tr v-for="(r, ri) in (q.props?.rows||[{title:'行1'},{title:'行2'}])" :key="typeof r==='string'?r:r.title"><td class="matrix-label">{{ typeof r==='string'?r:r.title }}</td><td v-for="c in (q.props?.columns||[{title:'列A'},{title:'列B'}])" :key="typeof c==='string'?c:(c.title||c.label)"><el-radio-group :model-value="val[ri]" @update:model-value="(v: any) => val[ri] = v"><el-radio :value="typeof c==='string'?c:(c.title||c.label)" /></el-radio-group></td></tr></tbody></table>
      </div>
      <div v-else-if="q.type==='matrixCheckbox'" class="preview-matrix">
        <table><thead><tr><th class="corner">行\\列</th><th v-for="c in (q.props?.columns||[{title:'列A'},{title:'列B'}])" :key="typeof c==='string'?c:(c.title||c.label)">{{ typeof c==='string'?c:(c.title||c.label) }}</th></tr></thead><tbody><tr v-for="(r, ri) in (q.props?.rows||[{title:'行1'},{title:'行2'}])" :key="typeof r==='string'?r:r.title"><td class="matrix-label">{{ typeof r==='string'?r:r.title }}</td><td v-for="c in (q.props?.columns||[{title:'列A'},{title:'列B'}])" :key="typeof c==='string'?c:(c.title||c.label)"><el-checkbox-group :model-value="val[ri]||[]" @update:model-value="(v: any) => val[ri] = v"><el-checkbox :value="typeof c==='string'?c:(c.title||c.label)" /></el-checkbox-group></td></tr></tbody></table>
      </div>
      <div v-else-if="q.type==='matrixFillBlank'" class="preview-matrix">
        <table><thead><tr><th class="corner">行\\列</th><th v-for="c in (q.props?.columns||[{title:'列A'},{title:'列B'}])" :key="typeof c==='string'?c:(c.title||c.label)">{{ typeof c==='string'?c:(c.title||c.label) }}</th></tr></thead><tbody><tr v-for="(r, ri) in (q.props?.rows||[{title:'行1'},{title:'行2'}])" :key="typeof r==='string'?r:r.title"><td class="matrix-label">{{ typeof r==='string'?r:r.title }}</td><td v-for="(c, ci) in (q.props?.columns||[{title:'列A'},{title:'列B'}])" :key="typeof c==='string'?c:(c.title||c.label)"><el-input :model-value="val[ri]?.[ci]" @update:model-value="(v: any) => { if(!val[ri]) val[ri]={}; val[ri][ci]=v }" placeholder="填空" size="small" style="width:100%" /></td></tr></tbody></table>
      </div>
      <div v-else-if="q.type==='matrixAuto'" class="preview-matrix">
        <table><thead><tr><th class="corner">#</th><th v-for="c in (q.props?.columns||[])" :key="c.label||c.id||c">{{ c.label||c }}</th></tr></thead><tbody><tr v-for="(r, ri) in (q.props?.rows||[])" :key="ri"><td class="matrix-label">{{ ri+1 }}</td><td v-for="(c, ci) in (q.props?.columns||[])" :key="c.label||c.id||c"><el-input :model-value="val[ri]?.[ci]" @update:model-value="(v: any) => { if(!val[ri]) val[ri]={}; val[ri][ci]=v }" size="small" :placeholder="c.label||'值'" style="width:100%" /></td></tr></tbody></table>
        <div style="display:flex;align-items:center;gap:8px;padding:6px 8px;border-top:1px solid #e0e0e0"><el-button text size="small" disabled>+ 添加行</el-button><span style="flex:1"></span><el-button text size="small" type="danger" disabled><svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="9"/><line x1="8" y1="12" x2="16" y2="12"/></svg> 删除行</el-button></div>
      </div>
      <div v-else-if="q.type==='questionSet'" class="preview-plain">问题组（内部题）</div>
      <div v-else-if="q.type==='pagination'" class="preview-plain">—— 分页 ——</div>
      <el-cascader v-else-if="q.type==='user'||q.type==='dept'" v-model="val" :placeholder="q.type==='user'?'选择成员':'选择部门'" style="width:100%" :options="userDeptTreeOptions" :props="{ multiple: !!q.multiple, emitPath: false }" clearable />
      <div v-else-if="q.type==='richText'" style="border:1px solid #dcdfe6;border-radius:4px;overflow:hidden">
        <QuillEditor v-model:content="val" content-type="html" :options="{ theme: 'snow', placeholder: q.placeholder || '输入富文本内容...' }" style="min-height:150px" />
      </div>
      <div v-else-if="q.type==='autopop'" class="preview-plain"><el-input placeholder="自动填充" v-model="val" /></div>
      <div v-else-if="q.type==='nps'" class="preview-nps">
        <div class="nps-labels"><span>0</span><span>10</span></div>
        <el-rate v-model="val" :max="10" show-score score-template="{value}" />
      </div>
    </div>

    <!-- 扫码弹窗 -->
    <el-dialog v-if="showScanner" v-model="showScanner" title="扫码" width="400px" :close-on-click-modal="false" destroy-on-close @opened="onScannerOpen" @close="onScannerClose">
      <div ref="scannerRef" style="width:100%;aspect-ratio:1;overflow:hidden;background:#000;border-radius:8px"></div>
      <template #footer>
        <el-button @click="showScanner = false">取消</el-button>
      </template>
    </el-dialog>

    <!-- 题干高级编辑弹窗 -->
    <el-dialog v-if="showTitleAdvancedEditor" v-model="showTitleAdvancedEditor" title="编辑题干" width="700px" destroy-on-close draggable @close="cancelTitleAdvancedEdit" class="title-full-editor-dialog">
      <QuillEditor v-model:content="titleEditBuffer" content-type="html"
        :options="fullTitleEditorOptions"
        style="min-height:50vh"
        @ready="onFullEditorReady" />
      <template #footer>
        <el-button @click="showTitleAdvancedEditor = false">取消</el-button>
        <el-button type="primary" @click="confirmTitleAdvancedEdit">确定</el-button>
      </template>
    </el-dialog>

    <!-- 选项高级编辑弹窗 -->
    <el-dialog v-if="showOptionAdvancedEditor" v-model="showOptionAdvancedEditor" title="编辑选项" width="700px" destroy-on-close draggable @close="cancelOptionAdvancedEdit">
      <QuillEditor v-model:content="optionEditBuffer" content-type="html"
        :options="fullTitleEditorOptions"
        style="min-height:50vh"
        @ready="onOptionFullEditorReady" />
      <template #footer>
        <el-button @click="showOptionAdvancedEditor = false">取消</el-button>
        <el-button type="primary" @click="confirmOptionAdvancedEdit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { QuillEditor } from '@vueup/vue-quill'
import '@vueup/vue-quill/dist/vue-quill.snow.css'
import { Html5Qrcode } from 'html5-qrcode'

const props = defineProps<{ q: any; editing?: boolean }>()
const emit = defineEmits<{
  'update:title': [v: string]
  'update:description': [v: string]
  remove: []
  'open-logic': []
  copy: []
  'upload-bank': []
  'select-option': [idx: number]
}>()

const descRef = ref<HTMLElement>()

const showTitleAdvancedEditor = ref(false)
const titleEditBuffer = ref('')
function openTitleAdvancedEditor() {
  titleEditBuffer.value = props.q.title || ''
  showTitleAdvancedEditor.value = true
}
function confirmTitleAdvancedEdit() {
  props.q.title = titleEditBuffer.value
  showTitleAdvancedEditor.value = false
  emit('update:title', titleEditBuffer.value)
}
function cancelTitleAdvancedEdit() {
  titleEditBuffer.value = ''
}

const showOptionAdvancedEditor = ref(false)
const optionAdvEditIndex = ref(-1)
const optionEditBuffer = ref('')
function openOptionAdvancedEditor(idx: number) {
  const opts = props.q.props?.options || []
  optionAdvEditIndex.value = idx
  optionEditBuffer.value = opts[idx]?.label || ''
  showOptionAdvancedEditor.value = true
}
function confirmOptionAdvancedEdit() {
  const opts = props.q.props?.options
  if (opts && opts[optionAdvEditIndex.value]) {
    opts[optionAdvEditIndex.value].label = optionEditBuffer.value
  }
  showOptionAdvancedEditor.value = false
  optionAdvEditIndex.value = -1
}
function cancelOptionAdvancedEdit() {
  optionEditBuffer.value = ''
  optionAdvEditIndex.value = -1
}
function onTitleEditorReady(quill: any) {
  const toolbar = quill.getModule('toolbar')
  if (!toolbar || !toolbar.container) return
  const btn = document.createElement('button')
  btn.className = 'ql-advanced'
  btn.innerHTML = '<svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>'
  btn.title = '高级编辑'
  btn.addEventListener('click', (e: MouseEvent) => {
    e.stopPropagation()
    openTitleAdvancedEditor()
  })
  toolbar.container.appendChild(btn)
}
function onOptionEditorReady(quill: any, idx: number) {
  const toolbar = quill.getModule('toolbar')
  if (!toolbar || !toolbar.container) return
  const btn = document.createElement('button')
  btn.className = 'ql-advanced'
  btn.innerHTML = '<svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>'
  btn.title = '高级编辑'
  btn.addEventListener('click', (e: MouseEvent) => {
    e.stopPropagation()
    openOptionAdvancedEditor(idx)
  })
  toolbar.container.appendChild(btn)
}
function setupImageHandler(quill: any) {
  const toolbar = quill.getModule('toolbar')
  if (!toolbar) return
  toolbar.addHandler('image', () => {
    const input = document.createElement('input')
    input.type = 'file'
    input.accept = 'image/*'
    input.onchange = async () => {
      const file = input.files?.[0]
      if (!file) return
      const fd = new FormData()
      fd.append('file', file)
      try {
        const resp = await fetch('/upload', { method: 'POST', body: fd })
        const json = await resp.json()
        if (json.code === 0 && json.data?.url) {
          const fullUrl = (json.data.domain || '') + json.data.url
          const range = quill.getSelection(true)
          quill.insertEmbed(range.index, 'image', fullUrl)
          quill.setSelection(range.index + 1)
        }
      } catch { /* ignore */ }
      input.remove()
    }
    input.click()
  })
}
function onFullEditorReady(quill: any) {
  const toolbar = quill.getModule('toolbar')
  if (!toolbar || !toolbar.container) return
  setupImageHandler(quill)
  const btn = document.createElement('button')
  btn.className = 'ql-fullscreen'
  let fs = false
  const origStyles: Record<string, string> = {}
  btn.innerHTML = '<svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><path d="M8 3H5a2 2 0 0 0-2 2v3m18 0V5a2 2 0 0 0-2-2h-3m0 18h3a2 2 0 0 0 2-2v-3M3 16v3a2 2 0 0 0 2 2h3"/></svg>'
  btn.title = '全屏'
  btn.addEventListener('click', (e: MouseEvent) => {
    e.stopPropagation()
    const dialogEl = (quill.container as HTMLElement).closest('.el-dialog') as HTMLElement
    if (!dialogEl) return
    if (!fs) {
      origStyles.position = dialogEl.style.position
      origStyles.top = dialogEl.style.top
      origStyles.left = dialogEl.style.left
      origStyles.width = dialogEl.style.width
      origStyles.height = dialogEl.style.height
      origStyles.maxWidth = dialogEl.style.maxWidth
      origStyles.margin = dialogEl.style.margin
      origStyles.borderRadius = dialogEl.style.borderRadius
      origStyles.zIndex = dialogEl.style.zIndex
      dialogEl.style.position = 'fixed'
      dialogEl.style.top = '0'
      dialogEl.style.left = '0'
      dialogEl.style.width = '100vw'
      dialogEl.style.height = '100vh'
      dialogEl.style.maxWidth = '100vw'
      dialogEl.style.maxHeight = '100vh'
      dialogEl.style.margin = '0'
      dialogEl.style.borderRadius = '0'
      dialogEl.style.zIndex = '9999'
      const body = dialogEl.querySelector('.el-dialog__body') as HTMLElement
      const qlContainer = dialogEl.querySelector('.ql-container') as HTMLElement
      if (body && qlContainer) {
        const toolbarEl = dialogEl.querySelector('.ql-toolbar') as HTMLElement
        const footerEl = dialogEl.querySelector('.el-dialog__footer') as HTMLElement
        const headerEl = dialogEl.querySelector('.el-dialog__header') as HTMLElement
        const toolH = toolbarEl?.offsetHeight || 0
        const footH = footerEl?.offsetHeight || 0
        const headH = headerEl?.offsetHeight || 0
        const bodyPad = body.offsetHeight - body.clientHeight
        qlContainer.style.height = (window.innerHeight - headH - footH - toolH - bodyPad) + 'px'
      }
      btn.innerHTML = '<svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><path d="M8 3v3a2 2 0 0 1-2 2H3m18 0h-3a2 2 0 0 1-2-2V3m0 18v-3a2 2 0 0 1 2-2h3M3 16h3a2 2 0 0 1 2 2v3"/></svg>'
      btn.title = '退出全屏'
      fs = true
    } else {
      dialogEl.style.position = origStyles.position
      dialogEl.style.top = origStyles.top
      dialogEl.style.left = origStyles.left
      dialogEl.style.width = origStyles.width
      dialogEl.style.height = origStyles.height
      dialogEl.style.maxWidth = origStyles.maxWidth
      dialogEl.style.margin = origStyles.margin
      dialogEl.style.borderRadius = origStyles.borderRadius
      dialogEl.style.zIndex = origStyles.zIndex
      const qlContainer = dialogEl.querySelector('.ql-container') as HTMLElement
      if (qlContainer) qlContainer.style.height = ''
      btn.innerHTML = '<svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><path d="M8 3H5a2 2 0 0 0-2 2v3m18 0V5a2 2 0 0 0-2-2h-3m0 18h3a2 2 0 0 0 2-2v-3M3 16v3a2 2 0 0 0 2 2h3"/></svg>'
      btn.title = '全屏'
      fs = false
    }
  })
  toolbar.container.appendChild(btn)
}
const titleEditorOptions = {
  theme: 'snow',
  modules: {
    toolbar: [
      ['bold', 'italic', 'underline', 'strike'],
      [{ color: [] }, { background: [] }],
      [{ list: 'ordered' }, { list: 'bullet' }],
      ['clean']
    ]
  }
}
const fullTitleEditorOptions = {
  theme: 'snow',
  modules: {
    toolbar: [
      [{ header: [1, 2, 3, false] }],
      ['bold', 'italic', 'underline', 'strike'],
      [{ color: [] }, { background: [] }],
      [{ list: 'ordered' }, { list: 'bullet' }, { list: 'check' }],
      [{ indent: '-1' }, { indent: '+1' }],
      [{ align: [] }],
      ['blockquote', 'code-block'],
      ['link', 'image'],
      ['clean']
    ]
  }
}
const optionEditorOptions = {
  theme: 'snow',
  modules: {
    toolbar: [
      ['bold', 'italic', 'underline', 'strike'],
      [{ color: [] }, { background: [] }],
      ['clean']
    ]
  }
}
function onOptionFullEditorReady(quill: any) {
  const toolbar = quill.getModule('toolbar')
  if (!toolbar || !toolbar.container) return
  setupImageHandler(quill)
  const btn = document.createElement('button')
  btn.className = 'ql-fullscreen'
  let fs = false
  const origStyles: Record<string, string> = {}
  btn.innerHTML = '<svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><path d="M8 3H5a2 2 0 0 0-2 2v3m18 0V5a2 2 0 0 0-2-2h-3m0 18h3a2 2 0 0 0 2-2v-3M3 16v3a2 2 0 0 0 2 2h3"/></svg>'
  btn.title = '全屏'
  btn.addEventListener('click', (e: MouseEvent) => {
    e.stopPropagation()
    const dialogEl = (quill.container as HTMLElement).closest('.el-dialog') as HTMLElement
    if (!dialogEl) return
    if (!fs) {
      origStyles.position = dialogEl.style.position
      origStyles.top = dialogEl.style.top
      origStyles.left = dialogEl.style.left
      origStyles.width = dialogEl.style.width
      origStyles.height = dialogEl.style.height
      origStyles.maxWidth = dialogEl.style.maxWidth
      origStyles.margin = dialogEl.style.margin
      origStyles.borderRadius = dialogEl.style.borderRadius
      origStyles.zIndex = dialogEl.style.zIndex
      dialogEl.style.position = 'fixed'
      dialogEl.style.top = '0'
      dialogEl.style.left = '0'
      dialogEl.style.width = '100vw'
      dialogEl.style.height = '100vh'
      dialogEl.style.maxWidth = '100vw'
      dialogEl.style.maxHeight = '100vh'
      dialogEl.style.margin = '0'
      dialogEl.style.borderRadius = '0'
      dialogEl.style.zIndex = '9999'
      const body = dialogEl.querySelector('.el-dialog__body') as HTMLElement
      const qlContainer = dialogEl.querySelector('.ql-container') as HTMLElement
      if (body && qlContainer) {
        const toolbarEl = dialogEl.querySelector('.ql-toolbar') as HTMLElement
        const footerEl = dialogEl.querySelector('.el-dialog__footer') as HTMLElement
        const headerEl = dialogEl.querySelector('.el-dialog__header') as HTMLElement
        const toolH = toolbarEl?.offsetHeight || 0
        const footH = footerEl?.offsetHeight || 0
        const headH = headerEl?.offsetHeight || 0
        const bodyPad = body.offsetHeight - body.clientHeight
        qlContainer.style.height = (window.innerHeight - headH - footH - toolH - bodyPad) + 'px'
      }
      btn.innerHTML = '<svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><path d="M8 3v3a2 2 0 0 1-2 2H3m18 0h-3a2 2 0 0 1-2-2V3m0 18v-3a2 2 0 0 1 2-2h3M3 16h3a2 2 0 0 1 2 2v3"/></svg>'
      btn.title = '退出全屏'
      fs = true
    } else {
      dialogEl.style.position = origStyles.position
      dialogEl.style.top = origStyles.top
      dialogEl.style.left = origStyles.left
      dialogEl.style.width = origStyles.width
      dialogEl.style.height = origStyles.height
      dialogEl.style.maxWidth = origStyles.maxWidth
      dialogEl.style.margin = origStyles.margin
      dialogEl.style.borderRadius = origStyles.borderRadius
      dialogEl.style.zIndex = origStyles.zIndex
      const qlContainer = dialogEl.querySelector('.ql-container') as HTMLElement
      if (qlContainer) qlContainer.style.height = ''
      btn.innerHTML = '<svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><path d="M8 3H5a2 2 0 0 0-2 2v3m18 0V5a2 2 0 0 0-2-2h-3m0 18h3a2 2 0 0 0 2-2v-3M3 16v3a2 2 0 0 0 2 2h3"/></svg>'
      btn.title = '全屏'
      fs = false
    }
  })
  toolbar.container.appendChild(btn)
}

function getInitVal(type: string): any {
  if (type === 'checkbox') return []
  if (type === 'switch') return false
  if (type === 'number') return undefined
  if (type === 'rating' || type === 'nps') return 0
  if (type === 'dateRange') return ['', '']
  if (['matrixRadio','matrixCheckbox','matrixFillBlank','matrixAuto'].includes(type)) return {}
  if (['cascade','user','dept'].includes(type)) {
    if (type === 'user' || type === 'dept') {
      return props.q.multiple ? [] : ''
    }
    return []
  }
  if (['multiInput','hInput'].includes(type)) return (props.q.props?.fields||[]).map(() => '')
  return ''
}
const val = ref<any>(getInitVal(props.q.type))

watch(() => props.q.type, (t) => { val.value = getInitVal(t) })

const sigCanvasRef = ref<HTMLCanvasElement>()
const sigDrawing = ref(false)
function initSigCanvas() {
  const c = sigCanvasRef.value
  if (!c) return
  const rect = c.getBoundingClientRect()
  const dpr = Math.min(window.devicePixelRatio || 1, 2)
  if (c.width !== Math.round(rect.width * dpr) || c.height !== Math.round(rect.height * dpr)) {
    const oldData = c.toDataURL()
    c.width = Math.round(rect.width * dpr)
    c.height = Math.round(rect.height * dpr)
    const ctx = c.getContext('2d')
    if (ctx) {
      ctx.scale(dpr, dpr)
      if (oldData && oldData !== 'data:,') {
        const img = new Image()
        img.onload = () => { ctx.setTransform(1, 0, 0, 1, 0, 0); ctx.drawImage(img, 0, 0); ctx.scale(dpr, dpr) }
        img.src = oldData
      }
    }
  }
}
import { onMounted } from 'vue'
onMounted(() => { setTimeout(initSigCanvas, 50) })
function getSigCtx() {
  const c = sigCanvasRef.value
  if (!c) return null
  if (!c.getContext) return null
  return c.getContext('2d')!
}
function getSigPos(e: MouseEvent) {
  const r = sigCanvasRef.value!.getBoundingClientRect()
  return { x: e.clientX - r.left, y: e.clientY - r.top }
}
function onSigMouseDown(e: MouseEvent) {
  sigDrawing.value = true
  const ctx = getSigCtx()
  if (!ctx) return
  const p = getSigPos(e)
  ctx.beginPath()
  ctx.moveTo(p.x, p.y)
}
function onSigMouseMove(e: MouseEvent) {
  if (!sigDrawing.value) return
  const ctx = getSigCtx()
  if (!ctx) return
  const p = getSigPos(e)
  ctx.lineTo(p.x, p.y)
  ctx.strokeStyle = '#303133'
  ctx.lineWidth = 2
  ctx.lineCap = 'round'
  ctx.stroke()
}
function onSigMouseUp() {
  if (!sigDrawing.value) return
  sigDrawing.value = false
  val.value = sigCanvasRef.value?.toDataURL() || ''
}
function onSigTouchStart(e: TouchEvent) {
  sigDrawing.value = true
  const ctx = getSigCtx()
  if (!ctx) return
  const t = e.touches[0]
  const r = sigCanvasRef.value!.getBoundingClientRect()
  ctx.beginPath()
  ctx.moveTo(t.clientX - r.left, t.clientY - r.top)
}
function onSigTouchMove(e: TouchEvent) {
  if (!sigDrawing.value) return
  const ctx = getSigCtx()
  if (!ctx) return
  const t = e.touches[0]
  const r = sigCanvasRef.value!.getBoundingClientRect()
  ctx.lineTo(t.clientX - r.left, t.clientY - r.top)
  ctx.strokeStyle = '#303133'
  ctx.lineWidth = 2
  ctx.lineCap = 'round'
  ctx.stroke()
}
function clearSignature() {
  const ctx = getSigCtx()
  if (!ctx || !sigCanvasRef.value) return
  ctx.clearRect(0, 0, sigCanvasRef.value.width, sigCanvasRef.value.height)
  val.value = ''
}

const showScanner = ref(false)
const scannerRef = ref<HTMLDivElement>()
let scanner: Html5Qrcode | null = null
function onScannerOpen() {
  if (!scannerRef.value) return
  scanner = new Html5Qrcode('scanner-ref-fallback')
  scannerRef.value.id = 'scanner-ref-fallback'
  scanner.start(
    { facingMode: 'environment' },
    { fps: 10, qrbox: { width: 250, height: 250 } },
    (decodedText) => {
      val.value = decodedText
      showScanner.value = false
    },
    () => {}
  ).catch(() => {})
}
function onScannerClose() {
  if (scanner) {
    scanner.stop().catch(() => {})
    scanner = null
  }
}

const hasOptions = computed(() => ['select','radio','checkbox','picker','judge','cascade','matrixRadio','matrixCheckbox'].includes(props.q.type))
const isInputSubFieldType = computed(() => ['input','textarea','signature','scanCode','file'].includes(props.q.type))
const isSingleInput = computed(() => ['input','textarea'].includes(props.q.type))
const hasFields = computed(() => ['multiInput','hInput'].includes(props.q.type))
const isChoiceType = computed(() => ['select','radio','checkbox','picker','judge','cascade'].includes(props.q.type))
const isMatrixChoiceType = computed(() => ['matrixRadio','matrixCheckbox'].includes(props.q.type))
const isMatrixType = computed(() => ['matrixRadio','matrixCheckbox','matrixFillBlank','matrixAuto'].includes(props.q.type))

const newMatrixRowName = ref('')
const newMatrixColName = ref('')
const userDeptTreeOptions = computed(() => {
  const q = props.q
  const options = q.props?.options || []
  if (q.type === 'user') {
    const deptMap: Record<string, any> = {}
    options.forEach((o: any) => {
      const deptId = o.deptId || ''
      if (!deptMap[deptId]) {
        deptMap[deptId] = { value: '__d__' + deptId, label: o.deptName || deptId || '未分组', children: [] }
      }
      deptMap[deptId].children.push({ value: o.value, label: o.label || '成员' })
    })
    return Object.values(deptMap)
  }
  const map: Record<string, any> = {}
  options.forEach((o: any) => { map[o.value] = { ...o, children: [] } })
  const roots: any[] = []
  options.forEach((o: any) => {
    if (o.parentId && map[o.parentId]) {
      map[o.parentId].children.push(map[o.value])
    } else {
      roots.push(map[o.value])
    }
  })
  return roots
})
const matrixGridCols = computed(() => {
  const n = props.q.props?.columns?.length || 2
  if (props.q.type === 'matrixAuto') return `20px 130px repeat(${n}, minmax(100px, 1fr))`
  return `20px 130px repeat(${n}, minmax(100px, 1fr)) 32px`
})
function ensureRows() {
  if (!props.q.props) props.q.props = {}
  if (!Array.isArray(props.q.props.rows)) props.q.props.rows = []
}
function ensureCols() {
  if (!props.q.props) props.q.props = {}
  if (!Array.isArray(props.q.props.columns)) props.q.props.columns = []
}
function onMatrixRowLabelBlur(idx: number, e: FocusEvent) {
  const text = (e.target as HTMLElement).innerText.trim()
  if (!text) return
  ensureRows()
  const existing = props.q.props.rows[idx] || {}
  props.q.props.rows[idx] = typeof existing === 'string' ? text : { ...existing, title: text }
}
function onMatrixColLabelBlur(idx: number, e: FocusEvent) {
  const text = (e.target as HTMLElement).innerText.trim()
  if (!text) return
  ensureCols()
  const existing = props.q.props.columns[idx] || {}
  props.q.props.columns[idx] = typeof existing === 'string' ? text : { ...existing, title: text }
}
function addMatrixRow() {
  ensureRows()
  const name = `行${props.q.props.rows.length + 1}`
  props.q.props.rows.push({ title: name, id: Date.now() + '_' + Math.random().toString(36).slice(2,6), width: 150 })
}
function removeMatrixRow(idx: number) {
  ensureRows()
  props.q.props.rows.splice(idx, 1)
}
function removeMatrixLastRow() {
  ensureRows()
  if (props.q.props.rows.length > 0) props.q.props.rows.pop()
}
function addMatrixCol() {
  ensureCols()
  const name = `列${props.q.props.columns.length + 1}`
  props.q.props.columns.push({ title: name, id: Date.now() + '_' + Math.random().toString(36).slice(2,6), width: 150 })
}
function removeMatrixCol(idx: number) {
  ensureCols()
  props.q.props.columns.splice(idx, 1)
}

const fieldIcon = computed(() => {
  if (props.q.type === 'checkbox') return '☐'
  if (props.q.type === 'judge') return '○'
  if (isChoiceType.value) return '○'
  return '▸'
})

const examCorrectAnswer = computed(() => props.q.examCorrectAnswer)

function onDescBlur(e: FocusEvent) {
  const text = (e.target as HTMLElement).innerText.trim()
  emit('update:description', text)
}
/** 直接修改父级 reactive 数据，绕过多层 emit 闭包问题 */
function ensureOpts() {
  if (!props.q.props) props.q.props = {}
  if (!Array.isArray(props.q.props.options)) props.q.props.options = []
}
function onOptionLabelBlur(idx: number, e: FocusEvent) {
  const text = (e.target as HTMLElement).innerText.trim()
  if (!text) return
  ensureOpts()
  const existing = props.q.props.options[idx] || {}
  props.q.props.options[idx] = { ...existing, label: text }
}
function removeOption(idx: number) {
  ensureOpts()
  props.q.props.options.splice(idx, 1)
}
function addOption() {
  ensureOpts()
  const n = props.q.props.options.length + 1
  const letter = n <= 26 ? String.fromCharCode(64 + n) : `opt${n}`
  props.q.props.options.push({ label: `选项${n}`, value: letter })
}

function ensureFields() {
  if (!props.q.props) props.q.props = {}
  if (!Array.isArray(props.q.props.fields)) props.q.props.fields = []
}
function onFieldLabelBlur(idx: number, e: FocusEvent) {
  const text = (e.target as HTMLElement).innerText.trim()
  if (!text) return
  ensureFields()
  const existing = props.q.props.fields[idx] || {}
  props.q.props.fields[idx] = { ...existing, label: text }
}
function addField() {
  ensureFields()
  props.q.props.fields.push({ label: '字段' + (props.q.props.fields.length + 1), placeholder: '' })
}
function removeField(idx: number) {
  ensureFields()
  props.q.props.fields.splice(idx, 1)
}

const typeName = (t: string) => {
  const map: Record<string, string> = {
    input: '单行文本', text: '文本', textarea: '多行', number: '数字',
    select: '下拉', radio: '单选', checkbox: '多选', picker: '选择器',
    cascade: '级联选择', judge: '判断',
    multiInput: '多项填空', hInput: '横向填空', scanCode: '扫码',
    signature: '签名', file: '上传文件',
    rating: '评分', nps: 'NPS评分',
    date: '日期', time: '时间', dateRange: '日期范围',
    phone: '手机', email: '邮箱', idCard: '身份证', password: '密码',
    switch: '开关', location: '地理位置',
    matrixRadio: '矩阵单选', matrixCheckbox: '矩阵多选',
    matrixFillBlank: '矩阵填空', matrixAuto: '表格自增',
    divider: '分割线', description: '说明',
    questionSet: '问题组', pagination: '分页',
    user: '成员', dept: '部门', richText: '富文本',
    autopop: '自动填充',
  }
  return map[t] || t
}
const tagType = (t: string) => {
  const map: Record<string, string> = {
    input: '', text: '', textarea: '', number: '', multiInput: '', hInput: '', scanCode: '',
    select: 'warning', radio: 'warning', checkbox: 'warning', picker: 'warning', judge: 'warning', cascade: 'warning',
    rating: 'success', nps: 'success',
    file: '', signature: '', location: '',
    phone: '', email: '', idCard: '', password: '', switch: '',
    matrixRadio: '', matrixCheckbox: '', matrixFillBlank: '', matrixAuto: '',
    divider: 'info', description: 'info',
    questionSet: 'info', pagination: 'info',
    user: '', dept: '', richText: '', autopop: '',
    date: '', time: '', dateRange: '',
  }
  return map[t] || ''
}
function rateIconSvg(icon: string) {
  const gray = '#c0c4cc'
  if (icon === 'heart') {
    return `<svg viewBox="0 0 24 24" width="18" height="18" fill="${gray}"><path d="M12 21.35l-1.45-1.32C5.4 15.36 2 12.28 2 8.5 2 5.42 4.42 3 7.5 3c1.74 0 3.41.81 4.5 2.09C13.09 3.81 14.76 3 16.5 3 19.58 3 22 5.42 22 8.5c0 3.78-3.4 6.86-8.55 11.54L12 21.35z"/></svg>`
  }
  if (icon === 'smiley') {
    return `<svg viewBox="0 0 24 24" width="18" height="18" fill="${gray}"><path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm5 8h-2v-2h2v2zm-6 0H9v-2h2v2zm1 6c-2.33 0-4.31-1.46-5.11-3.5h10.22c-.8 2.04-2.78 3.5-5.11 3.5z"/></svg>`
  }
  return `<svg viewBox="0 0 24 24" width="18" height="18" fill="${gray}"><path d="M12 17.27L18.18 21l-1.64-7.03L22 9.24l-7.19-.61L12 2 9.19 8.63 2 9.24l5.46 4.73L5.82 21z"/></svg>`
}
function optionGrid(q: any) {
  const cols = q.optionLayout
  if (!cols || cols <= 1) return {}
  return { display: 'grid', gridTemplateColumns: `repeat(${cols}, 1fr)`, gap: '4px' }
}
</script>

<style scoped>
.question-preview {
  padding: 0 0 6px;
}
.question-preview.editing {
  padding: 0 8px 8px;
  border: 1px dashed #fb454c;
  border-radius: 8px;
  background: #fff;
  overflow: visible;
}
.required-tag { border:1px solid #fb454c !important; }
.question-preview.is-hidden {
  opacity: 0.4;
  filter: grayscale(0.6);
  position: relative;
}
.question-preview.is-hidden::after {
  content: '已隐藏';
  position: absolute;
  left: 50%;
  top: 50%;
  transform: translate(-50%, -50%);
  font-size: 12px;
  color: #909399;
  background: rgba(255,255,255,0.9);
  padding: 2px 10px;
  border-radius: 4px;
  z-index: 1;
  pointer-events: none;
}
.preview-header {
  margin-bottom: 4px;
  overflow: hidden;
}
.preview-type-tag {
  float: left;
  margin-right: 6px;
}
.preview-title {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
  word-break: break-word;
  line-height: 1.5;
  padding-top: -1px;
}
.preview-title :deep(p) {
  margin: 0;
}
.preview-actions {
  position: absolute; right: 0; top: 2px;
  display: flex;
  background: #f5f5f5; border-radius: 5px; overflow: hidden; font-size: 0;
}
.preview-actions :deep(.el-button) { padding: 2px 3px; }
.preview-desc {
  font-size: 12px;
  color: #909399;
  margin-bottom: 6px;
}
.preview-body {
  min-height: 28px;
}
.preview-media {
  margin-bottom: 8px;
}
.media-preview {
  border-radius: 6px;
  border: 1px solid #eee;
}
.preview-options {
  display: flex;
  flex-direction: column;
  gap: 4px;
  width: 100%;
}
.preview-radio-group,
.preview-checkbox-group {
  align-items: flex-start;
  text-align: left;
}
.preview-plain {
  font-size: 13px;
  color: #606266;
  padding: 4px 0;
  white-space: pre-wrap;
  word-break: break-word;
}
.preview-matrix {
  border: 1px solid #d0d0d0;
  border-radius: 6px;
  overflow-x: auto;
}
.preview-matrix table {
  width: 100%;
  border-collapse: collapse;
  white-space: nowrap;
}
.preview-matrix th,
.preview-matrix td {
  padding: 6px 10px;
  border: 1px solid #e0e0e0;
  font-size: 12px;
  text-align: center;
  min-width: 70px;
}
.preview-matrix th {
  background: #f0f2f5;
  font-weight: 500;
  color: #666;
}
.preview-matrix th.corner {
  color: #999;
  font-weight: 400;
}
.preview-matrix td.matrix-label {
  background: #fafafa;
  color: #606266;
  font-weight: 400;
  text-align: left;
  min-width: 60px;
}
.preview-nps {
  padding: 4px 0;
}
.preview-rate-icons,
.rate-icons-preview {
  display: inline-flex;
  align-items: center;
  gap: 4px;
}
.preview-rate-icons span,
.rate-icons-preview {
  line-height: 1;
}
.nps-labels {
  display: flex;
  justify-content: space-between;
  font-size: 11px;
  color: #909399;
  margin-bottom: 2px;
}

/* ======= 编辑模式 ======= */
.edit-title-line {
  display: flex;
  align-items: flex-start;
  gap: 4px;
  margin-bottom: 6px;
  position: relative;
  padding-right: 140px;
  overflow: visible;
}
.title-editor-wrap {
  flex: 1;
  min-width: 0;
  font-size: 14px;
  font-weight: 500;
  position: relative;
  overflow: visible;
  min-height: 28px;
}
.title-editor-wrap :deep(.ql-container) {
  min-height: 28px;
  font-size: 14px;
  font-family: inherit;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  height: auto;
}
.title-editor-wrap :deep(.ql-editor) {
  padding: 4px 6px;
  line-height: 1.5;
  color: #303133;
  height: auto;
}
.title-editor-wrap :deep(.ql-toolbar) {
  position: absolute;
  top: -30px;
  left: 0;
  right: 0;
  z-index: 10;
  padding: 2px 4px;
  background: rgba(255,255,255,0.95);
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  opacity: 0;
  pointer-events: none;
  transition: opacity 0.2s;
}
.title-editor-wrap:focus-within :deep(.ql-toolbar) {
  opacity: 1;
  pointer-events: auto;
}
.title-editor-wrap :deep(.ql-toolbar .ql-formats) {
  margin-right: 6px;
}
.edit-desc-line {
  margin-bottom: 6px;
}
.desc-editable {
  font-size: 12px;
  color: #909399;
  padding: 3px 6px;
  border: 1px solid transparent;
  border-radius: 4px;
  outline: none;
  min-height: 20px;
}
.desc-editable:focus {
  border-color: #ddd;
  background: #fff;
}
.desc-editable:empty::before {
  content: attr(placeholder);
  color: #ddd;
}
.signature-pad-wrap {
  border: 1px dashed #d9d9d9;
  border-radius: 6px;
  overflow: hidden;
  padding: 4px;
}
.sig-canvas {
  display: block;
  width: 100%;
  height: 120px;
  cursor: crosshair;
  touch-action: none;
}
.edit-options-area {
  margin: 8px 0;
  padding: 8px;
  background: #fafafa;
  border-radius: 6px;
}
.edit-option-row {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-bottom: 4px;
  padding: 3px 4px;
  border-radius: 4px;
  transition: background 0.15s;
}
.edit-option-row:hover {
  background: #fff;
}
.option-icon {
  flex-shrink: 0;
  font-size: 14px;
  color: #666;
  width: 16px;
  text-align: center;
}
.option-editable {
  flex: 1;
  font-size: 13px;
  color: #303133;
  padding: 3px 6px;
  border: 1px solid transparent;
  border-radius: 4px;
  outline: none;
  min-width: 0;
  line-height: 1.5;
}
.option-editable:focus {
  border-color: #fb454c;
  background: #fff;
}
.option-editor-wrap {
  flex: 1;
  min-width: 0;
  position: relative;
  overflow: visible;
}
.option-editor-wrap :deep(.ql-toolbar) {
  position: absolute;
  top: -30px;
  left: 0;
  right: 0;
  z-index: 10;
  padding: 2px 4px;
  background: rgba(255,255,255,0.95);
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  opacity: 0;
  pointer-events: none;
  transition: opacity 0.2s;
}
.option-editor-wrap:focus-within :deep(.ql-toolbar) {
  opacity: 1;
  pointer-events: auto;
}
.option-editor-wrap :deep(.ql-container) {
  border: 1px solid transparent;
  border-radius: 4px;
  font-size: 13px;
  font-family: inherit;
  height: auto;
  min-height: 28px;
}
.option-editor-wrap :deep(.ql-editor) {
  padding: 3px 6px;
  line-height: 1.5;
  height: auto;
}
.option-editor-wrap:focus-within :deep(.ql-container) {
  border-color: #fb454c;
  background: #fff;
}
.option-editor-wrap :deep(.ql-toolbar .ql-formats) {
  margin-right: 6px;
}
.opt-del-btn {
  opacity: 0.3;
  transition: opacity 0.15s;
}
.edit-option-row:hover .opt-del-btn {
  opacity: 1;
}
.add-option-btn {
  margin-top: 6px;
}
</style>
