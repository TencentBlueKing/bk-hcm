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
        children: [
          {
            path: '',
            component: () => import('@/views/business/business-manage.vue'),
            meta: {
              activeKey: 'businessHost',
              breadcrumb: ['计算', '主机'],
            },
          },
          {
            path: 'detail',
            name: 'hostBusinessDetail',
            component: () => import('@/views/business/business-detail.vue'),
            meta: {
              activeKey: 'businessHost',
              breadcrumb: ['计算', '主机', '详情'],
            },
          }
        ],
        meta: {
          activeKey: 'businessHost',
          breadcrumb: ['计算', '主机'],
        },
      },
      {
        path: '/business/drive',
        name: '硬盘',
        children: [
          {
            path: '',
            component: () => import('@/views/business/business-manage.vue'),
            meta: {
              activeKey: 'businessDisk',
              breadcrumb: ['计算', '硬盘'],
            },
          },
          {
            path: 'detail',
            name: 'driveBusinessDetail',
            component: () => import('@/views/business/business-detail.vue'),
            meta: {
              activeKey: 'businessDisk',
              breadcrumb: ['计算', '硬盘', '详情'],
            },
          }
        ],
        meta: {
          activeKey: 'businessDisk',
          breadcrumb: ['计算', '硬盘'],
        },
      },
      {
        path: '/business/image',
        name: '镜像',
        children: [
          {
            path: '',
            component: () => import('@/views/business/business-manage.vue'),
            meta: {
              activeKey: 'businessImage',
              breadcrumb: ['计算', '镜像'],
            },
          },
          {
            path: 'detail',
            name: 'imageBusinessDetail',
            component: () => import('@/views/business/business-detail.vue'),
            meta: {
              activeKey: 'businessImage',
              breadcrumb: ['计算', '镜像', '详情'],
            },
          }
        ],
        meta: {
          activeKey: 'businessImage',
          breadcrumb: ['计算', '镜像'],
        },
      },
      // {
      //   path: '/business/snapshot',
      //   name: '快照',
      //   component: () => import('@/views/business/demo2'),
      //   meta: {
      //     activeKey: 'businessSnapshot',
      //   },
      // },
    ],
  },
  {
    path: '/business',
    name: '网络',
    children: [
      {
        path: '/business/vpc',
        name: 'VPC',
        children: [
          {
            path: '',
            component: () => import('@/views/business/business-manage.vue'),
            meta: {
              activeKey: 'businessVpc',
              breadcrumb: ['网络', 'VPC'],
            },
          },
          {
            path: 'detail',
            name: 'vpcBusinessDetail',
            component: () => import('@/views/business/business-detail.vue'),
            meta: {
              activeKey: 'businessVpc',
              breadcrumb: ['计算', 'VPC', '详情'],
            },
          }
        ],
        meta: {
          activeKey: 'businessVpc',
          breadcrumb: ['网络', 'VPC'],
        },
      },
      {
        path: '/business/subnet',
        name: '子网',
        children: [
          {
            path: '',
            component: () => import('@/views/business/business-manage.vue'),
            meta: {
              activeKey: 'businessSubnet',
              breadcrumb: ['网络', '子网'],
            },
          },
          {
            path: 'detail',
            name: 'subnetBusinessDetail',
            component: () => import('@/views/business/business-detail.vue'),
            meta: {
              activeKey: 'businessSubnet',
              breadcrumb: ['计算', '子网', '详情'],
            },
          }
        ],
        meta: {
          activeKey: 'businessSubnet',
          breadcrumb: ['网络', '子网'],
        },
      },
      {
        path: '/business/ip',
        name: '弹性IP',
        children: [
          {
            path: '',
            component: () => import('@/views/business/business-manage.vue'),
            meta: {
              activeKey: 'businessElasticIP',
              breadcrumb: ['网络', '弹性IP'],
            },
          },
          {
            path: 'detail',
            name: 'ipBusinessDetail',
            component: () => import('@/views/business/business-detail.vue'),
            meta: {
              activeKey: 'businessElasticIP',
              breadcrumb: ['计算', '弹性IP', '详情'],
            },
          }
        ],
        meta: {
          activeKey: 'businessElasticIP',
          breadcrumb: ['网络', '弹性IP'],
        },
      },
      {
        path: '/business/network-interface',
        name: '网络接口',
        children: [
          {
            path: '',
            component: () => import('@/views/business/business-manage.vue'),
            meta: {
              activeKey: 'businessNetwork',
              breadcrumb: ['网络', '网络接口'],
            },
          },
          {
            path: 'detail',
            name: 'network-interfaceBusinessDetail',
            component: () => import('@/views/business/business-detail.vue'),
            meta: {
              activeKey: 'businessNetwork',
              breadcrumb: ['计算', '网络接口', '详情'],
            },
          }
        ],
        meta: {
          activeKey: 'businessNetwork',
          breadcrumb: ['网络', '网络接口'],
        },
      },
      {
        path: '/business/routing',
        name: '路由表',
        children: [
          {
            path: '',
            component: () => import('@/views/business/business-manage.vue'),
            meta: {
              activeKey: 'businessRoutingTable',
              breadcrumb: ['网络', '路由表'],
            },
          },
          {
            path: 'detail',
            name: 'routingBusinessDetail',
            component: () => import('@/views/business/business-detail.vue'),
            meta: {
              activeKey: 'businessRoutingTable',
              breadcrumb: ['计算', '路由表', '详情'],
            },
          }
        ],
        meta: {
          activeKey: 'businessRoutingTable',
          breadcrumb: ['网络', '路由表'],
        },
      },
      {
        path: '/business/security',
        name: '安全组',
        children: [
          {
            path: '',
            component: () => import('@/views/business/business-manage.vue'),
            meta: {
              activeKey: 'businessSecurityGroup',
              breadcrumb: ['网络', '安全组'],
            },
          },
          {
            path: 'detail',
            name: 'securityBusinessDetail',
            component: () => import('@/views/business/business-detail.vue'),
            meta: {
              activeKey: 'businessSecurityGroup',
              breadcrumb: ['计算', '安全组', '详情'],
            },
          }
        ],
        meta: {
          activeKey: 'businessSecurityGroup',
          breadcrumb: ['网络', '安全组'],
        },
      },
    ],
  },
  // {
  //   path: '/business-storage',
  //   name: '存储',
  //   children: [
  //     {
  //       path: '/business-storage/object-storage',
  //       name: '对象存储',
  //       component: () => import('@/views/business/business-manage'),
  //       meta: {
  //         activeKey: 'businessObjectStorage',
  //       },
  //     },
  //     {
  //       path: '/business-storage/file-storage',
  //       name: '文件存储',
  //       component: () => import('@/views/business/business-manage'),
  //       meta: {
  //         activeKey: 'businessFileStorage',
  //       },
  //     },
  //   ],
  // },
];

export default businesseMenus;
