import type { RouteRecordRaw } from 'vue-router';
import i18n from '@/language/i18n';

const { t } = i18n.global;


const serviceMenus: RouteRecordRaw[] = [
  {
    path: '/service/service-apply',
    name: t('服务申请'),
    component: () => import('@/views/service/service-apply/index.vue'),
    meta: {
      activeKey: 'serviceApply',
      breadcrumb: [t('服务'), t('服务申请')],
    },
  },
  {
    path: '/service/my-apply',
    name: t('我的申请'),
    component: () => import('@/views/service/my-apply/index.vue'),
    meta: {
      activeKey: 'myApply',
      breadcrumb: [t('服务'), t('我的申请')],
    },
  },
  {
    path: '/service/my-approval',
    name: t('我的审批'),
    component: () => import('@/views/service/my-approval/index.vue'),
    meta: {
      breadcrumb: [t('服务'), t('我的审批')],
    },
  },
];

export default serviceMenus;
