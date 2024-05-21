// import { CogShape } from 'bkui-vue/lib/icon';
import type { RouteRecordRaw } from 'vue-router';
import i18n from '@/language/i18n';

const { t } = i18n.global;

const resourceMenus: RouteRecordRaw[] = [
  {
    path: '/resource',
    component: () => import('@/views/resource/resource-manage/resource-entry.vue'),
    children: [
      {
        path: 'resource',
        component: () => import('@/views/resource/resource-manage/resource-manage.vue'),
        children: [
          {
            path: 'record',
            name: 'operationRecord',
            component: () => import('@/views/resource/resource-manage/operationRecord/index'),
          },
          {
            path: 'account',
            name: t('账号详情'),
            component: () => import('@/views/resource/resource-manage/accountInfo/index'),
            children: [
              {
                path: 'detail',
                name: '基本信息',
                component: () => import('@/views/resource/accountmanage/account-detail'),
              },
              {
                path: 'resource',
                name: '资源状态',
                component: () => import('@/views/resource/resource-manage/accountInfo/component/resourceStatus/index'),
              },
              {
                path: 'manage',
                name: '用户列表',
                component: () => import('@/views/resource/resource-manage/accountInfo/component/usersList/index'),
              },
            ],
          },
          {
            path: 'recycle',
            name: '资源接入回收站',
            // component: () => import('@/views/resource/resource-manage/recycleBin/index'),
            component: () => import('@/views/resource/recyclebin-manager/recyclebin-manager.vue'),
          },
        ],
        meta: {
          notMenu: true,
        },
      },
      {
        path: '/resource/account',
        name: t('账户'),
        alias: '',
        component: () => import('@/views/resource/accountmanage/index.vue'),
        meta: {
          activeKey: 'resourceAccount',
          breadcrumb: [t('云管'), t('账户')],
          action: 'account_find',
        },
      },
      {
        path: '/resource/detail/:type',
        name: 'resourceDetail',
        component: () => import('@/views/resource/resource-manage/resource-detail.vue'),
        meta: {
          activeKey: 'resourceResource',
          breadcrumb: [t('云管'), t('资源'), '详情'],
          notMenu: true,
        },
      },
      {
        path: '/resource/record/detail',
        name: 'resourceRecordDetail',
        component: () => import('@/views/resource/resource-manage/operationRecord/RecordDetail/index'),
        meta: {
          activeKey: 'resourceResource',
          breadcrumb: [t('云管'), t('资源'), '详情'],
          notMenu: true,
        },
      },
      {
        path: '/resource/service-apply/cvm',
        name: 'resourceApplyCvm',
        component: () => import('@/views/service/service-apply/cvm'),
        meta: {
          activeKey: 'resourceResource',
          breadcrumb: [t('云管'), t('资源'), '新建主机'],
          notMenu: true,
        },
      },
      {
        path: '/resource/service-apply/vpc',
        name: 'resourceApplyVPC',
        component: () => import('@/views/service/service-apply/vpc'),
        meta: {
          activeKey: 'resourceResource',
          breadcrumb: [t('云管'), t('资源'), '新建VPC'],
          notMenu: true,
        },
      },
      {
        path: '/resource/service-apply/disk',
        name: 'resourceApplyDisk',
        component: () => import('@/views/service/service-apply/disk'),
        meta: {
          activeKey: 'resourceResource',
          breadcrumb: [t('云管'), t('资源'), '新建云硬盘'],
          notMenu: true,
        },
      },
      {
        path: '/resource/service-apply/subnet',
        name: 'resourceApplySubnet',
        component: () => import('@/views/service/service-apply/subnet'),
        meta: {
          backRouter: -1,
          activeKey: 'resourceResource',
          breadcrumb: [t('云管'), t('资源'), '新建子网'],
          notMenu: true,
        },
      },
      {
        path: '/resource/service-apply/clb',
        name: 'resourceApplyClb',
        component: () => import('@/views/service/service-apply/clb'),
        meta: {
          backRouter: -1,
          activeKey: 'resourceResource',
          breadcrumb: [t('云管'), t('资源'), '新建负载均衡'],
          notMenu: true,
          applyRes: 'lb',
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
];

export default resourceMenus;
