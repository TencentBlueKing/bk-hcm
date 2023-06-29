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

// 账号类型
export const ACCOUNT_TYPES = [
  {
    id: 'resource',
    name: '资源账号',
  },
  {
    id: 'registration',
    name: '登记账号',
  },
  {
    id: 'security_audit',
    name: '安全审计账号',
  },
];

// 站点类型
export const SITE_TYPES = [
  {
    id: 'china',
    name: '中国站',
  },
  {
    id: 'international',
    name: '国际站',
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
  PROVISIONING: '准备资源中',
  STAGING: '启动中',
  RUNNING: '运行中',
  STOPPING: '停止中',
  REPAIRING: '修复中',
  TERMINATED: '已关机',
  SUSPENDING: '暂停中',
  SUSPENDED: '已暂停',
};

export const AZURE_CLOUD_HOST_STATUS = {
  'PowerState/creating': '创建中',
  'PowerState/starting': '启动中',
  'PowerState/running': '运行中',
  'PowerState/stopping': '停止中',
  'PowerState/stopped': '已关机',
  'PowerState/deallocating': '已停止(从主机分离中)',
  'PowerState/deallocated': '已停止(已从主机分离)',
};

export const HUAWEI_CLOUD_HOST_STATUS = {
  BUILD: '创建中',
  REBOOT: '重启中',
  HARD_REBOOT: '强制重启中',
  REBUILD: '重建中',
  MIGRATING: '热迁移中',
  RESIZE: '变更中',
  ACTIVE: '运行中',
  SHUTOFF: '已停止',
  REVERT_RESIZE: '回退变更规格',
  VERIFY_RESIZE: '校验变更配置',
  ERROR: '异常',
  DELETED: '删除中',
  SHELVED: '启动镜像异常',
  SHELVED_OFFLOADED: '启动磁盘异常',
  UNKNOWN: '未知状态',
};

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

export const CLOUD_AREA_REGION_GCP = {
  'northamerica-northeast1': '蒙特利尔',
  'northamerica-northeast2': '多伦多',
  'southamerica-east1': '圣保罗',
  'southamerica-west1': '圣地亚哥',
  'us-central1': '爱荷华',
  'us-east1': '南卡罗来纳',
  'us-east4': '北弗吉尼亚',
  'us-east5': '哥伦布',
  'us-south1': '达拉斯',
  'us-west1': '俄勒冈',
  'us-west2': '洛杉矶',
  'us-west3': '盐湖城',
  'us-west4': '拉斯维加斯',
  'europe-central2': '华沙',
  'europe-north1': '芬兰',
  'europe-southwest1': '马德里',
  'europe-west1': '比利时',
  'europe-west12': '都灵',
  'europe-west2': '伦敦',
  'europe-west3': '法兰克福',
  'europe-west4': '荷兰',
  'europe-west6': '苏黎世',
  'europe-west8': '米兰',
  'europe-west9': '巴黎',
  'me-central1': 'Doha',
  'me-west1': '特拉维夫',
  'asia-east1': '台湾',
  'asia-east2': '香港',
  'asia-northeast1': '东京',
  'asia-northeast2': '大阪',
  'asia-northeast3': '首尔',
  'asia-south1': '孟买',
  'asia-south2': '德里',
  'asia-southeast1': '新加坡',
  'asia-southeast2': '雅加达',
  'australia-southeast1': '悉尼',
  'australia-southeast2': '墨尔本',
};

export const CLOUD_AREA_REGION_AWS = {
  'us-east-2': 'US East (Ohio)',
  'us-east-1': '美国东部（弗吉尼亚北部）',
  'us-west-1': '美国西部（加利福尼亚北部）',
  'us-west-2': '美国西部（俄勒冈）',
  'af-south-1': 'Africa (Cape Town)',
  'ap-east-1': 'Asia Pacific (Hong Kong)',
  'ap-south-2': '亚太地区（海得拉巴）',
  'ap-southeast-3': '亚太地区（雅加达）',
  'ap-southeast-4': '亚太地区（墨尔本）',
  'ap-south-1': 'Asia Pacific (Mumbai)',
  'ap-northeast-3': 'Asia Pacific (Osaka)',
  'ap-northeast-2': 'Asia Pacific (Seoul)',
  'ap-southeast-1': '亚太地区（新加坡）',
  'ap-southeast-2': '亚太地区（悉尼）',
  'ap-northeast-1': '亚太区域（东京）',
  'ca-central-1': 'Canada (Central)',
  'eu-central-1': 'Europe (Frankfurt)',
  'eu-west-1': '欧洲（爱尔兰）',
  'eu-west-2': 'Europe (London)',
  'eu-south-1': 'Europe (Milan)',
  'eu-west-3': 'Europe (Paris)',
  'eu-south-2': '欧洲（西班牙）',
  'eu-north-1': '欧洲（斯德哥尔摩）',
  'eu-central-2': '欧洲（苏黎世）',
  'me-south-1': '中东（巴林）',
  'me-central-1': '中东（阿联酋）',
  'sa-east-1': '南美洲（圣保罗）',
};

export const INSTANCE_CHARGE_MAP = {
  PREPAID: '包年包月',
  POSTPAID_BY_HOUR: '按量计费',
  CDHPAID: '专用宿主机付费',
  SPOTPAID: '竞价实例',
};

export const NET_CHARGE_MAP = {
  BANDWIDTH_PREPAID: '按带宽包年包月计费',
  TRAFFIC_POSTPAID_BY_HOUR: '按流量计费',
  BANDWIDTH_POSTPAID_BY_HOUR: '按带宽使用时长计费',
  BANDWIDTH_PACKAGE: '按带宽包计费',
};

export const SITE_TYPE_MAP = {
  china: '中国站',
  international: '国际站',
};

export const LANGUAGE_TYPE = {
  zh_cn: 'zh_cn',
  en: 'en',
};

export const SEARCH_VALUE_IDS = [
  'cloud_id', // 云ID
];
