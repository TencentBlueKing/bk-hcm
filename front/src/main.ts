import { createApp } from 'vue';
import { createPinia } from 'pinia';

import bus from './common/bus';
import http from './http';
import router from './router';
import App from './app';
import i18n from './language/i18n';
import directive from '@/directive/index';
import components from '@/components/index';
import './style/index.scss';
// 全量引入自定义图标
import './assets/iconfont/style.css';

// 全量引入 bkui-vue
import bkui from 'bkui-vue';
// 全量引入 bkui-vue 样式
import 'bkui-vue/dist/style.css';

const app = createApp(App);

const pinia = createPinia();

app.config.globalProperties.$bus = bus;
app.config.globalProperties.$http = http;

app.use(i18n).use(directive).use(router).use(components).use(pinia).use(bkui);

router.isReady().then(() => {
  app.mount('#app');
});
