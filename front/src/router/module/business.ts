// import { CogShape } from 'bkui-vue/lib/icon';
import { LBRouteName } from '@/constants';
import type { RouteRecordRaw } from 'vue-router';

const businesseMenus: RouteRecordRaw[] = [
  {
    path: '/business',
    children: [
      {
        path: '/business/host',
        name: 'businessHost',
        alias: '',
        children: [
          {
            path: '',
            name: 'hostBusinessList',
            component: () => import('@/views/business/business-manage.vue'),
            meta: {
              activeKey: 'businessHost',
              // breadcrumb: ['资源', '主机'],
            },
          },
          {
            path: 'detail',
            name: 'hostBusinessDetail',
            component: () => import('@/views/business/business-detail.vue'),
            meta: {
              activeKey: 'businessHost',
              // breadcrumb: ['资源', '主机', '详情'],
            },
          },
          {
            path: 'recyclebin/:type',
            name: 'hostBusinessRecyclebin',
            component: () => import('@/views/business/business-manage.vue'),
            meta: {
              backRouter: 'hostBusinessList',
              activeKey: 'businessHost',
              // breadcrumb: ['资源', '主机', '回收记录'],
            },
          },
        ],
        meta: {
          title: '主机',
          activeKey: 'businessHost',
          // breadcrumb: ['资源', '主机'],
          isShowBreadcrumb: true,
          icon: 'hcm-icon bkhcm-icon-host',
        },
      },
      {
        path: '/business/drive',
        name: 'businessDisk',
        children: [
          {
            path: '',
            name: 'businessDiskList',
            component: () => import('@/views/business/business-manage.vue'),
            meta: {
              activeKey: 'businessDisk',
              // breadcrumb: ['资源', '硬盘'],
            },
          },
          {
            path: 'detail',
            name: 'driveBusinessDetail',
            component: () => import('@/views/business/business-detail.vue'),
            meta: {
              activeKey: 'businessDisk',
              // breadcrumb: ['资源', '硬盘', '详情'],
            },
          },
          {
            path: 'recyclebin/:type',
            name: 'diskBusinessRecyclebin',
            component: () => import('@/views/business/business-manage.vue'),
            meta: {
              backRouter: 'businessDiskList',
              activeKey: 'businessDisk',
              // breadcrumb: ['资源', '硬盘', '回收记录'],
            },
          },
        ],
        meta: {
          title: '硬盘',
          activeKey: 'businessDisk',
          // breadcrumb: ['资源', '硬盘'],
          isShowBreadcrumb: true,
          icon: 'hcm-icon bkhcm-icon-disk',
        },
      },
      {
        path: '/business/image',
        name: 'businessImage',
        children: [
          {
            path: '',
            component: () => import('@/views/business/business-manage.vue'),
            meta: {
              activeKey: 'businessImage',
              // breadcrumb: ['资源', '镜像'],
            },
          },
          {
            path: 'detail',
            name: 'imageBusinessDetail',
            component: () => import('@/views/business/business-detail.vue'),
            meta: {
              activeKey: 'businessImage',
              // breadcrumb: ['资源', '镜像', '详情'],
            },
          },
        ],
        meta: {
          title: '镜像',
          activeKey: 'businessImage',
          // breadcrumb: ['资源', '镜像'],
          notMenu: true,
          icon: 'hcm-icon bkhcm-icon-image',
        },
      },
      {
        path: '/business/vpc',
        name: 'businessVpc',
        children: [
          {
            path: '',
            component: () => import('@/views/business/business-manage.vue'),
            meta: {
              activeKey: 'businessVpc',
              // breadcrumb: ['资源', 'VPC'],
            },
          },
          {
            path: 'detail',
            name: 'vpcBusinessDetail',
            component: () => import('@/views/business/business-detail.vue'),
            meta: {
              activeKey: 'businessVpc',
              // breadcrumb: ['资源', 'VPC', '详情'],
            },
          },
        ],
        meta: {
          title: 'VPC',
          activeKey: 'businessVpc',
          // breadcrumb: ['资源', 'VPC'],
          isShowBreadcrumb: true,
          icon: 'hcm-icon bkhcm-icon-vpc',
        },
      },
      {
        path: '/business/subnet',
        name: 'businessSubnet',
        children: [
          {
            path: '',
            component: () => import('@/views/business/business-manage.vue'),
            meta: {
              activeKey: 'businessSubnet',
              // breadcrumb: ['资源', '子网'],
            },
          },
          {
            path: 'detail',
            name: 'subnetBusinessDetail',
            component: () => import('@/views/business/business-detail.vue'),
            meta: {
              activeKey: 'businessSubnet',
              // breadcrumb: ['资源', '子网', '详情'],
            },
          },
        ],
        meta: {
          title: '子网',
          activeKey: 'businessSubnet',
          // breadcrumb: ['资源', '子网'],
          isShowBreadcrumb: true,
          icon: 'hcm-icon bkhcm-icon-subnet',
        },
      },
      {
        path: '/business/ip',
        name: 'businessElasticIP',
        children: [
          {
            path: '',
            component: () => import('@/views/business/business-manage.vue'),
            meta: {
              activeKey: 'businessElasticIP',
              // breadcrumb: ['资源', '弹性IP'],
            },
          },
          {
            path: 'detail',
            name: 'eipsBusinessDetail',
            component: () => import('@/views/business/business-detail.vue'),
            meta: {
              activeKey: 'businessElasticIP',
              // breadcrumb: ['资源', '弹性IP', '详情'],
            },
          },
        ],
        meta: {
          title: '弹性IP',
          activeKey: 'businessElasticIP',
          // breadcrumb: ['资源', '弹性IP'],
          isShowBreadcrumb: true,
          icon: 'hcm-icon bkhcm-icon-eip',
        },
      },
      {
        path: '/business/network-interface',
        name: 'businessNetwork',
        children: [
          {
            path: '',
            component: () => import('@/views/business/business-manage.vue'),
            meta: {
              activeKey: 'businessNetwork',
              // breadcrumb: ['资源', '网络接口'],
            },
          },
          {
            path: 'detail',
            name: 'network-interfaceBusinessDetail',
            component: () => import('@/views/business/business-detail.vue'),
            meta: {
              activeKey: 'businessNetwork',
              // breadcrumb: ['资源', '网络接口', '详情'],
            },
          },
        ],
        meta: {
          title: '网络接口',
          activeKey: 'businessNetwork',
          // breadcrumb: ['资源', '网络接口'],
          isShowBreadcrumb: true,
          icon: 'hcm-icon bkhcm-icon-network-interface',
        },
      },
      {
        path: '/business/routing',
        name: 'businessRoutingTable',
        children: [
          {
            path: '',
            component: () => import('@/views/business/business-manage.vue'),
            meta: {
              activeKey: 'businessRoutingTable',
              // breadcrumb: ['资源', '路由表'],
            },
          },
          {
            path: 'detail',
            name: 'routeBusinessDetail',
            component: () => import('@/views/business/business-detail.vue'),
            meta: {
              activeKey: 'businessRoutingTable',
              // breadcrumb: ['资源', '路由表', '详情'],
            },
          },
        ],
        meta: {
          title: '路由表',
          activeKey: 'businessRoutingTable',
          // breadcrumb: ['资源', '路由表'],
          isShowBreadcrumb: true,
          icon: 'hcm-icon bkhcm-icon-route-table',
        },
      },
      {
        path: '/business/security',
        name: 'businessSecurityGroup',
        children: [
          {
            path: '',
            component: () => import('@/views/business/business-manage.vue'),
            meta: {
              activeKey: 'businessSecurityGroup',
              // breadcrumb: ['资源', '安全组'],
            },
          },
          {
            path: 'detail',
            name: 'securityBusinessDetail',
            component: () => import('@/views/business/business-detail.vue'),
            meta: {
              activeKey: 'businessSecurityGroup',
              // breadcrumb: ['资源', '安全组', '详情'],
            },
          },
        ],
        meta: {
          title: '安全组',
          activeKey: 'businessSecurityGroup',
          // breadcrumb: ['资源', '安全组'],
          isShowBreadcrumb: true,
          icon: 'hcm-icon bkhcm-icon-security-group',
        },
      },
      {
        path: 'gcp/detail',
        name: 'gcpBusinessDetail',
        component: () => import('@/views/business/business-detail.vue'),
        meta: {
          activeKey: 'businessSecurityGroup',
          // breadcrumb: ['资源', 'gcp防火墙', '详情'],
          notMenu: true,
        },
      },
      {
        path: 'template/detail',
        name: 'templateBusinessDetail',
        component: () => import('@/views/business/business-detail.vue'),
        meta: {
          activeKey: 'businessSecurityGroup',
          // breadcrumb: ['资源', '参数模板', '详情'],
          notMenu: true,
        },
      },
      {
        path: '/business/loadbalancer',
        name: 'businessClb',
        component: () => import('@/views/business/load-balancer/index'),
        redirect: '/business/loadbalancer/clb-view',
        children: [
          {
            path: 'clb-view',
            name: 'loadbalancer-view',
            component: () => import('@/views/business/load-balancer/clb-view/index'),
            children: [
              {
                path: '',
                name: LBRouteName.allLbs,
                component: () => import('@/views/business/load-balancer/clb-view/all-clbs-manager/index'),
                props(route) {
                  return route.query;
                },
                meta: {
                  type: 'all',
                },
              },
              {
                path: 'lb/:id',
                name: LBRouteName.lb,
                component: () => import('@/views/business/load-balancer/clb-view/specific-clb-manager/index'),
                props(route) {
                  return { ...route.params, ...route.query };
                },
                meta: {
                  type: 'lb',
                  rootRoutePath: '/business/loadbalancer/clb-view',
                },
              },
              {
                path: 'listener/:id',
                name: LBRouteName.listener,
                component: () => import('@/views/business/load-balancer/clb-view/specific-listener-manager/index'),
                props(route) {
                  return { ...route.params, ...route.query };
                },
                meta: {
                  type: 'listener',
                  rootRoutePath: '/business/loadbalancer/clb-view',
                },
              },
              {
                path: 'domain/:id',
                name: LBRouteName.domain,
                component: () => import('@/views/business/load-balancer/clb-view/specific-domain-manager/index'),
                props(route) {
                  return { ...route.params, ...route.query };
                },
                meta: {
                  type: 'domain',
                  rootRoutePath: '/business/loadbalancer/clb-view',
                },
              },
            ],
          },
          {
            path: 'group-view',
            name: 'target-group-view',
            component: () => import('@/views/business/load-balancer/group-view/index'),
            children: [
              {
                path: '',
                name: LBRouteName.allTgs,
                component: () => import('@/views/business/load-balancer/group-view/all-groups-manager/index'),
                props(route) {
                  return route.query;
                },
              },
              {
                path: ':id',
                name: LBRouteName.tg,
                component: () =>
                  import('@/views/business/load-balancer/group-view/specific-target-group-manager/index'),
                props(route) {
                  return { ...route.params, ...route.query };
                },
                meta: {
                  rootRoutePath: '/business/loadbalancer/group-view',
                },
              },
            ],
            meta: {
              applyRes: 'targetGroup',
            },
          },
        ],
        meta: {
          title: '负载均衡',
          activeKey: 'businessClb',
          icon: 'hcm-icon bkhcm-icon-loadbalancer',
        },
      },
      {
        path: '/business/cert',
        name: 'businessCert',
        component: () => import('@/views/business/cert-manager/index'),
        meta: {
          title: '证书托管',
          activeKey: 'businessCert',
          isShowBreadcrumb: true,
          icon: 'hcm-icon bkhcm-icon-cert',
        },
      },
    ],
    meta: {
      groupTitle: '资源',
    },
  },
  {
    path: '/business',
    children: [
      {
        path: '/business/record',
        name: 'businessRecord',
        children: [
          {
            path: '',
            name: 'operationRecords',
            component: () => import('@/views/resource/resource-manage/operationRecord/index'),
            meta: {
              activeKey: 'businessRecord',
              isShowBreadcrumb: true,
              icon: 'hcm-icon bkhcm-icon-operation-record',
            },
          },
          {
            path: 'detail',
            name: 'operationRecordsDetail',
            component: () => import('@/views/resource/resource-manage/operationRecord/RecordDetail/index'),
            meta: {
              activeKey: 'businessRecord',
              isShowBreadcrumb: true,
              icon: 'hcm-icon bkhcm-icon-cert',
            },
          },
        ],
        meta: {
          title: '操作记录',
          activeKey: 'businessRecord',
          isShowBreadcrumb: true,
          icon: 'hcm-icon bkhcm-icon-operation-record',
        },
      },
    ],
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
          // breadcrumb: ['业务', '回收站'],
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
          // breadcrumb: ['资源管理', '主机'],
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
          // breadcrumb: ['资源管理', 'VPC'],
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
          // breadcrumb: ['资源管理', '云硬盘'],
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
          // breadcrumb: ['资源管理', '子网'],
          notMenu: true,
        },
      },
      {
        path: '/business/service/service-apply/clb',
        name: 'applyClb',
        component: () => import('@/views/service/service-apply/clb'),
        meta: {
          backRouter: -1,
          activeKey: 'businessClb',
          // breadcrumb: ['资源管理', '负载均衡'],
          notMenu: true,
          applyRes: 'lb',
        },
      },
    ],
    meta: {
      groupTitle: '回收站',
    },
  },
];

export default businesseMenus;
