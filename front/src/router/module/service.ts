import type { RouteRecordRaw } from 'vue-router';
import i18n from '@/language/i18n';
import { ticketRoutes } from '@/views/ticket/route-config';

const { t } = i18n.global;

const serviceMenus: RouteRecordRaw[] = [
  {
    path: '/service',
    children: [
      ...ticketRoutes,
      {
        path: '/service/service-apply',
        name: 'serviceApply',
        component: () => import('@/views/service/service-apply/index.vue'),
        meta: {
          title: t('服务申请'),
          activeKey: 'serviceApply',
          // breadcrumb: [t('服务'), t('服务申请')],
          notMenu: true,
          isShowBreadcrumb: true,
        },
      },
      {
        path: '/service/my-approval',
        name: t('我的审批'),
        component: () => import('@/views/service/my-approval/page'),
        meta: {
          // breadcrumb: [t('服务'), t('我的审批')],
          isShowBreadcrumb: true,
          notMenu: true,
        },
      },
    ],
    meta: {
      groupTitle: '资源',
    },
  },
];

export default serviceMenus;
