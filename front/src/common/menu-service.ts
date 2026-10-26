import type { RouteLocationNormalizedLoaded, RouteRecordRaw } from 'vue-router';
import i18n from '@/language/i18n';
import {
  MENU_BUSINESS,
  MENU_BUSINESS_CERT_MANAGEMENT,
  MENU_BUSINESS_CLOUD_ACCOUNT,
  MENU_BUSINESS_DISK_MANAGEMENT,
  MENU_BUSINESS_EIP_MANAGEMENT,
  MENU_BUSINESS_HOST_MANAGEMENT,
  MENU_BUSINESS_LOAD_BALANCER,
  MENU_BUSINESS_NETWORK_INTERFACE_MANAGEMENT,
  MENU_BUSINESS_OPERATION_LOG,
  MENU_BUSINESS_RECYCLEBIN,
  MENU_BUSINESS_ROUTEING_TABLE_MANAGEMENT,
  MENU_BUSINESS_SECURITY_GROUP_MANAGEMENT,
  MENU_BUSINESS_SUBNET_MANAGEMENT,
  MENU_BUSINESS_TASK_MANAGEMENT,
  MENU_BUSINESS_TICKET_MANAGEMENT,
  MENU_BUSINESS_VPC_MANAGEMENT,
  MENU_SERVICE,
  MENU_SERVICE_PERMISSION_POLICY,
  MENU_SERVICE_REQUEST,
  MENU_SERVICE_TICKET_MANAGEMENT,
  MENU_RESOURCE,
  MENU_RESOURCE_ACCESS,
  MENU_RESOURCE_ACCOUNT_MANAGEMENT,
  MENU_RESOURCE_MANAGE,
  MENU_RESOURCE_OPERATION_LOG,
  MENU_RESOURCE_RECYCLEBIN,
  MENU_SCHEME,
  MENU_BILL_ROOT_ACCOUNT,
  MENU_BILL_MAIN_ACCOUNT,
  MENU_BILL_MANAGE,
} from '@/constants/menu-symbol';
import { businessViews, resourceViews, serviceViews } from '@/views';

/**
 * 菜单定义，与路由解耦：菜单只按 name 引用路由，显示名、图标、分组、显隐归菜单。
 * 一级菜单 id 即顶层路由 name，页面所属一级菜单按 route.matched[0].name 判定。
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

const { ENABLE_CLOUD_SELECTION, ENABLE_ACCOUNT_BILL } = window.PROJECT_CONFIG;
const { t } = i18n.global;

const findRoute = (views: RouteRecordRaw[], id: symbol | string): RouteRecordRaw | undefined => {
  for (const view of views) {
    if (view.name === id) return view;
    const child = view.children?.length ? findRoute(view.children, id) : undefined;
    if (child) return child;
  }
  return undefined;
};

export const getMenuRoute = (views: RouteRecordRaw[], id: symbol | string) => {
  const view = findRoute(views, id);
  return view ? { name: view.name, path: view.path } : { name: '', path: '' };
};

export const toMenuKey = (id: symbol | string) => (typeof id === 'symbol' ? id.description : id);

const businessMenus: IMenu[] = [
  {
    id: MENU_BUSINESS_HOST_MANAGEMENT,
    i18n: '主机',
    group: '云资源',
    groupIcon: 'bkhcm-icon-vpc',
    route: getMenuRoute(businessViews, MENU_BUSINESS_HOST_MANAGEMENT),
  },
  {
    id: MENU_BUSINESS_DISK_MANAGEMENT,
    i18n: '硬盘',
    group: '云资源',
    route: getMenuRoute(businessViews, MENU_BUSINESS_DISK_MANAGEMENT),
  },
  {
    id: MENU_BUSINESS_VPC_MANAGEMENT,
    i18n: 'VPC',
    group: '云资源',
    route: getMenuRoute(businessViews, MENU_BUSINESS_VPC_MANAGEMENT),
  },
  {
    id: MENU_BUSINESS_SUBNET_MANAGEMENT,
    i18n: '子网',
    group: '云资源',
    route: getMenuRoute(businessViews, MENU_BUSINESS_SUBNET_MANAGEMENT),
  },
  {
    id: MENU_BUSINESS_EIP_MANAGEMENT,
    i18n: '弹性IP',
    group: '云资源',
    route: getMenuRoute(businessViews, MENU_BUSINESS_EIP_MANAGEMENT),
  },
  {
    id: MENU_BUSINESS_NETWORK_INTERFACE_MANAGEMENT,
    i18n: '网络接口',
    group: '云资源',
    route: getMenuRoute(businessViews, MENU_BUSINESS_NETWORK_INTERFACE_MANAGEMENT),
  },
  {
    id: MENU_BUSINESS_ROUTEING_TABLE_MANAGEMENT,
    i18n: '路由表',
    group: '云资源',
    route: getMenuRoute(businessViews, MENU_BUSINESS_ROUTEING_TABLE_MANAGEMENT),
  },
  {
    id: MENU_BUSINESS_SECURITY_GROUP_MANAGEMENT,
    i18n: '安全组',
    group: '云资源',
    route: getMenuRoute(businessViews, MENU_BUSINESS_SECURITY_GROUP_MANAGEMENT),
  },
  {
    id: MENU_BUSINESS_LOAD_BALANCER,
    i18n: '负载均衡',
    group: '云资源',
    route: getMenuRoute(businessViews, MENU_BUSINESS_LOAD_BALANCER),
  },
  {
    id: MENU_BUSINESS_CERT_MANAGEMENT,
    i18n: '证书托管',
    group: '云资源',
    route: getMenuRoute(businessViews, MENU_BUSINESS_CERT_MANAGEMENT),
  },
  {
    id: MENU_BUSINESS_TICKET_MANAGEMENT,
    i18n: t('单据管理'),
    group: '操作管理',
    groupIcon: 'bkhcm-icon-operation-record',
    route: getMenuRoute(businessViews, MENU_BUSINESS_TICKET_MANAGEMENT),
  },
  {
    id: MENU_BUSINESS_TASK_MANAGEMENT,
    i18n: '任务管理',
    group: '操作管理',
    route: getMenuRoute(businessViews, MENU_BUSINESS_TASK_MANAGEMENT),
  },
  {
    id: MENU_BUSINESS_OPERATION_LOG,
    i18n: '操作记录',
    group: '操作管理',
    route: getMenuRoute(businessViews, MENU_BUSINESS_OPERATION_LOG),
  },
  {
    id: MENU_BUSINESS_CLOUD_ACCOUNT,
    i18n: '云账号管理',
    icon: 'bkhcm-icon-account-manage',
    route: getMenuRoute(businessViews, MENU_BUSINESS_CLOUD_ACCOUNT),
  },
  {
    id: MENU_BUSINESS_RECYCLEBIN,
    i18n: '回收站',
    icon: 'bkhcm-icon-recyclebin',
    route: getMenuRoute(businessViews, MENU_BUSINESS_RECYCLEBIN),
  },
];

// 工作台复用 service 视图：分组对齐 PR #1897，个人事务放单据，公共事务放机房裁撤 / 资源选型 / 账号管理 / 权限策略库
const serviceMenus: IMenu[] = [
  {
    id: MENU_SERVICE_TICKET_MANAGEMENT,
    i18n: t('单据管理'),
    group: '个人事务',
    groupIcon: 'bkhcm-icon-user-8',
    route: getMenuRoute(serviceViews, MENU_SERVICE_TICKET_MANAGEMENT),
  },
  {
    id: MENU_SCHEME,
    i18n: '资源选型',
    group: '公共事务',
    groupIcon: 'bkhcm-icon-user-8',
    visibility: () => ENABLE_CLOUD_SELECTION === 'true',
    route: getMenuRoute(serviceViews, MENU_SCHEME),
  },
  // 权限策略库过渡放在这里，四期随服务请求一起迁
  {
    id: MENU_SERVICE_PERMISSION_POLICY,
    i18n: t('权限策略库'),
    group: '公共事务',
    route: getMenuRoute(serviceViews, MENU_SERVICE_PERMISSION_POLICY),
  },
];

const isAccountBillEnabled = () => ENABLE_ACCOUNT_BILL === 'true';

const resourceMenus: IMenu[] = [
  {
    id: MENU_RESOURCE_MANAGE,
    i18n: '资源纳管',
    group: '云资源',
    groupIcon: 'bkhcm-icon-vpc',
    route: getMenuRoute(resourceViews, MENU_RESOURCE_MANAGE),
  },
  // 过渡位置：三期迁至工作台
  {
    id: MENU_RESOURCE_ACCOUNT_MANAGEMENT,
    i18n: '账号管理',
    group: '云资源',
    route: getMenuRoute(resourceViews, MENU_RESOURCE_ACCOUNT_MANAGEMENT),
  },
  {
    id: MENU_RESOURCE_RECYCLEBIN,
    i18n: '回收管理',
    group: '云资源',
    route: getMenuRoute(resourceViews, MENU_RESOURCE_RECYCLEBIN),
  },
  {
    id: MENU_BILL_ROOT_ACCOUNT,
    i18n: '一级账号',
    group: '云账单管理',
    groupIcon: 'bkhcm-icon-account-manage',
    visibility: isAccountBillEnabled,
    route: getMenuRoute(resourceViews, MENU_BILL_ROOT_ACCOUNT),
  },
  {
    id: MENU_BILL_MAIN_ACCOUNT,
    i18n: '二级账号',
    group: '云账单管理',
    visibility: isAccountBillEnabled,
    route: getMenuRoute(resourceViews, MENU_BILL_MAIN_ACCOUNT),
  },
  {
    id: MENU_BILL_MANAGE,
    i18n: '账单管理',
    group: '云账单管理',
    visibility: isAccountBillEnabled,
    route: getMenuRoute(resourceViews, MENU_BILL_MANAGE),
  },
  {
    id: MENU_RESOURCE_OPERATION_LOG,
    i18n: '操作记录',
    icon: 'bkhcm-icon-operation-record',
    route: getMenuRoute(resourceViews, MENU_RESOURCE_OPERATION_LOG),
  },
];

// 不声明 menu 的一级菜单不显示侧栏
const menus: IMenu[] = [
  { id: MENU_BUSINESS, i18n: '业务资源', menu: businessMenus },
  { id: MENU_SERVICE, i18n: '工作台', menu: serviceMenus },
  // 服务请求只做跳转（跳工作台的单据管理），自身没有侧栏，四期连入口一并删除
  {
    id: MENU_SERVICE_REQUEST,
    i18n: '服务请求',
    route: getMenuRoute(serviceViews, MENU_SERVICE_TICKET_MANAGEMENT),
  },
  // 只做跳转，资源接入下线时删除
  { id: MENU_RESOURCE_ACCESS, i18n: '资源接入', route: getMenuRoute(resourceViews, MENU_RESOURCE_MANAGE) },
  { id: MENU_RESOURCE, i18n: '资源运营', menu: resourceMenus },
];

const filter = (items: IMenu[]): IMenu[] =>
  items
    .filter((item) => {
      if (!Object.prototype.hasOwnProperty.call(item, 'visibility')) return true;
      return typeof item.visibility === 'function' ? item.visibility() : !!item.visibility;
    })
    .map((item) => (item.menu?.length ? { ...item, menu: filter(item.menu) } : item));

/** 顶栏：一级菜单按 visibility 过滤 */
export const getMenus = () => filter(menus);

/**
 * 当前页面所属的一级菜单（顶栏激活与侧栏）。一级菜单本身不按 visibility 过滤：
 * 开关只作用于顶栏入口，直接访问地址时侧栏照常显示。
 */
export const getCurrentTopMenu = (route: RouteLocationNormalizedLoaded) => {
  const top = menus.find((item) => item.id === route.matched[0]?.name);
  return top && { ...top, menu: filter(top.menu ?? []) };
};
