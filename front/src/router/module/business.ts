// import { CogShape } from 'bkui-vue/lib/icon';
import type { RouteRecordRaw } from 'vue-router';

const businesseMenus: RouteRecordRaw[] = [
  {
    path: '/business',
    name: '业务',
    children: [
      {
        path: '/business/auto',
        name: 'agent自动化',
        alias: '',
        component: () => import('@/views/business/demo2'),
        meta: {
          activeKey: 'agentAuto',
        },
      },
    ],
  },
  {
    path: '/business/audit',
    name: '审计',
    component: () => import('@/views/business/demo'),
    meta: {
      activeKey: 'businessAudit',
    },
  },
];

export default businesseMenus;
