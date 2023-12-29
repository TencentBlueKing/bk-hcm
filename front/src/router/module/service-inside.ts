/**
 * 服务模块除了侧边栏的内部导航
 */
import type { RouteRecordRaw } from 'vue-router';

const serviceInside: RouteRecordRaw[] = [
  {
    path: '/service/service-apply/account-add',
    name: 'applyAccount',
    component: () => import('@/views/service/service-apply/account-add/index'),
    meta: {
      backRouter: -1,
      activeKey: 'serviceApply',
      breadcrumb: ['服务', '服务申请', '账号'],
    },
  },
  // {
  //   path: '/service/service-apply/cvm',
  //   name: 'applyCvm',
  //   component: () => import('@/views/service/service-apply/cvm'),
  //   meta: {
  //     backRouter: -1,
  //     activeKey: 'serviceApply',
  //     breadcrumb: ['服务', '服务申请', '主机'],
  //   },
  // },
  // {
  //   path: '/service/service-apply/vpc',
  //   name: 'applyVPC',
  //   component: () => import('@/views/service/service-apply/vpc'),
  //   meta: {
  //     backRouter: -1,
  //     activeKey: 'serviceApply',
  //     breadcrumb: ['服务', '服务申请', 'VPC'],
  //   },
  // },
  // {
  //   path: '/service/service-apply/disk',
  //   name: 'applyDisk',
  //   component: () => import('@/views/service/service-apply/disk'),
  //   meta: {
  //     backRouter: -1,
  //     activeKey: 'serviceApply',
  //     breadcrumb: ['服务', '服务申请', '云硬盘'],
  //   },
  // },
];

export default serviceInside;
