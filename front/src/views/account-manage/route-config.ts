import type { LocationQuery, RouteLocationNormalized, RouteRecordRaw } from 'vue-router';
import Meta from '@/router/meta';
import { AUTH_FIND_ACCOUNT } from '@/constants/auth-symbols';
import {
  MENU_RESOURCE,
  MENU_RESOURCE_ACCOUNT_CHECK_IN,
  MENU_RESOURCE_ACCOUNT_DETAIL,
  MENU_RESOURCE_ACCOUNT_DETAIL_BASIC,
  MENU_RESOURCE_ACCOUNT_MANAGEMENT,
  MENU_RESOURCE_RESOURCE_LIST,
} from '@/constants/menu-symbol';

// 详情父路由与子路由共用同一份 meta：to.meta 由 matched 从父到子浅合并，子路由 new Meta 的默认值
// （auth: {}、owner / activeKey: Symbol.for('')、menu: {}）会整项盖掉父路由的，守卫读 to.meta.auth.view 时就丢了查看权限
const getDetailMeta = () =>
  new Meta({
    owner: MENU_RESOURCE,
    activeKey: MENU_RESOURCE_ACCOUNT_MANAGEMENT,
    auth: { view: { type: AUTH_FIND_ACCOUNT } },
    menu: { i18n: '账号详情', relative: MENU_RESOURCE_ACCOUNT_MANAGEMENT },
    layout: { breadcrumb: { show: true } },
  });

const accountManageRsc: RouteRecordRaw[] = [
  {
    name: MENU_RESOURCE_ACCOUNT_MANAGEMENT,
    path: 'account',
    component: () => import('@/views/account-manage/list.vue'),
    meta: {
      ...new Meta({
        owner: MENU_RESOURCE,
        activeKey: MENU_RESOURCE_ACCOUNT_MANAGEMENT,
        auth: { view: { type: AUTH_FIND_ACCOUNT } },
        menu: { i18n: '账号管理' },
        layout: { breadcrumb: { show: true }, contentWidth: 'fit' },
      }),
    },
  },
  {
    name: MENU_RESOURCE_ACCOUNT_CHECK_IN,
    path: 'account/check-in',
    component: () => import('@/views/service/service-apply/account-add/index'),
    meta: {
      ...new Meta({
        owner: MENU_RESOURCE,
        activeKey: MENU_RESOURCE_ACCOUNT_MANAGEMENT,
        menu: { i18n: '账号登记', relative: MENU_RESOURCE_ACCOUNT_MANAGEMENT },
        layout: { breadcrumb: { show: false } },
      }),
    },
  },
  {
    name: MENU_RESOURCE_ACCOUNT_DETAIL,
    path: 'account/details/:accountId',
    redirect: { name: MENU_RESOURCE_ACCOUNT_DETAIL_BASIC },
    component: () => import('@/views/account-manage/details/index'),
    meta: { ...getDetailMeta() },
    children: [
      {
        name: MENU_RESOURCE_ACCOUNT_DETAIL_BASIC,
        path: 'basic',
        component: () => import('@/views/account-manage/details/basic'),
        meta: { ...getDetailMeta() },
      },
    ],
  },
];

// query 值可能是数组，取第一个非空值
const pickQuery = (value: LocationQuery[string]) => {
  const val = Array.isArray(value) ? value.find(Boolean) : value;
  return val || undefined;
};

// 原「账号信息」页签与「资源状态」：带账号进资源纳管并打开账号状态侧滑，不带账号进账号列表
const toAccountStatus = (to: RouteLocationNormalized) => {
  const { accountId, id, isDetail, ...query } = to.query;
  const targetId = pickQuery(accountId) ?? pickQuery(id);
  if (!targetId) return { name: MENU_RESOURCE_ACCOUNT_MANAGEMENT, query };
  return {
    name: MENU_RESOURCE_RESOURCE_LIST,
    params: { resourceType: 'host' },
    query: { ...query, accountId: targetId, showAccountStatus: '1' },
  };
};

// 原「基本信息」「用户列表」与旧账号详情：进账号管理里该账号的详情
const toDetail = (to: RouteLocationNormalized) => {
  const { accountId, id, isDetail, ...query } = to.query;
  const targetId = pickQuery(accountId) ?? pickQuery(id);
  if (!targetId) return { name: MENU_RESOURCE_ACCOUNT_MANAGEMENT, query };
  return { name: MENU_RESOURCE_ACCOUNT_DETAIL, params: { accountId: targetId }, query };
};

const accountManageLegacy: RouteRecordRaw[] = [
  { path: '/resource/resource/account', redirect: toAccountStatus },
  { path: '/resource/resource/account/resource', redirect: toAccountStatus },
  { path: '/resource/resource/account/detail', redirect: toDetail },
  { path: '/resource/resource/account/manage', redirect: toDetail },
  { path: '/resource/account/detail', redirect: toDetail },
  { path: '/resource/account/add', redirect: { name: MENU_RESOURCE_ACCOUNT_CHECK_IN } },
  { path: '/service/service-apply/account-add', redirect: { name: MENU_RESOURCE_ACCOUNT_CHECK_IN } },
];

export { accountManageRsc, accountManageLegacy };
