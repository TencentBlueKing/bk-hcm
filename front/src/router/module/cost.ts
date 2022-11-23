import type { RouteRecordRaw } from 'vue-router';

const costMenus: RouteRecordRaw[] = [
  {
    path: '/cost/resourceAnalyze',
    name: '资源分析',
    component: () => import('@/views/resource/demo'),
    meta: {
      activeKey: 'resourceAnalyze',
    },
  },
  {
    path: '/cost',
    name: '成本分析',
    children: [
      {
        path: '/cost/costReport',
        name: '成本报表',
        alias: '',
        component: () => import('@/views/resource/demo'),
        meta: {
          activeKey: 'costReport',
        },
      },
      {
        path: '/cost/publicCloudBill',
        name: '公有云账单',
        alias: '',
        component: () => import('@/views/resource/demo'),
        meta: {
          activeKey: 'publicCloudBill',
        },
      },
    ],
  },
];

export default costMenus;
