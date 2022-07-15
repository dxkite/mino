import { createApp } from 'vue'
import App from './App.vue'
import installElementPlus from './plugins/element'
import VueForm from '@lljj/vue3-form-element'
import router from './router'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'

const app = createApp(App).use(router)
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
    app.component(key, component)
}

installElementPlus(app)

app.component('VueForm', VueForm);
app.mount('#app')