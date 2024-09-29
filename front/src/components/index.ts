import { type App } from 'vue';
import PermissionDialog from '@/components/permission-dialog';

// 搜索组件
import SearchAccount from './search/account.vue';
import SearchEnum from './search/enum.vue';
import SearchDatetime from './search/datetime.vue';
import SearchUser from './search/user.vue';

// 展示值组件
import DisplayValue from './display-value/index.vue';

const components = [PermissionDialog, SearchAccount, SearchEnum, SearchDatetime, SearchUser, DisplayValue];
export default {
  install(app: App) {
    components.forEach((component) => {
      app.component(component.name, component);
    });
  },
};
