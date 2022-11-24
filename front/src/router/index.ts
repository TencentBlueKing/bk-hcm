import {
  createRouter,
  RouteRecordRaw,
  NavigationGuardNext,
  createWebHashHistory,
  RouteLocationNormalized,
} from 'vue-router';
import common from './module/common';
import workbench from './module/workbench';
import resource from './module/resource';
import resourceInside from './module/resource-inside';
import service from './module/service';
import business from './module/business';

const routes: RouteRecordRaw[] = [
  ...common,
  ...workbench,
  ...resource,
  ...resourceInside,
  ...service,
  ...business,
  {
    // path: '/',
    // name: 'index',
    // component: () => import('@/views/resource/demo'),
    path: '/',
    redirect: '/resource/account',
    meta: {
      activeKey: 'resourceAccount',
      breadcrumb: ['云管', '账户'],
    },
  },
];

const router = createRouter({
  history: createWebHashHistory(),
  routes,
});

router.beforeEach((to: RouteLocationNormalized, from: RouteLocationNormalized, next: NavigationGuardNext) => {
  next();
});


export default router;
