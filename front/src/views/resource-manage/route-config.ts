import type { RouteRecordRaw } from 'vue-router';
import Meta from '@/router/meta';
import { AUTH_FIND_IAAS_RESOURCE } from '@/constants/auth-symbols';
import {
  MENU_RESOURCE,
  MENU_RESOURCE_MANAGE,
  MENU_RESOURCE_RESOURCE_LIST,
  MENU_RESOURCE_DETAIL,
  MENU_RESOURCE_HOST_APPLY,
  MENU_RESOURCE_VPC_APPLY,
  MENU_RESOURCE_DISK_APPLY,
  MENU_RESOURCE_SUBNET_APPLY,
  MENU_RESOURCE_LOAD_BALANCER_APPLY,
} from '@/constants/menu-symbol';

/**
 * 资源纳管模块 —— 唯一需要 AccountVendorGroup 侧边栏的模块
 * resource-entry.vue 提供左侧云厂商&账号选择器 + 账号头部信息 + RouterView
 */
const resourceManageRoutes: RouteRecordRaw[] = [
  {
    name: MENU_RESOURCE_MANAGE,
    path: 'manage',
    component: () => import('@/views/resource-manage/resource-entry.vue'),
    redirect: { name: MENU_RESOURCE_RESOURCE_LIST, params: { resourceType: 'host' } },
    meta: {
      ...new Meta({
        owner: MENU_RESOURCE,
        activeKey: MENU_RESOURCE_MANAGE,
        auth: {
          view: { type: AUTH_FIND_IAAS_RESOURCE },
        },
        menu: {
          i18n: '资源纳管',
        },
        layout: {
          breadcrumb: {
            show: false,
          },
        },
      }),
    },
    children: [
      // 资源详情（更具体的路径需放在列表之前）
      {
        name: MENU_RESOURCE_DETAIL,
        path: ':resourceType/details/:id',
        component: () => import('@/views/resource/resource-manage/resource-detail.vue'),
        meta: {
          ...new Meta({
            owner: MENU_RESOURCE,
            activeKey: MENU_RESOURCE_MANAGE,
            menu: {
              i18n: '资源详情',
              relative: MENU_RESOURCE_MANAGE,
            },
            layout: {
              breadcrumb: {
                show: false,
              },
            },
          }),
        },
      },
      // 资源列表（资源类型 tab：主机、VPC、子网等）
      {
        name: MENU_RESOURCE_RESOURCE_LIST,
        path: ':resourceType',
        component: () => import('@/views/resource-manage/resource-list.vue'),
        meta: {
          ...new Meta({
            owner: MENU_RESOURCE,
            activeKey: MENU_RESOURCE_MANAGE,
            menu: {
              i18n: '资源列表',
              relative: MENU_RESOURCE_MANAGE,
            },
            layout: {
              breadcrumb: {
                show: false,
              },
            },
          }),
        },
      },
      // 申请主机
      {
        name: MENU_RESOURCE_HOST_APPLY,
        path: 'service-apply/cvm',
        component: () => import('@/views/service/service-apply/cvm'),
        meta: {
          ...new Meta({
            owner: MENU_RESOURCE,
            activeKey: MENU_RESOURCE_MANAGE,
            menu: { i18n: '申请主机', relative: MENU_RESOURCE_MANAGE },
            layout: { breadcrumb: { show: false } },
          }),
        },
      },
      // 申请 VPC
      {
        name: MENU_RESOURCE_VPC_APPLY,
        path: 'service-apply/vpc',
        component: () => import('@/views/service/service-apply/vpc'),
        meta: {
          ...new Meta({
            owner: MENU_RESOURCE,
            activeKey: MENU_RESOURCE_MANAGE,
            menu: { i18n: '申请VPC', relative: MENU_RESOURCE_MANAGE },
            layout: { breadcrumb: { show: false } },
          }),
        },
      },
      // 申请硬盘
      {
        name: MENU_RESOURCE_DISK_APPLY,
        path: 'service-apply/disk',
        component: () => import('@/views/service/service-apply/disk'),
        meta: {
          ...new Meta({
            owner: MENU_RESOURCE,
            activeKey: MENU_RESOURCE_MANAGE,
            menu: { i18n: '申请硬盘', relative: MENU_RESOURCE_MANAGE },
            layout: { breadcrumb: { show: false } },
          }),
        },
      },
      // 申请子网
      {
        name: MENU_RESOURCE_SUBNET_APPLY,
        path: 'service-apply/subnet',
        component: () => import('@/views/service/service-apply/subnet'),
        meta: {
          ...new Meta({
            owner: MENU_RESOURCE,
            activeKey: MENU_RESOURCE_MANAGE,
            menu: { i18n: '申请子网', relative: MENU_RESOURCE_MANAGE },
            layout: { breadcrumb: { show: false } },
          }),
        },
      },
      // 申请负载均衡
      {
        name: MENU_RESOURCE_LOAD_BALANCER_APPLY,
        path: 'service-apply/clb',
        component: () => import('@/views/service/service-apply/clb'),
        meta: {
          ...new Meta({
            owner: MENU_RESOURCE,
            activeKey: MENU_RESOURCE_MANAGE,
            menu: { i18n: '申请负载均衡', relative: MENU_RESOURCE_MANAGE },
            layout: { breadcrumb: { show: false } },
          }),
        },
      },
    ],
  },
];

export default resourceManageRoutes;
