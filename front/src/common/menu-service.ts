import type { RouteRecordRaw } from 'vue-router';
import {
  MENU_BUSINESS,
  MENU_SERVICE,
  MENU_RESOURCE_MANAGE,
  MENU_RESOURCE,
  MENU_RESOURCE_RECYCLEBIN,
  MENU_RESOURCE_OPERATION_LOG,
  MENU_SCHEME,
  MENU_BUSINESS_HOST_MANAGEMENT,
  MENU_BUSINESS_DISK_MANAGEMENT,
  MENU_BUSINESS_VPC_MANAGEMENT,
  MENU_BUSINESS_SUBNET_MANAGEMENT,
  MENU_BUSINESS_EIP_MANAGEMENT,
  MENU_BUSINESS_NETWORK_INTERFACE_MANAGEMENT,
  MENU_BUSINESS_ROUTEING_TABLE_MANAGEMENT,
  MENU_BUSINESS_SECURITY_GROUP_MANAGEMENT,
  MENU_BUSINESS_LOAD_BALANCER,
  MENU_BUSINESS_CERT_MANAGEMENT,
  MENU_BUSINESS_OPERATION_LOG,
  MENU_BUSINESS_TASK_MANAGEMENT,
  MENU_BUSINESS_RECYCLEBIN,
  MENU_SERVICE_APPLY_MANAGEMENT,
  MENU_SERVICE_ACCOUNT_MANAGE,
  MENU_BILL_ROOT_ACCOUNT,
  MENU_BILL_MAIN_ACCOUNT,
  MENU_BILL_MANAGE,
} from '../constants/menu-symbol';
import { businessViews, serviceViews, resourceViews } from '@/views';

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

const getMenuRoute = (views: RouteRecordRaw[], symbol: symbol | string) => {
  const menuView = Array.isArray(views) ? views.find((view) => view.name === symbol) : views;
  if (menuView) {
    return {
      name: menuView.name,
      path: menuView.path,
    };
  }
  return { name: '', path: '' };
};

const menus: IMenu[] = [
  {
    id: MENU_BUSINESS,
    i18n: '业务资源',
    // false 为隐藏菜单，通过地址栏访问可见子菜单
    visibility: true,
    menu: [
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
        id: MENU_BUSINESS_OPERATION_LOG,
        i18n: '操作记录',
        group: '操作管理',
        groupIcon: 'bkhcm-icon-operation-record',
        route: getMenuRoute(businessViews, MENU_BUSINESS_OPERATION_LOG),
      },
      {
        id: MENU_BUSINESS_TASK_MANAGEMENT,
        i18n: '任务管理',
        group: '操作管理',
        route: getMenuRoute(businessViews, MENU_BUSINESS_TASK_MANAGEMENT),
      },
      {
        id: MENU_BUSINESS_RECYCLEBIN,
        i18n: '回收站',
        icon: 'bkhcm-icon-recyclebin',
        route: getMenuRoute(businessViews, MENU_BUSINESS_RECYCLEBIN),
      },
    ],
  },
  {
    id: MENU_SERVICE,
    i18n: '工作台',
    menu: [
      {
        id: MENU_SERVICE_APPLY_MANAGEMENT,
        i18n: '单据管理',
        group: '个人事务',
        groupIcon: 'bkhcm-icon-user-8',
        route: getMenuRoute(serviceViews, MENU_SERVICE_APPLY_MANAGEMENT),
      },
      {
        id: MENU_SCHEME,
        i18n: '资源选型',
        visibility: ENABLE_CLOUD_SELECTION === 'true',
        group: '公共事务',
        groupIcon: 'bkhcm-icon-user-8',
        route: getMenuRoute(serviceViews, MENU_SCHEME),
      },
      {
        id: MENU_SERVICE_ACCOUNT_MANAGE,
        i18n: '账号管理',
        group: '公共事务',
        groupIcon: 'bkhcm-icon-user-8',
        route: getMenuRoute(serviceViews, MENU_SERVICE_ACCOUNT_MANAGE),
      },
    ],
  },
  {
    id: MENU_RESOURCE,
    i18n: '资源运营',
    menu: [
      {
        id: MENU_RESOURCE_MANAGE,
        i18n: '资源纳管',
        group: '云资源',
        groupIcon: 'bkhcm-icon-vpc',
        route: getMenuRoute(resourceViews, MENU_RESOURCE_MANAGE),
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
        visibility: () => ENABLE_ACCOUNT_BILL === 'true',
        group: '云账号管理',
        groupIcon: 'bkhcm-icon-account-manage',
        route: getMenuRoute(resourceViews, MENU_BILL_ROOT_ACCOUNT),
      },
      {
        id: MENU_BILL_MAIN_ACCOUNT,
        i18n: '二级账号',
        visibility: () => ENABLE_ACCOUNT_BILL === 'true',
        group: '云账号管理',
        route: getMenuRoute(resourceViews, MENU_BILL_MAIN_ACCOUNT),
      },
      {
        id: MENU_BILL_MANAGE,
        i18n: '账单管理',
        visibility: () => ENABLE_ACCOUNT_BILL === 'true',
        group: '云账号管理',
        route: getMenuRoute(resourceViews, MENU_BILL_MANAGE),
      },
      {
        id: MENU_RESOURCE_OPERATION_LOG,
        i18n: '操作记录',
        icon: 'bkhcm-icon-operation-record',
        route: getMenuRoute(resourceViews, MENU_RESOURCE_OPERATION_LOG),
      },
    ],
  },
];

export const getMenus = () => {
  const filter = (items: IMenu[]): IMenu[] => {
    return items
      .filter((menu) => {
        if (!Object.prototype.hasOwnProperty.call(menu, 'visibility')) return true;
        return typeof menu.visibility === 'function' ? menu.visibility() : !!menu.visibility;
      })
      .map((menu) => (menu.menu?.length ? { ...menu, menu: filter(menu.menu) } : menu));
  };
  return filter(menus);
};
