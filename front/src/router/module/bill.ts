import type { RouteRecordRaw } from 'vue-router';

const bill: RouteRecordRaw[] = [
  {
    path: '/bill/account-manage',
    name: '云账号管理',
    component: () => import('@/views/bill/account/account-manage/index'),
    meta: {
      activeKey: 'account-manage',
      icon: 'hcm-icon bkhcm-icon-host',
    },
  },
  {
    path: '/bill/bill-manage',
    name: '云账单',
    component: () => import('@/views/bill/bill/bill-manage/index'),
    meta: {
      activeKey: 'bill-manage',
      icon: 'hcm-icon bkhcm-icon-host',
    },
  },
];

export default bill;