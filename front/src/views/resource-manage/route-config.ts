import type { RouteLocationNormalized, RouteRecordRaw } from 'vue-router';
import Meta from '@/router/meta';
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

// 模块内的面包屑由 resource-entry.vue 在右栏渲染，整页顶部不再画一条
const getMeta = (title: string, relative?: symbol) =>
  new Meta({
    owner: MENU_RESOURCE,
    activeKey: MENU_RESOURCE_MANAGE,
    title,
    menu: { relative },
    layout: { breadcrumb: { show: false } },
  });

const resourceManageRoutes: RouteRecordRaw[] = [
  {
    name: MENU_RESOURCE_MANAGE,
    path: 'manage',
    component: () => import('@/views/resource-manage/resource-entry.vue'),
    redirect: { name: MENU_RESOURCE_RESOURCE_LIST, params: { resourceType: 'host' } },
    meta: { ...getMeta('资源纳管') },
    children: [
      // 无 _f 返回时按 name 回退会沿用当前的 resourceType，回到同类型列表
      {
        name: MENU_RESOURCE_DETAIL,
        path: ':resourceType/details/:id',
        component: () => import('@/views/resource/resource-manage/resource-detail.vue'),
        meta: { ...getMeta('资源详情', MENU_RESOURCE_RESOURCE_LIST) },
      },
      {
        name: MENU_RESOURCE_RESOURCE_LIST,
        path: ':resourceType',
        component: () => import('@/views/resource-manage/resource-list.vue'),
        meta: { ...getMeta('资源列表', MENU_RESOURCE_MANAGE) },
      },
      {
        name: MENU_RESOURCE_HOST_APPLY,
        path: 'service-apply/cvm',
        component: () => import('@/views/service/service-apply/cvm'),
        meta: { ...getMeta('购买主机', MENU_RESOURCE_RESOURCE_LIST) },
      },
      {
        name: MENU_RESOURCE_VPC_APPLY,
        path: 'service-apply/vpc',
        component: () => import('@/views/service/service-apply/vpc'),
        meta: { ...getMeta('购买VPC', MENU_RESOURCE_RESOURCE_LIST) },
      },
      {
        name: MENU_RESOURCE_DISK_APPLY,
        path: 'service-apply/disk',
        component: () => import('@/views/service/service-apply/disk'),
        meta: { ...getMeta('购买硬盘', MENU_RESOURCE_RESOURCE_LIST) },
      },
      {
        name: MENU_RESOURCE_SUBNET_APPLY,
        path: 'service-apply/subnet',
        component: () => import('@/views/service/service-apply/subnet'),
        meta: { ...getMeta('新建子网', MENU_RESOURCE_RESOURCE_LIST) },
      },
      {
        name: MENU_RESOURCE_LOAD_BALANCER_APPLY,
        path: 'service-apply/clb',
        component: () => import('@/views/service/service-apply/clb'),
        // components/account-selector 据此只列腾讯云账号
        meta: { ...getMeta('购买负载均衡', MENU_RESOURCE_RESOURCE_LIST), isFilterAccount: true },
      },
    ],
  },
];

// 旧资源列表 /resource/resource?type=&accountId=：type 进 params，其余参数保留
export const toResourceList = (to: RouteLocationNormalized) => {
  const { type, ...query } = to.query;
  return { name: MENU_RESOURCE_RESOURCE_LIST, params: { resourceType: (type as string) || 'host' }, query };
};

const applyPairs: [string, symbol][] = [
  ['cvm', MENU_RESOURCE_HOST_APPLY],
  ['vpc', MENU_RESOURCE_VPC_APPLY],
  ['disk', MENU_RESOURCE_DISK_APPLY],
  ['subnet', MENU_RESOURCE_SUBNET_APPLY],
  ['clb', MENU_RESOURCE_LOAD_BALANCER_APPLY],
];

// 旧新建页地址 /resource/service-apply/*，参数保留
export const resourceManageLegacy: RouteRecordRaw[] = applyPairs.map(([slug, name]) => ({
  path: `/resource/service-apply/${slug}`,
  redirect: (to: RouteLocationNormalized) => ({ name, query: to.query }),
}));

const legacyDetailTypeMap: Record<string, string> = { eips: 'eip', route: 'routing' };

// 旧资源详情 /resource/detail/:resourceType?id=&type=<云厂商>：类型改用新名，云厂商改用 vendor；缺 id 回对应列表
export const resourceDetailLegacy: RouteRecordRaw[] = [
  {
    path: '/resource/detail/:resourceType',
    redirect: (to: RouteLocationNormalized) => {
      const { id, type, _f, ...query } = to.query;
      const resourceType = legacyDetailTypeMap[to.params.resourceType as string] || to.params.resourceType;
      if (!id) {
        return { name: MENU_RESOURCE_RESOURCE_LIST, params: { resourceType }, query: { accountId: query.accountId } };
      }
      return {
        name: MENU_RESOURCE_DETAIL,
        params: { resourceType, id: id as string },
        query: { ...query, vendor: type },
      };
    },
  },
];

export default resourceManageRoutes;
