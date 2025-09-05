import type { RouteRecordRaw } from 'vue-router';
import { ticketRoutes } from '@/views/ticket/route-config';

const serviceMenus: RouteRecordRaw[] = [
  {
    path: '/service',
    children: [...ticketRoutes],
    meta: {
      groupTitle: '资源',
    },
  },
];

export default serviceMenus;
