// import { CogShape } from 'bkui-vue/lib/icon';
import { RouteRecordRaw } from 'vue-router';

const resourceMenus: RouteRecordRaw[] = [
  {
    path: '/resource',
    name: '主机',
    children: [
      {
        path: '/resource/vm',
        name: '虚拟机',
        alias: '',
        component: () => import('@/views/resource/demo2'),
        meta: {
          activeKey: 'vm',
        },
      },
      {
        path: '/resource/demo3',
        name: 'demo3-template',
        alias: '',
        component: () => import('@/views/resource/demo3.vue'),
        meta: {
          activeKey: 'demo3',
        },
      },
      {
        path: '/resource/demo4',
        name: 'demo4-template-setup',
        alias: '',
        component: () => import('@/views/resource/demo4.vue'),
        meta: {
          activeKey: 'demo4',
        },
      },
      {
        path: '/resource/snapshot',
        name: '快照',
        alias: '',
        component: () => import('@/views/resource/demo2'),
        meta: {
          activeKey: 'snapshot',
        },
      },
      {
        path: '/resource/image',
        name: '镜像',
        alias: '',
        component: () => import('@/views/resource/demo'),
        meta: {
          activeKey: 'image',
        },
      },
      {
        path: '/resource/securityGroup',
        name: '安全组',
        alias: '',
        component: () => import('@/views/resource/demo'),
        meta: {
          activeKey: 'securityGroup',
        },
      },
      {
        path: '/resource/blockStorage',
        name: '快存储',
        alias: '',
        component: () => import('@/views/resource/demo'),
        meta: {
          activeKey: 'blockStorage',
        },
      },
    ],
  },
  {
    path: '/resource/net',
    name: '网络',
    children: [
      {
        path: '/resource/net/vpc',
        name: 'vpc',
        alias: '',
        component: () => import('@/views/resource/demo'),
        meta: {
          activeKey: 'vpc',
        },
      },
      {
        path: '/resource/net/eip',
        name: '弹性IP',
        alias: '',
        component: () => import('@/views/resource/demo'),
        meta: {
          activeKey: 'eip',
        },
      },
      {
        path: '/resource/net/subnet',
        name: '子网',
        alias: '',
        component: () => import('@/views/resource/demo'),
        meta: {
          activeKey: 'subnet',
        },
      },
      {
        path: '/resource/net/routerTable',
        name: '路由表',
        alias: '',
        component: () => import('@/views/resource/demo'),
        meta: {
          activeKey: 'routerTable',
        },
      },
    ],
  },
  {
    path: '/resource/loadBalance',
    name: '负载均衡',
    component: () => import('@/views/resource/demo'),
    meta: {
      activeKey: 'loadBalance',
    },
  },
  {
    path: '/resource/storage',
    name: '存储',
    children: [
      {
        path: '/resource/storage/fileStorage',
        name: '文件存储',
        alias: '',
        component: () => import('@/views/resource/demo'),
        meta: {
          activeKey: 'fileStorage',
        },
      },
      {
        path: '/resource/storage/objectStorage',
        name: '对象存储',
        alias: '',
        component: () => import('@/views/resource/demo'),
        meta: {
          activeKey: 'objectStorage',
        },
      },
    ],
  },
];

export default resourceMenus;
