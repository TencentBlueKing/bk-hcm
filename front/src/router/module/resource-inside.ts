/**
 * 资源模块除了侧边栏的内部导航
 */
import type { RouteRecordRaw } from 'vue-router';

const resourceInside: RouteRecordRaw[] = [
  {
    path: '/resource/account/add',
    name: 'accountAdd',
    component: () => import('@/views/resource/accountmanage/account-add.vue'),
    meta: {
      activeKey: 'resourceAccount',
      breadcrumb: ['云管', '账户', '新增账户'],
    },
  },
];

export default resourceInside;
