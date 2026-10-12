// import { CogShape } from 'bkui-vue/lib/icon';
import type { RouteRecordRaw } from 'vue-router';
import { toResourceList } from '@/views/resource-manage/route-config';
import { MENU_RESOURCE_MANAGE, MENU_RESOURCE_RESOURCE_MANAGEMENT } from '@/constants/menu-symbol';

// 原资源接入页面，只承接旧地址（侧栏无入口），随资源接入下线删除
const resourceMenus: RouteRecordRaw[] = [
  {
    path: '/resource',
    component: () => import('@/views/resource/resource-manage/resource-entry.vue'),
    // 与资源运营根同路径，vue-router 优先匹配子记录，不转址时裸 /resource 会渲染空白外壳
    redirect: { name: MENU_RESOURCE_MANAGE },
    meta: { legacyResourceAccess: true },
    children: [
      {
        path: 'resource',
        name: MENU_RESOURCE_RESOURCE_MANAGEMENT,
        component: () => import('@/views/resource/resource-manage/resource-manage.vue'),
        children: [
          // 资源列表已迁到资源纳管
          { path: '', redirect: toResourceList },
        ],
        meta: {
          notMenu: true,
        },
      },
    ],
  },
];

export default resourceMenus;
