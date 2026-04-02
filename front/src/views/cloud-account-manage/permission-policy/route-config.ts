import type { RouteRecordRaw } from 'vue-router';
import i18n from '@/language/i18n';
import Meta from '@/router/meta';
import { MENU_SERVICE_PERMISSION_POLICY } from '@/constants/menu-symbol';

const { t } = i18n.global;

export const permissionPolicyRoutes: RouteRecordRaw[] = [
  {
    path: 'permission-policy',
    name: MENU_SERVICE_PERMISSION_POLICY,
    component: () => import('@/views/cloud-account-manage/permission-policy/index.vue'),
    meta: {
      ...new Meta({
        activeKey: MENU_SERVICE_PERMISSION_POLICY,
        title: t('权限策略库'),
        isShowBreadcrumb: true,
        icon: 'hcm-icon bkhcm-icon-my-apply',
        checkAuth: 'permission_policy_library',
      }),
    },
  },
];
