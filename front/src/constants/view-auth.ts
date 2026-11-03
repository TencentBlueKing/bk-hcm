import type { RouteRecordName } from 'vue-router';
import type { IAuthSign } from '@/common/auth-service';
import {
  MENU_BILL_ROOT_ACCOUNT,
  MENU_BILL_MAIN_ACCOUNT,
  MENU_BILL_MANAGE,
  MENU_RESOURCE_ACCOUNT_MANAGEMENT,
  MENU_RESOURCE_OPERATION_LOG,
  MENU_RESOURCE_RECYCLEBIN,
  MENU_SERVICE_PERMISSION_POLICY,
} from './menu-symbol';
import {
  AUTH_ACCESS_BIZ,
  AUTH_BIZ_FIND_AUDIT,
  AUTH_FIND_ACCOUNT,
  AUTH_FIND_ROOT_ACCOUNT,
  AUTH_FIND_MAIN_ACCOUNT,
  AUTH_FIND_ACCOUNT_BILL,
  AUTH_FIND_PERMISSION_POLICY_LIBRARY,
  AUTH_FIND_RECYCLE_BIN,
} from './auth-symbols';

export interface IViewAuthConfig {
  // 路由 name，同时是侧栏菜单 id
  viewId: RouteRecordName;
  authSign: IAuthSign;
  // 无权限时侧栏隐藏同 id 的菜单项；页面是否拦截只看路由 meta.auth.view
  hideMenu?: boolean;
}

// 任一业务的访问权限，业务资源无权限页（noBiz）按它判定
export const VIEW_AUTH_BIZ_ACCESS = Symbol('view_auth_biz_access');

/**
 * 视图权限预鉴权配置
 * 在 preload 中合成一次请求批量鉴权；authSign 只用静态 sign，transform 不能返回数组（结果按下标对齐）
 */
export const viewAuthConfig: IViewAuthConfig[] = [
  { viewId: MENU_BILL_ROOT_ACCOUNT, authSign: { type: AUTH_FIND_ROOT_ACCOUNT }, hideMenu: true },
  { viewId: MENU_BILL_MAIN_ACCOUNT, authSign: { type: AUTH_FIND_MAIN_ACCOUNT } },
  { viewId: MENU_BILL_MANAGE, authSign: { type: AUTH_FIND_ACCOUNT_BILL }, hideMenu: true },
  { viewId: MENU_SERVICE_PERMISSION_POLICY, authSign: { type: AUTH_FIND_PERMISSION_POLICY_LIBRARY }, hideMenu: true },
  { viewId: MENU_RESOURCE_ACCOUNT_MANAGEMENT, authSign: { type: AUTH_FIND_ACCOUNT }, hideMenu: true },
  { viewId: MENU_RESOURCE_RECYCLEBIN, authSign: { type: AUTH_FIND_RECYCLE_BIN }, hideMenu: true },
  { viewId: MENU_RESOURCE_OPERATION_LOG, authSign: { type: AUTH_BIZ_FIND_AUDIT }, hideMenu: true },
  { viewId: VIEW_AUTH_BIZ_ACCESS, authSign: { type: AUTH_ACCESS_BIZ } },
];
