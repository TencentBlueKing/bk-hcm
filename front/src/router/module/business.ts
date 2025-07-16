// import { CogShape } from 'bkui-vue/lib/icon';
import type { RouteRecordRaw } from 'vue-router';
import { operationLogBiz as operationLogBizRouteConfig } from '@/views/operation-log/route-config';
import { loadBalancerBiz as loadBalancerBizRouteConfig } from '@/views/load-balancer/route-config';
import taskRouteConfig from '@/views/task/route-config';
import Meta from '../meta';
import {
  MENU_BUSINESS_CERT_MANAGEMENT,
  MENU_BUSINESS_DISK_MANAGEMENT,
  MENU_BUSINESS_EIP_MANAGEMENT,
  MENU_BUSINESS_HOST_MANAGEMENT,
  MENU_BUSINESS_IMAGE_MANAGEMENT,
  MENU_BUSINESS_NETWORK_INTERFACE_MANAGEMENT,
  MENU_BUSINESS_RECYCLE_BIN_MANAGEMENT,
  MENU_BUSINESS_ROUTEING_TABLE_MANAGEMENT,
  MENU_BUSINESS_SECURITY_GROUP_MANAGEMENT,
  MENU_BUSINESS_SUBNET_MANAGEMENT,
  MENU_BUSINESS_VPC_MANAGEMENT,
} from '@/constants/menu-symbol';

const businessMenus: RouteRecordRaw[] = [
  {
    path: '/business',
    children: [
      {
        path: '/business/host',
        alias: '',
        children: [
          {
            path: '',
            name: MENU_BUSINESS_HOST_MANAGEMENT,
            component: () => import('@/views/business/business-manage.vue'),
            meta: {
              activeKey: MENU_BUSINESS_HOST_MANAGEMENT,
            },
          },
          {
            path: 'detail',
            // TODO: details后续优化name，注意use-column里面的跳转
            name: 'hostBusinessDetail',
            component: () => import('@/views/business/business-detail.vue'),
            meta: {
              ...new Meta({
                activeKey: MENU_BUSINESS_HOST_MANAGEMENT,
                layout: {
                  breadcrumbs: {
                    show: false,
                  },
                },
              }),
            },
          },
          {
            path: 'recyclebin/:type',
            name: MENU_BUSINESS_RECYCLE_BIN_MANAGEMENT,
            component: () => import('@/views/business/business-manage.vue'),
            meta: {
              ...new Meta({
                activeKey: MENU_BUSINESS_HOST_MANAGEMENT,
                isShowBreadcrumb: false,
              }),
            },
          },
        ],
        meta: {
          title: '主机',
          activeKey: MENU_BUSINESS_HOST_MANAGEMENT,
          isShowBreadcrumb: true,
          icon: 'hcm-icon bkhcm-icon-host',
        },
      },
      {
        path: '/business/drive',
        children: [
          {
            path: '',
            name: MENU_BUSINESS_DISK_MANAGEMENT,
            component: () => import('@/views/business/business-manage.vue'),
            meta: {
              activeKey: MENU_BUSINESS_DISK_MANAGEMENT,
            },
          },
          {
            path: 'detail',
            name: 'driveBusinessDetail',
            component: () => import('@/views/business/business-detail.vue'),
            meta: {
              ...new Meta({
                activeKey: MENU_BUSINESS_DISK_MANAGEMENT,
                isShowBreadcrumb: false,
              }),
            },
          },
          {
            path: 'recyclebin/:type',
            component: () => import('@/views/business/business-manage.vue'),
            meta: {
              ...new Meta({
                activeKey: MENU_BUSINESS_DISK_MANAGEMENT,
                isShowBreadcrumb: false,
              }),
            },
          },
        ],
        meta: {
          title: '硬盘',
          activeKey: MENU_BUSINESS_DISK_MANAGEMENT,
          isShowBreadcrumb: true,
          icon: 'hcm-icon bkhcm-icon-disk',
        },
      },
      {
        path: '/business/image',
        children: [
          {
            path: '',
            name: MENU_BUSINESS_IMAGE_MANAGEMENT,
            component: () => import('@/views/business/business-manage.vue'),
            meta: {
              activeKey: MENU_BUSINESS_IMAGE_MANAGEMENT,
            },
          },
          {
            path: 'detail',
            name: 'imageBusinessDetail',
            component: () => import('@/views/business/business-detail.vue'),
            meta: {
              activeKey: MENU_BUSINESS_IMAGE_MANAGEMENT,
            },
          },
        ],
        meta: {
          title: '镜像',
          activeKey: MENU_BUSINESS_IMAGE_MANAGEMENT,
          notMenu: true,
          icon: 'hcm-icon bkhcm-icon-image',
        },
      },
      {
        path: '/business/vpc',
        children: [
          {
            path: '',
            name: MENU_BUSINESS_VPC_MANAGEMENT,
            component: () => import('@/views/business/business-manage.vue'),
            meta: {
              activeKey: MENU_BUSINESS_VPC_MANAGEMENT,
            },
          },
          {
            path: 'detail',
            name: 'vpcBusinessDetail',
            component: () => import('@/views/business/business-detail.vue'),
            meta: {
              ...new Meta({
                activeKey: MENU_BUSINESS_VPC_MANAGEMENT,
                isShowBreadcrumb: false,
              }),
            },
          },
        ],
        meta: {
          title: 'VPC',
          activeKey: MENU_BUSINESS_VPC_MANAGEMENT,
          isShowBreadcrumb: true,
          icon: 'hcm-icon bkhcm-icon-vpc',
        },
      },
      {
        path: '/business/subnet',
        children: [
          {
            path: '',
            name: MENU_BUSINESS_SUBNET_MANAGEMENT,
            component: () => import('@/views/business/business-manage.vue'),
            meta: {
              activeKey: MENU_BUSINESS_SUBNET_MANAGEMENT,
            },
          },
          {
            path: 'detail',
            name: 'subnetBusinessDetail',
            component: () => import('@/views/business/business-detail.vue'),
            meta: {
              ...new Meta({
                activeKey: MENU_BUSINESS_SUBNET_MANAGEMENT,
                isShowBreadcrumb: false,
              }),
            },
          },
        ],
        meta: {
          title: '子网',
          activeKey: MENU_BUSINESS_SUBNET_MANAGEMENT,
          isShowBreadcrumb: true,
          icon: 'hcm-icon bkhcm-icon-subnet',
        },
      },
      {
        path: '/business/ip',
        children: [
          {
            path: '',
            name: MENU_BUSINESS_EIP_MANAGEMENT,
            component: () => import('@/views/business/business-manage.vue'),
            meta: {
              activeKey: MENU_BUSINESS_EIP_MANAGEMENT,
            },
          },
          {
            path: 'detail',
            name: 'eipsBusinessDetail',
            component: () => import('@/views/business/business-detail.vue'),
            meta: {
              activeKey: MENU_BUSINESS_EIP_MANAGEMENT,
              isShowBreadcrumb: false,
            },
          },
        ],
        meta: {
          title: '弹性IP',
          activeKey: MENU_BUSINESS_EIP_MANAGEMENT,
          isShowBreadcrumb: true,
          icon: 'hcm-icon bkhcm-icon-eip',
        },
      },
      {
        path: '/business/network-interface',
        children: [
          {
            path: '',
            name: MENU_BUSINESS_NETWORK_INTERFACE_MANAGEMENT,
            component: () => import('@/views/business/business-manage.vue'),
            meta: {
              activeKey: MENU_BUSINESS_NETWORK_INTERFACE_MANAGEMENT,
            },
          },
          {
            path: 'detail',
            name: 'network-interfaceBusinessDetail',
            component: () => import('@/views/business/business-detail.vue'),
            meta: {
              activeKey: MENU_BUSINESS_NETWORK_INTERFACE_MANAGEMENT,
              isShowBreadcrumb: false,
            },
          },
        ],
        meta: {
          title: '网络接口',
          activeKey: MENU_BUSINESS_NETWORK_INTERFACE_MANAGEMENT,
          isShowBreadcrumb: true,
          icon: 'hcm-icon bkhcm-icon-network-interface',
        },
      },
      {
        path: '/business/routing',
        children: [
          {
            path: '',
            name: MENU_BUSINESS_ROUTEING_TABLE_MANAGEMENT,
            component: () => import('@/views/business/business-manage.vue'),
            meta: {
              activeKey: MENU_BUSINESS_ROUTEING_TABLE_MANAGEMENT,
            },
          },
          {
            path: 'detail',
            name: 'routeBusinessDetail',
            component: () => import('@/views/business/business-detail.vue'),
            meta: {
              activeKey: MENU_BUSINESS_ROUTEING_TABLE_MANAGEMENT,
              isShowBreadcrumb: false,
            },
          },
        ],
        meta: {
          title: '路由表',
          activeKey: MENU_BUSINESS_ROUTEING_TABLE_MANAGEMENT,
          isShowBreadcrumb: true,
          icon: 'hcm-icon bkhcm-icon-route-table',
        },
      },
      {
        path: '/business/security',
        children: [
          {
            path: '',
            name: MENU_BUSINESS_SECURITY_GROUP_MANAGEMENT,
            component: () => import('@/views/business/business-manage.vue'),
            meta: {
              activeKey: MENU_BUSINESS_SECURITY_GROUP_MANAGEMENT,
            },
          },
          {
            path: 'detail',
            name: 'securityBusinessDetail',
            component: () => import('@/views/business/business-detail.vue'),
            meta: {
              activeKey: MENU_BUSINESS_SECURITY_GROUP_MANAGEMENT,
              isShowBreadcrumb: false,
            },
          },
        ],
        meta: {
          title: '安全组',
          activeKey: MENU_BUSINESS_SECURITY_GROUP_MANAGEMENT,
          isShowBreadcrumb: true,
          icon: 'hcm-icon bkhcm-icon-security-group',
        },
      },
      {
        path: 'gcp/detail',
        name: 'gcpBusinessDetail',
        component: () => import('@/views/business/business-detail.vue'),
        meta: {
          activeKey: MENU_BUSINESS_SECURITY_GROUP_MANAGEMENT,
          notMenu: true,
        },
      },
      {
        path: 'template/detail',
        name: 'templateBusinessDetail',
        component: () => import('@/views/business/business-detail.vue'),
        meta: {
          activeKey: MENU_BUSINESS_SECURITY_GROUP_MANAGEMENT,
          notMenu: true,
        },
      },
      loadBalancerBizRouteConfig[0],
      {
        path: '/business/cert',
        name: MENU_BUSINESS_CERT_MANAGEMENT,
        component: () => import('@/views/business/cert-manager/index'),
        meta: {
          title: '证书托管',
          activeKey: MENU_BUSINESS_CERT_MANAGEMENT,
          isShowBreadcrumb: true,
          icon: 'hcm-icon bkhcm-icon-cert',
          isFilterAccount: true,
        },
      },
    ],
    meta: {
      groupTitle: '资源',
    },
  },
  {
    path: '/business',
    children: [...operationLogBizRouteConfig, ...taskRouteConfig],
    meta: {
      groupTitle: '其他',
    },
  },
  {
    path: '/business',
    children: [
      {
        path: '/business/recyclebin',
        name: 'businessRecyclebin',
        component: () => import('@/views/business/business-manage.vue'),
        meta: {
          title: '回收站',
          activeKey: 'businessRecyclebin',
          isShowBreadcrumb: true,
          icon: 'hcm-icon bkhcm-icon-recyclebin',
        },
      },
      {
        path: '/business/service/service-apply/cvm',
        name: 'applyCvm',
        component: () => import('@/views/service/service-apply/cvm'),
        meta: {
          ...new Meta({
            activeKey: MENU_BUSINESS_HOST_MANAGEMENT,
            notMenu: true,
            menu: {
              relative: MENU_BUSINESS_HOST_MANAGEMENT,
            },
          }),
        },
      },
      {
        path: '/business/service/service-apply/vpc',
        name: 'applyVPC',
        component: () => import('@/views/service/service-apply/vpc'),
        meta: {
          ...new Meta({
            activeKey: MENU_BUSINESS_VPC_MANAGEMENT,
            notMenu: true,
            menu: {
              relative: MENU_BUSINESS_VPC_MANAGEMENT,
            },
          }),
        },
      },
      {
        path: '/business/service/service-apply/disk',
        name: 'applyDisk',
        component: () => import('@/views/service/service-apply/disk'),
        meta: {
          ...new Meta({
            activeKey: MENU_BUSINESS_DISK_MANAGEMENT,
            notMenu: true,
            menu: {
              relative: MENU_BUSINESS_DISK_MANAGEMENT,
            },
          }),
        },
      },
      {
        path: '/business/service/service-apply/subnet',
        name: 'applySubnet',
        component: () => import('@/views/service/service-apply/subnet'),
        meta: {
          ...new Meta({
            activeKey: MENU_BUSINESS_SUBNET_MANAGEMENT,
            notMenu: true,
            menu: {
              relative: MENU_BUSINESS_SUBNET_MANAGEMENT,
            },
          }),
        },
      },
      loadBalancerBizRouteConfig[1],
    ],
    meta: {
      groupTitle: '回收站',
    },
  },
];

export default businessMenus;
