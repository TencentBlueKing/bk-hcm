import { NetworkAccountType } from '@/api/load_balancers/apply-clb/types';

// 负载均衡-路由组件名称
export enum LBRouteName {
  allLbs = 'all-lbs-manager',
  lb = 'specific-lb-manager',
  listener = 'specific-listener-manager',
  domain = 'specific-domain-manager',
  allTgs = 'all-tgs-manager',
  tg = 'specific-tg-manager',
}
// 负载均衡-路由组件名称映射
export const LB_ROUTE_NAME_MAP = {
  all: 'all-lbs-manager',
  lb: 'specific-lb-manager',
  listener: 'specific-listener-manager',
  domain: 'specific-domain-manager',
};

// 网络类型
export const LOAD_BALANCER_TYPE = [
  {
    label: '公网',
    value: 'OPEN',
  },
  {
    label: '内网',
    value: 'INTERNAL',
  },
];
// IP版本
export const ADDRESS_IP_VERSION = [
  {
    label: 'IPv4',
    value: 'IPV4',
  },
  {
    label: 'IPv6',
    value: 'IPv6FullChain',
  },
  {
    label: 'IPv6 NAT64',
    value: 'IPV6',
    isDisabled: (region: string) => !WHITE_LIST_REGION_IPV6_NAT64.includes(region),
  },
];
// 可用区类型
export const ZONE_TYPE = [
  {
    label: '单可用区',
    value: '0',
  },
  {
    label: '主备可用区',
    value: '1',
    isDisabled: (region: string, accountType: NetworkAccountType) =>
      !WHITE_LIST_REGION_PRIMARY_STAND_ZONE.includes(region) || accountType !== 'STANDARD',
  },
];
// 网络计费模式
export const INTERNET_CHARGE_TYPE = [
  {
    label: '包月',
    value: undefined,
  },
  {
    label: '按流量',
    value: 'TRAFFIC_POSTPAID_BY_HOUR',
  },
  {
    label: '按带宽',
    value: 'BANDWIDTH_POSTPAID_BY_HOUR',
  },
  {
    label: '共享带宽包',
    value: 'BANDWIDTH_PACKAGE',
  },
];

// 支持IPv6 NAT64的地域
export const WHITE_LIST_REGION_IPV6_NAT64 = ['ap-beijing', 'ap-shanghai', 'ap-guangzhou'];
// 支持主备可用区的地域
export const WHITE_LIST_REGION_PRIMARY_STAND_ZONE = [
  'ap-guangzhou',
  'ap-shanghai',
  'ap-nanjing',
  'ap-beijing',
  'ap-hongkong',
  'ap-seoul',
];

// 会话类型映射
export const SESSION_TYPE_MAP = {
  NORMAL: '基于源 IP ',
  QUIC_CID: '基于源端口',
};

// 证书认证方式映射
export const SSL_MODE_MAP = {
  UNIDIRECTIONAL: '单向认证',
  MUTUAL: '双向认证',
};

// 均衡方式映射
export const SCHEDULER_MAP = {
  WRR: '按权重轮询',
  LEAST_CONN: '最小连接数',
  IP_HASH: 'IP Hash',
};
// 均衡方式映射 - 反向映射
export const SCHEDULER_REVERSE_MAP = {
  按权重轮询: 'WRR',
  最小连接数: 'LEAST_CONN',
  IP_HASH: 'IP_HASH',
};

// 传输层协议, 如 TCP, UDP
export const TRANSPORT_LAYER_LIST = ['TCP', 'UDP'];
// 应用层协议, 如 HTTP, HTTPS
export const APPLICATION_LAYER_LIST = ['HTTP', 'HTTPS'];

// 负载均衡网络类型映射
export const LB_NETWORK_TYPE_MAP = {
  OPEN: '公网',
  INTERNAL: '内网',
};

// 负载均衡网络类型映射 - 反向映射
export const LB_NETWORK_TYPE_REVERSE_MAP = {
  公网: 'OPEN',
  内网: 'INTERNAL',
};

// 腾讯云负载均衡状态映射
export const CLB_STATUS_MAP = {
  '1': '正常运行',
  '0': '创建中',
};

// 负载均衡规格映射 - 反向映射
export const CLB_SPECS_REVERSE_MAP = {
  简约型: 'clb.c1.small',
  标准型规格: 'clb.c2.medium',
  高阶型1规格: 'clb.c3.small',
  高阶型2规格: 'clb.c3.medium',
  超强型1规格: 'clb.c4.small',
  超强型2规格: 'clb.c4.medium',
  超强型3规格: 'clb.c4.large',
  超强型4规格: 'clb.c4.xlarge',
};

// 腾讯云CLB规格列表映射
export const CLB_SPEC_TYPE_COLUMNS_MAP = {
  'clb.c1.small': {
    connectionsPerMinute: '100,000',
    newConnectionsPerSecond: '10,000',
    queriesPerSecond: '10,000',
    bandwidthLimit: '1Gbps',
  },
  'clb.c2.medium': {
    connectionsPerMinute: '100,000',
    newConnectionsPerSecond: '10,000',
    queriesPerSecond: '10,000',
    bandwidthLimit: '2Gbps',
  },
  'clb.c3.small': {
    connectionsPerMinute: '200,000',
    newConnectionsPerSecond: '20,000',
    queriesPerSecond: '20,000',
    bandwidthLimit: '4Gbps',
  },
  'clb.c3.medium': {
    connectionsPerMinute: '500,000',
    newConnectionsPerSecond: '50,000',
    queriesPerSecond: '30,000',
    bandwidthLimit: '6Gbps',
  },
  'clb.c4.small': {
    connectionsPerMinute: '1,000,000',
    newConnectionsPerSecond: '100,000',
    queriesPerSecond: '50,000',
    bandwidthLimit: '10Gbps',
  },
  'clb.c4.medium': {
    connectionsPerMinute: '2,000,000',
    newConnectionsPerSecond: '200,000',
    queriesPerSecond: '100,000',
    bandwidthLimit: '20Gbps',
  },
  'clb.c4.large': {
    connectionsPerMinute: '5,000,000',
    newConnectionsPerSecond: '500,000',
    queriesPerSecond: '200,000',
    bandwidthLimit: '40Gbps',
  },
  'clb.c4.xlarge': {
    connectionsPerMinute: '10,000,000',
    newConnectionsPerSecond: '1,000,000',
    queriesPerSecond: '300,000',
    bandwidthLimit: '60Gbps',
  },
};

// 监听器同步状态映射 - 反向映射
export const LISTENER_BINDING_STATUS_REVERSE_MAP = {
  绑定中: 'binding',
  已绑定: 'success',
};

// 编辑目标组操作场景映射
export const TG_OPERATION_SCENE_MAP = {
  add: '新增目标组',
  edit: '编辑目标组基本信息',
  BatchDelete: '批量删除目标组',
  AddRs: '添加RS',
  BatchAddRs: '批量添加RS',
  BatchDeleteRs: '批量删除RS',
  port: '批量修改端口',
  weight: '批量修改权重',
};

// IP版本映射 - 前端展示使用
export const IP_VERSION_MAP = {
  ipv4: 'IPv4',
  ipv6: 'IPv6',
  ipv6_nat64: 'IPv6_NAT',
  ipv6_dual_stack: 'IPv6双栈',
};

// 运营商类型
export const ISP_TYPES = ['BGP', 'CTCC', 'CUCC', 'CMCC'];

// 带宽包状态
export enum BANDWIDTH_PACKAGE_STATUS {
  CREATING = 'CREATING',
  CREATED = 'CREATED',
  DELETING = 'DELETING',
  DELETED = 'DELETED',
}

// 带宽包类型映射
export const BANDWIDTH_PACKAGE_NETWORK_TYPE_MAP = {
  BGP: '普通BGP共享带宽包',
  HIGH_QUALITY_BGP: '精品BGP共享带宽包',
  SINGLEISP_CMCC: '中国移动共享带宽包',
  SINGLEISP_CTCC: '中国电信共享带宽包',
  SINGLEISP_CUCC: '中国联通共享带宽包',
  SINGLEISP: '单线',
  ANYCAST: 'ANYCAST加速',
};

// 带宽包计费类型映射
export const BANDWIDTH_PACKAGE_CHARGE_TYPE_MAP = {
  TOP5_POSTPAID_BY_MONTH: '按月后付费TOP5计费',
  PERCENT95_POSTPAID_BY_MONTH: '按月后付费月95计费',
  ENHANCED95_POSTPAID_BY_MONTH: '按月后付费增强型95计费',
  FIXED_PREPAID_BY_MONTH: '包月预付费计费',
  PEAK_BANDWIDTH_POSTPAID_BY_DAY: '后付费日结按带宽计费',
};
