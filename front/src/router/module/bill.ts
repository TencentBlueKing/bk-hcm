import type { RouteRecordRaw } from 'vue-router';

const bill: RouteRecordRaw[] = [
  {
    path: '/bill/account-manage',
    name: 'account-manage',
    component: () => import('@/views/bill/account/account-manage/index'),
    meta: {
      title: '云账号管理',
      activeKey: 'account-manage',
      icon: 'hcm-icon bkhcm-icon-user-8',
    },
  },
  {
    path: '/bill/bill-manage',
    name: 'bill-manage',
    component: () => import('@/views/bill/bill/index'),
    redirect: '/bill/bill-manage/summary',
    children: [
      {
        path: 'summary',
        name: 'billSummary',
        component: () => import('@/views/bill/bill/summary'),
        redirect: '/bill/bill-manage/summary/manage',
        children: [
          {
            path: 'manage',
            name: 'billSummaryManage',
            component: () => import('@/views/bill/bill/summary/manage'),
          },
          {
            path: 'operation-record',
            name: 'billSummaryOperationRecord',
            component: () => import('@/views/bill/bill/summary/operation-record'),
          },
        ],
      },
      {
        path: 'detail',
        name: 'billDetail',
        component: () => import('@/views/bill/bill/detail'),
      },
      {
        path: 'adjust',
        name: 'billAdjust',
        component: () => import('@/views/bill/bill/adjust'),
      },
    ],
    meta: {
      title: '云账单管理',
      activeKey: 'bill-manage',
      icon: 'hcm-icon bkhcm-icon-host',
      hasPageRoute: true,
      checkAuth: 'account_bill_find',
    },
  },
  {
    path: '/bill/account-manage/first-account',
    name: '录入一级账号',
    component: () => import('@/views/bill/account/create-account/create-first-account'),
    meta: {
      notMenu: true,
      activeKey: 'account-manage',
    },
  },
  {
    path: '/bill/account-manage/second-account',
    name: '创建二级账号',
    component: () => import('@/views/bill/account/create-account/create-second-account'),
    meta: {
      notMenu: true,
      activeKey: 'account-manage',
    },
  },
];

export default bill;
