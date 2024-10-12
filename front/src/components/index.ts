import { type App } from 'vue';
import PermissionDialog from '@/components/permission-dialog';

// 搜索组件
import SearchAccount from './search/account.vue';
import SearchEnum from './search/enum.vue';
import SearchDatetime from './search/datetime.vue';
import SearchUser from './search/user.vue';

// 展示值组件
import DisplayValue from './display-value/index.vue';

// 表单元素组件
import FormBool from './form/bool.vue';
import FormEnum from './form/enum.vue';
import FormString from './form/string.vue';
import FormArray from './form/array.vue';
import FormNumber from './form/number.vue';
import FormCert from './form/cert.vue';
import FormCa from './form/ca.vue';

const components = [
  PermissionDialog,
  SearchAccount,
  SearchEnum,
  SearchDatetime,
  SearchUser,
  DisplayValue,
  FormBool,
  FormEnum,
  FormString,
  FormArray,
  FormNumber,
  FormCert,
  FormCa,
];
export default {
  install(app: App) {
    components.forEach((component) => {
      app.component(component.name, component);
    });
  },
};
