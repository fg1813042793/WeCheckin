import { createApp } from 'vue'
import ElementPlus from 'element-plus'
import zhCn from 'element-plus/es/locale/lang/zh-cn'
import 'element-plus/dist/index.css'
import '../styles/admin-ui-tokens.css'
import './preview.css'
import './org-approver-preview.css'
import OrgApproverPreview from './OrgApproverPreview.vue'

createApp(OrgApproverPreview).use(ElementPlus, { locale: zhCn }).mount('#app')
