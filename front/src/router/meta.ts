import type { RouteLocationNormalized } from 'vue-router';
import type { IAuthSign, IPermission } from '@/common/auth-service';

interface Menu {
  i18n?: string;
  parent?: any;
  relative?: string | symbol;
}

// 视图权限配置函数，接收完整的路由信息
export type AuthViewFn = (to: RouteLocationNormalized) => IAuthSign;

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
  available?: boolean;
  owner?: symbol;
  /** @deprecated 使用 menu.i18n 代替 */
  title?: string;
  authKey?: string;
  view?: string;
  extra?: Extra;
  menu?: Menu;
  auth?: Auth;
  layout?: Layout;
  activeKey?: symbol;
  /** @deprecated 使用 layout.breadcrumb.show 代替 */
  isShowBreadcrumb?: boolean;
  permissionData?: IPermission;
}

export default class Meta {
  available = true;

  owner = Symbol.for('');

  authKey = 'view';

  view = 'default';

  extra: Extra = {};

  menu: Menu = {};

  auth: Auth = {};

  layout: Layout = {
    breadcrumb: {
      show: true,
      back: true,
    },
  };

  activeKey = Symbol.for('');

  permissionData: IPermission = null;

  constructor(data: RouteMetaConfig) {
    // 设置非嵌套对象属性
    this.available = data.available ?? this.available;
    this.owner = data.owner ?? this.owner;
    this.authKey = data.authKey ?? this.authKey;
    this.view = data.view ?? this.view;
    this.activeKey = data.activeKey ?? this.activeKey;
    this.permissionData = data.permissionData ?? this.permissionData;

    // 合并嵌套对象（保留默认值，只覆盖传入的属性）
    this.menu = Object.assign(this.menu, data.menu);

    this.auth = Object.assign(this.auth, data.auth);

    this.layout = Object.assign(this.layout, data.layout);

    this.extra = Object.assign(this.extra, data.extra);
  }
}
