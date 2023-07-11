// import { CogShape } from 'bkui-vue/lib/icon';
import type { RouteRecordRaw } from 'vue-router';
import i18n from '@/language/i18n';

const { t } = i18n.global;

const resourceMenus: RouteRecordRaw[] = [
  {
    path: '/resource',
    name: t('云管'),
    children: [
      {
        path: '/resource/account',
        name: t('账号'),
        alias: '',
        component: () => import('@/views/resource/accountmanage/index.vue'),
        meta: {
          activeKey: 'resourceAccount',
          breadcrumb: [t('云管'), t('账号')],
        },
      },
      {
        path: '/resource/resource',
        name: t('资源'),
        component: () => import('@/views/resource/resource-manage/resource-entry.vue'),
        children: [
          {
            path: '/resource/resource',
            component: () => import('@/views/resource/resource-manage/resource-manage.vue'),
            meta: {
              activeKey: 'resourceResource',
              breadcrumb: [t('云管'), t('资源')],
            },
          },
          {
            path: '/resource/detail/:type',
            name: 'resourceDetail',
            component: () => import('@/views/resource/resource-manage/resource-detail.vue'),
            meta: {
              activeKey: 'resourceResource',
              breadcrumb: [t('云管'), t('资源'), '详情'],
            },
          },
          {
            path: '/resource/service-apply/cvm',
            name: 'resourceApplyCvm',
            component: () => import('@/views/service/service-apply/cvm'),
            meta: {
              activeKey: 'resourceResource',
              breadcrumb: [t('云管'), t('资源'), '新建主机'],
            },
          },
          {
            path: '/resource/service-apply/vpc',
            name: 'resourceApplyVPC',
            component: () => import('@/views/service/service-apply/vpc'),
            meta: {
              activeKey: 'resourceResource',
              breadcrumb: [t('云管'), t('资源'), '新建VPC'],
            },
          },
          {
            path: '/resource/service-apply/disk',
            name: 'resourceApplyDisk',
            component: () => import('@/views/service/service-apply/disk'),
            meta: {
              activeKey: 'resourceResource',
              breadcrumb: [t('云管'), t('资源'), '新建云硬盘'],
            },
          },
        ],
        meta: {
          activeKey: 'resourceResource',
          breadcrumb: [t('云管'), t('资源')],
        },
      },
      {
        path: '/resource/recyclebin',
        name: t('回收站'),
        component: () => import('@/views/resource/recyclebin-manager/recyclebin-manager.vue'),
        meta: {
          activeKey: 'resourceRecyclebin',
          breadcrumb: [t('云管'), t('回收站')],
        },
      },
    ],
  },
  // {
  //   path: '/resource-net',
  //   name: t('网络'),
  //   children: [
  //     {
  //       path: '/resource-net/survey',
  //       name: t('概况'),
  //       component: () => import('@/views/workbench/demo'),
  //       meta: {
  //         activeKey: 'survey',
  //         breadcrumb: [t('网络'), t('概况')],
  //       },
  //     },
  //     {
  //       path: '/resource-net/planning',
  //       name: t('规划'),
  //       component: () => import('@/views/workbench/demo'),
  //       meta: {
  //         activeKey: 'planning',
  //         breadcrumb: [t('网络'), t('规划')],
  //       },
  //     },
  //     {
  //       path: '/resource-net/recycle',
  //       name: t('回收'),
  //       component: () => import('@/views/workbench/demo'),
  //       meta: {
  //         activeKey: 'recycle',
  //         breadcrumb: [t('网络'), t('规划')],
  //       },
  //     },
  //   ],
  // },
];

export default resourceMenus;
