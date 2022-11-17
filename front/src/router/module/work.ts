// import { CogShape } from 'bkui-vue/lib/icon';
import { RouteRecordRaw } from 'vue-router';

const workMenus: RouteRecordRaw[] = [
  {
    path: '/workbench/projectManage',
    name: '项目管理',
    component: () => import('@/views/resource/demo'),
    meta: {
      activeKey: 'projectManage',
    },
  },
  {
    path: '/workbench/userManage',
    name: '用户管理',
    component: () => import('@/views/resource/demo'),
    meta: {
      activeKey: 'userManage',
    },
  },
  {
    path: '/workbench/roleManage',
    name: '角色管理',
    component: () => import('@/views/resource/demo'),
    meta: {
      activeKey: 'roleManage',
    },
  },
  {
    path: '/workbench/tenant',
    name: '配额管理',
    children: [
      {
        path: '/workbench/tenant/projectTenant',
        name: '项目配额',
        alias: '',
        component: () => import('@/views/resource/demo'),
        meta: {
          activeKey: 'projectTenant',
        },
      },
    ],
  },
  {
    path: '/workbench/cloudManage',
    name: '云管理',
    component: () => import('@/views/resource/demo'),
    meta: {
      activeKey: 'cloudManage',
    },
  },
  {
    path: '/workbench/system',
    name: '系统设置',
    children: [
      {
        path: '/workbench/system/log',
        name: '审计',
        alias: '',
        component: () => import('@/views/resource/demo'),
        meta: {
          activeKey: 'log',
        },
      },
    ],
  },
];

export default workMenus;
