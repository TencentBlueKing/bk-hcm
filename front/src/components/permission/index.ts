import permissionDialog from '@/components/permission-dialog';

import { App } from 'vue';

export default {
  install(app: App) {
    // 此处形参为main.js文件中use()方法自动传进来的Vue实例
    app.component(permissionDialog.name, permissionDialog);
  },
};
