import { createApp } from 'vue'
import ElementPlus from 'element-plus'
import zhCn from 'element-plus/es/locale/lang/zh-cn'
import 'element-plus/dist/index.css'
import '../styles/admin-ui-tokens.css'
import UserPreview from './UserPreview.vue'
import './preview.css'
import './user-preview.css'

createApp(UserPreview).use(ElementPlus, { locale: zhCn }).mount('#app')
