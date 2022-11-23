import {
  createRouter,
  RouteRecordRaw,
  NavigationGuardNext,
  createWebHashHistory,
  RouteLocationNormalized,
} from 'vue-router';
import common from './module/common';
import workbench from './module/workbench';
import cost from './module/cost';
import resource from './module/resource';
import service from './module/service';
import business from './module/business';

const routes: RouteRecordRaw[] = [
  ...common,
  ...workbench,
  ...cost,
  ...resource,
  ...service,
  ...business,
  {
    path: '/',
    name: 'index',
    component: () => import('@/views/resource/demo'),
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
