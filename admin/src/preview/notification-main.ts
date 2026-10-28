import { createApp } from 'vue'
import ElementPlus from 'element-plus'
import zhCn from 'element-plus/es/locale/lang/zh-cn'
import 'element-plus/dist/index.css'
import '../styles/admin-ui-tokens.css'
import NotificationPreview from './NotificationPreview.vue'
import './preview.css'
import './notification-preview.css'

createApp(NotificationPreview).use(ElementPlus, { locale: zhCn }).mount('#app')
