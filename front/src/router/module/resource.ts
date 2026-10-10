// import { CogShape } from 'bkui-vue/lib/icon';
import type { RouteRecordRaw } from 'vue-router';
import i18n from '@/language/i18n';
import { operationLogRsc as operationLogRscRouteConfig } from '@/views/operation-log/route-config';
import { toResourceList } from '@/views/resource-manage/route-config';
import {
  MENU_RESOURCE_MANAGE,
  MENU_RESOURCE_RESOURCE_LIST,
  MENU_RESOURCE_RESOURCE_MANAGEMENT,
} from '@/constants/menu-symbol';
import { AUTH_FIND_ACCOUNT, AUTH_FIND_RECYCLE_BIN } from '@/constants/auth-symbols';

const { t } = i18n.global;

// 原资源接入页面，只承接旧地址（侧栏无入口），随账号管理、回收管理与操作记录、资源详情迁出后逐条删除
const resourceMenus: RouteRecordRaw[] = [
  {
    path: '/resource',
    component: () => import('@/views/resource/resource-manage/resource-entry.vue'),
    // 与资源运营根同路径，vue-router 优先匹配子记录，不转址时裸 /resource 会渲染空白外壳
    redirect: { name: MENU_RESOURCE_MANAGE },
    meta: { legacyResourceAccess: true },
    children: [
      {
        path: 'resource',
        name: MENU_RESOURCE_RESOURCE_MANAGEMENT,
        component: () => import('@/views/resource/resource-manage/resource-manage.vue'),
        children: [
          // 资源列表已迁到资源纳管
          { path: '', redirect: toResourceList },
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
        component: () => import('@/views/resource/accountmanage/index.vue'),
        meta: {
          action: 'account_find',
          auth: { view: { type: AUTH_FIND_ACCOUNT } },
        },
      },
      {
        path: '/resource/detail/:type',
        name: 'resourceDetail',
        component: () => import('@/views/resource/resource-manage/resource-detail.vue'),
        meta: {
          notMenu: true,
          menu: { i18n: '资源详情', relative: MENU_RESOURCE_RESOURCE_LIST },
        },
      },
      operationLogRscRouteConfig[1],
      {
        path: '/resource/recyclebin',
        name: t('回收站'),
        component: () => import('@/views/resource/recyclebin-manager/recyclebin-manager.vue'),
        meta: {
          auth: { view: { type: AUTH_FIND_RECYCLE_BIN } },
        },
      },
    ],
  },
];

export default resourceMenus;
