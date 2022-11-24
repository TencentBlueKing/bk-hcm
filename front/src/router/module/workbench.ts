// import { CogShape } from 'bkui-vue/lib/icon';
import type { RouteRecordRaw } from 'vue-router';

const workbenchMenus: RouteRecordRaw[] = [
  {
    path: '/workbench',
    name: '配置',
    children: [
      {
        path: '/workbench/auto',
        name: 'agent自动化',
        alias: '',
        component: () => import('@/views/workbench/demo2'),
        meta: {
          activeKey: 'workbenchAuto',
        },
      },
    ],
  },
  {
    path: '/workbench-audit',
    name: '审计',
    component: () => import('@/views/workbench/demo'),
    meta: {
      activeKey: 'workbenchAudit',
    },
  },
];

export default workbenchMenus;
