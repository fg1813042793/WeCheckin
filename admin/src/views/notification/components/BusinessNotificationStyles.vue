<script setup lang="ts">
import { Check, InfoFilled, Picture, Microphone, Document, Promotion, ChatDotRound } from '@element-plus/icons-vue'
import WorkflowUserTreePicker from '../../workflow/components/WorkflowUserTreePicker.vue'
import { useNotificationStyles } from '../useNotificationStyles'
const props = defineProps<{ canEdit: boolean; canSend: boolean; canSendDingTalk: boolean }>()
const emit = defineEmits<{ sent: [channel: 'in_app' | 'dingtalk'] }>()
const { loading, loadError, loadEditor, saving, testSending, styles, selectedType, previewChannel, recipientOptionsLoading, recipientOptions, testForm, selectedStyle, selectedIconComponent, dingTalkPreview, dingTalkHeadColor, iconOptions, toneOptions, dingTalkMessageTypeOptions, previewChannelOptions, selectType, saveStyles, sendTest, handleDingTalkTypeChange, dingTalkFieldVisible, typeName, toneName, dingTalkMessageTypeName, tagType } = useNotificationStyles(props, channel => emit('sent', channel))
</script>
<template>
  <div v-if="loadError" role="alert">
    <el-alert :title="loadError" type="error" :closable="false" />
    <el-button :loading="loading" @click="loadEditor">重试</el-button>
  </div>
    <div v-loading="loading" class="style-editor">
      <nav class="style-types" aria-label="消息类型">
        <button
          v-for="style in styles"
          :key="style.type"
          type="button"
          :class="['style-type', { 'is-active': style.type === selectedType }]"
          @click="selectType(style.type)"
        >
          <span :class="['tone-dot', `tone-dot--${style.tone}`]" />
          <span class="style-type__text">
            <strong>{{ typeName(style.type) }}</strong>
            <small>{{ style.type }}</small>
          </span>
        </button>
      </nav>

      <div v-if="selectedStyle" class="style-workspace">
        <section class="style-form">
          <div class="section-heading">
            <div>
              <h3>{{ typeName(selectedStyle.type) }}</h3>
              <code>{{ selectedStyle.type }}</code>
            </div>
            <el-tag :type="tagType(selectedStyle.tone)" effect="light">{{ toneName(selectedStyle.tone) }}</el-tag>
          </div>

          <el-form label-position="top" @submit.prevent>
            <div class="form-grid">
              <el-form-item label="类型标签" required>
                <el-input v-model="selectedStyle.label" maxlength="20" show-word-limit :disabled="!canEdit" />
              </el-form-item>
              <el-form-item label="图标" required>
                <el-select v-model="selectedStyle.icon" :disabled="!canEdit" class="full-width">
                  <el-option v-for="option in iconOptions" :key="option.value" :label="option.label" :value="option.value">
                    <div class="icon-option">
                      <el-icon><component :is="option.component" /></el-icon>
                      <span>{{ option.label }}</span>
                    </div>
                  </el-option>
                </el-select>
              </el-form-item>
            </div>
            <el-form-item label="语义色" required>
              <el-radio-group v-model="selectedStyle.tone" :disabled="!canEdit" class="tone-selector">
                <el-radio-button v-for="option in toneOptions" :key="option.value" :value="option.value">
                  <span :class="['tone-dot', `tone-dot--${option.value}`]" />{{ option.label }}
                </el-radio-button>
              </el-radio-group>
            </el-form-item>

            <div class="template-heading">
              <h3>钉钉消息模板</h3>
              <el-popover placement="bottom-start" :width="460" trigger="hover">
                <div class="template-help">
                  <p>业务通知先生成标题和正文，消息样式再使用以下变量组装钉钉消息。</p>
                  <dl>
                    <div><dt><code v-text="'{{title}}'" /></dt><dd>业务通知生成的标题；工作流中是节点模板替换后的标题。</dd></div>
                    <div><dt><code v-text="'{{content}}'" /></dt><dd>业务通知生成的正文；工作流中是节点模板替换后的正文。</dd></div>
                    <div><dt><code v-text="'{{url}}'" /></dt><dd>当前通知的跳转地址；工作流消息会指向对应实例并在钉钉工作台内打开。</dd></div>
                    <div><dt><code v-text="'{{sourceName}}'" /></dt><dd>通知来源名称，例如 WeCheckin 流程。</dd></div>
                    <div><dt><code v-text="'{{picUrl}}'" /></dt><dd>工作流取已发布流程定义的 Logo；绩效通知取 DINGTALK_H5_LOGO_URL。</dd></div>
                    <div><dt><code v-text="'{{mediaId}}'" /></dt><dd>图片、语音或文件消息使用的钉钉媒体 ID。</dd></div>
                    <div><dt><code v-text="'{{duration}}'" /></dt><dd>语音消息时长，单位为秒。</dd></div>
                  </dl>
                </div>
                <template #reference>
                  <el-button link :icon="InfoFilled" aria-label="查看钉钉模板变量说明" />
                </template>
              </el-popover>
            </div>
            <el-form-item label="消息类型" required>
              <el-select
                v-model="selectedStyle.dingTalk.messageType"
                class="full-width"
                :disabled="!canEdit"
                @change="handleDingTalkTypeChange"
              >
                <el-option v-for="option in dingTalkMessageTypeOptions" :key="option.value" :label="option.label" :value="option.value" />
              </el-select>
            </el-form-item>
            <div class="form-grid">
              <el-form-item v-if="dingTalkFieldVisible('title')" label="标题模板" required>
                <el-input v-model="selectedStyle.dingTalk.title" :disabled="!canEdit" maxlength="5000" />
              </el-form-item>
              <el-form-item v-if="dingTalkFieldVisible('url')" label="跳转地址模板" required>
                <el-input v-model="selectedStyle.dingTalk.url" :disabled="!canEdit" maxlength="5000" />
                <div class="field-help field-help--url">
                  <p>推荐保留 <code v-text="'{{url}}'" />，系统会按企业配置自动生成钉钉应用内跳转地址。</p>
                  <p>手工配置时使用以下格式，不是将协议直接拼在 HTTPS 地址前：</p>
                  <code class="field-help__protocol">dingtalk://dingtalkclient/action/openapp?corpid=&lt;CorpID&gt;&amp;container_type=work_platform&amp;app_id=&lt;UnifiedAppID&gt;&amp;redirect_type=jump&amp;redirect_url=&lt;URL 编码后的 HTTPS 地址&gt;</code>
                  <p><code>app_id</code> 优先填 UnifiedAppID；只有 AgentID 时填 <code>0_&lt;AgentID&gt;</code>，<code>redirect_url</code> 必须先进行 URL 编码。</p>
                </div>
              </el-form-item>
              <el-form-item v-if="dingTalkFieldVisible('sourceName')" label="来源名称模板">
                <el-input v-model="selectedStyle.dingTalk.sourceName" :disabled="!canEdit" maxlength="5000" />
              </el-form-item>
              <el-form-item v-if="dingTalkFieldVisible('picUrl')" label="图片地址模板">
                <el-input v-model="selectedStyle.dingTalk.picUrl" :disabled="!canEdit" maxlength="5000" />
                <div class="field-help">
                  可填写固定的 HTTPS 图片地址；使用 <code v-text="'{{picUrl}}'" /> 时按通知来源取图，值为空则不展示图片。
                </div>
              </el-form-item>
              <el-form-item v-if="dingTalkFieldVisible('mediaId')" label="钉钉 mediaId" required>
                <el-input v-model="selectedStyle.dingTalk.mediaId" :disabled="!canEdit" maxlength="5000" />
                <div class="field-help">请填写通过钉钉媒体上传接口获得的 mediaId。</div>
              </el-form-item>
              <el-form-item v-if="dingTalkFieldVisible('duration')" label="语音时长（秒）" required>
                <el-input-number v-model="selectedStyle.dingTalk.duration" :disabled="!canEdit" :min="1" :max="3600" controls-position="right" />
              </el-form-item>
              <el-form-item v-if="dingTalkFieldVisible('buttonTitle')" label="按钮文字" required>
                <el-input v-model="selectedStyle.dingTalk.buttonTitle" :disabled="!canEdit" maxlength="20" />
                <div class="field-help">按钮颜色由钉钉客户端固定，不支持自定义。</div>
              </el-form-item>
              <el-form-item v-if="dingTalkFieldVisible('headColor')" label="OA 顶部颜色">
                <el-color-picker v-model="dingTalkHeadColor" :disabled="!canEdit" />
              </el-form-item>
            </div>
            <el-form-item v-if="dingTalkFieldVisible('content')" label="正文模板" required>
              <el-input
                v-model="selectedStyle.dingTalk.content"
                type="textarea"
                :rows="4"
                :disabled="!canEdit"
                maxlength="5000"
                show-word-limit
                resize="vertical"
              />
            </el-form-item>
          </el-form>

          <div class="preview-heading">
            <h3>实时预览</h3>
            <el-segmented v-model="previewChannel" :options="previewChannelOptions" size="small" />
          </div>
          <div v-if="previewChannel === 'in_app'" :class="['message-preview', `message-preview--${selectedStyle.tone}`]">
            <div :class="['message-preview__icon', `message-preview__icon--${selectedStyle.tone}`]">
              <el-icon><component :is="selectedIconComponent" /></el-icon>
            </div>
            <div class="message-preview__body">
              <div class="message-preview__meta">
                <el-tag :type="tagType(selectedStyle.tone)" size="small" effect="light">{{ selectedStyle.label }}</el-tag>
                <span>刚刚</span>
              </div>
              <strong>{{ testForm.title || '测试通知标题' }}</strong>
              <p>{{ testForm.content || '测试通知正文' }}</p>
            </div>
          </div>
          <div v-else class="dingtalk-preview">
            <div class="dingtalk-preview__brand">
              <span>钉钉工作通知</span>
              <el-tag size="small" effect="plain">{{ dingTalkMessageTypeName(dingTalkPreview.messageType) }}</el-tag>
            </div>
            <template v-if="['image', 'voice', 'file'].includes(dingTalkPreview.messageType)">
              <div class="media-preview">
                <el-icon><component :is="dingTalkPreview.messageType === 'image' ? Picture : dingTalkPreview.messageType === 'voice' ? Microphone : Document" /></el-icon>
                <span>{{ dingTalkPreview.mediaId || 'mediaId' }}</span>
                <small v-if="dingTalkPreview.messageType === 'voice'">{{ dingTalkPreview.duration }} 秒</small>
              </div>
            </template>
            <template v-else>
              <strong>{{ dingTalkPreview.title || testForm.title || '测试通知标题' }}</strong>
              <p>{{ dingTalkPreview.content || testForm.content || '测试通知正文' }}</p>
              <el-button v-if="['link', 'oa', 'action_card'].includes(dingTalkPreview.messageType)" size="small" plain>
                {{ dingTalkPreview.buttonTitle || '查看流程' }}
              </el-button>
            </template>
          </div>
        </section>

        <section v-if="canSend || canSendDingTalk" class="test-section">
          <h3>发送测试</h3>
          <el-form label-position="top" @submit.prevent>
            <el-form-item label="测试标题" required>
              <el-input v-model="testForm.title" maxlength="255" show-word-limit />
            </el-form-item>
            <el-form-item label="测试正文" required>
              <el-input v-model="testForm.content" type="textarea" :rows="3" maxlength="5000" show-word-limit resize="vertical" />
            </el-form-item>
            <el-form-item label="测试收件人" required>
              <WorkflowUserTreePicker
                v-model="testForm.userIds"
                :departments="recipientOptions.departments"
                :users="recipientOptions.users"
                :multiple="true"
                :disabled="recipientOptionsLoading"
                placeholder="请选择测试收件人"
              />
            </el-form-item>
          </el-form>
          <div class="test-actions">
            <el-button
              v-if="canSend"
              type="primary"
              :icon="Promotion"
              :loading="testSending === 'in_app'"
              :disabled="Boolean(testSending)"
              @click="sendTest('in_app')"
            >发送站内信测试</el-button>
            <el-button
              v-if="canSendDingTalk"
              type="primary"
              plain
              :icon="ChatDotRound"
              :loading="testSending === 'dingtalk'"
              :disabled="Boolean(testSending)"
              @click="sendTest('dingtalk')"
            >发送钉钉测试</el-button>
          </div>
        </section>
      </div>
    </div>

    <div class="style-actions">
      <el-button v-if="canEdit" type="primary" :icon="Check" :loading="saving" :disabled="loading || Boolean(loadError)" @click="saveStyles()">保存样式</el-button>
    </div>
</template>
<style scoped>
.style-actions { display: flex; justify-content: flex-end; margin-top: var(--admin-ui-space-4); }

.style-editor { display: grid; grid-template-columns: 248px minmax(0, 1fr); min-height: 620px; border: 1px solid var(--admin-border); }
.style-types { overflow-y: auto; border-right: 1px solid var(--admin-border); background: #f7f9fc; }
.style-type { display: flex; width: 100%; min-height: 54px; align-items: center; gap: 10px; padding: 8px 12px; border: 0; border-bottom: 1px solid var(--admin-border); background: transparent; color: var(--admin-text-secondary); text-align: left; cursor: pointer; }
.style-type:hover { background: #eef5ff; }
.style-type.is-active { background: #eaf3ff; color: #2563eb; box-shadow: inset 3px 0 #2563eb; }
.style-type__text { display: flex; min-width: 0; flex-direction: column; gap: 2px; }
.style-type__text strong { font-size: 13px; font-weight: 600; }
.style-type__text small { overflow: hidden; color: var(--admin-muted); font-size: 11px; text-overflow: ellipsis; white-space: nowrap; }
.style-workspace { min-width: 0; overflow-y: auto; }
.style-form, .test-section { padding: 20px 22px; }
.test-section { border-top: 8px solid var(--admin-bg); }
.section-heading, .preview-heading { display: flex; align-items: center; justify-content: space-between; gap: 16px; margin-bottom: 18px; }
.section-heading h3, .preview-heading h3, .test-section h3 { margin: 0; color: var(--admin-text); font-size: 16px; }
.section-heading code { display: block; margin-top: 3px; color: var(--admin-muted); font-size: 11px; }
.form-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 14px; }
.full-width { width: 100%; }
.icon-option { display: flex; align-items: center; gap: 8px; }
.tone-selector { display: flex; flex-wrap: wrap; }
.tone-selector :deep(.el-radio-button__inner) { display: inline-flex; align-items: center; gap: 6px; }
.tone-dot { width: 8px; height: 8px; flex: 0 0 8px; border-radius: 50%; }
.tone-dot--primary { background: #2563eb; }
.tone-dot--success { background: #00875a; }
.tone-dot--warning { background: #b45309; }
.tone-dot--danger { background: #c93756; }
.tone-dot--info { background: #64748b; }
.template-heading { display: flex; align-items: center; gap: 8px; margin: 8px 0 16px; padding-top: 18px; border-top: 1px solid var(--admin-border); }
.template-heading h3 { margin: 0; color: var(--admin-text); font-size: 15px; }
.template-heading .el-button { color: var(--admin-muted); }
.template-help { color: var(--admin-text-secondary); font-size: 12px; line-height: 18px; }
.template-help p { margin: 0 0 10px; }
.template-help dl { display: grid; gap: 7px; margin: 0; }
.template-help dl > div { display: grid; grid-template-columns: 92px minmax(0, 1fr); gap: 8px; }
.template-help dt, .template-help dd { margin: 0; }
.template-help code, .field-help code { color: #2563eb; }
.field-help { margin-top: 5px; color: var(--admin-muted); font-size: 12px; line-height: 18px; }
.field-help--url p { margin: 0 0 5px; }
.field-help--url p:last-child { margin: 5px 0 0; }
.field-help__protocol { display: block; overflow-wrap: anywhere; word-break: break-word; }
.preview-heading { margin-top: 14px; }
.message-preview { display: flex; min-height: 116px; gap: 14px; padding: 16px; border: 1px solid var(--admin-border); border-left-width: 3px; border-radius: 6px; background: #fff; }
.message-preview--primary { border-left-color: #2563eb; }
.message-preview--success { border-left-color: #00875a; }
.message-preview--warning { border-left-color: #b45309; }
.message-preview--danger { border-left-color: #c93756; }
.message-preview--info { border-left-color: #64748b; }
.message-preview__icon { display: flex; width: 34px; height: 34px; flex: 0 0 34px; align-items: center; justify-content: center; border-radius: 6px; font-size: 18px; }
.message-preview__icon--primary { background: #eaf2ff; color: #2563eb; }
.message-preview__icon--success { background: #e7f7f1; color: #00875a; }
.message-preview__icon--warning { background: #fff4df; color: #b45309; }
.message-preview__icon--danger { background: #fff0f3; color: #c93756; }
.message-preview__icon--info { background: #eef2f6; color: #64748b; }
.message-preview__body { min-width: 0; flex: 1; }
.message-preview__meta { display: flex; align-items: center; justify-content: space-between; gap: 10px; margin-bottom: 8px; color: var(--admin-muted); font-size: 12px; }
.message-preview__body > strong { display: block; overflow: hidden; color: var(--admin-text); font-size: 14px; text-overflow: ellipsis; white-space: nowrap; }
.message-preview__body > p, .dingtalk-preview p { display: -webkit-box; overflow: hidden; margin: 7px 0 0; color: var(--admin-text-secondary); line-height: 20px; -webkit-box-orient: vertical; -webkit-line-clamp: 2; }
.dingtalk-preview { min-height: 116px; padding: 16px; border: 1px solid var(--admin-border); border-radius: 6px; background: #fff; }
.dingtalk-preview__brand { display: flex; align-items: center; justify-content: space-between; gap: 12px; margin-bottom: 12px; color: #1476ff; font-size: 12px; font-weight: 600; }
.dingtalk-preview > strong { color: var(--admin-text); font-size: 14px; }
.dingtalk-preview .el-button { margin-top: 12px; }
.media-preview { display: flex; min-height: 64px; align-items: center; gap: 10px; color: var(--admin-text-secondary); }
.media-preview .el-icon { font-size: 28px; color: #1476ff; }
.media-preview span { min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.media-preview small { margin-left: auto; color: var(--admin-muted); }
.test-section h3 { margin-bottom: 16px; }
.test-actions { display: flex; flex-wrap: wrap; justify-content: flex-end; gap: 8px; }
@media (max-width: 760px) {
  .style-editor { grid-template-columns: 1fr; }
  .style-types { display: flex; overflow-x: auto; border-right: 0; border-bottom: 1px solid var(--admin-border); }
  .style-type { width: 170px; flex: 0 0 170px; border-right: 1px solid var(--admin-border); border-bottom: 0; }
  .style-workspace { max-height: 64vh; }
  .form-grid { grid-template-columns: 1fr; gap: 0; }
}
</style>
