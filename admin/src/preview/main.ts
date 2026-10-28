import { createApp } from 'vue'
import ElementPlus from 'element-plus'
import zhCn from 'element-plus/es/locale/lang/zh-cn'
import 'element-plus/dist/index.css'
import '../styles/admin-ui-tokens.css'
import PreviewApp from './PreviewApp.vue'
import './preview.css'

createApp(PreviewApp).use(ElementPlus, { locale: zhCn }).mount('#app')
