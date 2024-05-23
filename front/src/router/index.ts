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
import serviceInside from './module/service-inside';
import business from './module/business';
import scheme from './module/scheme';
import i18n from '@/language/i18n';
import { useCommonStore } from '@/store';
import { useVerify } from '@/hooks';

const { t } = i18n.global;

const routes: RouteRecordRaw[] = [
  ...common,
  ...workbench,
  ...resource,
  ...resourceInside,
  ...service,
  ...serviceInside,
  ...business,
  ...scheme,
  {
    // path: '/',
    // name: 'index',
    // component: () => import('@/views/resource/demo'),
    path: '/',
    redirect: '/business/host',
    meta: {
      activeKey: 'businessHost',
      breadcrumb: [t('计算'), t('主机')],
    },
  },
  {
    // path: '/',
    // name: 'index',
    // component: () => import('@/views/resource/demo'),
    path: '/403',
    redirect: '/403',
  },
];

const router = createRouter({
  history: createWebHashHistory(),
  routes,
});

// 进入目标页面
// eslint-disable-next-line max-len
const toCurrentPage = (
  authVerifyData: any,
  currentFindAuthData: any,
  next: NavigationGuardNext,
  to?: RouteLocationNormalized,
) => {
  if (currentFindAuthData) {
    // 当前页面需要鉴权
    if (authVerifyData && !authVerifyData?.permissionAction[currentFindAuthData.id]) {
      // 当前页面没有权限
      next({
        name: '403',
        params: {
          id: currentFindAuthData.id,
        },
      });
    } else if (/403/.test(to?.path)) {
      next({
        path: '/',
      });
    } else {
      next();
    }
  } else {
    if (to && to.name === '403' && authVerifyData && authVerifyData?.permissionAction?.account_find) {
      // 无权限用户切换到有权限用户时需要判断
      next({
        path: '/resource/account',
      });
    } else {
      next();
    }
  }
};

router.beforeEach((to: RouteLocationNormalized, from: RouteLocationNormalized, next: NavigationGuardNext) => {
  const commonStore = useCommonStore();
  const { pageAuthData, authVerifyData } = commonStore; // 所有需要检验的查看权限数据
  const currentFindAuthData = pageAuthData.find((e: any) => e.path === to.path || e?.path?.includes(to.path));

  // if (to.path === '/service/my-approval') {
  //   window.open(`${BK_ITSM_URL}/#/workbench/ticket/approval`);
  //   window.location.reload();
  // }
  if (from.path === '/') {
    // 刷新或者首次进入请求权限接口
    const { getAuthVerifyData } = useVerify(); // 权限中心权限
    getAuthVerifyData(pageAuthData).then(() => {
      const { authVerifyData } = commonStore;
      toCurrentPage(authVerifyData, currentFindAuthData, next, to);
    });
  } else if (['/scheme/recommendation', '/scheme/deployment/list'].includes(to.path)) {
    next();
  } else {
    toCurrentPage(authVerifyData, currentFindAuthData, next);
  }
});

export default router;
