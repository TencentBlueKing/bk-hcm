import type { RouteRecordRaw } from 'vue-router';

// 资源选型模块路由
const scheme: RouteRecordRaw[] = [
  {
    path: '/scheme',
    name: 'resource-scheme',
    component: () => import('@/views/scheme/index'),
    children: [
      {
        path: 'recommendation',
        name: 'scheme-recommendation',
        component: () => import('@/views/scheme/scheme-recommendation/index'),
        meta: {
          notMenu: true,
        },
      },
      {
        path: 'deployment/list',
        name: 'scheme-list',
        component: () => import('@/views/scheme/scheme-list/index'),
        meta: {
          notMenu: true,
        },
      },
      {
        path: 'deployment/detail',
        name: 'scheme-detail',
        component: () => import('@/views/scheme/scheme-detail/index'),
        meta: {
          notMenu: true,
        },
      },
    ],
  },
];

export default scheme;
