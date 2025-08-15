export enum ActiveQueryKey {
  DETAILS = 'details_active',
}

export enum ClbDetailsTabKey {
  LISTENER = 'listener',
  INFO = 'info',
  SECURITY = 'security',
}

export enum LoadBalancerActionType {
  PURCHASE = 'purchase',
  BATCH_OPERATION = 'batch_operation',
  CREATE_LISTENER_OR_RULES = 'create_listener_or_rules',
  BIND_RS = 'bind_rs',
  REMOVE = 'remove',
  SYNC = 'sync',
  COPY = 'copy',
  BATCH_EXPORT = 'batch_export',
}

export enum ListenerActionType {
  ADD = 'add',
  REMOVE = 'remove',
  SYNC = 'sync',
  BATCH_EXPORT = 'batch_export',
}

export type ResourceActionType = LoadBalancerActionType | ListenerActionType;

export enum LoadBalancerType {
  OPEN = 'OPEN',
  INTERNAL = 'INTERNAL',
}
export const LB_TYPE_NAME = {
  [LoadBalancerType.OPEN]: '公网',
  [LoadBalancerType.INTERNAL]: '内网',
};

export enum IpVersionType {
  IPV4 = 'ipv4',
  IPV6 = 'ipv6',
  IPV6_NAT64 = 'ipv6_nat64',
  IPV6_DUAL_STACK = 'ipv6_dual_stack',
}
export const IP_VERSION_DISPLAY_NAME = {
  [IpVersionType.IPV4]: 'IPv4',
  [IpVersionType.IPV6]: 'IPv6',
  [IpVersionType.IPV6_NAT64]: 'IPv6_NAT',
  [IpVersionType.IPV6_DUAL_STACK]: 'IPv6双栈',
};

export enum ClbStatusType {
  RUNNING = 1,
  CREATING = 0,
}
export const CLB_STATUS_NAME = {
  [ClbStatusType.RUNNING]: '正常运行',
  [ClbStatusType.CREATING]: '创建中',
};

export enum LoadBalancerIsp {
  CMCC = 'CMCC',
  CUCC = 'CUCC',
  CTCC = 'CTCC',
  BGP = 'BGP',
  INTERNAL = 'INTERNAL',
}
export const LOAD_BALANCER_ISP_NAME = {
  [LoadBalancerIsp.CMCC]: '中国移动',
  [LoadBalancerIsp.CUCC]: '中国联通',
  [LoadBalancerIsp.CTCC]: '中国电信',
  [LoadBalancerIsp.BGP]: 'BGP',
  [LoadBalancerIsp.INTERNAL]: '内网流量',
};

export enum LoadBalancerBatchImportOperationType {
  create_layer4_listener = 'create_layer4_listener',
  create_layer7_listener = 'create_layer7_listener',
  create_layer7_rule = 'create_layer7_rule',
  binding_layer4_rs = 'binding_layer4_rs',
  binding_layer7_rs = 'binding_layer7_rs',
}

export enum ListenerProtocol {
  TCP = 'TCP',
  UDP = 'UDP',
  HTTP = 'HTTP',
  HTTPS = 'HTTPS',
}
export const LISTENER_PROTOCOL_NAME = {
  [ListenerProtocol.TCP]: 'TCP',
  [ListenerProtocol.UDP]: 'UDP',
  [ListenerProtocol.HTTP]: 'HTTP',
  [ListenerProtocol.HTTPS]: 'HTTPS',
};
export const LAYER_4_LISTENER_PROTOCOL = [ListenerProtocol.TCP, ListenerProtocol.UDP];
export const LAYER_7_LISTENER_PROTOCOL = [ListenerProtocol.HTTP, ListenerProtocol.HTTPS];
export const LISTENER_PROTOCOL_LIST = [...LAYER_4_LISTENER_PROTOCOL, ...LAYER_7_LISTENER_PROTOCOL];

export enum Scheduler {
  WRR = 'WRR',
  LEAST_CONN = 'LEAST_CONN',
  IP_HASH = 'IP_HASH',
}
export const SCHEDULER_NAME = {
  [Scheduler.WRR]: '按权重轮询',
  [Scheduler.LEAST_CONN]: '最小连接数',
  [Scheduler.IP_HASH]: 'IP Hash',
};
export const SCHEDULER_LIST = [Scheduler.WRR, Scheduler.LEAST_CONN, Scheduler.IP_HASH];

export enum BindingStatusType {
  BINDING = 'binding',
  SUCCESS = 'success',
  FAILED = 'failed',
  UNBINDING = 'unbinding',
}
export const BINDING_STATUS_NAME = {
  [BindingStatusType.BINDING]: '绑定中',
  [BindingStatusType.SUCCESS]: '已绑定',
  [BindingStatusType.FAILED]: '绑定失败',
  [BindingStatusType.UNBINDING]: '未绑定',
};

export enum SessionType {
  NORMAL = 'NORMAL',
  QUIC_CID = 'QUIC_CID',
}
export const SESSION_TYPE_NAME = {
  [SessionType.NORMAL]: '基于源 IP',
  [SessionType.QUIC_CID]: '基于源端口',
};

export enum SSLMode {
  UNIDIRECTIONAL = 'UNIDIRECTIONAL',
  MUTUAL = 'MUTUAL',
}
export const SSL_MODE_NAME = {
  [SSLMode.UNIDIRECTIONAL]: '单向认证',
  [SSLMode.MUTUAL]: '双向认证',
};

export enum RsInstType {
  CVM = 'CVM',
  ENI = 'ENI',
}
