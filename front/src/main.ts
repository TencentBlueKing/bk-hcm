import { createApp } from 'vue';
import { createPinia } from 'pinia';
import { gotoLoginPage } from '@/utils/login-helper';
import { watchVersion } from '@/utils/check-version';

import 'reflect-metadata';

import bus from './common/bus';
import http from './http';
import router from './router';
import App from './app.vue';
import i18n from './language/i18n';
import directive from '@/directive/index';
import components from '@/components/index';
import { useUserStore, preload } from '@/store';
import './style/index.scss';
// 全量引入自定义图标
import './assets/iconfont/style.css';
import '@blueking/ediatable/vue3/vue3.css';

// 全量引入 bkui-vue
import bkui from 'bkui-vue';
// 全量引入 bkui-vue 样式
import 'bkui-vue/dist/style.css';

const app = createApp(App);

const pinia = createPinia();

app.config.globalProperties.$bus = bus;
app.config.globalProperties.$http = http;

app.use(i18n).use(directive).use(components).use(pinia).use(bkui);

const { userInfo } = useUserStore();

userInfo()
  .then(() => {
    preload().finally(() => {
      app.use(router);
      app.mount('#app');
      if (process.env.NODE_ENV === 'production') {
        watchVersion();
      }
    });
  })
  .catch((err) => {
    console.error(err);
    gotoLoginPage();
  });
