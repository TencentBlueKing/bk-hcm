import type { IAuthSign } from '@/common/auth-service';
import {
  MENU_RESOURCE_MANAGE,
  MENU_RESOURCE_RECYCLEBIN,
  MENU_RESOURCE_OPERATION_LOG,
  MENU_SCHEME_RECOMMENDATION,
  MENU_SCHEME_LIST,
  MENU_SERVICE_ACCOUNT_MANAGE,
  MENU_BILL_ROOT_ACCOUNT,
  MENU_BILL_MAIN_ACCOUNT,
  MENU_BILL_MANAGE,
} from './menu-symbol';
import {
  AUTH_FIND_IAAS_RESOURCE,
  AUTH_FIND_RECYCLE_BIN,
  AUTH_BIZ_FIND_AUDIT,
  AUTH_CREATE_CLOUD_SELECTION_SCHEME,
  AUTH_FIND_CLOUD_SELECTION_SCHEME,
  AUTH_FIND_ACCOUNT,
  AUTH_FIND_ROOT_ACCOUNT,
  AUTH_FIND_MAIN_ACCOUNT,
  AUTH_FIND_ACCOUNT_BILL,
} from './auth-symbols';

export interface IViewAuthConfig {
  viewId: symbol;
  authSign: IAuthSign;
}

/**
 * 资源/工作台视图权限配置
 * 在 preload 中统一批量鉴权，不需要 bizId
 */
export const viewAuthConfig: IViewAuthConfig[] = [
  { viewId: MENU_RESOURCE_MANAGE, authSign: { type: AUTH_FIND_IAAS_RESOURCE } },
  { viewId: MENU_RESOURCE_RECYCLEBIN, authSign: { type: AUTH_FIND_RECYCLE_BIN } },
  { viewId: MENU_RESOURCE_OPERATION_LOG, authSign: { type: AUTH_BIZ_FIND_AUDIT } },
  { viewId: MENU_SCHEME_RECOMMENDATION, authSign: { type: AUTH_CREATE_CLOUD_SELECTION_SCHEME } },
  { viewId: MENU_SCHEME_LIST, authSign: { type: AUTH_FIND_CLOUD_SELECTION_SCHEME } },
  { viewId: MENU_SERVICE_ACCOUNT_MANAGE, authSign: { type: AUTH_FIND_ACCOUNT } },
  { viewId: MENU_BILL_ROOT_ACCOUNT, authSign: { type: AUTH_FIND_ROOT_ACCOUNT } },
  { viewId: MENU_BILL_MAIN_ACCOUNT, authSign: { type: AUTH_FIND_MAIN_ACCOUNT } },
  { viewId: MENU_BILL_MANAGE, authSign: { type: AUTH_FIND_ACCOUNT_BILL } },
];
