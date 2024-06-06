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
      notMenu: true,
      isShowBreadcrumb: true,
    },
  },
  {
    path: '/service/my-apply',
    name: t('我的申请'),
    component: () => import('@/views/service/apply-list/index'),
    meta: {
      activeKey: 'myApply',
      breadcrumb: [t('服务'), t('我的申请')],
      isShowBreadcrumb: true,
    },
  },
  {
    path: '/service/my-apply/detail',
    name: '申请单据详情',
    component: () => import('@/views/service/apply-detail/index'),
    meta: {
      activeKey: 'myApply',
      notMenu: true,
    },
  },
  {
    path: '/service/my-approval',
    name: t('我的审批'),
    component: () => import('@/views/service/my-approval/page'),
    meta: {
      breadcrumb: [t('服务'), t('我的审批')],
      isShowBreadcrumb: true,
      notMenu: true,
    },
  },
];

export default serviceMenus;
