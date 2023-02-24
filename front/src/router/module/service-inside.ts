/**
 * 服务模块除了侧边栏的内部导航
 */
import type { RouteRecordRaw } from 'vue-router';

const serviceInside: RouteRecordRaw[] = [
  {
    path: '/service/service-apply/account-add',
    name: 'applyAccount',
    component: () => import('@/views/service/service-apply/account-add/index'),
    meta: {
      backRouter: -1,
      activeKey: 'serviceApply',
      breadcrumb: ['服务', '服务申请', '账号'],
    },
  },
];

export default serviceInside;
