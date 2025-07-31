import { ConstantMapRecord } from '@/typings';

// 全局业务id
export const GLOBAL_BIZS_KEY = 'bizs';
export const GLOBAL_BIZS_VERSION = '1.6.3';
export const GLOBAL_BIZS_VERSION_KEY = 'bizs_version';

// 账号校验接口类型
export enum AccountVerifyEnum {
  ROOT = 'root_accounts',
  ACCOUNT = 'accounts',
}
export enum VendorEnum {
  TCLOUD = 'tcloud',
  AWS = 'aws',
  AZURE = 'azure',
  GCP = 'gcp',
  HUAWEI = 'huawei',
  ZENLAYER = 'zenlayer',
  KAOPU = 'kaopu',
  OTHER = 'other',
}

export enum ResourceTypeEnum {
  CVM = 'cvm',
  VPC = 'vpc',
  DISK = 'disk',
  SUBNET = 'subnet',
  CLB = 'clb',
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
  {
    name: '负载均衡',
    type: 'clb',
  },
  {
    name: '证书管理',
    type: 'certs',
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
  {
    id: 'other',
    name: '其他云厂商',
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
  //   移除 ID 搜索条件
  // {
  //   name: 'ID',
  //   id: 'id',
  // },
  // 资源ID需给出对应的提示文案, 如主机ID, VPC ID
  // {
  //   name: '资源ID',
  //   id: 'cloud_id',
  // },
  {
    name: '名称',
    id: 'name',
    meta: {
      search: {
        filterRules: () => ({}),
      },
    },
  },
  {
    name: '云厂商',
    id: 'vendor',
    children: VENDORS,
    async: false,
    meta: {
      search: {
        filterRules: () => ({}),
      },
    },
  },
  {
    name: '云账号ID',
    id: 'account_id',
    children: [],
    meta: {
      search: {
        filterRules: () => ({}),
      },
    },
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

export const CIDRDATARANGE = {
  10: { min: 0, max: 255 },
  172: { min: 16, max: 31 },
  192: { min: 168, max: 168 },
};

export const TCLOUDCIDRMASKRANGE = {
  10: { min: 12, max: 28 },
  172: { min: 12, max: 28 },
  192: { min: 16, max: 28 },
};

export const CIDRMASKRANGE = {
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

export const CLOUD_HOST_STATUS: ConstantMapRecord = {
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
  '': '未获取',
};

export const CLOUD_AREA_REGION_GCP: ConstantMapRecord = {
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

export const CLOUD_AREA_REGION_GCP_EN = {
  蒙特利尔: 'northamerica-northeast1',
  多伦多: 'northamerica-northeast2',
  圣保罗: 'southamerica-east1',
  圣地亚哥: 'southamerica-west1',
  爱荷华: 'us-central1',
  南卡罗来纳: 'us-east1',
  北弗吉尼亚: 'us-east4',
  哥伦布: 'us-east5',
  达拉斯: 'us-south1',
  俄勒冈: 'us-west1',
  洛杉矶: 'us-west2',
  盐湖城: 'us-west3',
  拉斯维加斯: 'us-west4',
  华沙: 'europe-central2',
  芬兰: 'europe-north1',
  马德里: 'europe-southwest1',
  比利时: 'europe-west1',
  都灵: 'europe-west12',
  伦敦: 'europe-west2',
  法兰克福: 'europe-west3',
  荷兰: 'europe-west4',
  苏黎世: 'europe-west6',
  米兰: 'europe-west8',
  巴黎: 'europe-west9',
  Doha: 'me-central1',
  特拉维夫: 'me-west1',
  台湾: 'asia-east1',
  香港: 'asia-east2',
  东京: 'asia-northeast1',
  大阪: 'asia-northeast2',
  首尔: 'asia-northeast3',
  孟买: 'asia-south1',
  德里: 'asia-south2',
  新加坡: 'asia-southeast1',
  雅加达: 'asia-southeast2',
  悉尼: 'australia-southeast1',
  墨尔本: 'australia-southeast2',
};

export const CLOUD_AREA_REGION_AWS: ConstantMapRecord = {
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

export const CLOUD_AREA_REGION_AWS_EN = {
  'US East (Ohio)': 'us-east-2',
  '美国东部（弗吉尼亚北部）': 'us-east-1',
  '美国西部（加利福尼亚北部）': 'us-west-1',
  '美国西部（俄勒冈）': 'us-west-2',
  'Africa (Cape Town)': 'af-south-1',
  'Asia Pacific (Hong Kong)': 'ap-east-1',
  '亚太地区（海得拉巴）': 'ap-south-2',
  '亚太地区（雅加达）': 'ap-southeast-3',
  '亚太地区（墨尔本）': 'ap-southeast-4',
  'Asia Pacific (Mumbai)': 'ap-south-1',
  'Asia Pacific (Osaka)': 'ap-northeast-3',
  'Asia Pacific (Seoul)': 'ap-northeast-2',
  '亚太地区（新加坡）': 'ap-southeast-1',
  '亚太地区（悉尼）': 'ap-southeast-2',
  '亚太区域（东京）': 'ap-northeast-1',
  'Canada (Central)': 'ca-central-1',
  'Europe (Frankfurt)': 'eu-central-1',
  '欧洲（爱尔兰）': 'eu-west-1',
  'Europe (London)': 'eu-west-2',
  'Europe (Milan)': 'eu-south-1',
  'Europe (Paris)': 'eu-west-3',
  '欧洲（西班牙）': 'eu-south-2',
  '欧洲（斯德哥尔摩）': 'eu-north-1',
  '欧洲（苏黎世）': 'eu-central-2',
  '中东（巴林）': 'me-south-1',
  '中东（阿联酋）': 'me-central-1',
  '南美洲（圣保罗）': 'sa-east-1',
};

export const INSTANCE_CHARGE_MAP: ConstantMapRecord = {
  PREPAID: '包年包月',
  POSTPAID_BY_HOUR: '按量计费',
  CDHPAID: '专用宿主机付费',
  SPOTPAID: '竞价实例',
};

export const NET_CHARGE_MAP: ConstantMapRecord = {
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
  zh_cn: 'zh-cn',
  en: 'en',
};

export const SEARCH_VALUE_IDS = [
  'cloud_id', // 云ID
];

export const RESOURCE_TABS = [
  {
    key: '/resource/resource/',
    label: '资源管理',
  },
  {
    key: '/resource/resource/account',
    label: '账号信息',
  },
  {
    key: '/resource/resource/record',
    label: '操作记录',
  },
  {
    key: '/resource/resource/recycle',
    label: '回收站',
  },
];

export const RESOURCE_DETAIL_TABS = [
  {
    key: '/resource/resource/account/detail',
    label: '基本信息',
  },
  {
    key: '/resource/resource/account/resource',
    label: '资源状态',
  },
  {
    key: '/resource/resource/account/manage',
    label: '用户列表',
  },
];

export const RESOURCE_TYPES_MAP = {
  disk: '硬盘',
  vpc: 'VPC',
  subnet: '子网',
  eip: '弹性IP',
  security_group: '安全组',
  cvm: '主机',
  route_table: '路由表',
  sub_account: '云账号',
  account: '账号',
  gcp_firewall_rule: '防火墙规则',
  route: '路由表',
  network_interface: '网络接口',
  region: '地域',
  image: '镜像',
  zone: '可用区',
  azure_resource_group: '微软云资源组',
  argument_template: '参数模板',
  cert: '证书',
  load_balancer: '负载均衡',
  security_group_usage_biz_rel: '安全组使用业务',
  cvm_cc_info: '主机资产数据',
};

export const RESOURCES_SYNC_STATUS_MAP = {
  sync_success: '已同步',
  sync_failed: '同步失败',
  syncing: '同步中',
};

export enum SECURITY_GROUP_RULE_TYPE {
  INGRESS = 'ingress',
  EGRESS = 'egress',
}

export const VendorMap: Record<string, string> = {
  [VendorEnum.AWS]: '亚马逊云',
  [VendorEnum.AZURE]: '微软云',
  [VendorEnum.GCP]: '谷歌云',
  [VendorEnum.HUAWEI]: '华为云',
  [VendorEnum.TCLOUD]: '腾讯云',
  [VendorEnum.ZENLAYER]: 'Zenlayer',
  [VendorEnum.KAOPU]: '靠谱云',
  [VendorEnum.OTHER]: '其他云厂商',
};

export const VendorReverseMap: ConstantMapRecord = {
  亚马逊云: VendorEnum.AWS,
  微软云: VendorEnum.AZURE,
  谷歌云: VendorEnum.GCP,
  华为云: VendorEnum.HUAWEI,
  腾讯云: VendorEnum.TCLOUD,
  Zenlayer: VendorEnum.ZENLAYER,
  靠谱云: VendorEnum.KAOPU,
  其他云厂商: VendorEnum.OTHER,
};

export const SYNC_STAUS_MAP = {
  a: '绑定中',
  b: '成功',
  c: '失败',
  d: '部分成功',
};

export const TARGET_GROUP_PROTOCOLS = ['TCP', 'UDP', 'HTTP', 'HTTPS'];

export const LB_TYPE_MAP: ConstantMapRecord = {
  OPEN: '公网',
  INTERNAL: '内网',
};

export const CHARGE_TYPE: ConstantMapRecord = {
  PREPAID: '包年包月',
  POSTPAID_BY_HOUR: '按量计费',
};

export const LB_ISP: ConstantMapRecord = {
  CMCC: '中国移动',
  CUCC: '中国联通',
  CTCC: '中国电信',
  BGP: 'BGP',
  INTERNAL: '内网流量',
};

export const CLB_SPECS: ConstantMapRecord = {
  'clb.c1.small': '简约型',
  'clb.c2.medium': '标准型规格',
  'clb.c3.small': '高阶型1规格',
  'clb.c3.medium': '高阶型2规格',
  'clb.c4.small': '超强型1规格',
  'clb.c4.medium': '超强型2规格',
  'clb.c4.large': '超强型3规格',
  'clb.c4.xlarge': '超强型4规格',
};

export const CLB_BINDING_STATUS: ConstantMapRecord = {
  binding: '绑定中',
  success: '已绑定',
  failed: '未绑定',
};
