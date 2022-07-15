import ElementPlus from 'element-plus'
import 'element-plus/lib/theme-chalk/index.css'
import locale from 'element-plus/lib/locale/lang/zh-cn'
import { App } from 'vue'

export default (app: App) => {
  app.use(ElementPlus, { locale })
}
