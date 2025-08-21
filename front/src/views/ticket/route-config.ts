import type { RouteRecordRaw } from 'vue-router';
import i18n from '@/language/i18n';

const { t } = i18n.global;

export const ticketRoutes: RouteRecordRaw[] = [
  // 兼容老路由
  {
    path: '/service/my-apply',
    redirect: '/service/ticket',
    // meta是必要的，如果不想在侧边栏显示，需要设置notMenu为true
    meta: {
      notMenu: true,
    },
  },
  {
    path: '/service/my-apply/detail',
    redirect: '/service/ticket/detail',
    // meta是必要的，如果不想在侧边栏显示，需要设置notMenu为true
    meta: {
      notMenu: true,
    },
  },
  {
    path: 'ticket',
    name: 'menu_ticket_manage',
    component: () => import('@/views/ticket/entry-srv.vue'),
    meta: {
      activeKey: 'menu_ticket_manage',
      title: t('单据管理'),
      // breadcrumb: [t('服务'), t('我的申请')],
      isShowBreadcrumb: true,
      icon: 'hcm-icon bkhcm-icon-my-apply',
    },
  },
  {
    path: 'ticket/detail',
    name: 'menu_ticket_detail',
    component: () => import('@/views/ticket/detail/apply-detail'),
    meta: {
      activeKey: 'menu_ticket_manage',
      notMenu: true,
    },
  },
];
