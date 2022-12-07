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
        path: '/resource/resource',
        name: '资源',
        component: () => import('@/views/resource/resource-manage/resource-entry.vue'),
        children: [
          {
            path: '/resource/resource',
            component: () => import('@/views/resource/resource-manage/resource-manage.vue'),
            meta: {
              activeKey: 'resourceResource',
              breadcrumb: ['云管', '资源'],
            },
          },
          {
            path: '/resource/detail/:type',
            name: 'resourceDetail',
            component: () => import('@/views/resource/resource-manage/resource-detail.vue'),
            meta: {
              activeKey: 'resourceResource',
              breadcrumb: ['云管', '资源', '详情'],
            },
          },
        ],
        meta: {
          activeKey: 'resourceResource',
          breadcrumb: ['云管', '资源'],
        },
      },
      {
        path: '/resource/recyclebin',
        name: '回收站',
        component: () => import('@/views/workbench/demo2'),
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
        component: () => import('@/views/workbench/demo'),
        meta: {
          activeKey: 'survey',
          breadcrumb: ['网络', '概况'],
        },
      },
      {
        path: '/resource-net/planning',
        name: '规划',
        component: () => import('@/views/workbench/demo'),
        meta: {
          activeKey: 'planning',
          breadcrumb: ['网络', '规划'],
        },
      },
      {
        path: '/resource-net/recycle',
        name: '回收',
        component: () => import('@/views/workbench/demo'),
        meta: {
          activeKey: 'recycle',
          breadcrumb: ['网络', '规划'],
        },
      },
    ],
  },
];

export default resourceMenus;
