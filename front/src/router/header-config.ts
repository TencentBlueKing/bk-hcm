/**
 * 头部导航配置
 */

export const headRouteConfig = [
  {
    id: 'business',
    name: '资源管理',
    route: 'business',
    href: '#/business/host',
  },
  {
    id: 'service',
    name: '我的单据',
    route: 'service',
    href: '#/service/my-apply',
  },
  {
    id: 'resource',
    name: '资源接入',
    route: 'resource',
    href: '#/resource/resource',
  },
  {
    id: 'scheme',
    name: '资源选型',
    route: 'scheme',
    href: '#/scheme/recommendation',
  },

  // 接下来是 资源选型、平台管理
  // {
  //   id: 'workbench',
  //   name: '工作台',
  //   route: 'workbench',
  //   href: '#/workbench/audit',
  // },
];
