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
import i18n from '@/language/i18n';
import http from '@/http';

import { useCommonStore } from '@/store';
const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

const { t } = i18n.global;

const authPagePath = ['/resource/account'];   // 需要鉴权的列表页
const action = [{ type: 'account', action: 'find', id: 'account_find' }];

// 修改符合格式的参数
const getParams = (action: any) => {
  const params = action?.reduce((p: any, v: any) => {
    p.resources.push({
      action: v.action,
      resource_type: v.type,
    });
    return p;
  }, { resources: [] });
  return params;
};


// 管理全局列表权限变量
const addAuthVerifyListParams = (data: any) => {
  const commonStore = useCommonStore();
  commonStore.addAuthVerifyParams(data);
};


const routes: RouteRecordRaw[] = [
  ...common,
  ...workbench,
  ...resource,
  ...resourceInside,
  ...service,
  ...serviceInside,
  ...business,
  {
    // path: '/',
    // name: 'index',
    // component: () => import('@/views/resource/demo'),
    path: '/',
    redirect: '/resource/account',
    meta: {
      activeKey: 'resourceAccount',
      breadcrumb: [t('云管'), t('账号')],
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

// const showPermissionDialog = ref(false);    // 无权限弹窗

// const {
//   authVerifyData,
// } = useVerify(showPermissionDialog, ['find']);

router.beforeEach(async (to: RouteLocationNormalized, from: RouteLocationNormalized, next: NavigationGuardNext) => {
  if (authPagePath.includes(to.path)) {
    const params = getParams(action);
    const res = await http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/web/auth/verify`, params);
    const { permission } = res.data;
    if (permission) {   // 没有权限
      const actionItem = permission.actions.filter((e: any) => e.id === to.meta.action);
      const routerParams = {
        system_id: permission.system_id,
        actions: actionItem,
      };
      addAuthVerifyListParams(routerParams);
      next({
        path: '/403',
      });
    } else {
      next();
    }
  } else {
    next();
  }
});


export default router;
