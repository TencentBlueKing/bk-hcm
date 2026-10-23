import type { RouteRecordRaw } from 'vue-router';
import i18n from '@/language/i18n';
import Meta from '@/router/meta';
import { MENU_SERVICE_PERMISSION_POLICY } from '@/constants/menu-symbol';
import { AUTH_FIND_PERMISSION_POLICY_LIBRARY } from '@/constants/auth-symbols';

const { t } = i18n.global;

export const permissionPolicyRoutes: RouteRecordRaw[] = [
  {
    path: 'permission-policy',
    name: MENU_SERVICE_PERMISSION_POLICY,
    component: () => import('@/views/cloud-account-manage/permission-policy/index.vue'),
    meta: {
      ...new Meta({
        activeKey: MENU_SERVICE_PERMISSION_POLICY,
        menu: { i18n: t('权限策略库') },
        auth: { view: { type: AUTH_FIND_PERMISSION_POLICY_LIBRARY } },
      }),
    },
  },
];
