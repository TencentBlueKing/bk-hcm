// import { CogShape } from 'bkui-vue/lib/icon';
import type { RouteRecordRaw } from 'vue-router';

const businesseMenus: RouteRecordRaw[] = [
  {
    path: '/business',
    name: '资源',
    children: [
      {
        path: '/business/host',
        name: '主机',
        alias: '',
        children: [
          {
            path: '',
            name: 'hostBusinessList',
            component: () => import('@/views/business/business-manage.vue'),
            meta: {
              activeKey: 'businessHost',
              breadcrumb: ['资源', '主机'],
            },
          },
          {
            path: 'detail',
            name: 'hostBusinessDetail',
            component: () => import('@/views/business/business-detail.vue'),
            meta: {
              activeKey: 'businessHost',
              breadcrumb: ['资源', '主机', '详情'],
            },
          },
          {
            path: 'recyclebin/:type',
            name: 'hostBusinessRecyclebin',
            component: () => import('@/views/business/business-manage.vue'),
            meta: {
              backRouter: 'hostBusinessList',
              activeKey: 'businessHost',
              breadcrumb: ['资源', '主机', '回收记录'],
            },
          },
        ],
        meta: {
          activeKey: 'businessHost',
          breadcrumb: ['资源', '主机'],
          isShowBreadcrumb: true,
          icon: 'hcm-icon bkhcm-icon-host',
        },
      },
      {
        path: '/business/drive',
        name: '硬盘',
        children: [
          {
            path: '',
            name: 'businessDiskList',
            component: () => import('@/views/business/business-manage.vue'),
            meta: {
              activeKey: 'businessDisk',
              breadcrumb: ['资源', '硬盘'],
            },
          },
          {
            path: 'detail',
            name: 'driveBusinessDetail',
            component: () => import('@/views/business/business-detail.vue'),
            meta: {
              activeKey: 'businessDisk',
              breadcrumb: ['资源', '硬盘', '详情'],
            },
          },
          {
            path: 'recyclebin/:type',
            name: 'diskBusinessRecyclebin',
            component: () => import('@/views/business/business-manage.vue'),
            meta: {
              backRouter: 'businessDiskList',
              activeKey: 'businessDisk',
              breadcrumb: ['资源', '硬盘', '回收记录'],
            },
          },
        ],
        meta: {
          activeKey: 'businessDisk',
          breadcrumb: ['资源', '硬盘'],
          isShowBreadcrumb: true,
          icon: 'hcm-icon bkhcm-icon-disk',
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
              breadcrumb: ['资源', '镜像'],
            },
          },
          {
            path: 'detail',
            name: 'imageBusinessDetail',
            component: () => import('@/views/business/business-detail.vue'),
            meta: {
              activeKey: 'businessImage',
              breadcrumb: ['资源', '镜像', '详情'],
            },
          },
        ],
        meta: {
          activeKey: 'businessImage',
          breadcrumb: ['资源', '镜像'],
          notMenu: true,
          icon: 'hcm-icon bkhcm-icon-image',
        },
      },
      {
        path: '/business/vpc',
        name: 'VPC',
        children: [
          {
            path: '',
            component: () => import('@/views/business/business-manage.vue'),
            meta: {
              activeKey: 'businessVpc',
              breadcrumb: ['资源', 'VPC'],
            },
          },
          {
            path: 'detail',
            name: 'vpcBusinessDetail',
            component: () => import('@/views/business/business-detail.vue'),
            meta: {
              activeKey: 'businessVpc',
              breadcrumb: ['资源', 'VPC', '详情'],
            },
          },
        ],
        meta: {
          activeKey: 'businessVpc',
          breadcrumb: ['资源', 'VPC'],
          isShowBreadcrumb: true,
          icon: 'hcm-icon bkhcm-icon-vpc',
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
              breadcrumb: ['资源', '子网'],
            },
          },
          {
            path: 'detail',
            name: 'subnetBusinessDetail',
            component: () => import('@/views/business/business-detail.vue'),
            meta: {
              activeKey: 'businessSubnet',
              breadcrumb: ['资源', '子网', '详情'],
            },
          },
        ],
        meta: {
          activeKey: 'businessSubnet',
          breadcrumb: ['资源', '子网'],
          isShowBreadcrumb: true,
          icon: 'hcm-icon bkhcm-icon-subnet',
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
              breadcrumb: ['资源', '弹性IP'],
            },
          },
          {
            path: 'detail',
            name: 'eipsBusinessDetail',
            component: () => import('@/views/business/business-detail.vue'),
            meta: {
              activeKey: 'businessElasticIP',
              breadcrumb: ['资源', '弹性IP', '详情'],
            },
          },
        ],
        meta: {
          activeKey: 'businessElasticIP',
          breadcrumb: ['资源', '弹性IP'],
          isShowBreadcrumb: true,
          icon: 'hcm-icon bkhcm-icon-eip',
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
              breadcrumb: ['资源', '网络接口'],
            },
          },
          {
            path: 'detail',
            name: 'network-interfaceBusinessDetail',
            component: () => import('@/views/business/business-detail.vue'),
            meta: {
              activeKey: 'businessNetwork',
              breadcrumb: ['资源', '网络接口', '详情'],
            },
          },
        ],
        meta: {
          activeKey: 'businessNetwork',
          breadcrumb: ['资源', '网络接口'],
          isShowBreadcrumb: true,
          icon: 'hcm-icon bkhcm-icon-network-interface',
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
              breadcrumb: ['资源', '路由表'],
            },
          },
          {
            path: 'detail',
            name: 'routeBusinessDetail',
            component: () => import('@/views/business/business-detail.vue'),
            meta: {
              activeKey: 'businessRoutingTable',
              breadcrumb: ['资源', '路由表', '详情'],
            },
          },
        ],
        meta: {
          activeKey: 'businessRoutingTable',
          breadcrumb: ['资源', '路由表'],
          isShowBreadcrumb: true,
          icon: 'hcm-icon bkhcm-icon-route-table',
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
              breadcrumb: ['资源', '安全组'],
            },
          },
          {
            path: 'detail',
            name: 'securityBusinessDetail',
            component: () => import('@/views/business/business-detail.vue'),
            meta: {
              activeKey: 'businessSecurityGroup',
              breadcrumb: ['资源', '安全组', '详情'],
            },
          },
        ],
        meta: {
          activeKey: 'businessSecurityGroup',
          breadcrumb: ['资源', '安全组'],
          isShowBreadcrumb: true,
          icon: 'hcm-icon bkhcm-icon-security-group',
        },
      },
    ],
  },
  {
    path: '/business',
    name: '其他',
    children: [
      {
        path: '/business/record',
        name: '操作记录',
        component: () => import('@/views/resource/resource-manage/operationRecord/index'),
        meta: {
          activeKey: 'record',
          isShowBreadcrumb: true,
          icon: 'hcm-icon bkhcm-icon-operation-record',
        },
      },
    ],
  },
  {
    path: '/business',
    name: '回收站',
    children: [
      {
        path: '/business/recyclebin',
        name: '回收站',
        component: () => import('@/views/business/business-manage.vue'),
        meta: {
          activeKey: 'recyclebin',
          breadcrumb: ['业务', '回收站'],
          isShowBreadcrumb: true,
          icon: 'hcm-icon bkhcm-icon-recyclebin',
        },
      },
      {
        path: '/business/service/service-apply/cvm',
        name: 'applyCvm',
        component: () => import('@/views/service/service-apply/cvm'),
        meta: {
          backRouter: -1,
          activeKey: 'businessHost',
          breadcrumb: ['资源管理', '主机'],
          notMenu: true,
        },
      },
      {
        path: '/business/service/service-apply/vpc',
        name: 'applyVPC',
        component: () => import('@/views/service/service-apply/vpc'),
        meta: {
          backRouter: -1,
          activeKey: 'businessVpc',
          breadcrumb: ['资源管理', 'VPC'],
          notMenu: true,
        },
      },
      {
        path: '/business/service/service-apply/disk',
        name: 'applyDisk',
        component: () => import('@/views/service/service-apply/disk'),
        meta: {
          backRouter: -1,
          activeKey: 'businessDisk',
          breadcrumb: ['资源管理', '云硬盘'],
          notMenu: true,
        },
      },
      {
        path: '/business/service/service-apply/subnet',
        name: 'applySubnet',
        component: () => import('@/views/service/service-apply/subnet'),
        meta: {
          backRouter: -1,
          activeKey: 'businessSubnet',
          breadcrumb: ['资源管理', '子网'],
          notMenu: true,
        },
      },
    ],
  },
];

export default businesseMenus;
