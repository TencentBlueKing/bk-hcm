import type { RouteRecordRaw } from 'vue-router';
import i18n from '@/language/i18n';
import {
  MENU_SERVICE_HOST_APPLICATION,
  MENU_SERVICE_HOST_RECYCLE_ENTRY,
  MENU_SERVICE_HOST_RECYCLE,
  MENU_SERVICE_RESOURCE_PLAN_CVM,
} from '@/constants/menu-symbol';
import ticketRoutes from '@/views/ticket/route-config';
import { gpuDemandSrv as gpuDemandSrvRouteConfig } from '@/views/resource-plan/route-config';

const { t } = i18n.global;

const serviceMenus: RouteRecordRaw[] = [
  {
    path: '/service',
    children: [
      ...ticketRoutes,
      // 单据管理 tab 资源预测详情
      {
        path: '/service/my-apply/resource-plan/detail',
        redirect: '/service/ticket/resource-plan/detail',
        meta: {
          notMenu: true,
        },
      },

      {
        path: '/service/dissolve',
        component: () => import('@/views/ziyanScr/recycle-server-room'),
        children: [],
        meta: {
          title: t('机房裁撤'),
          activeKey: 'dissolve',
          breadcrumb: ['资源', '主机'],
          isShowBreadcrumb: true,
          icon: 'hcm-icon bkhcm-icon-dissolve',
        },
      },
    ],
    meta: {
      groupTitle: '资源',
    },
  },
  {
    path: '/service',
    children: [
      {
        path: '/service/resource-plan/cvm',
        name: 'opResourcePlan-redirect',
        redirect: '/service/resource-plan/cvm/home',
        meta: {
          activeKey: 'opResourcePlan',
          title: t('资源预测'),
          isShowBreadcrumb: true,
          icon: 'hcm-icon bkhcm-icon-resource-plan',
          notMenu: true,
        },
      },
      {
        path: '/service/resource-plan/cvm/home',
        name: MENU_SERVICE_RESOURCE_PLAN_CVM,
        component: () => import('@/views/resource-plan/entry-srv.vue'),
        meta: {
          activeKey: 'opResourcePlan',
          title: t('资源预测'),
          isShowBreadcrumb: true,
          icon: 'hcm-icon bkhcm-icon-resource-plan',
          checkAuth: 'ziyan_resource_plan_manage',
        },
      },
      {
        path: '/service/resource-plan/cvm/detail',
        name: 'opResourcePlanDetail',
        component: () => import('@/views/service/resource-plan/resource-manage/detail'),
        meta: {
          activeKey: 'opResourcePlanDetail',
          notMenu: true,
        },
      },
      {
        path: '/service/resource-plan/cvm/mod',
        name: 'modPlanList',
        component: () => import('@/views/service/resource-plan/resource-manage/mod'),
        meta: {
          activeKey: 'planlist',
          notMenu: true,
        },
      },
      ...gpuDemandSrvRouteConfig,
      {
        path: '/service/resource-plan',
        redirect: (to) => ({ path: '/service/resource-plan/cvm/home', query: to.query }),
        meta: { notMenu: true },
      },
      {
        path: '/service/resource-plan/home',
        redirect: (to) => ({ path: '/service/resource-plan/cvm/home', query: to.query }),
        meta: { notMenu: true },
      },
      {
        path: '/service/resource-plan/detail',
        redirect: (to) => ({ path: '/service/resource-plan/cvm/detail', query: to.query }),
        meta: { notMenu: true },
      },
      {
        path: '/service/resource-plan/mod',
        redirect: (to) => ({ path: '/service/resource-plan/cvm/mod', query: to.query }),
        meta: { notMenu: true },
      },
      {
        path: '/service/hostInventory',
        component: () => import('@/views/ziyanScr/hostInventory/index'),
        meta: {
          title: t('主机库存'),
          activeKey: 'inventory',
          breadcrumb: ['资源', '主机'],
          isShowBreadcrumb: true,
          icon: 'hcm-icon bkhcm-icon-host-inventory',
          checkAuth: 'ziyan_resource_inventory_find',
        },
      },
      {
        path: '/service/hostApplication',
        component: () => import('@/views/ziyanScr/hostApplication'),
        name: MENU_SERVICE_HOST_APPLICATION,
        meta: {
          title: t('主机申领'),
          activeKey: 'apply',
          breadcrumb: ['资源', '主机'],
          isShowBreadcrumb: true,
          icon: 'hcm-icon bkhcm-icon-host-application',
          checkAuth: 'ziyan_resource_create',
        },
      },
      {
        path: '/service/hostApplication/detail/:id',
        name: 'host-application-detail',
        component: () => import('@/views/ziyanScr/hostApplication/components/application-detail/index'),
        meta: {
          activeKey: 'apply',
          notMenu: true,
          menu: {
            relative: MENU_SERVICE_HOST_APPLICATION,
          },
        },
      },
      {
        path: '/service/hostApplication/apply',
        name: '提交主机申请',
        component: () => import('@/views/ziyanScr/hostApplication/components/application-form/index'),
        meta: {
          activeKey: 'apply',
          notMenu: true,
        },
      },
      {
        path: '/service/hostApplication/modify',
        name: '修改主机申请',
        component: () => import('@/views/ziyanScr/hostApplication/components/application-modify/index.vue'),
        meta: {
          activeKey: 'apply',
          notMenu: true,
        },
      },
      {
        path: '/service/hostRecycling',
        name: MENU_SERVICE_HOST_RECYCLE_ENTRY,
        children: [
          {
            path: '',
            name: MENU_SERVICE_HOST_RECYCLE,
            component: () => import('@/views/ziyanScr/host-recycle'),
            meta: {
              activeKey: 'recovery',
              isShowBreadcrumb: true,
              breadcrumb: ['资源', '主机'],
            },
          },
          {
            path: 'resources',
            name: 'resources',
            component: () => import('@/views/ziyanScr/RecyclingResources'),
            meta: {
              activeKey: 'recovery',
              breadcrumb: ['资源', '主机'],
            },
          },
          {
            path: 'preDetail',
            name: 'PreDetail',
            component: () => import('@/views/ziyanScr/host-recycle/pre-details'),
            meta: {
              activeKey: 'recovery',
              breadcrumb: ['资源', '主机'],
            },
          },
          {
            path: 'docDetail',
            name: 'docDetail',
            component: () => import('@/views/ziyanScr/host-recycle/bill-detail'),
            meta: {
              activeKey: 'recovery',
              breadcrumb: ['资源', '主机'],
            },
          },
        ],
        meta: {
          activeKey: 'recovery',
          title: t('主机回收'),
          breadcrumb: ['资源', '主机'],
          icon: 'hcm-icon bkhcm-icon-host-recycle',
          checkAuth: 'ziyan_resource_recycle',
        },
      },
      {
        path: '/service/cvm-model',
        component: () => import('@/views/ziyanScr/cvm-model'),
        name: 'CVM机型',
        children: [],
        meta: {
          activeKey: 'model',
          title: t('CVM机型'),
          breadcrumb: ['资源', '主机'],
          isShowBreadcrumb: true,
          icon: 'hcm-icon bkhcm-icon-cvm-type',
          checkAuth: 'ziyan_cvm_type_find',
        },
      },
      {
        path: '/service/cvm-subnet',
        name: 'CVM子网',
        component: () => import('@/views/ziyanScr/cvm-web'),
        children: [],
        meta: {
          title: t('CVM子网'),
          activeKey: 'subnet',
          breadcrumb: ['资源', '主机'],
          isShowBreadcrumb: true,
          icon: 'hcm-icon bkhcm-icon-subnet',
          checkAuth: 'ziyan_cvm_subnet_find',
        },
      },
      {
        path: '/service/resource-manage',
        name: '资源上下架',
        children: [
          {
            path: '',
            name: 'resourceManage',
            component: () => import('@/views/ziyanScr/resource-manage'),
            meta: {
              activeKey: 'scr-resource-manage',
              isShowBreadcrumb: true,
            },
          },
          {
            path: 'detail/:id',
            name: 'scrResourceManageDetail',
            component: () => import('@/views/ziyanScr/resource-manage/detail'),
            props(route) {
              return { ...route.params, ...route.query };
            },
            meta: {
              activeKey: 'scr-resource-manage',
            },
          },
          {
            path: 'create',
            name: 'scrResourceManageCreate',
            component: () => import('@/views/ziyanScr/resource-manage/create'),
            props(route) {
              return { ...route.query };
            },
            meta: {
              activeKey: 'scr-resource-manage',
            },
          },
        ],
        meta: {
          title: t('资源上下架'),
          activeKey: 'scr-resource-manage',
          icon: 'hcm-icon bkhcm-icon-res-shelves',
          checkAuth: 'ziyan_res_shelves_find',
        },
      },
      {
        path: '/service/cvm-produce',
        name: 'CVM生产',
        component: () => import('@/views/ziyanScr/cvm-produce'),
        children: [],
        meta: {
          title: t('CVM生产'),
          activeKey: 'produce',
          breadcrumb: ['资源', '主机'],
          isShowBreadcrumb: true,
          icon: 'hcm-icon bkhcm-icon-cvm-produce',
          checkAuth: 'ziyan_cvm_create_find',
        },
      },
    ],
    meta: {
      groupTitle: '管理',
    },
  },
];

export default serviceMenus;
