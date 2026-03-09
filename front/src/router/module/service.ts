/**
 * @deprecated 此文件已废弃，路由配置已迁移到 views 目录下的各模块 route-config.ts
 * 请使用 views/index.ts 中的 serviceViews
 */
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
