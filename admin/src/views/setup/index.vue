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
            <el-form-item label="图片验证码">
              <el-switch
                v-model="tokenConfig.adminCaptchaEnabled"
                :active-value="1"
                :inactive-value="0"
                active-text="开启"
                inactive-text="关闭"
              />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="saveTokenConfig" :loading="savingToken">保存</el-button>
            </el-form-item>
          </el-form>
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
    const tokenConfig = reactive({ userExpire: '', userPrefix: '', userSingleLogin: 0, adminExpire: '', adminPrefix: '', adminSingleLogin: 0, adminCaptchaEnabled: 0 })
    const savingToken = ref(false)
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
      try {
        const resCaptcha = await request.get('/api/v2/home/setup', { params: { key: 'ADMIN_LOGIN_CAPTCHA_ENABLED' } })
        tokenConfig.adminCaptchaEnabled = resCaptcha.data === '1' ? 1 : 0
      } catch (e) {
        tokenConfig.adminCaptchaEnabled = 0
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
          }),
          request.put('/api/v2/admin/settings/content', {
            key: 'ADMIN_LOGIN_CAPTCHA_ENABLED',
            value: String(tokenConfig.adminCaptchaEnabled)
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
    })

    return { activeTab, textTabs, contents, savingKey, formFields, savingForm, homeConfig, savingHome, saveContent, addField, delField, saveFormFields, loadHomeConfig, saveHomeConfig, staticDomain, savingDomain, loadStaticDomain, saveStaticDomain, tokenConfig, savingToken, saveTokenConfig }
  }
}
</script>

<style scoped>
.setup-page {
  padding: 20px;
}

</style>
