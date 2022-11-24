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
        component: () => import('@/views/resource/accountmanage/index.vue'),
        meta: {
          activeKey: 'resourceAccount',
          breadcrumb: ['云管', '账户'],
        },
      },
      {
        path: '/resource/res',
        name: '资源',
        component: () => import('@/views/resource/demo2'),
        meta: {
          activeKey: 'resourceRes',
          breadcrumb: ['云管', '资源'],
        },
      },
      {
        path: '/resource/recyclebin',
        name: '回收站',
        component: () => import('@/views/resource/demo2'),
        meta: {
          activeKey: 'resourceRecyclebin',
          breadcrumb: ['云管', '回收站'],
        },
      },
    ],
  },
  {
    path: '/resource-net',
    name: '网络',
    children: [
      {
        path: '/resource-net/survey',
        name: '概况',
        component: () => import('@/views/resource/demo'),
        meta: {
          activeKey: 'survey',
          breadcrumb: ['网络', '概况'],
        },
      },
      {
        path: '/resource-net/planning',
        name: '规划',
        component: () => import('@/views/resource/demo'),
        meta: {
          activeKey: 'planning',
          breadcrumb: ['网络', '规划'],
        },
      },
      {
        path: '/resource-net/recycle',
        name: '回收',
        component: () => import('@/views/resource/demo'),
        meta: {
          activeKey: 'recycle',
          breadcrumb: ['网络', '规划'],
        },
      },
    ],
  },
];

export default resourceMenus;
