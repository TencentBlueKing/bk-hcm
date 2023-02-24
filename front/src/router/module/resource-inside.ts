/**
 * 资源模块除了侧边栏的内部导航
 */
import type { RouteRecordRaw } from 'vue-router';

const resourceInside: RouteRecordRaw[] = [
  {
    path: '/resource/account/add',
    name: 'accountAdd',
    component: () => import('@/views/resource/accountmanage/account-add'),
    meta: {
      backRouter: -1,
      activeKey: 'resourceAccount',
      breadcrumb: ['云管', '账号', '新增账号'],
    },
  },
  {
    path: '/resource/account/detail',
    name: 'accountDetail',
    component: () => import('@/views/resource/accountmanage/account-detail'),
    meta: {
      backRouter: -1,
      activeKey: 'resourceAccount',
      breadcrumb: ['云管', '账号', '详情'],
    },
  },
];

export default resourceInside;
