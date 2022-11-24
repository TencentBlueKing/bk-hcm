import type { RouteRecordRaw } from 'vue-router';

const serviceMenus: RouteRecordRaw[] = [
  {
    path: '/service/serviceApply',
    name: '服务申请',
    component: () => import('@/views/resource/demo'),
    meta: {
      activeKey: 'serviceApply',
    },
  },
  {
    path: '/service/myApply',
    name: '我的申请',
    component: () => import('@/views/resource/demo'),
    meta: {
      activeKey: 'myApply',
    },
  },
  {
    path: '/service/myApproval',
    name: '我的审批',
    component: () => import('@/views/resource/demo'),
    meta: {
      activeKey: 'myApproval',
    },
  },
];

export default serviceMenus;
