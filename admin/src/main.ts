import { createApp } from 'vue'
import ElementPlus from 'element-plus'
import zhCn from 'element-plus/es/locale/lang/zh-cn'
import 'element-plus/dist/index.css'
import './styles/admin.css'
import { registerAdminIcons } from './icons'
import App from './App.vue'
import router from './router'
const app = createApp(App)
registerAdminIcons(app)
app.use(ElementPlus, { locale: zhCn })
app.use(router)
app.mount('#app')
