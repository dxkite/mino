import { createApp } from 'vue';
import App from './App.vue';
import VueForm from '@lljj/vue3-form-element';
import router from './router';
import ElementPlus from 'element-plus';
import 'element-plus/dist/index.css';
import * as ElementPlusIconsVue from '@element-plus/icons-vue';

const app = createApp(App).use(router)
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
    app.component(key, component)
}

app.use(ElementPlus);
app.component('VueForm', VueForm);
app.mount('#app');