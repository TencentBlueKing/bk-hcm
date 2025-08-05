import { NetworkAccountType } from '@/api/load_balancers/apply-clb/types';
import { ConstantMapRecord } from '@/typings';

export const BGP_VIP_ISP_TYPES: string[] = ['BGP'];

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
    value: 'BANDWIDTH_PREPAID',
    // 云平台当前API接口暂不支持包月参数
    isDisabled: () => true,
    tipsContent: '云平台当前API接口暂不支持包月参数',
  },
  {
    label: '按流量',
    value: 'TRAFFIC_POSTPAID_BY_HOUR',
    isDisabled: (vipIsp: string) => !BGP_VIP_ISP_TYPES.includes(vipIsp),
    tipsContent: '仅支持BGP线路',
  },
  {
    label: '按带宽',
    value: 'BANDWIDTH_POSTPAID_BY_HOUR',
    isDisabled: (vipIsp: string) => !BGP_VIP_ISP_TYPES.includes(vipIsp),
    tipsContent: '仅支持BGP线路',
  },
  {
    label: '共享带宽包',
    value: 'BANDWIDTH_PACKAGE',
    isDisabled: () => false,
  },
];

// 支持IPv6 NAT64的地域
export const WHITE_LIST_REGION_IPV6_NAT64 = ['ap-beijing', 'ap-shanghai'];
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
export const SESSION_TYPE_MAP: ConstantMapRecord = {
  NORMAL: '基于源 IP ',
  QUIC_CID: '基于源端口',
};

// 均衡方式映射 - 反向映射
export const SCHEDULER_REVERSE_MAP: ConstantMapRecord = {
  按权重轮询: 'WRR',
  最小连接数: 'LEAST_CONN',
  IP_HASH: 'IP_HASH',
};

// 负载均衡网络类型映射
export const LB_NETWORK_TYPE_MAP: ConstantMapRecord = {
  OPEN: '公网',
  INTERNAL: '内网',
};

// 负载均衡网络类型映射 - 反向映射
export const LB_NETWORK_TYPE_REVERSE_MAP: ConstantMapRecord = {
  公网: 'OPEN',
  内网: 'INTERNAL',
};

// 腾讯云负载均衡状态映射
export const CLB_STATUS_MAP: ConstantMapRecord = {
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
export const CLB_SPEC_TYPE_COLUMNS_MAP: Record<
  string,
  {
    connectionsPerMinute?: number;
    newConnectionsPerSecond?: number;
    queriesPerSecond?: number;
    bandwidthLimit: number;
  }
> = {
  shared: { bandwidthLimit: 10240 },
  'clb.c1.small': {
    connectionsPerMinute: 100000,
    newConnectionsPerSecond: 10000,
    queriesPerSecond: 10000,
    bandwidthLimit: 1024,
  },
  'clb.c2.medium': {
    connectionsPerMinute: 100000,
    newConnectionsPerSecond: 10000,
    queriesPerSecond: 10000,
    bandwidthLimit: 2048,
  },
  'clb.c3.small': {
    connectionsPerMinute: 200000,
    newConnectionsPerSecond: 20000,
    queriesPerSecond: 20000,
    bandwidthLimit: 4096,
  },
  'clb.c3.medium': {
    connectionsPerMinute: 500000,
    newConnectionsPerSecond: 50000,
    queriesPerSecond: 30000,
    bandwidthLimit: 6144,
  },
  'clb.c4.small': {
    connectionsPerMinute: 1000000,
    newConnectionsPerSecond: 100000,
    queriesPerSecond: 50000,
    bandwidthLimit: 10240,
  },
  'clb.c4.medium': {
    connectionsPerMinute: 2000000,
    newConnectionsPerSecond: 200000,
    queriesPerSecond: 100000,
    bandwidthLimit: 20480,
  },
  'clb.c4.large': {
    connectionsPerMinute: 5000000,
    newConnectionsPerSecond: 500000,
    queriesPerSecond: 200000,
    bandwidthLimit: 40960,
  },
  'clb.c4.xlarge': {
    connectionsPerMinute: 10000000,
    newConnectionsPerSecond: 1000000,
    queriesPerSecond: 300000,
    bandwidthLimit: 61440,
  },
};

// 监听器同步状态映射 - 反向映射
export const LISTENER_BINDING_STATUS_REVERSE_MAP: ConstantMapRecord = {
  绑定中: 'binding',
  已绑定: 'success',
};

export enum TargetGroupOperationScene {
  ADD = 'add',
  EDIT = 'edit',
  BATCH_DELETE = 'batch_delete',
  ADD_RS = 'add_rs',
  BATCH_ADD_RS = 'batch_add_rs',
  BATCH_DELETE_RS = 'batch_delete_rs',
  SINGLE_UPDATE_PORT = 'single_update_port',
  SINGLE_UPDATE_WEIGHT = 'single_update_weight',
  BATCH_UPDATE_PORT = 'batch_update_port',
  BATCH_UPDATE_WEIGHT = 'batch_update_weight',
}
// 编辑目标组操作场景映射
export const TG_OPERATION_SCENE_MAP = {
  [TargetGroupOperationScene.ADD]: '新增目标组',
  [TargetGroupOperationScene.EDIT]: '编辑目标组基本信息',
  [TargetGroupOperationScene.BATCH_DELETE]: '批量删除目标组',
  [TargetGroupOperationScene.ADD_RS]: '添加RS',
  [TargetGroupOperationScene.BATCH_ADD_RS]: '批量添加RS',
  [TargetGroupOperationScene.BATCH_DELETE_RS]: '批量删除RS',
  [TargetGroupOperationScene.SINGLE_UPDATE_PORT]: '修改单个端口',
  [TargetGroupOperationScene.SINGLE_UPDATE_WEIGHT]: '修改单个端口',
  [TargetGroupOperationScene.BATCH_UPDATE_PORT]: '批量修改端口',
  [TargetGroupOperationScene.BATCH_UPDATE_WEIGHT]: '批量修改权重',
};

// IP版本映射 - 前端展示使用
export const IP_VERSION_MAP: ConstantMapRecord = {
  ipv4: 'IPv4',
  ipv6: 'IPv6',
  ipv6_nat64: 'IPv6_NAT',
  ipv6_dual_stack: 'IPv6双栈',
};

// 运营商类型
export const ISP_TYPES = ['BGP', 'CTCC', 'CUCC', 'CMCC'];

// 安全组放通模式
export const LOAD_BALANCER_PASS_TO_TARGET_LIST = [
  {
    label: '启用默认放通',
    description: '启用后，CLB 和 CVM 之间默认放通，来自 CLB 的流量，仅通过 CLB 上安全组的校验',
    value: true,
  },
  {
    label: '不启用默认放通',
    description: '不启用，来自 CLB 的流量，需同时通过 CLB 和 CVM 上安全组的校验',
    value: false,
  },
];

// 带宽包状态
export enum BANDWIDTH_PACKAGE_STATUS {
  CREATING = 'CREATING',
  CREATED = 'CREATED',
  DELETING = 'DELETING',
  DELETED = 'DELETED',
}

// 带宽包类型映射
export const BANDWIDTH_PACKAGE_NETWORK_TYPE_MAP: Record<string, string> = {
  BGP: '普通BGP共享带宽包',
  HIGH_QUALITY_BGP: '精品BGP共享带宽包',
  SINGLEISP_CMCC: '中国移动共享带宽包',
  SINGLEISP_CTCC: '中国电信共享带宽包',
  SINGLEISP_CUCC: '中国联通共享带宽包',
  SINGLEISP: '单线',
  ANYCAST: 'ANYCAST加速',
};

// 带宽包计费类型映射
export const BANDWIDTH_PACKAGE_CHARGE_TYPE_MAP: Record<string, string> = {
  TOP5_POSTPAID_BY_MONTH: '按月后付费TOP5计费',
  PERCENT95_POSTPAID_BY_MONTH: '按月后付费月95计费',
  ENHANCED95_POSTPAID_BY_MONTH: '按月后付费增强型95计费',
  FIXED_PREPAID_BY_MONTH: '包月预付费计费',
  PEAK_BANDWIDTH_POSTPAID_BY_DAY: '后付费日结按带宽计费',
};

// 负载均衡运营商和带宽包网络类型映射
export const LOADBALANCER_BANDWIDTH_PACKAGE_NETWORK_TYPES_MAP: Record<string, string[]> = {
  BGP: ['BGP'],
  CMCC: ['SINGLEISP', 'SINGLEISP_CMCC'],
  CTCC: ['SINGLEISP', 'SINGLEISP_CTCC'],
  CUCC: ['SINGLEISP', 'SINGLEISP_CUCC'],
};
