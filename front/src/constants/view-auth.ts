import type { IAuthSign } from '@/common/auth-service';
import { MENU_SCHEME_RECOMMENDATION } from './menu-symbol';
import { AUTH_CREATE_CLOUD_SELECTION_SCHEME } from './auth-symbols';

export interface IViewAuthConfig {
  viewId: symbol;
  authSign: IAuthSign;
}

/**
 * 资源视图权限配置
 * 在 preload 中统一获取，不需要 bizId
 *
 * TODO: 后续补充完整的视图权限配置
 * 示例：
 * { viewId: MENU_RESOURCE_ACCOUNT, authSign: { type: AUTH_FIND_ACCOUNT } },
 * { viewId: MENU_RESOURCE_RECYCLEBIN, authSign: { type: AUTH_FIND_RECYCLE_BIN } },
 */
export const viewAuthConfig: IViewAuthConfig[] = [
  { viewId: MENU_SCHEME_RECOMMENDATION, authSign: { type: AUTH_CREATE_CLOUD_SELECTION_SCHEME } },
];
