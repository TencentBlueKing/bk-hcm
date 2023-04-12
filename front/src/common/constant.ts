export enum VendorEnum {
  TCLOUD = 'tcloud',
  AWS = 'aws',
  AZURE = 'azure',
  GCP = 'gcp',
  HUAWEI = 'huawei',
}

export enum ResourceTypeEnum {
  CVM = 'cvm',
  VPC = 'vpc',
  DISK = 'disk',
}

// 资源类型
export const RESOURCE_TYPES = [
  {
    name: '主机',
    type: 'host',
  },
  {
    name: 'VPC',
    type: 'vpc',
  },
  {
    name: '子网',
    type: 'subnet',
  },
  {
    name: '安全组',
    type: 'security',
  },
  {
    name: '云硬盘',
    type: 'drive',
  },
  {
    name: '网络接口',
    type: 'network-interface',
  },
  {
    name: '弹性 IP',
    type: 'ip',
  },
  {
    name: '路由表',
    type: 'routing',
  },
  {
    name: '镜像',
    type: 'image',
  },
];

// 云厂商
export const VENDORS = [
  {
    id: 'tcloud',
    name: '腾讯云',
  },
  {
    id: 'aws',
    name: '亚马逊云',
  },
  {
    id: 'azure',
    name: '微软云',
  },
  {
    id: 'gcp',
    name: '谷歌云',
  },
  {
    id: 'huawei',
    name: '华为云',
  },
];

// 审计资源类型（与资源类型暂时独立开）
export const AUDIT_RESOURCE_TYPES = [
  {
    name: '云账号',
    type: 'account',
  },
  {
    name: 'CVM',
    type: 'cvm',
  },
  {
    name: 'VPC',
    type: 'vpc',
  },
  {
    name: '安全组',
    type: 'security_group',
  },
  {
    name: 'EIP',
    type: 'eip',
  },
  {
    name: '硬盘',
    type: 'disk',
  },
  {
    name: 'GCP防火墙',
    type: 'gcp_firewall_rule',
  },
  {
    name: '路由',
    type: 'route_table',
  },
  {
    name: '镜像',
    type: 'image',
  },
  {
    name: '网络接口',
    type: 'network_interface',
  },
  {
    name: '子网',
    type: 'subnet',
  },
];


export const FILTER_DATA = [
  {
    name: 'ID',
    id: 'id',
  },
  {
    name: '资源ID',
    id: 'cloud_id',
  },
  {
    name: '名称',
    id: 'name',
  },
  {
    name: '云厂商',
    id: 'vendor',
    children: VENDORS,
  },
  {
    name: '云账号ID',
    id: 'account_id',
    children: [],
  },
  // {
  //   name: '状态',
  //   id: 'status',
  // },
];
