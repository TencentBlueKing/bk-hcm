// import { CogShape } from 'bkui-vue/lib/icon';
import type { RouteRecordRaw } from 'vue-router';
import i18n from '@/language/i18n';

const { t } = i18n.global;

const workbenchMenus: RouteRecordRaw[] = [
  // {
  //   path: '/workbench',
  //   name: '配置',
  //   children: [
  //     {
  //       path: '/workbench/auto',
  //       name: 'agent自动化',
  //       alias: '',
  //       component: () => import('@/views/workbench/demo2'),
  //       meta: {
  //         activeKey: 'workbenchAuto',
  //       },
  //     },
  //   ],
  // },
  {
    path: '/workbench/audit',
    name: '审计',
    component: () => import('@/views/workbench/audit'),
    meta: {
      activeKey: 'workbenchAudit',
      breadcrumb: [t('工作台'), t('审计')],
    },
  },
];

export default workbenchMenus;
