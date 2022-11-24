// import { CogShape } from 'bkui-vue/lib/icon';
import type { RouteRecordRaw } from 'vue-router';

const businesseMenus: RouteRecordRaw[] = [
  {
    path: '/business',
    name: '计算',
    children: [
      {
        path: '/business/host',
        name: '主机',
        alias: '',
        component: () => import('@/views/business/demo2'),
        meta: {
          activeKey: 'businessHost',
        },
      },
      {
        path: '/business/disk',
        name: '硬盘',
        component: () => import('@/views/business/demo2'),
        meta: {
          activeKey: 'businessDisk',
        },
      },
      {
        path: '/business/image',
        name: '镜像',
        component: () => import('@/views/business/demo2'),
        meta: {
          activeKey: 'businessImage',
        },
      },
      {
        path: '/business/snapshot',
        name: '快照',
        component: () => import('@/views/business/demo2'),
        meta: {
          activeKey: 'businessSnapshot',
        },
      },
    ],
  },
  {
    path: '/business-net',
    name: '网络',
    children: [
      {
        path: '/business-net/security-group',
        name: '安全组',
        component: () => import('@/views/business/demo2'),
        meta: {
          activeKey: 'businessSecurityGroup',
        },
      },
      {
        path: '/business-net/vpc',
        name: 'VPC',
        component: () => import('@/views/business/demo2'),
        meta: {
          activeKey: 'businessVpc',
        },
      },
      {
        path: '/business-net/subnet',
        name: '子网',
        component: () => import('@/views/business/demo2'),
        meta: {
          activeKey: 'businessSubnet',
        },
      },
      {
        path: '/business-net/elastic-ip',
        name: '弹性IP',
        component: () => import('@/views/business/demo2'),
        meta: {
          activeKey: 'businessElasticIP',
        },
      },
      {
        path: '/business-net/routing-table',
        name: '路由表',
        component: () => import('@/views/business/demo2'),
        meta: {
          activeKey: 'businessRoutingTable',
        },
      },
    ],
  },
  {
    path: '/business-storage',
    name: '存储',
    children: [
      {
        path: '/business-storage/object-storage',
        name: '对象存储',
        component: () => import('@/views/business/demo2'),
        meta: {
          activeKey: 'businessObjectStorage',
        },
      },
      {
        path: '/business-storage/file-storage',
        name: '文件存储',
        component: () => import('@/views/business/demo2'),
        meta: {
          activeKey: 'businessFileStorage',
        },
      },
    ],
  },
];

export default businesseMenus;
