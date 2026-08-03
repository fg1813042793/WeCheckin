<template>
  <div class="setup-page">
    <el-card shadow="never">
      <el-tabs v-model="activeTab">

        <el-tab-pane label="静态域名" name="STATIC_DOMAIN">
          <el-form label-width="120px" style="max-width: 500px;">
            <el-form-item label="静态资源域名">
              <el-input v-model="staticDomain" placeholder="https://cdn.example.com" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="saveStaticDomain" :loading="savingDomain">保存</el-button>
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <el-tab-pane label="用户表单配置" name="SETUP_USER_FORM_FIELDS">
          <el-button type="primary" size="small" @click="addField" style="margin-bottom: 12px;">新增字段</el-button>
          <el-table :data="formFields" border stripe size="small" style="margin-bottom: 16px;">
            <el-table-column label="排序" width="100">
              <template #default="{ row }">
                <el-input-number v-model="row.sort" :min="0" size="small" controls-position="right" style="width: 80px;" />
              </template>
            </el-table-column>
            <el-table-column label="字段名称" min-width="120">
              <template #default="{ row }">
                <el-input v-model="row.label" size="small" placeholder="字段名称" />
              </template>
            </el-table-column>
            <el-table-column label="字段类型" width="120">
              <template #default="{ row }">
                <el-select v-model="row.type" size="small" style="width: 100%;">
                  <el-option label="文本" value="文本" />
                  <el-option label="数字" value="数字" />
                  <el-option label="多行文本" value="多行文本" />
                  <el-option label="选择" value="选择" />
                  <el-option label="日期" value="日期" />
                  <el-option label="时间" value="时间" />
                  <el-option label="日期时间" value="日期时间" />
                  <el-option label="图片" value="图片" />
                  <el-option label="定位" value="定位" />
                </el-select>
              </template>
            </el-table-column>
            <el-table-column label="必填" width="60">
              <template #default="{ row }">
                <el-checkbox v-model="row.required" :true-value="1" :false-value="0" />
              </template>
            </el-table-column>
            <el-table-column label="选项(逗号分隔)" min-width="150">
              <template #default="{ row }">
                <el-input v-if="row.type === '选择'" v-model="row.options" size="small" placeholder="选择类型时填写" />
              </template>
            </el-table-column>
            <el-table-column label="操作" width="80">
              <template #default="{ $index }">
                <el-button type="danger" size="small" link @click="delField($index)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
          <el-button type="primary" @click="saveFormFields" :loading="savingForm">保存表单配置</el-button>
        </el-tab-pane>

        <el-tab-pane label="登录配置" name="LOGIN_CONFIG">
          <el-form label-width="180px" style="max-width: 500px;">
            <el-divider content-position="left">管理员</el-divider>
            <el-form-item label="Token 过期时间">
              <el-input v-model="tokenConfig.adminExpire" placeholder="24h">
                <template #append>例: 24h / 12h</template>
              </el-input>
            </el-form-item>
            <el-form-item label="Redis Key 前缀">
              <el-input v-model="tokenConfig.adminPrefix" placeholder="admin_token:" />
            </el-form-item>
            <el-form-item label="单点登录">
              <el-switch
                v-model="tokenConfig.adminSingleLogin"
                :active-value="1"
                :inactive-value="0"
                active-text="开启（同一账号仅允许一处登录）"
                inactive-text="关闭（允许多设备同时在线）"
                inline-prompt
                style="--el-switch-on-color: #f56c6c; --el-switch-off-color: #67c23a"
              />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="saveTokenConfig" :loading="savingToken">保存</el-button>
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <el-tab-pane label="问卷通知模版" name="SURVEY_TEMPLATE_PRESETS">
          <el-tabs v-model="presetTab">
            <el-tab-pane label="内置模版" name="builtin">
              <div style="margin-bottom:12px;color:#666;font-size:13px">
                编辑问卷交卷后的通知消息模版，模版参数说明：<code>{title}</code> 问卷标题、<code>{questionCount}</code> 题目数、<code>{total}</code> 总答卷数、<code>{submitter}</code> 提交人、<code>{date}</code> 提交时间、<code>{result}</code> 统计结果
              </div>
              <el-button type="primary" size="small" @click="openAdd('builtin')" style="margin-bottom:12px">新增模版</el-button>
              <el-table :data="builtinPresets" border stripe size="small" style="margin-bottom:16px" v-if="builtinPresets.length">
                <el-table-column label="名称" width="180" prop="label" />
                <el-table-column label="模版内容" min-width="400">
                  <template #default="{ row }">
                    <div style="white-space:pre-wrap;font-size:12px;line-height:1.6;max-height:200px;overflow-y:auto">{{ row.value }}</div>
                  </template>
                </el-table-column>
                <el-table-column label="操作" width="130">
                  <template #default="{ $index }">
                    <el-button type="primary" size="small" link @click="openEdit('builtin', $index)">编辑</el-button>
                    <el-button type="danger" size="small" link @click="delBuiltinPreset($index)">删除</el-button>
                  </template>
                </el-table-column>
              </el-table>
              <el-empty v-else description="暂无内置模版" />
            </el-tab-pane>
            <el-tab-pane label="自定义模版" name="custom">
              <div style="margin-bottom:12px;color:#666;font-size:13px">当前管理员的自定义通知模版，与「内置模版」功能相同，仅在问卷设计器下拉菜单中分组显示。</div>
              <el-button type="primary" size="small" @click="openAdd('custom')" style="margin-bottom:12px">新增模版</el-button>
              <el-table :data="customPresets" border stripe size="small" style="margin-bottom:16px" v-if="customPresets.length">
                <el-table-column label="名称" width="180" prop="label" />
                <el-table-column label="模版内容" min-width="400">
                  <template #default="{ row }">
                    <div style="white-space:pre-wrap;font-size:12px;line-height:1.6;max-height:200px;overflow-y:auto">{{ row.value }}</div>
                  </template>
                </el-table-column>
                <el-table-column label="操作" width="130">
                  <template #default="{ $index }">
                    <el-button type="primary" size="small" link @click="openEdit('custom', $index)">编辑</el-button>
                    <el-button type="danger" size="small" link @click="delCustomPreset($index)">删除</el-button>
                  </template>
                </el-table-column>
              </el-table>
              <el-empty v-else description="暂无自定义模版" />
            </el-tab-pane>
          </el-tabs>

          <!-- 新增/编辑模版弹窗 -->
          <el-dialog v-model="presetDialogVisible" :title="presetDialogTitle" width="500px" :close-on-click-modal="false">
            <el-form label-position="top" size="small">
              <el-form-item label="模版名称">
                <el-input v-model="presetForm.label" placeholder="模版名称" maxlength="50" />
              </el-form-item>
              <el-form-item label="模版内容">
                <el-input v-model="presetForm.value" type="textarea" :rows="6" placeholder="模版内容，支持 {title} {questionCount} {total} {submitter} {date} {result} 等变量" />
              </el-form-item>
            </el-form>
            <template #footer>
              <el-button @click="presetDialogVisible = false">取消</el-button>
              <el-button type="primary" @click="confirmPresetDialog">确定</el-button>
            </template>
          </el-dialog>
        </el-tab-pane>
      </el-tabs>
    </el-card>
  </div>
</template>

<script>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import request from '../../utils/request'

export default {
  name: 'Setup',
  setup() {
    const activeTab = ref('STATIC_DOMAIN')
    const savingKey = ref('')
    const contents = reactive({})
    const formFields = ref([])
    const savingForm = ref(false)
    const homeConfig = reactive({ vouch_limit: 10, new_limit: 10, hot_limit: 10 })
    const savingHome = ref(false)
    const staticDomain = ref('')
    const savingDomain = ref(false)
    const tokenConfig = reactive({ userExpire: '', userPrefix: '', userSingleLogin: 0, adminExpire: '', adminPrefix: '', adminSingleLogin: 0 })
    const savingToken = ref(false)
    const builtinPresets = ref([])
    const savingBuiltin = ref(false)
    const customPresets = ref([])
    const presetTab = ref('builtin')
    const presetDialogVisible = ref(false)
    const presetDialogTitle = ref('')
    const presetForm = reactive({ label: '', value: '' })
    const presetEditMode = ref('')  // 'add' | 'edit'
    const presetEditType = ref('')  // 'builtin' | 'custom'
    const presetEditIndex = ref(-1)

    const defaultBuiltinPresets = [
      { label: '简洁', value: '📋 问卷「{title}」收到新答卷\n新提交人：{submitter}\n时间：{date}\n共 {total} 份提交\n\n{result}' },
      { label: '详细', value: '📊 问卷统计报告\n━━━━━━━━━━━━━━━━━━\n📌 问卷：{title}\n👤 新提交人：{submitter}\n🕐 时间：{date}\n📈 总提交数：{total}\n\n{result}\n━━━━━━━━━━━━━━━━━━' },
      { label: '仅统计结果', value: '{result}' }
    ]

    const loadBuiltinPresets = async () => {
      try {
        const res = await request.get('/api/v2/home/setup', { params: { key: 'BUILTIN_TEMPLATE_PRESETS' } })
        if (res.code === 0 && res.data) {
          const parsed = typeof res.data === 'string' ? JSON.parse(res.data) : res.data
          builtinPresets.value = Array.isArray(parsed) && parsed.length ? parsed : [...defaultBuiltinPresets]
        } else {
          builtinPresets.value = [...defaultBuiltinPresets]
        }
      } catch (e) {
        builtinPresets.value = [...defaultBuiltinPresets]
      }
    }

    const delBuiltinPreset = async (index) => {
      builtinPresets.value.splice(index, 1)
      await saveBuiltinPresets()
    }

    const saveBuiltinPresets = async () => {
      savingBuiltin.value = true
      try {
        await request.put('/api/v2/admin/settings/content', { key: 'BUILTIN_TEMPLATE_PRESETS', value: JSON.stringify(builtinPresets.value) })
        ElMessage.success('保存成功')
      } catch (e) {
        console.error(e)
      } finally {
        savingBuiltin.value = false
      }
    }

    const loadCustomPresets = async () => {
      try {
        const res = await request.get('/api/v2/admin/survey-template-presets')
        if (res.code === 0 && Array.isArray(res.data)) {
          customPresets.value = res.data
        }
      } catch {}
    }

    const savingCustom = ref(false)
    const delCustomPreset = async (index) => {
      customPresets.value.splice(index, 1)
      await saveCustomPresets()
    }
    const saveCustomPresets = async () => {
      savingCustom.value = true
      try {
        await request.put('/api/v2/admin/survey-template-presets', JSON.stringify({ presets: customPresets.value }), {
          headers: { 'Content-Type': 'application/json' },
          transformRequest: [(d) => d]
        })
        ElMessage.success('保存成功')
      } catch (e) {
        console.error(e)
      } finally {
        savingCustom.value = false
      }
    }

    const openAdd = (type) => {
      presetEditMode.value = 'add'
      presetEditType.value = type
      presetEditIndex.value = -1
      presetForm.label = ''
      presetForm.value = ''
      presetDialogTitle.value = '新增模版'
      presetDialogVisible.value = true
    }
    const openEdit = (type, index) => {
      const list = type === 'builtin' ? builtinPresets.value : customPresets.value
      const item = list[index]
      if (!item) return
      presetEditMode.value = 'edit'
      presetEditType.value = type
      presetEditIndex.value = index
      presetForm.label = item.label
      presetForm.value = item.value
      presetDialogTitle.value = '编辑模版'
      presetDialogVisible.value = true
    }
    const confirmPresetDialog = async () => {
      const name = presetForm.label.trim()
      const content = presetForm.value.trim()
      if (!name) { ElMessage.warning('请输入模版名称'); return }
      if (!content) { ElMessage.warning('请输入模版内容'); return }
      const list = presetEditType.value === 'builtin' ? builtinPresets.value : customPresets.value
      if (presetEditMode.value === 'edit') {
        const idx = presetEditIndex.value
        if (idx >= 0 && idx < list.length) {
          list[idx] = { label: name, value: content }
        }
      } else {
        list.push({ label: name, value: content })
      }
      presetDialogVisible.value = false
      if (presetEditType.value === 'builtin') {
        await saveBuiltinPresets()
      } else {
        await saveCustomPresets()
      }
    }

    const textTabs = [
      { title: '用户协议', key: 'SETUP_CONTENT_AGREEMENT' },
      { title: '隐私政策', key: 'SETUP_CONTENT_PRIVACY' },
      { title: '关于我们', key: 'SETUP_CONTENT_ABOUT' }
    ]

    const loadContent = async (key) => {
      try {
        const res = await request.get('/api/v2/home/setup', { params: { key } })
        contents[key] = res.data || ''
      } catch (e) {
        contents[key] = ''
      }
    }

    const saveContent = async (key) => {
      savingKey.value = key
      try {
        await request.put('/api/v2/admin/settings/content', { key, value: contents[key] || '' })
        ElMessage.success('保存成功')
      } catch (e) {
        console.error(e)
      } finally {
        savingKey.value = ''
      }
    }

    const loadFormFields = async () => {
      try {
        const res = await request.get('/api/v2/user-form-fields')
        formFields.value = res.data || []
      } catch (e) {
        formFields.value = []
      }
    }

    const addField = () => {
      formFields.value.push({ label: '', type: '文本', required: 0, options: '', sort: formFields.value.length })
    }

    const delField = (index) => {
      formFields.value.splice(index, 1)
    }

    const saveFormFields = async () => {
      savingForm.value = true
      try {
        await request.put('/api/v2/admin/settings/content', { key: 'SETUP_USER_FORM_FIELDS', value: JSON.stringify(formFields.value) })
        ElMessage.success('保存成功')
      } catch (e) {
        console.error(e)
      } finally {
        savingForm.value = false
      }
    }

    const loadHomeConfig = async () => {
      try {
        const res = await request.get('/api/v2/home/setup', { params: { key: 'HOME_PAGE_CONFIG' } })
        if (res.data) {
          const parsed = typeof res.data === 'string' ? JSON.parse(res.data) : res.data
          Object.assign(homeConfig, parsed)
        }
      } catch (e) {
        // defaults
      }
    }

    const saveHomeConfig = async () => {
      savingHome.value = true
      try {
        await request.put('/api/v2/admin/settings/content', { key: 'HOME_PAGE_CONFIG', value: JSON.stringify(homeConfig) })
        ElMessage.success('保存成功')
      } catch (e) {
        console.error(e)
      } finally {
        savingHome.value = false
      }
    }

    const loadStaticDomain = async () => {
      try {
        const res = await request.get('/api/v2/home/setup', { params: { key: 'STATIC_DOMAIN' } })
        staticDomain.value = res.data || 'http://localhost:8083'
      } catch (e) {
        staticDomain.value = 'http://localhost:8083'
      }
    }

    const saveStaticDomain = async () => {
      savingDomain.value = true
      try {
        await request.put('/api/v2/admin/settings/content', { key: 'STATIC_DOMAIN', value: staticDomain.value })
        ElMessage.success('保存成功')
      } catch (e) {
        console.error(e)
      } finally {
        savingDomain.value = false
      }
    }

    const tokenConfigKeys = {
      userExpire: 'TOKEN_USER_EXPIRE',
      userPrefix: 'TOKEN_USER_REDIS_PREFIX',
      adminExpire: 'TOKEN_ADMIN_EXPIRE',
      adminPrefix: 'TOKEN_ADMIN_REDIS_PREFIX'
    }

    const loadTokenConfig = async () => {
      const results = await Promise.allSettled(
        Object.entries(tokenConfigKeys).map(([field, key]) =>
          request.get('/api/v2/home/setup', { params: { key } }).then(res => ({ field, value: res.data || '' }))
        )
      )
      for (const r of results) {
        if (r.status === 'fulfilled' && r.value) {
          tokenConfig[r.value.field] = r.value.value
        }
      }
      // 单点登录开关
      try {
        const resUser = await request.get('/api/v2/home/setup', { params: { key: 'USER_SINGLE_LOGIN' } })
        tokenConfig.userSingleLogin = resUser.data === '1' ? 1 : 0
      } catch (e) {
        tokenConfig.userSingleLogin = 0
      }
      try {
        const resAdmin = await request.get('/api/v2/home/setup', { params: { key: 'ADMIN_SINGLE_LOGIN' } })
        tokenConfig.adminSingleLogin = resAdmin.data === '1' ? 1 : 0
      } catch (e) {
        tokenConfig.adminSingleLogin = 0
      }
    }

    const saveTokenConfig = async () => {
      savingToken.value = true
      try {
        await Promise.all([
          ...Object.entries(tokenConfigKeys).map(([field, key]) =>
            request.put('/api/v2/admin/settings/content', { key, value: tokenConfig[field] || '' })
          ),
          request.put('/api/v2/admin/settings/content', {
            key: 'USER_SINGLE_LOGIN',
            value: String(tokenConfig.userSingleLogin)
          }),
          request.put('/api/v2/admin/settings/content', {
            key: 'ADMIN_SINGLE_LOGIN',
            value: String(tokenConfig.adminSingleLogin)
          })
        ])
        ElMessage.success('保存成功')
      } catch (e) {
        console.error(e)
      } finally {
        savingToken.value = false
      }
    }

    onMounted(() => {
      textTabs.forEach(tab => loadContent(tab.key))
      loadFormFields()
      loadHomeConfig()
      loadStaticDomain()
      loadTokenConfig()
      loadBuiltinPresets()
      loadCustomPresets()
    })

    return { activeTab, textTabs, contents, savingKey, formFields, savingForm, homeConfig, savingHome, saveContent, addField, delField, saveFormFields, loadHomeConfig, saveHomeConfig, staticDomain, savingDomain, loadStaticDomain, saveStaticDomain, tokenConfig, savingToken, saveTokenConfig, builtinPresets, savingBuiltin, delBuiltinPreset, saveBuiltinPresets, customPresets, presetTab, savingCustom, delCustomPreset, saveCustomPresets, presetDialogVisible, presetDialogTitle, presetForm, openAdd, openEdit, confirmPresetDialog }
  }
}
</script>

<style scoped>
.setup-page {
  padding: 20px;
}
</style>
