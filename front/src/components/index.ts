import { App } from 'vue';
import permissionDialog from '@/components/permission-dialog/install-permission';

const components = [permissionDialog];
export default {
  install(app: App) {
    // eslint-disable-next-line array-callback-return
    components.map((item) => {
      app.use(item);
    });
  },
};
