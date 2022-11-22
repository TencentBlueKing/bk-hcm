import { createApp } from 'vue';
import { createPinia } from 'pinia';

import bus from './common/bus';
import http from './http';
import router from './router';
import App from './App';
import i18n from './language/i18n';
import './style/index.scss';

// 全量引入 bkui-vue
import bkui from 'bkui-vue';
// 全量引入 bkui-vue 样式
import 'bkui-vue/dist/style.css';

const app = createApp(App);

app.config.globalProperties.$bus = bus;
app.config.globalProperties.$http = http;

app.use(i18n)
  .use(router)
  .use(createPinia())
  .use(bkui)
  .mount('#app');
