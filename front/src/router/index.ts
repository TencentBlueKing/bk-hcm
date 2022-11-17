import {
  createRouter,
  RouteRecordRaw,
  NavigationGuardNext,
  createWebHashHistory,
  RouteLocationNormalized,
} from 'vue-router';
import common from './module/common';
import work from './module/work';
import cost from './module/cost';
import resources from './module/resources';
import services from './module/services';

const routes: RouteRecordRaw[] = [
  ...common,
  ...work,
  ...cost,
  ...resources,
  ...services,
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
