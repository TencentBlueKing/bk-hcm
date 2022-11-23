// import { CogShape } from 'bkui-vue/lib/icon';
import type { RouteRecordRaw } from 'vue-router';

const resourceMenus: RouteRecordRaw[] = [
  {
    path: '/resource',
    name: '云管',
    children: [
      {
        path: '/resource/account',
        name: '账户',
        alias: '',
        component: () => import('@/views/resource/demo2'),
        meta: {
          activeKey: 'account',
        },
      },
      {
        path: '/resource/resource',
        name: '资源',
        component: () => import('@/views/resource/demo2'),
        meta: {
          activeKey: 'resource',
        },
      },
      {
        path: '/resource/recyclebin',
        name: '回收站',
        component: () => import('@/views/resource/demo2'),
        meta: {
          activeKey: 'recyclebin',
        },
      },
    ],
  },
  {
    path: '/resource/net',
    name: '网络',
    children: [
      {
        path: '/resource/net/survey',
        name: '概况',
        component: () => import('@/views/resource/demo'),
        meta: {
          activeKey: 'survey',
        },
      },
      {
        path: '/resource/net/planning',
        name: '规划',
        component: () => import('@/views/resource/demo'),
        meta: {
          activeKey: 'planning',
        },
      },
      {
        path: '/resource/net/recycle',
        name: '回收',
        component: () => import('@/views/resource/demo'),
        meta: {
          activeKey: 'recycle',
        },
      },
    ],
  },
];

export default resourceMenus;
