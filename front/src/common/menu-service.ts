import type { RouteRecordRaw } from 'vue-router';
import { MENU_BILL_ROOT_ACCOUNT, MENU_BILL_MAIN_ACCOUNT, MENU_BILL_MANAGE } from '@/constants/menu-symbol';
import { billViews } from '@/views';

/**
 * 菜单定义，与路由解耦：菜单只按 name 引用路由，显示名、图标、分组、显隐归菜单。
 * 目前只收录一级菜单「资源运营」的侧栏，其它一级视图的菜单仍由路由结构（groupTitle + children）表达。
 */
export interface IMenu {
  id: symbol | string;
  i18n: string;
  icon?: string;
  groupIcon?: string;
  group?: string;
  route?: {
    name?: symbol | string;
    path: string;
  };
  menu?: IMenu[];
  visibility?: boolean | (() => boolean);
}

export const getMenuRoute = (views: RouteRecordRaw[], id: symbol | string) => {
  const view = views.find((item) => item.name === id);
  return view ? { name: view.name, path: view.path } : { name: '', path: '' };
};

const filterVisible = (items: IMenu[]): IMenu[] =>
  items
    .filter((item) => {
      if (!Object.prototype.hasOwnProperty.call(item, 'visibility')) return true;
      return typeof item.visibility === 'function' ? item.visibility() : !!item.visibility;
    })
    .map((item) => (item.menu?.length ? { ...item, menu: filterVisible(item.menu) } : item));

const billMenus: IMenu[] = [
  {
    id: MENU_BILL_ROOT_ACCOUNT,
    i18n: '一级账号',
    icon: 'bkhcm-icon-account-manage',
    group: '云账单管理',
    route: getMenuRoute(billViews, MENU_BILL_ROOT_ACCOUNT),
  },
  {
    id: MENU_BILL_MAIN_ACCOUNT,
    i18n: '二级账号',
    icon: 'bkhcm-icon-account-manage',
    group: '云账单管理',
    route: getMenuRoute(billViews, MENU_BILL_MAIN_ACCOUNT),
  },
  {
    id: MENU_BILL_MANAGE,
    i18n: '云账单管理',
    icon: 'bkhcm-icon-bill-manage',
    group: '云账单管理',
    route: getMenuRoute(billViews, MENU_BILL_MANAGE),
  },
];

export const getBillMenus = () => filterVisible(billMenus);
