<template>
  <div class="form-designer">
    <aside class="field-palette">
      <div class="panel-heading">
        <div>
          <strong>字段组件</strong>
          <span>{{ fieldTypes.length }} 个组件</span>
        </div>
      </div>
      <div class="palette-content">
        <section v-for="group in fieldGroups" :key="group.label" class="palette-group">
          <h3>{{ group.label }}</h3>
          <div class="palette-grid">
            <button
              v-for="item in group.items"
              :key="item.type"
              type="button"
              :disabled="readonly"
              :class="{ dragging: isPaletteDragging(item.type) }"
              draggable="true"
              @click="addField(item.type)"
              @dragstart="handlePaletteDragStart(item.type, $event)"
              @dragend="handleDragEnd"
            >
              <span class="palette-type-icon"><el-icon><component :is="item.icon" /></el-icon></span>
              <span>{{ item.label }}</span>
              <el-icon class="palette-add-icon"><Plus /></el-icon>
            </button>
          </div>
        </section>
      </div>
    </aside>

    <main class="form-canvas">
      <div class="canvas-heading">
        <div>
          <strong>流程表单</strong>
          <span>共 {{ fields.length }} 个字段</span>
        </div>
        <el-button v-if="fields.length && !readonly" size="small" icon="Plus" @click="addField('text')">添加字段</el-button>
      </div>

      <div class="canvas-stage">
        <section class="form-sheet">
          <div
            v-if="fields.length === 0"
            class="empty-canvas-drop"
            :class="{ active: isDropTarget(null, 0) }"
            @dragover.prevent
            @dragenter="handleDragEnter(null, 0)"
            @drop.prevent="handleDrop(null, 0, $event)"
          >
            <el-empty :image-size="88" description="从左侧选择或拖入字段开始设计表单" />
          </div>
          <div v-else class="field-list">
            <article
              v-for="(field, index) in fields"
              :key="field.key"
              class="field-item"
              :class="{
                active: selectedField === field,
                dragging: dragState?.fieldKey === field.key,
                'drop-before': isDropTarget(null, index),
                'field-item--compact': fieldSpan(field) <= 8,
                'field-item--group': field.type === 'group',
              }"
              :style="{ gridColumn: `span ${fieldSpan(field)}` }"
              :draggable="!readonly"
              @click="selectField(field)"
              @dragstart="handleDragStart(field, null, index, $event)"
              @dragend="handleDragEnd"
              @dragover.prevent
              @dragenter="handleDragEnter(null, index)"
              @drop.prevent="handleDrop(null, index, $event)"
            >
              <div class="field-item__content">
                <div class="field-item__heading">
                  <div class="field-item__main">
                    <button
                      v-if="!readonly"
                      type="button"
                      class="field-drag-handle"
                      draggable="true"
                      title="拖动调整字段顺序"
                      @click.stop
                      @dragstart.stop="handleDragStart(field, null, index, $event)"
                      @dragend.stop="handleDragEnd"
                    >
                      <el-icon><Rank /></el-icon>
                    </button>
                    <span class="field-type-icon"><el-icon><component :is="fieldTypeMeta(field.type).icon" /></el-icon></span>
                    <div>
                      <strong>{{ field.label || '未命名字段' }}<i v-if="field.required">*</i></strong>
                      <p>{{ fieldTypeMeta(field.type).label }} · {{ field.key }} · {{ fieldSpanLabel(field) }}</p>
                    </div>
                  </div>
                  <div v-if="!readonly" class="field-actions" @click.stop>
                    <el-button circle size="small" type="danger" plain icon="Delete" title="删除" @click="removeField(field)" />
                  </div>
                </div>

                <div v-if="field.type === 'group'" class="group-fields" @click.stop>
                  <div class="group-fields__grid">
                    <article
                      v-for="(child, childIndex) in groupFields(field)"
                      :key="child.key"
                      class="field-item field-item--nested"
                      :class="{
                        active: selectedField === child,
                        dragging: dragState?.fieldKey === child.key,
                        'drop-before': isDropTarget(field.key, childIndex),
                        'field-item--compact': fieldSpan(child) <= 8,
                      }"
                      :style="{ gridColumn: `span ${fieldSpan(child)}` }"
                      :draggable="!readonly"
                      @click.stop="selectField(child)"
                      @dragstart.stop="handleDragStart(child, field.key, childIndex, $event)"
                      @dragend.stop="handleDragEnd"
                      @dragover.prevent.stop
                      @dragenter.stop="handleDragEnter(field.key, childIndex)"
                      @drop.stop.prevent="handleDrop(field.key, childIndex, $event)"
                    >
                      <div class="field-item__heading">
                        <div class="field-item__main">
                          <button
                            v-if="!readonly"
                            type="button"
                            class="field-drag-handle"
                            draggable="true"
                            title="拖动调整组件位置"
                            @click.stop
                            @dragstart.stop="handleDragStart(child, field.key, childIndex, $event)"
                            @dragend.stop="handleDragEnd"
                          >
                            <el-icon><Rank /></el-icon>
                          </button>
                          <span class="field-type-icon"><el-icon><component :is="fieldTypeMeta(child.type).icon" /></el-icon></span>
                          <div>
                            <strong>{{ child.label || '未命名组件' }}<i v-if="child.required">*</i></strong>
                            <p>{{ fieldTypeMeta(child.type).label }} · {{ child.key }} · {{ fieldSpanLabel(child) }}</p>
                          </div>
                        </div>
                        <div v-if="!readonly" class="field-actions" @click.stop>
                          <el-button circle size="small" type="danger" plain icon="Delete" title="删除" @click="removeField(child)" />
                        </div>
                      </div>
                      <WorkflowFormFieldPreview :field="child" />
                    </article>
                    <div
                      class="field-drop-zone field-drop-zone--tail group-drop-zone"
                      :class="{
                        active: isDropTarget(field.key, groupFields(field).length),
                        'group-drop-zone--empty': groupFields(field).length === 0,
                        'is-dragging': Boolean(dragState && canDropIntoGroup(field)),
                      }"
                      @dragover.prevent.stop
                      @dragenter.stop="handleDragEnter(field.key, groupFields(field).length)"
                      @drop.stop.prevent="handleDrop(field.key, groupFields(field).length, $event)"
                    >
                      <el-icon><Plus /></el-icon>
                      <span v-if="dragState && canDropIntoGroup(field)">放置到分组</span>
                    </div>
                  </div>
                </div>

                <div v-else class="field-preview" @click.stop="selectField(field)">
                  <el-input
                    v-if="field.type === 'text'"
                    :model-value="stringDefault(field.default)"
                    :placeholder="field.placeholder || '请输入内容'"
                    disabled
                  />
                  <el-input
                    v-else-if="field.type === 'textarea'"
                    :model-value="stringDefault(field.default)"
                    :placeholder="field.placeholder || '请输入内容'"
                    type="textarea"
                    :autosize="textareaAutosize(field, 3, 8)"
                    resize="none"
                    disabled
                  />
                  <el-input-number
                    v-else-if="field.type === 'number' || field.type === 'amount'"
                    :model-value="numberDefault(field.default)"
                    :placeholder="field.placeholder || (field.type === 'amount' ? '请输入金额' : '请输入数字')"
                    :precision="field.type === 'amount' ? 2 : undefined"
                    controls-position="right"
                    disabled
                  />
                  <el-input
                    v-else-if="field.type === 'phone' || field.type === 'email'"
                    :model-value="stringDefault(field.default)"
                    :placeholder="field.placeholder || (field.type === 'phone' ? '请输入手机号' : '请输入邮箱')"
                    disabled
                  >
                    <template #suffix><el-icon><component :is="field.type === 'phone' ? 'Cellphone' : 'Message'" /></el-icon></template>
                  </el-input>
                  <el-tree-select
                    v-else-if="isDropdownField(field) && shouldUseTreeSelect(field)"
                    :model-value="selectDefault(field)"
                    :data="fieldOptions(field)"
                    :multiple="field.type === 'multi_select'"
                    :props="optionTreeProps"
                    node-key="value"
                    check-strictly
                    :placeholder="field.placeholder || '请选择'"
                    disabled
                  />
                  <el-select
                    v-else-if="isDropdownField(field)"
                    :model-value="selectDefault(field)"
                    :multiple="field.type === 'multi_select'"
                    :placeholder="field.placeholder || '请选择'"
                    disabled
                  >
                    <el-option v-for="option in fieldOptions(field)" :key="option.value" :label="option.label" :value="option.value" />
                  </el-select>
                  <el-radio-group v-else-if="field.type === 'radio'" :model-value="stringDefault(field.default)" disabled>
                    <el-radio v-for="option in flatFieldOptions(field)" :key="option.value" :value="option.value">{{ option.label }}</el-radio>
                  </el-radio-group>
                  <el-checkbox-group v-else-if="field.type === 'checkbox'" :model-value="arrayDefault(field.default)" disabled>
                    <el-checkbox v-for="option in flatFieldOptions(field)" :key="option.value" :value="option.value">{{ option.label }}</el-checkbox>
                  </el-checkbox-group>
                  <el-input
                    v-else-if="field.type === 'date' || field.type === 'datetime'"
                    :model-value="stringDefault(field.default)"
                    :placeholder="field.placeholder || (field.type === 'date' ? '请选择日期' : '请选择日期时间')"
                    disabled
                  >
                    <template #suffix><el-icon><component :is="field.type === 'date' ? 'Calendar' : 'Clock'" /></el-icon></template>
                  </el-input>
                  <el-time-picker
                    v-else-if="field.type === 'time'"
                    :model-value="null"
                    :placeholder="field.placeholder || '请选择时间'"
                    disabled
                  />
                  <el-date-picker
                    v-else-if="field.type === 'date_range'"
                    :model-value="[]"
                    type="daterange"
                    range-separator="至"
                    start-placeholder="开始日期"
                    end-placeholder="结束日期"
                    disabled
                  />
                  <el-input
                    v-else-if="['user', 'user_multi', 'department', 'department_multi'].includes(field.type)"
                    :placeholder="field.placeholder || organizationPlaceholder(field.type)"
                    disabled
                  >
                    <template #suffix><el-icon><component :is="field.type.startsWith('user') ? 'User' : 'OfficeBuilding'" /></el-icon></template>
                  </el-input>
                  <el-button v-else-if="field.type === 'attachment'" icon="Upload" disabled>选择附件</el-button>
                  <div v-else-if="field.type === 'detail_list'" class="detail-preview">
                    <div class="detail-preview__grid">
                      <div
                        v-for="column in detailColumns(field)"
                        :key="column.key"
                        class="detail-preview__cell"
                        :style="{ gridColumn: `span ${fieldSpan(column)}` }"
                      >
                        <span class="detail-preview__label">{{ column.label || column.key }}</span>
                        <span class="detail-preview__control">{{ detailColumnPlaceholder(column) }}</span>
                      </div>
                    </div>
                    <el-button size="small" icon="Plus" disabled>新增行</el-button>
                  </div>
                  <el-switch v-else :model-value="Boolean(field.default)" disabled />
                </div>
              </div>
            </article>
            <div
              class="field-drop-zone field-drop-zone--tail root-tail-drop-zone"
              :class="{ active: isDropTarget(null, fields.length), 'is-dragging': Boolean(dragState) }"
              @dragover.prevent
              @dragenter="handleDragEnter(null, fields.length)"
              @drop.prevent="handleDrop(null, fields.length, $event)"
            >
              <el-icon><Plus /></el-icon>
              <span v-if="dragState">放置到表单末尾</span>
            </div>
          </div>
        </section>
      </div>
    </main>

    <aside class="property-panel">
      <div class="panel-heading">
        <div>
          <strong>字段属性</strong>
          <span>{{ selectedField ? fieldTypeMeta(selectedField.type).label : '请选择字段' }}</span>
        </div>
      </div>
      <el-empty v-if="!selectedField" :image-size="72" description="选择中间字段后配置属性" />
      <el-form v-else label-position="top" class="property-form">
        <section class="property-section">
          <h3>基础信息</h3>
          <el-form-item :label="componentNameLabel(selectedField)" required>
            <el-input v-model="selectedField.label" :maxlength="fieldLabelMaxLength(selectedField)" :disabled="readonly" @input="emitChange" />
          </el-form-item>
          <el-form-item label="组件编码" required>
            <el-input v-model="selectedField.key" maxlength="100" :disabled="readonly" @input="emitChange">
              <template #append><el-tooltip content="以字母开头，可使用字母、数字、点、下划线和中划线"><el-icon><QuestionFilled /></el-icon></el-tooltip></template>
            </el-input>
          </el-form-item>
          <el-form-item v-if="isWorkflowDataField(selectedField)" label="提示文字">
            <el-input v-model="selectedField.placeholder" maxlength="120" :disabled="readonly" @input="emitChange" />
          </el-form-item>
        </section>

        <section v-if="selectedField.type === 'description'" class="property-section">
          <h3>说明内容</h3>
          <el-form-item label="静态说明" required>
            <el-input v-model="selectedField.content" type="textarea" :rows="8" maxlength="2000" show-word-limit resize="vertical" :disabled="readonly" @input="emitChange" />
          </el-form-item>
        </section>

        <section v-if="selectedField.type === 'button'" class="property-section">
          <h3>说明设置</h3>
          <el-form-item label="说明标题" required>
            <el-input :model-value="selectedField.help?.title || ''" maxlength="100" :disabled="readonly" @input="updateHelpField('title', $event)" />
          </el-form-item>
          <el-form-item label="说明内容" required>
            <el-input :model-value="selectedField.help?.content || ''" type="textarea" :rows="8" maxlength="2000" show-word-limit resize="vertical" :disabled="readonly" @input="updateHelpField('content', $event)" />
          </el-form-item>
        </section>

        <section v-else-if="supportsFieldHelp(selectedField)" class="property-section">
          <h3>说明设置</h3>
          <div class="required-setting">
            <span>显示查看说明按钮</span>
            <el-switch :model-value="Boolean(selectedField.help)" :disabled="readonly" @change="toggleFieldHelp" />
          </div>
          <template v-if="selectedField.help">
            <el-form-item label="按钮文字">
              <el-input v-model="selectedField.help.buttonText" maxlength="30" placeholder="查看说明" :disabled="readonly" @input="emitChange" />
            </el-form-item>
            <el-form-item label="说明标题" required>
              <el-input v-model="selectedField.help.title" maxlength="100" :disabled="readonly" @input="emitChange" />
            </el-form-item>
            <el-form-item label="说明内容" required>
              <el-input v-model="selectedField.help.content" type="textarea" :rows="8" maxlength="2000" show-word-limit resize="vertical" :disabled="readonly" @input="emitChange" />
            </el-form-item>
          </template>
        </section>

        <section class="property-section layout-setting">
          <h3>布局设置</h3>
          <el-form-item label="字段宽度">
            <el-radio-group :model-value="fieldSpan(selectedField)" :disabled="readonly" @change="updateFieldSpan">
              <el-radio-button v-for="item in fieldSpanOptions" :key="item.value" :value="item.value">
                {{ item.label }}
              </el-radio-button>
            </el-radio-group>
          </el-form-item>
          <p>PC 端按栅格同行排列，移动端自动切换为整行。</p>
        </section>

        <section v-if="isWorkflowDataField(selectedField)" class="property-section">
          <h3>校验规则</h3>
          <div class="required-setting">
            <span>必填字段</span>
            <el-switch v-model="selectedField.required" :disabled="readonly" @change="emitChange" />
          </div>
          <el-form-item v-if="['text', 'textarea', 'phone', 'email'].includes(selectedField.type)" label="最大长度">
            <el-input-number
              v-model="selectedField.maxLength"
              :min="0"
              :max="100000"
              :disabled="readonly"
              controls-position="right"
              @change="emitChange"
            />
          </el-form-item>
          <div v-if="selectedField.type === 'textarea'" class="textarea-visible-rows number-range">
            <el-form-item label="最小显示行数">
              <el-input-number
                :model-value="textareaAutosize(selectedField, 3, 8).minRows"
                :min="1"
                :max="30"
                :disabled="readonly"
                controls-position="right"
                @change="updateTextareaVisibleRows(selectedField, 'minVisibleRows', $event, 3, 8)"
              />
            </el-form-item>
            <el-form-item label="最大显示行数">
              <el-input-number
                :model-value="textareaAutosize(selectedField, 3, 8).maxRows"
                :min="1"
                :max="30"
                :disabled="readonly"
                controls-position="right"
                @change="updateTextareaVisibleRows(selectedField, 'maxVisibleRows', $event, 3, 8)"
              />
            </el-form-item>
          </div>
          <div class="number-range">
            <el-form-item v-if="selectedField.type === 'number' || selectedField.type === 'amount'" label="最小值"><el-input-number v-model="selectedField.min" placeholder="不限制" :disabled="readonly" controls-position="right" @change="emitChange" /></el-form-item>
            <el-form-item v-if="selectedField.type === 'number' || selectedField.type === 'amount'" label="最大值"><el-input-number v-model="selectedField.max" placeholder="不限制" :disabled="readonly" controls-position="right" @change="emitChange" /></el-form-item>
          </div>
          <WorkflowValidationRulesEditor
            :field="selectedField"
            :fields="props.fields"
            :readonly="readonly"
            @change="emitChange"
          />
        </section>

        <section v-if="supportsDefault(selectedField)" class="property-section">
          <h3>默认值</h3>
          <el-form-item>
            <el-switch
              v-if="selectedField.type === 'boolean'"
              :model-value="Boolean(selectedField.default)"
              :disabled="readonly"
              @change="updateDefault"
            />
            <el-input-number
              v-else-if="selectedField.type === 'number' || selectedField.type === 'amount'"
              :model-value="numberDefault(selectedField.default)"
              :precision="selectedField.type === 'amount' ? 2 : undefined"
              :disabled="readonly"
              controls-position="right"
              @change="updateDefault"
            />
            <el-tree-select
              v-else-if="isOptionField(selectedField) && shouldUseTreeSelect(selectedField)"
              :model-value="selectedField.default"
              :data="fieldOptions(selectedField)"
              :multiple="selectedField.type === 'multi_select' || selectedField.type === 'checkbox'"
              :props="optionTreeProps"
              node-key="value"
              check-strictly
              clearable
              :disabled="readonly"
              style="width: 100%"
              @change="updateDefault"
            />
            <el-select
              v-else-if="isOptionField(selectedField)"
              :model-value="selectedField.default"
              :multiple="selectedField.type === 'multi_select' || selectedField.type === 'checkbox'"
              clearable
              :disabled="readonly"
              style="width: 100%"
              @change="updateDefault"
            >
              <el-option v-for="option in flatFieldOptions(selectedField)" :key="option.value" :label="option.label" :value="option.value" />
            </el-select>
            <el-input
              v-else
              :model-value="stringDefault(selectedField.default)"
              clearable
              :disabled="readonly"
              @input="updateDefault"
            />
          </el-form-item>
        </section>

        <section v-if="isOptionField(selectedField)" class="property-section option-editor">
          <h3>选项配置</h3>
          <el-form-item label="选项来源">
            <el-radio-group :model-value="optionSourceType(selectedField)" :disabled="readonly" @change="updateOptionSourceType">
              <el-radio-button value="static">静态选项</el-radio-button>
              <el-radio-button value="api">后端接口</el-radio-button>
            </el-radio-group>
          </el-form-item>

          <template v-if="optionSourceType(selectedField) === 'static'">
            <div class="option-editor__heading">
              <h3>静态选项</h3>
              <el-button v-if="!readonly" link type="primary" icon="Plus" @click="addOption">新增选项</el-button>
            </div>
            <div v-for="(option, index) in selectedField.options" :key="index" class="option-row">
              <span class="option-index">{{ index + 1 }}</span>
              <div class="option-row__inputs">
                <el-input v-model="option.label" placeholder="选项名称" :disabled="readonly" @input="emitChange" />
                <el-input v-model="option.value" placeholder="选项值" :disabled="readonly" @input="emitChange" />
              </div>
              <el-button circle size="small" type="danger" plain icon="Delete" :disabled="readonly || (selectedField.options?.length || 0) <= 1" @click="removeOption(index)" />
            </div>

            <el-form-item label="JSON" class="option-json-input">
              <el-input
                v-model="optionJsonText"
                type="textarea"
                :rows="8"
                resize="vertical"
                :disabled="readonly"
                placeholder='[{"label":"华东","value":"east","children":[{"label":"上海","value":"shanghai"}]}]'
              />
            </el-form-item>
            <div class="option-json-actions">
              <el-button size="small" :disabled="readonly" @click="formatOptionsJson">格式化</el-button>
              <el-button size="small" type="primary" :disabled="readonly" @click="applyOptionsJson">应用 JSON</el-button>
            </div>
          </template>

          <div v-else class="option-api-config">
            <div class="option-editor__heading">
              <h3>接口配置</h3>
            </div>
            <el-form-item label="后端接口">
              <el-input v-model="selectedField.optionSource!.url" :disabled="readonly" placeholder="/api/v2/admin/departments/tree" @input="emitChange" />
            </el-form-item>
            <div class="option-api-grid">
              <el-form-item label="请求方式">
                <el-select v-model="selectedField.optionSource!.method" :disabled="readonly" @change="emitChange">
                  <el-option label="GET" value="GET" />
                  <el-option label="POST" value="POST" />
                </el-select>
              </el-form-item>
              <el-form-item label="responsePath">
                <el-input v-model="selectedField.optionSource!.responsePath" :disabled="readonly" placeholder="data" @input="emitChange" />
              </el-form-item>
              <el-form-item label="labelField">
                <el-input v-model="selectedField.optionSource!.labelField" :disabled="readonly" placeholder="name" @input="emitChange" />
              </el-form-item>
              <el-form-item label="valueField">
                <el-input v-model="selectedField.optionSource!.valueField" :disabled="readonly" placeholder="id" @input="emitChange" />
              </el-form-item>
              <el-form-item label="childrenField">
                <el-input v-model="selectedField.optionSource!.childrenField" :disabled="readonly" placeholder="children" @input="emitChange" />
              </el-form-item>
            </div>
          </div>
        </section>

        <section v-if="selectedField.type === 'detail_list'" class="property-section detail-editor">
          <h3>明细设置</h3>
          <div class="number-range">
            <el-form-item label="行标识">
              <el-input v-model="selectedField.rowKey" maxlength="40" :disabled="readonly" placeholder="id" @input="emitChange" />
            </el-form-item>
            <el-form-item label="最大行数">
              <el-input-number v-model="selectedField.maxRows" :min="0" :max="1000" :disabled="readonly" controls-position="right" @change="emitChange" />
            </el-form-item>
          </div>
          <el-form-item label="最小行数">
            <el-input-number v-model="selectedField.minRows" :min="0" :max="1000" :disabled="readonly" controls-position="right" @change="emitChange" />
          </el-form-item>
          <div class="option-editor__heading">
            <h3>明细列</h3>
            <el-button v-if="!readonly" link type="primary" icon="Plus" @click="addDetailColumn">新增列</el-button>
          </div>
          <div
            v-for="(column, index) in selectedField.columns"
            :key="`${column.key}_${index}`"
            class="detail-column-row"
            :class="{
              dragging: detailColumnDragIndex === index,
              'drop-before': detailColumnDropIndex === index,
            }"
            @dragenter.prevent="handleDetailColumnDragOver(index)"
            @dragover.prevent="handleDetailColumnDragOver(index)"
            @drop.prevent="handleDetailColumnDrop"
          >
            <button
              class="detail-column-drag-handle"
              type="button"
              :draggable="!readonly"
              :disabled="readonly"
              title="拖拽调整列顺序"
              @dragstart="handleDetailColumnDragStart(index, $event)"
              @dragend="handleDetailColumnDragEnd"
            >
              <el-icon><Rank /></el-icon>
              <span>{{ index + 1 }}</span>
            </button>
            <div class="detail-column-row__inputs">
              <el-input v-model="column.label" placeholder="列名称" :disabled="readonly" @input="emitChange" />
              <el-input v-model="column.key" placeholder="列编码" :disabled="readonly" @input="emitChange" />
              <el-select v-model="column.type" :disabled="readonly" @change="handleDetailColumnTypeChange(column)">
                <el-option v-for="item in detailColumnTypes" :key="item.type" :label="item.label" :value="item.type" />
              </el-select>
              <div class="detail-column-required">
                <span>必填</span>
                <el-switch v-model="column.required" :disabled="readonly" @change="emitChange" />
              </div>
              <div v-if="column.type === 'number' || column.type === 'amount'" class="detail-column-range">
                <el-form-item label="最小值">
                  <el-input-number
                    :model-value="column.min"
                    placeholder="不限制"
                    :disabled="readonly"
                    controls-position="right"
                    @change="updateDetailColumnRange(column, 'min', $event)"
                  />
                </el-form-item>
                <el-form-item label="最大值">
                  <el-input-number
                    :model-value="column.max"
                    placeholder="不限制"
                    :disabled="readonly"
                    controls-position="right"
                    @change="updateDetailColumnRange(column, 'max', $event)"
                  />
                </el-form-item>
              </div>
              <el-form-item v-if="['text', 'textarea', 'phone', 'email'].includes(column.type)" class="detail-column-max-length" label="最大长度">
                <el-input-number
                  :model-value="column.maxLength"
                  :min="1"
                  :max="100000"
                  placeholder="不限制"
                  :disabled="readonly"
                  controls-position="right"
                  @change="updateDetailColumnMaxLength(column, $event)"
                />
              </el-form-item>
              <div v-if="column.type === 'textarea'" class="textarea-visible-rows number-range">
                <el-form-item label="最小显示行数">
                  <el-input-number
                    :model-value="textareaAutosize(column, 2, 6).minRows"
                    :min="1"
                    :max="30"
                    :disabled="readonly"
                    controls-position="right"
                    @change="updateTextareaVisibleRows(column, 'minVisibleRows', $event, 2, 6)"
                  />
                </el-form-item>
                <el-form-item label="最大显示行数">
                  <el-input-number
                    :model-value="textareaAutosize(column, 2, 6).maxRows"
                    :min="1"
                    :max="30"
                    :disabled="readonly"
                    controls-position="right"
                    @change="updateTextareaVisibleRows(column, 'maxVisibleRows', $event, 2, 6)"
                  />
                </el-form-item>
              </div>
              <div class="detail-column-layout">
                <span>列宽</span>
                <el-radio-group :model-value="fieldSpan(column)" size="small" :disabled="readonly" @change="updateDetailColumnSpan(column, $event)">
                  <el-radio-button v-for="item in fieldSpanOptions" :key="item.value" :value="item.value">
                    {{ item.label }}
                  </el-radio-button>
                </el-radio-group>
              </div>
              <WorkflowValidationRulesEditor
                class="detail-column-validation"
                :field="column"
                :fields="detailColumns(selectedField)"
                :readonly="readonly"
                compact
                @change="emitChange"
              />
            </div>
            <div class="detail-column-actions">
              <el-button circle text size="small" icon="ArrowUp" title="上移一位" :disabled="readonly || index === 0" @click="moveDetailColumnByOffset(index, -1)" />
              <el-button circle text size="small" icon="ArrowDown" title="下移一位" :disabled="readonly || index === detailColumns(selectedField).length - 1" @click="moveDetailColumnByOffset(index, 1)" />
              <el-button circle text size="small" type="danger" icon="Delete" title="删除列" :disabled="readonly || detailColumns(selectedField).length <= 1" @click="removeDetailColumn(index)" />
            </div>
          </div>
          <div
            v-if="detailColumnDragIndex !== null"
            class="detail-column-tail-drop"
            :class="{ active: detailColumnDropIndex === detailColumns(selectedField).length }"
            @dragenter.prevent="handleDetailColumnTailDragOver"
            @dragover.prevent="handleDetailColumnTailDragOver"
            @drop.prevent="handleDetailColumnDrop"
          >
            <el-icon><Bottom /></el-icon>
            <span>放到末尾</span>
          </div>
        </section>
      </el-form>
    </aside>
  </div>
</template>

<script lang="ts" setup>
import { ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { WorkflowFormField, WorkflowFormFieldSpan, WorkflowFormFieldType, WorkflowFormOption, WorkflowOptionSourceType } from '../../types'
import { insertWorkflowField, isWorkflowDataField, moveWorkflowDetailColumn, moveWorkflowField, removeWorkflowField, workflowFieldByKey } from '../../formLayout'
import { flattenWorkflowOptions, hasWorkflowOptionChildren, normalizeWorkflowOptions, workflowTextareaAutosize as textareaAutosize } from '../../runtimeForm'
import WorkflowFormFieldPreview from './WorkflowFormFieldPreview.vue'
import WorkflowValidationRulesEditor from './WorkflowValidationRulesEditor.vue'

const props = defineProps<{ fields: WorkflowFormField[]; readonly?: boolean }>()
const emit = defineEmits<{ change: [] }>()

const fieldTypes: Array<{ type: WorkflowFormFieldType; label: string; icon: string }> = [
  { type: 'text', label: '单行文本', icon: 'EditPen' },
  { type: 'textarea', label: '多行文本', icon: 'Document' },
  { type: 'number', label: '数字', icon: 'Odometer' },
  { type: 'amount', label: '金额', icon: 'Money' },
  { type: 'phone', label: '手机号', icon: 'Cellphone' },
  { type: 'email', label: '邮箱', icon: 'Message' },
  { type: 'boolean', label: '开关', icon: 'Switch' },
  { type: 'select', label: '单选下拉', icon: 'Select' },
  { type: 'multi_select', label: '多选下拉', icon: 'Finished' },
  { type: 'radio', label: '单选框', icon: 'CircleCheck' },
  { type: 'checkbox', label: '复选框', icon: 'Checked' },
  { type: 'date', label: '日期', icon: 'Calendar' },
  { type: 'datetime', label: '日期时间', icon: 'Clock' },
  { type: 'time', label: '时间', icon: 'Timer' },
  { type: 'date_range', label: '日期区间', icon: 'Calendar' },
  { type: 'user', label: '人员', icon: 'User' },
  { type: 'user_multi', label: '多人', icon: 'UserFilled' },
  { type: 'department', label: '部门', icon: 'OfficeBuilding' },
  { type: 'department_multi', label: '多部门', icon: 'CopyDocument' },
  { type: 'attachment', label: '附件', icon: 'Paperclip' },
  { type: 'detail_list', label: '明细列表', icon: 'Tickets' },
  { type: 'group', label: '表单组', icon: 'FolderOpened' },
  { type: 'label', label: '标签', icon: 'CollectionTag' },
  { type: 'description', label: '说明', icon: 'InfoFilled' },
  { type: 'button', label: '说明按钮', icon: 'Pointer' },
]
const fieldGroups = [
  { label: '布局与说明', items: fieldTypes.filter(item => ['group', 'label', 'description', 'button'].includes(item.type)) },
  { label: '基础字段', items: fieldTypes.filter(item => ['text', 'textarea', 'number', 'amount', 'phone', 'email', 'boolean'].includes(item.type)) },
  { label: '选择与时间', items: fieldTypes.filter(item => ['select', 'multi_select', 'radio', 'checkbox', 'date', 'datetime', 'time', 'date_range'].includes(item.type)) },
  { label: '组织与附件', items: fieldTypes.filter(item => ['user', 'user_multi', 'department', 'department_multi', 'attachment'].includes(item.type)) },
  { label: '明细字段', items: fieldTypes.filter(item => ['detail_list'].includes(item.type)) },
]
const detailColumnTypes = fieldTypes.filter(item => !['detail_list', 'attachment', 'user', 'user_multi', 'department', 'department_multi', 'group', 'label', 'description', 'button'].includes(item.type))
const fieldSpanOptions: Array<{ label: string; value: WorkflowFormFieldSpan }> = [
  { label: '1/4', value: 6 },
  { label: '1/3', value: 8 },
  { label: '1/2', value: 12 },
  { label: '整行', value: 24 },
]

const selectedField = ref<WorkflowFormField | null>(null)
const dragState = ref<{
  kind: 'field' | 'palette'
  containerKey: string | null
  index: number
  fieldKey: string
  fieldType: WorkflowFormFieldType
} | null>(null)
const dropTarget = ref<{ containerKey: string | null; index: number } | null>(null)
const detailColumnDragIndex = ref<number | null>(null)
const detailColumnDropIndex = ref<number | null>(null)
const optionJsonText = ref('')
const optionTreeProps = { label: 'label', value: 'value', children: 'children' }

watch(() => props.fields, (fields) => {
  if (!fields.length) selectedField.value = null
  else if (!selectedField.value || workflowFieldByKey(fields, selectedField.value.key) !== selectedField.value) selectedField.value = fields[0]
}, { immediate: true, deep: false })

watch(selectedField, (field) => {
  syncOptionJsonText(field)
}, { immediate: true })

function fieldTypeMeta(type: WorkflowFormFieldType) {
  return fieldTypes.find(item => item.type === type) || fieldTypes[0]
}

function nextFieldKey() {
  let index = props.fields.length + 1
  while (workflowFieldByKey(props.fields, `field_${index}`)) index += 1
  return `field_${index}`
}

function buildField(type: WorkflowFormFieldType): WorkflowFormField {
  const meta = fieldTypeMeta(type)
  const key = nextFieldKey()
  if (type === 'group') return { key, label: '新建分组', type, span: 24, fields: [] }
  if (type === 'label') return { key, label: '标签文本', type, span: 24 }
  if (type === 'description') return { key, label: '填写提示', type, span: 24, content: '请输入说明内容' }
  if (type === 'button') return { key, label: '查看说明', type, span: 24, help: { title: '说明', content: '请输入说明内容' } }
  const field: WorkflowFormField = {
    key,
    label: meta.label,
    type,
    required: false,
    placeholder: '',
    span: defaultFieldSpan(type),
  }
  if (type === 'text' || type === 'textarea' || type === 'phone' || type === 'email') {
    field.maxLength = type === 'textarea' ? 2000 : type === 'email' ? 254 : type === 'phone' ? 20 : 200
  }
  if (type === 'textarea') {
    field.minVisibleRows = 3
    field.maxVisibleRows = 8
  }
  if (['select', 'multi_select', 'radio', 'checkbox'].includes(type)) {
    field.options = [
      { label: '选项一', value: 'option_1' },
      { label: '选项二', value: 'option_2' },
    ]
  }
  if (type === 'detail_list') {
    field.label = '我的目标'
    field.rowKey = 'id'
    field.minRows = 0
    field.maxRows = 20
    field.columns = [
      { key: 'target', label: '目标', type: 'textarea', required: true, maxLength: 200, minVisibleRows: 2, maxVisibleRows: 6, span: 24 },
      { key: 'weight', label: '权重', type: 'number', min: 0, max: 100, span: 12 },
      { key: 'result', label: '结果', type: 'textarea', maxLength: 500, minVisibleRows: 2, maxVisibleRows: 6, span: 24 },
    ]
  }
  return field
}

function defaultFieldSpan(type: WorkflowFormFieldType): WorkflowFormFieldSpan {
  return ['textarea', 'radio', 'checkbox', 'attachment', 'date_range', 'detail_list', 'group', 'label', 'description', 'button'].includes(type) ? 24 : 12
}

function fieldSpan(field: WorkflowFormField): WorkflowFormFieldSpan {
  return fieldSpanOptions.some(item => item.value === field.span) ? field.span as WorkflowFormFieldSpan : 24
}

function fieldSpanLabel(field: WorkflowFormField) {
  return fieldSpanOptions.find(item => item.value === fieldSpan(field))?.label || '整行'
}

function updateFieldSpan(value: string | number | boolean | undefined) {
  if (!selectedField.value) return
  const span = Number(value)
  if (!fieldSpanOptions.some(item => item.value === span)) return
  selectedField.value.span = span as WorkflowFormFieldSpan
  emitChange()
}

function addField(type: WorkflowFormFieldType) {
  if (props.readonly) return
  const field = buildField(type)
  props.fields.push(field)
  selectedField.value = field
  emitChange()
}

function selectField(field: WorkflowFormField) {
  selectedField.value = field
}

function groupFields(group: WorkflowFormField): WorkflowFormField[] {
  group.fields = Array.isArray(group.fields) ? group.fields : []
  return group.fields
}

function handlePaletteDragStart(type: WorkflowFormFieldType, event: DragEvent) {
  if (props.readonly) return
  dragState.value = { kind: 'palette', containerKey: null, index: -1, fieldKey: '', fieldType: type }
  dropTarget.value = null
  if (!event.dataTransfer) return
  event.dataTransfer.effectAllowed = 'copy'
  event.dataTransfer.setData('text/plain', type)
}

function isPaletteDragging(type: WorkflowFormFieldType) {
  return dragState.value?.kind === 'palette' && dragState.value.fieldType === type
}

function handleDragStart(field: WorkflowFormField, containerKey: string | null, index: number, event: DragEvent) {
  if (props.readonly) return
  dragState.value = { kind: 'field', containerKey, index, fieldKey: field.key, fieldType: field.type }
  dropTarget.value = null
  selectedField.value = field
  if (!event.dataTransfer) return
  event.dataTransfer.effectAllowed = 'move'
  event.dataTransfer.setData('text/plain', field.key)
  const card = (event.currentTarget as HTMLElement | null)?.closest('.field-item') as HTMLElement | null
  if (!card) return
  const rect = card.getBoundingClientRect()
  event.dataTransfer.setDragImage(card, Math.min(event.clientX - rect.left, rect.width), Math.min(event.clientY - rect.top, rect.height))
}

function handleDragEnter(containerKey: string | null, index: number) {
  const source = dragState.value
  if (!source) return
  if (containerKey !== null) {
    const group = workflowFieldByKey(props.fields, containerKey)
    if (!group || group.type !== 'group' || source.fieldType === 'group') {
      dropTarget.value = null
      return
    }
  }
  if (source.kind === 'field' && source.containerKey === containerKey && (index === source.index || index === source.index + 1)) {
    dropTarget.value = null
    return
  }
  dropTarget.value = { containerKey, index }
}

function handleDrop(containerKey: string | null, index: number, event: DragEvent) {
  event.preventDefault()
  const source = dragState.value
  if (!source) {
    handleDragEnd()
    return
  }
  const field = source.kind === 'palette'
    ? buildField(source.fieldType)
    : workflowFieldByKey(props.fields, source.fieldKey)
  const changed = field && (source.kind === 'palette'
    ? insertWorkflowField(props.fields, field, containerKey, index)
    : moveWorkflowField(props.fields, source.fieldKey, containerKey, index))
  if (changed && field) selectedField.value = field
  handleDragEnd()
  if (changed) emitChange()
}

function handleDragEnd() {
  dragState.value = null
  dropTarget.value = null
}

function isDropTarget(containerKey: string | null, index: number) {
  return dropTarget.value?.containerKey === containerKey && dropTarget.value.index === index
}

function canDropIntoGroup(group: WorkflowFormField) {
  if (!dragState.value || group.type !== 'group') return false
  return dragState.value.fieldType !== 'group'
}

async function removeField(field: WorkflowFormField) {
  if (props.readonly) return
  if (field.type === 'group' && groupFields(field).length > 0) {
    try {
      await ElMessageBox.confirm('删除分组将同时删除组内组件，确定继续？', '删除分组', { type: 'warning' })
    } catch {
      return
    }
  }
  const removesSelection = selectedField.value === field || (
    field.type === 'group' && Boolean(selectedField.value && workflowFieldByKey(groupFields(field), selectedField.value.key))
  )
  const removed = removeWorkflowField(props.fields, field.key)
  if (!removed) return
  if (removesSelection) selectedField.value = props.fields[0] || null
  emitChange()
}

function isOptionField(field: WorkflowFormField) {
  return isDropdownField(field) || field.type === 'radio' || field.type === 'checkbox'
}

function isDropdownField(field: WorkflowFormField) {
  return field.type === 'select' || field.type === 'multi_select'
}

function fieldOptions(field: WorkflowFormField): WorkflowFormOption[] {
  return normalizeWorkflowOptions(field.options || [], field.optionSource)
}

function flatFieldOptions(field: WorkflowFormField) {
  return flattenWorkflowOptions(fieldOptions(field))
}

function shouldUseTreeSelect(field: WorkflowFormField) {
  return isDropdownField(field) && (optionSourceType(field) === 'api' || hasWorkflowOptionChildren(fieldOptions(field)))
}

function optionSourceType(field: WorkflowFormField): WorkflowOptionSourceType {
  return field.optionSource?.type === 'api' ? 'api' : 'static'
}

function updateOptionSourceType(value: string | number | boolean | undefined) {
  if (!selectedField.value || !isOptionField(selectedField.value)) return
  if (value === 'api') {
    const previous = selectedField.value.optionSource
    selectedField.value.optionSource = {
      type: 'api',
      url: previous?.url || '/api/v2/admin/departments/tree',
      method: previous?.method === 'POST' ? 'POST' : 'GET',
      responsePath: previous?.responsePath || 'data',
      labelField: previous?.labelField || 'name',
      valueField: previous?.valueField || 'id',
      childrenField: previous?.childrenField || 'children',
    }
  } else {
    delete selectedField.value.optionSource
    ensureDefaultOptions(selectedField.value)
  }
  syncOptionJsonText(selectedField.value)
  emitChange()
}

function ensureDefaultOptions(field: WorkflowFormField) {
  if (Array.isArray(field.options) && field.options.length > 0) return
  field.options = [
    { label: '选项一', value: 'option_1' },
    { label: '选项二', value: 'option_2' },
  ]
}

function syncOptionJsonText(field: WorkflowFormField | null) {
  if (!field || !isOptionField(field)) {
    optionJsonText.value = ''
    return
  }
  optionJsonText.value = JSON.stringify(field.options || [], null, 2)
}

function applyOptionsJson() {
  if (props.readonly || !selectedField.value || !isOptionField(selectedField.value)) return
  try {
    const options = normalizeWorkflowOptions(JSON.parse(optionJsonText.value || '[]'))
    if (options.length === 0) {
      ElMessage.warning('JSON 选项不能为空')
      return
    }
    selectedField.value.options = options
    delete selectedField.value.optionSource
    optionJsonText.value = JSON.stringify(options, null, 2)
    emitChange()
    ElMessage.success('已应用 JSON 选项')
  } catch {
    ElMessage.error('JSON 格式无效')
  }
}

function formatOptionsJson() {
  if (!selectedField.value || !isOptionField(selectedField.value)) return
  try {
    const parsed = JSON.parse(optionJsonText.value || '[]')
    optionJsonText.value = JSON.stringify(parsed, null, 2)
  } catch {
    optionJsonText.value = JSON.stringify(selectedField.value.options || [], null, 2)
  }
}

function detailColumns(field: WorkflowFormField) {
  field.columns ||= []
  return field.columns
}

function nextDetailColumnKey(field: WorkflowFormField) {
  let index = detailColumns(field).length + 1
  while (detailColumns(field).some(item => item.key === `column_${index}`)) index += 1
  return `column_${index}`
}

function buildDetailColumn(field: WorkflowFormField): WorkflowFormField {
  return {
    key: nextDetailColumnKey(field),
    label: `明细列${detailColumns(field).length + 1}`,
    type: 'text',
    required: false,
    span: 12,
  }
}

function addDetailColumn() {
  if (props.readonly || !selectedField.value || selectedField.value.type !== 'detail_list') return
  selectedField.value.columns ||= []
  selectedField.value.columns.push(buildDetailColumn(selectedField.value))
  emitChange()
}

function removeDetailColumn(index: number) {
  if (props.readonly || !selectedField.value?.columns || selectedField.value.columns.length <= 1) return
  selectedField.value.columns.splice(index, 1)
  emitChange()
}

function handleDetailColumnDragStart(index: number, event: DragEvent) {
  if (props.readonly) return
  detailColumnDragIndex.value = index
  detailColumnDropIndex.value = null
  if (!event.dataTransfer) return
  event.dataTransfer.effectAllowed = 'move'
  event.dataTransfer.setData('text/plain', String(index))
  const row = (event.currentTarget as HTMLElement | null)?.closest('.detail-column-row') as HTMLElement | null
  if (!row) return
  const rect = row.getBoundingClientRect()
  event.dataTransfer.setDragImage(row, Math.min(24, rect.width), Math.min(18, rect.height))
}

function handleDetailColumnDragOver(index: number) {
  const sourceIndex = detailColumnDragIndex.value
  if (sourceIndex === null) return
  detailColumnDropIndex.value = index === sourceIndex || index === sourceIndex + 1
    ? null
    : index
}

function handleDetailColumnTailDragOver() {
  const sourceIndex = detailColumnDragIndex.value
  const field = selectedField.value
  if (sourceIndex === null || field?.type !== 'detail_list') return
  const targetIndex = detailColumns(field).length
  detailColumnDropIndex.value = sourceIndex === targetIndex - 1 ? null : targetIndex
}

function handleDetailColumnDrop() {
  const sourceIndex = detailColumnDragIndex.value
  const targetIndex = detailColumnDropIndex.value
  const field = selectedField.value
  const changed = sourceIndex !== null
    && targetIndex !== null
    && field?.type === 'detail_list'
    && moveWorkflowDetailColumn(detailColumns(field), sourceIndex, targetIndex)
  handleDetailColumnDragEnd()
  if (changed) emitChange()
}

function handleDetailColumnDragEnd() {
  detailColumnDragIndex.value = null
  detailColumnDropIndex.value = null
}

function moveDetailColumnByOffset(index: number, offset: -1 | 1) {
  const field = selectedField.value
  if (props.readonly || field?.type !== 'detail_list') return
  const columns = detailColumns(field)
  const targetIndex = offset < 0 ? index - 1 : index + 2
  if (moveWorkflowDetailColumn(columns, index, targetIndex)) emitChange()
}

function updateDetailColumnRange(column: WorkflowFormField, key: 'min' | 'max', value: number | undefined) {
  if (props.readonly) return
  if (typeof value === 'number' && Number.isFinite(value)) {
    column[key] = value
  } else {
    delete column[key]
  }
  emitChange()
}

function updateDetailColumnMaxLength(column: WorkflowFormField, value: number | undefined) {
  if (props.readonly) return
  if (typeof value === 'number' && Number.isFinite(value) && value > 0) {
    column.maxLength = value
  } else {
    delete column.maxLength
  }
  emitChange()
}

function updateTextareaVisibleRows(
  field: WorkflowFormField,
  key: 'minVisibleRows' | 'maxVisibleRows',
  value: number | undefined,
  defaultMinRows: number,
  defaultMaxRows: number,
) {
  if (props.readonly) return
  const current = textareaAutosize(field, defaultMinRows, defaultMaxRows)
  const nextValue = Math.min(30, Math.max(1, Math.floor(Number(value) || (key === 'minVisibleRows' ? current.minRows : current.maxRows))))
  if (key === 'minVisibleRows') {
    field.minVisibleRows = nextValue
    field.maxVisibleRows = Math.max(nextValue, current.maxRows)
  } else {
    field.maxVisibleRows = nextValue
    field.minVisibleRows = Math.min(current.minRows, nextValue)
  }
  emitChange()
}

function updateDetailColumnSpan(column: WorkflowFormField, value: string | number | boolean | undefined) {
  if (props.readonly) return
  const span = Number(value)
  if (!fieldSpanOptions.some(item => item.value === span)) return
  column.span = span as WorkflowFormFieldSpan
  emitChange()
}

function handleDetailColumnTypeChange(column: WorkflowFormField) {
  if (['select', 'multi_select', 'radio', 'checkbox'].includes(column.type)) {
    ensureDefaultOptions(column)
  } else {
    delete column.options
    delete column.optionSource
  }
  if (column.type !== 'number' && column.type !== 'amount') {
    delete column.min
    delete column.max
  }
  if (!['text', 'textarea', 'phone', 'email'].includes(column.type)) {
    delete column.maxLength
  }
  if (column.type === 'textarea') {
    column.minVisibleRows ??= 2
    column.maxVisibleRows ??= 6
  } else {
    delete column.minVisibleRows
    delete column.maxVisibleRows
  }
  emitChange()
}

function addOption() {
  if (!selectedField.value || !isOptionField(selectedField.value)) return
  selectedField.value.options ||= []
  const index = selectedField.value.options.length + 1
  selectedField.value.options.push({ label: `选项${index}`, value: `option_${index}` })
  syncOptionJsonText(selectedField.value)
  emitChange()
}

function removeOption(index: number) {
  if (!selectedField.value?.options || selectedField.value.options.length <= 1) return
  selectedField.value.options.splice(index, 1)
  syncOptionJsonText(selectedField.value)
  emitChange()
}

function componentNameLabel(field: WorkflowFormField) {
  if (field.type === 'group') return '分组标题'
  if (field.type === 'label') return '标签文字'
  if (field.type === 'button') return '按钮文字'
  return field.type === 'description' ? '说明名称' : '字段名称'
}

function fieldLabelMaxLength(field: WorkflowFormField) {
  if (field.type === 'button') return 30
  if (field.type === 'label') return 100
  return 60
}

function supportsFieldHelp(field: WorkflowFormField) {
  return field.type === 'group' || isWorkflowDataField(field)
}

function toggleFieldHelp(enabled: string | number | boolean) {
  if (!selectedField.value || props.readonly) return
  if (Boolean(enabled)) {
    selectedField.value.help ||= {
      buttonText: '查看说明',
      title: `${selectedField.value.label || '字段'}说明`,
      content: '请输入说明内容',
    }
  } else {
    delete selectedField.value.help
  }
  emitChange()
}

function updateHelpField(key: 'title' | 'content', value: string) {
  if (!selectedField.value || props.readonly) return
  selectedField.value.help ||= { title: '说明', content: '请输入说明内容' }
  selectedField.value.help[key] = value
  emitChange()
}

function supportsDefault(field: WorkflowFormField) {
  return isWorkflowDataField(field) && !['attachment', 'user', 'user_multi', 'department', 'department_multi', 'date_range', 'detail_list'].includes(field.type)
}

function stringDefault(value: unknown) {
  return typeof value === 'string' ? value : ''
}

function numberDefault(value: unknown) {
  return typeof value === 'number' ? value : undefined
}

function arrayDefault(value: unknown) {
  return Array.isArray(value) ? value.filter(item => typeof item === 'string') : []
}

function selectDefault(field: WorkflowFormField) {
  if (field.type === 'multi_select') return Array.isArray(field.default) ? field.default.filter(value => typeof value === 'string') : []
  return typeof field.default === 'string' ? field.default : undefined
}

function organizationPlaceholder(type: WorkflowFormFieldType) {
  if (type === 'user_multi') return '请选择人员，可多选'
  if (type === 'department_multi') return '请选择部门，可多选'
  return type === 'user' ? '请选择人员' : '请选择部门'
}

function detailColumnPlaceholder(column: WorkflowFormField) {
  if (column.type === 'number') return '0'
  if (column.type === 'amount') return '0.00'
  if (column.type === 'boolean') return '开关'
  if (isOptionField(column)) return '选项'
  return column.placeholder || '请输入'
}

function updateDefault(value: unknown) {
  if (!selectedField.value) return
  if (value === '' || value === undefined || value === null || (Array.isArray(value) && value.length === 0)) delete selectedField.value.default
  else selectedField.value.default = value
  emitChange()
}

function emitChange() {
  emit('change')
}
</script>

<style scoped>
.form-designer { display: grid; grid-template-columns: 320px minmax(480px, 1fr) 440px; width: 100%; min-width: 0; min-height: 0; height: 100%; overflow: hidden; background: #f2f5f8; }
.field-palette, .property-panel { display: flex; min-width: 0; min-height: 0; overflow: hidden; flex-direction: column; background: #fff; }
.field-palette { border-right: 1px solid #dfe6ee; }
.property-panel { border-left: 1px solid #dfe6ee; }
.panel-heading, .canvas-heading { display: flex; align-items: center; flex: 0 0 auto; justify-content: space-between; gap: 12px; min-height: 58px; padding: 0 16px; border-bottom: 1px solid #e7ebf0; background: #fff; }
.panel-heading > div { display: flex; align-items: baseline; justify-content: space-between; width: 100%; gap: 8px; }
.panel-heading strong, .canvas-heading strong { color: #1f2937; font-size: 14px; }
.panel-heading span, .canvas-heading span { color: #94a3b8; font-size: 11px; }
.canvas-heading > div { display: flex; align-items: baseline; gap: 8px; }
.palette-content { min-height: 0; overflow-y: auto; flex: 1; padding: 14px 12px 24px; }
.palette-group + .palette-group { margin-top: 20px; }
.palette-group h3 { margin: 0 0 9px 2px; color: #64748b; font-size: 11px; font-weight: 600; }
.palette-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 8px; }
.palette-grid button { position: relative; display: flex; align-items: center; gap: 7px; min-width: 0; height: 40px; padding: 0 8px; border: 1px solid #e2e8f0; border-radius: 6px; color: #475569; background: #fff; font: inherit; font-size: 12px; cursor: grab; transition: border-color .15s, background-color .15s, color .15s, opacity .15s; }
.palette-grid button:hover:not(:disabled) { border-color: #9fc2f7; color: #1677ff; background: #f6f9ff; }
.palette-grid button.dragging { opacity: .45; border-style: dashed; }
.palette-grid button:disabled { cursor: not-allowed; opacity: .55; }
.palette-type-icon { display: grid; flex: 0 0 auto; place-items: center; width: 24px; height: 24px; border-radius: 5px; color: #1677ff; background: #edf5ff; }
.palette-add-icon { position: absolute; right: 7px; opacity: 0; color: #1677ff; transition: opacity .15s; }
.palette-grid button:hover .palette-add-icon { opacity: 1; }
.form-canvas { display: flex; min-width: 0; min-height: 0; overflow: hidden; flex-direction: column; background: #f2f5f8; }
.canvas-stage { min-height: 0; overflow-y: auto; flex: 1; padding: 18px 22px 32px; }
.form-sheet { width: min(920px, 100%); min-height: calc(100% - 2px); margin: 0 auto; padding: 16px; border: 1px solid #dfe6ee; border-radius: 8px; background: #fff; box-shadow: 0 8px 24px rgb(15 23 42 / 4%); }
.empty-canvas-drop { display: grid; min-height: 360px; border: 1px dashed transparent; border-radius: 6px; transition: border-color .15s, background-color .15s; }
.empty-canvas-drop.active { border-color: #1677ff; background: #f2f7ff; }
.empty-canvas-drop > .el-empty { min-height: 360px; }
.field-list { display: grid; grid-template-columns: repeat(24, minmax(0, 1fr)); align-items: start; gap: 10px; }
.field-item { position: relative; display: block; min-height: 112px; padding: 14px; border: 1px solid #e3e8ef; border-radius: 7px; background: #fff; cursor: grab; transition: border-color .15s, box-shadow .15s, background-color .15s; }
.field-item:active { cursor: grabbing; }
.field-item:last-child { margin-bottom: 0; }
.field-item:hover { border-color: #b8cff1; background: #fcfdff; }
.field-item.active { border-color: #74a9f7; box-shadow: 0 0 0 2px rgb(22 119 255 / 7%); }
.field-item.active::before { position: absolute; top: 14px; bottom: 14px; left: -1px; width: 3px; border-radius: 0 2px 2px 0; background: #1677ff; content: ''; }
.field-item.dragging { opacity: .45; border-style: dashed; box-shadow: none; }
.field-item.drop-before::after { position: absolute; z-index: 2; top: -7px; right: 8px; left: 8px; height: 2px; border-radius: 2px; background: #1677ff; box-shadow: 0 0 0 3px rgb(22 119 255 / 9%); content: ''; }
.field-item__heading { display: flex; align-items: center; justify-content: space-between; min-width: 0; gap: 14px; }
.field-item__main { display: flex; align-items: center; min-width: 0; gap: 11px; }
.field-item__main > div { min-width: 0; }
.field-drag-handle { display: grid; flex: 0 0 auto; place-items: center; width: 24px; height: 34px; padding: 0; border: 0; border-radius: 4px; color: #a7b2c1; background: transparent; cursor: grab; }
.field-drag-handle:hover { color: #1677ff; background: #edf5ff; }
.field-drag-handle:active { cursor: grabbing; }
.field-type-icon { display: grid; flex: 0 0 auto; place-items: center; width: 34px; height: 34px; border-radius: 6px; color: #1677ff; background: #eaf3ff; }
.field-item strong { color: #273548; font-size: 13px; }
.field-item strong i { margin-left: 3px; color: #ef4444; font-style: normal; }
.field-item p { overflow: hidden; margin: 4px 0 0; color: #94a3b8; font-size: 11px; text-overflow: ellipsis; white-space: nowrap; }
.field-actions { display: flex; flex: 0 0 auto; gap: 5px; opacity: 0; transition: opacity .15s; }
.field-item:hover .field-actions, .field-item.active .field-actions, .field-actions:focus-within { opacity: 1; }
.field-item--compact .field-item__main { padding-right: 24px; gap: 6px; }
.field-item--compact .field-type-icon { width: 30px; height: 30px; }
.field-item--compact .field-actions { position: absolute; z-index: 3; top: 8px; right: 8px; padding: 3px; border: 1px solid #e5eaf0; border-radius: 6px; background: #fff; box-shadow: 0 5px 14px rgb(15 23 42 / 10%); }
.field-item--compact .field-preview { margin-left: 0; }
.field-item--group { grid-column: 1 / -1 !important; padding: 0 0 14px; border: 0; border-top: 1px solid #dfe6ee; border-radius: 0; background: transparent; }
.field-item--group:hover { border-color: #b8cff1; background: #f9fbfe; }
.field-item--group.active { border-color: #74a9f7; box-shadow: none; }
.field-item--group.active::before { top: 0; bottom: 0; }
.field-item--group > .field-item__content > .field-item__heading { padding: 14px 14px 4px; }
.group-fields { margin-top: 8px; padding: 0 14px; }
.group-fields__grid { display: grid; grid-template-columns: repeat(24, minmax(0, 1fr)); align-items: start; gap: 10px; }
.field-item--nested { min-height: 104px; }
.group-drop-zone { margin-top: 2px; }
.group-drop-zone--empty { min-height: 72px; border-color: #cbd5e1; background: #fff; }
.field-preview { max-width: 720px; margin: 12px 0 0 45px; pointer-events: none; }
.field-preview :deep(.el-select), .field-preview :deep(.el-input-number), .field-preview :deep(.el-date-editor) { width: 100%; }
.field-preview :deep(.el-radio-group), .field-preview :deep(.el-checkbox-group) { display: flex; flex-wrap: wrap; gap: 8px 18px; }
.field-preview :deep(.el-radio), .field-preview :deep(.el-checkbox) { margin-right: 0; }
.field-preview :deep(.is-disabled .el-input__wrapper), .field-preview :deep(.el-textarea.is-disabled .el-textarea__inner) { background: #f8fafc; box-shadow: 0 0 0 1px #e6ebf1 inset; }
.detail-preview {
  display: flex;
  flex-direction: column;
  gap: 8px;
  width: 100%;
  padding: 10px;
  border: 1px solid #e6ebf1;
  border-radius: 6px;
  background: #f8fafc;
}
.detail-preview__grid {
  display: grid;
  grid-template-columns: repeat(24, minmax(0, 1fr));
  gap: 8px;
}
.detail-preview__cell {
  display: flex;
  min-width: 0;
  flex-direction: column;
  gap: 6px;
}
.detail-preview__label {
  min-width: 0;
  overflow: hidden;
  color: #64748b;
  font-size: 11px;
  font-weight: 650;
  line-height: 1.4;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.detail-preview__control {
  display: block;
  min-height: 28px;
  padding: 6px 8px;
  overflow: hidden;
  border: 1px solid #e5eaf0;
  border-radius: 5px;
  color: #64748b;
  background: #fff;
  font-size: 11px;
  line-height: 1.4;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.field-drop-zone--tail { display: flex; grid-column: 1 / -1; align-items: center; justify-content: center; gap: 6px; min-height: 18px; border: 1px dashed transparent; border-radius: 6px; color: #94a3b8; background: transparent; font-size: 11px; transition: min-height .15s, border-color .15s, background-color .15s; }
.field-drop-zone--tail.is-dragging { min-height: 42px; border-color: #cbd5e1; background: #f8fafc; }
.field-drop-zone--tail.active { border-color: #1677ff; color: #1677ff; background: #f2f7ff; }
.property-panel > .panel-heading { position: sticky; top: 0; z-index: 2; }
.property-form { min-height: 0; overflow-y: auto; flex: 1; }
.property-form :deep(.el-form-item) { margin-bottom: 14px; }
.property-form :deep(.el-form-item__label) { margin-bottom: 5px; color: #475569; font-size: 12px; line-height: 20px; }
.property-form :deep(.el-input-number) { width: 100%; }
.property-section { padding: 16px; border-bottom: 1px solid #edf0f4; }
.property-section:last-child { border-bottom: 0; }
.property-section > h3, .option-editor__heading h3 { margin: 0 0 14px; color: #273548; font-size: 12px; font-weight: 650; }
.layout-setting :deep(.el-radio-group) { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); width: 100%; }
.layout-setting :deep(.el-radio-button), .layout-setting :deep(.el-radio-button__inner) { width: 100%; }
.layout-setting :deep(.el-radio-button__inner) { padding-right: 8px; padding-left: 8px; }
.layout-setting > p { margin: -3px 0 0; color: #94a3b8; font-size: 11px; line-height: 1.6; }
.required-setting { display: flex; align-items: center; justify-content: space-between; min-height: 34px; margin-bottom: 12px; color: #475569; font-size: 12px; }
.number-range { display: grid; grid-template-columns: 1fr 1fr; gap: 8px; }
.textarea-visible-rows { grid-column: 1 / -1; grid-template-columns: repeat(2, minmax(0, 1fr)); min-width: 0; }
.textarea-visible-rows :deep(.el-form-item) { min-width: 0; }
.option-editor__heading { display: flex; align-items: center; justify-content: space-between; min-height: 32px; }
.option-editor__heading h3 { margin-bottom: 0; }
.option-row { display: grid; grid-template-columns: 22px minmax(0, 1fr) 28px; align-items: center; gap: 7px; margin-top: 10px; }
.option-index { display: grid; place-items: center; width: 22px; height: 22px; border-radius: 50%; color: #64748b; background: #f1f5f9; font-size: 10px; }
.option-row__inputs { display: grid; grid-template-columns: 1fr 1fr; gap: 6px; }
.option-editor :deep(.el-radio-group) { width: 100%; }
.option-editor :deep(.el-radio-button) { flex: 1; }
.option-editor :deep(.el-radio-button__inner) { width: 100%; padding-right: 10px; padding-left: 10px; }
.option-json-input { margin-top: 14px; }
.option-json-input :deep(.el-textarea__inner) {
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 12px;
  line-height: 1.6;
}
.option-json-actions { display: flex; justify-content: flex-end; gap: 8px; margin-top: -4px; }
.option-api-config { display: flex; flex-direction: column; gap: 2px; }
.option-api-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 0 8px; }
.detail-editor .option-editor__heading { margin-top: 4px; }
.detail-column-row { position: relative; display: grid; grid-template-columns: 28px minmax(0, 1fr) 28px; align-items: start; gap: 7px; margin-top: 10px; padding: 8px 0; }
.detail-column-row.dragging { opacity: .45; }
.detail-column-row.drop-before::before { position: absolute; right: 0; left: 0; height: 2px; border-radius: 2px; background: #1677ff; box-shadow: 0 0 0 3px rgb(22 119 255 / 9%); content: ''; }
.detail-column-row.drop-before::before { top: -5px; }
.detail-column-drag-handle { display: flex; align-items: center; justify-content: center; flex-direction: column; gap: 1px; width: 28px; height: 40px; padding: 0; border: 1px solid #dfe6ee; border-radius: 4px; color: #64748b; background: #fff; cursor: grab; font-size: 10px; }
.detail-column-drag-handle:hover:not(:disabled) { color: #1677ff; background: #edf5ff; }
.detail-column-drag-handle:active { cursor: grabbing; }
.detail-column-drag-handle:disabled { cursor: not-allowed; opacity: .5; }
.detail-column-actions { display: flex; flex-direction: column; gap: 2px; }
.detail-column-actions :deep(.el-button) { margin-left: 0; }
.detail-column-tail-drop { display: flex; align-items: center; justify-content: center; gap: 6px; min-height: 38px; margin-top: 8px; border: 1px dashed #b8c4d4; border-radius: 6px; color: #64748b; background: #f8fafc; font-size: 11px; }
.detail-column-tail-drop.active { border-color: #1677ff; color: #1677ff; background: #edf5ff; }
.detail-column-row__inputs { display: grid; grid-template-columns: 1fr 1fr; gap: 6px; }
.detail-column-validation { grid-column: 1 / -1; }
.detail-column-range { display: grid; grid-column: 1 / -1; grid-template-columns: 1fr 1fr; gap: 8px; }
.detail-column-range :deep(.el-form-item) { margin-bottom: 0; }
.detail-column-max-length { grid-column: 1 / -1; }
.detail-column-layout { display: grid; grid-column: 1 / -1; grid-template-columns: 42px minmax(0, 1fr); align-items: center; gap: 8px; min-height: 32px; color: #475569; font-size: 12px; }
.detail-column-layout :deep(.el-radio-group) { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); width: 100%; }
.detail-column-layout :deep(.el-radio-button),
.detail-column-layout :deep(.el-radio-button__inner) { width: 100%; }
.detail-column-layout :deep(.el-radio-button__inner) { padding-right: 6px; padding-left: 6px; }
.detail-column-required {
  display: flex;
  align-items: center;
  justify-content: space-between;
  min-height: 32px;
  padding: 0 10px;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  color: #475569;
  font-size: 12px;
}
@media (max-width: 1380px) {
  .form-designer { grid-template-columns: 200px minmax(420px, 1fr) 380px; }
  .canvas-stage { padding-right: 16px; padding-left: 16px; }
}
@media (max-width: 1120px) {
  .form-designer { grid-template-columns: 176px minmax(360px, 1fr) 320px; }
  .palette-grid { grid-template-columns: 1fr; }
  .option-row__inputs, .detail-column-row__inputs { grid-template-columns: 1fr; }
  .field-preview { margin-left: 0; }
}
@media (max-width: 760px) {
  .field-item, .field-item--nested { grid-column: 1 / -1 !important; }
}
</style>
