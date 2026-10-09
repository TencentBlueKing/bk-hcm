interface Menu {
  i18n?: string;
  parent?: any;
  relative?: string | symbol;
}

interface AuthView {
  type: string;
  relation?: number[] | string[];
}

interface Auth {
  superView?: any;
  view?: AuthView | (() => AuthView);
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
  notMenu?: boolean;
  activeKey?: symbol | string;
  icon?: string;
  checkAuth?: string;
}

/**
 * layout、activeKey、owner 只在传入时出现在 meta 上：route.meta 是 matched 链上各记录 meta 的浅合并，
 * 子记录若产出默认值会遮住父记录的配置（如父记录的 layout.breadcrumb.show = false）。
 */
export default class Meta {
  available = true;

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
