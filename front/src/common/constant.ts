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
    name: '主机',
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
    name: 'GCP防火墙',
    type: 'gcp_firewall_rule',
  },
  {
    name: '弹性IP',
    type: 'eip',
  },
  {
    name: '硬盘',
    type: 'disk',
  },
  {
    name: '路由表',
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
export const CIDRLIST = [
  {
    name: '10',
    id: '10',
  },
  {
    name: '172',
    id: '172',
  },
  {
    name: '192',
    id: '192',
  },
];

export const CIDRDATARANGE  = {
  10: { min: 0, max: 255 },
  172: { min: 16, max: 31 },
  192: { min: 168, max: 168 },
};

export const TCLOUDCIDRMASKRANGE  = {
  10: { min: 12, max: 28 },
  172: { min: 12, max: 28 },
  192: { min: 16, max: 28 },
};

export const CIDRMASKRANGE  = {
  10: { min: 8, max: 28 },
  172: { min: 12, max: 28 },
  192: { min: 16, max: 28 },
};

export const GCP_CLOUD_HOST_STATUS = {
  'PROVISIONING': '准备资源中',
  'STAGING': '启动中',
  'RUNNING': '运行中',
  'STOPPING': '停止中',
  'REPAIRING': '修复中',
  'TERMINATED': '已关机',
  'SUSPENDING': '暂停中',
  'SUSPENDED': '已暂停'
};

export const AZURE_CLOUD_HOST_STATUS = {
  'PowerState/creating': '创建中',
  'PowerState/starting': '启动中',
  'PowerState/running': '运行中',
  'PowerState/stopping': '停止中',
  'PowerState/stopped': '已关机',
  'PowerState/deallocating': '已停止(从主机分离中)',
  'PowerState/deallocated': '已停止(已从主机分离)'
};

export const HUAWEI_CLOUD_HOST_STATUS = {
  'BUILD': '创建中',
  'REBOOT': '重启中',
  'HARD_REBOOT': '强制重启中',
  'REBUILD': '重建中',
  'MIGRATING': '热迁移中',
  'RESIZE': '变更中',
  'ACTIVE': '运行中',
  'SHUTOFF': '已停止',
  'REVERT_RESIZE': '回退变更规格',
  'VERIFY_RESIZE': '校验变更配置',
  'ERROR': '异常',
  'DELETED': '删除中',
  'SHELVED': '启动镜像异常',
  'SHELVED_OFFLOADED': '启动磁盘异常',
  'UNKNOWN': '未知状态'
}

export const CLOUD_HOST_STATUS = {
  PENDING: '创建中',
  LAUNCH_FAILED: '创建失败',
  RUNNING: '运行中',
  STOPPED: '关机',
  stopped: '关机',
  STARTING: '开机中',
  STOPPING: '关机中',
  REBOOTING: '重启中',
  SHUTDOWN: '停止待销毁',
  TERMINATING: '销毁中',
  running: '运行中',
  ...GCP_CLOUD_HOST_STATUS,
  ...AZURE_CLOUD_HOST_STATUS,
  ...HUAWEI_CLOUD_HOST_STATUS,
};