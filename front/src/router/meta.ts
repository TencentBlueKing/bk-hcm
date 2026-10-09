import type { RouteLocationNormalized } from 'vue-router';
import type { IAuthSign, IPermission } from '@/common/auth-service';

interface Menu {
  i18n?: string;
  parent?: any;
  relative?: string | symbol;
}

// 视图权限配置函数，接收完整的路由信息
export type AuthViewFn = (to: RouteLocationNormalized) => IAuthSign;

// available 支持静态布尔值或动态函数
export type AvailableFn = (to: RouteLocationNormalized) => boolean;

interface Auth {
  superView?: any;
  view?: IAuthSign | AuthViewFn;
  operation?: any;
  permission?: any;
}

interface Layout {
  breadcrumb?: {
    show?: boolean;
    back?: boolean;
  };
}

interface Extra {
  [key: string]: any;
}

export interface RouteMetaConfig {
  available?: boolean | AvailableFn;
  owner?: symbol;
  /** @deprecated 使用 menu.i18n 代替 */
  title?: string;
  authKey?: string;
  view?: string;
  extra?: Extra;
  menu?: Menu;
  auth?: Auth;
  layout?: Layout;
  notMenu?: boolean;
  activeKey?: symbol | string;
  icon?: string;
  permissionData?: IPermission;
  errorMessage?: string;
}

/**
 * layout、activeKey、owner 只在传入时出现在 meta 上：route.meta 是 matched 链上各记录 meta 的浅合并，
 * 子记录若产出默认值会遮住父记录的配置（如父记录的 layout.breadcrumb.show = false）。
 */
export default class Meta {
  available: boolean | AvailableFn = true;

  title = '';

  authKey = 'view';

  view = 'default';

  extra: Extra = {};

  menu: Menu = {};

  auth: Auth = {};

  constructor(data: RouteMetaConfig) {
    Object.keys(data).forEach((key) => {
      const typedKey = key as keyof RouteMetaConfig;
      Reflect.set(this, typedKey, data[typedKey]);
    });

    this.menu.i18n = this.title;

    this.menu = Object.assign(this.menu, data.menu);

    this.auth = Object.assign(this.auth, data.auth);

    this.extra = Object.assign(this.extra, data.extra);
  }
}
