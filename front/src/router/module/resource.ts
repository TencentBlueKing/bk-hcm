// import { CogShape } from 'bkui-vue/lib/icon';
import type { RouteRecordRaw } from 'vue-router';
import i18n from '@/language/i18n';
import { operationLogRsc as operationLogRscRouteConfig } from '@/views/operation-log/route-config';
import Meta from '../meta';
import {
  MENU_RESOURCE_LOAD_BALANCER_APPLY,
  MENU_RESOURCE_DISK_APPLY,
  MENU_RESOURCE_HOST_APPLY,
  MENU_RESOURCE_RESOURCE_MANAGEMENT,
  MENU_RESOURCE_SUBNET_APPLY,
  MENU_RESOURCE_VPC_APPLY,
} from '@/constants/menu-symbol';

const { t } = i18n.global;

const resourceMenus: RouteRecordRaw[] = [
  {
    path: '/resource',
    component: () => import('@/views/resource/resource-manage/resource-entry.vue'),
    children: [
      {
        path: 'resource',
        name: MENU_RESOURCE_RESOURCE_MANAGEMENT,
        component: () => import('@/views/resource/resource-manage/resource-manage.vue'),
        children: [
          operationLogRscRouteConfig[0],
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
          action: 'account_find',
        },
      },
      {
        path: '/resource/detail/:type',
        name: 'resourceDetail',
        component: () => import('@/views/resource/resource-manage/resource-detail.vue'),
        meta: {
          notMenu: true,
        },
      },
      operationLogRscRouteConfig[1],
      {
        path: '/resource/service-apply/cvm',
        name: MENU_RESOURCE_HOST_APPLY,
        component: () => import('@/views/service/service-apply/cvm'),
        meta: {
          ...new Meta({
            notMenu: true,
            menu: {
              relative: MENU_RESOURCE_RESOURCE_MANAGEMENT,
            },
          }),
        },
      },
      {
        path: '/resource/service-apply/vpc',
        name: MENU_RESOURCE_VPC_APPLY,
        component: () => import('@/views/service/service-apply/vpc'),
        meta: {
          ...new Meta({
            notMenu: true,
            menu: {
              relative: MENU_RESOURCE_RESOURCE_MANAGEMENT,
            },
          }),
        },
      },
      {
        path: '/resource/service-apply/disk',
        name: MENU_RESOURCE_DISK_APPLY,
        component: () => import('@/views/service/service-apply/disk'),
        meta: {
          ...new Meta({
            notMenu: true,
            menu: {
              relative: MENU_RESOURCE_RESOURCE_MANAGEMENT,
            },
          }),
        },
      },
      {
        path: '/resource/service-apply/subnet',
        name: MENU_RESOURCE_SUBNET_APPLY,
        component: () => import('@/views/service/service-apply/subnet'),
        meta: {
          ...new Meta({
            notMenu: true,
            menu: {
              relative: MENU_RESOURCE_RESOURCE_MANAGEMENT,
            },
          }),
        },
      },
      {
        path: '/resource/service-apply/clb',
        name: MENU_RESOURCE_LOAD_BALANCER_APPLY,
        component: () => import('@/views/service/service-apply/clb'),
        meta: {
          ...new Meta({
            notMenu: true,
            isFilterAccount: true,
            menu: {
              relative: MENU_RESOURCE_RESOURCE_MANAGEMENT,
            },
          }),
        },
      },
      {
        path: '/resource/recyclebin',
        name: t('回收站'),
        component: () => import('@/views/resource/recyclebin-manager/recyclebin-manager.vue'),
      },
    ],
  },
];

export default resourceMenus;
