import type { RouteRecordRaw } from 'vue-router';
import { ticketRoutes } from '@/views/ticket/route-config';
import { permissionPolicyRoutes } from '@/views/cloud-account-manage/permission-policy/route-config';

const serviceMenus: RouteRecordRaw[] = [
  {
    path: '/service',
    children: [...ticketRoutes, ...permissionPolicyRoutes],
    meta: {
      groupTitle: '资源',
    },
  },
];

export default serviceMenus;
