import { IQueryResData } from './common';

export type IOriginPage = 'lb' | 'listener' | 'domain';

// 腾讯云账号的负载均衡配额名称
export enum CLB_QUOTA_NAME {
  // 用户当前地域下的公网CLB配额
  TOTAL_OPEN_CLB_QUOTA = 'TOTAL_OPEN_CLB_QUOTA',
  // 用户当前地域下的内网CLB配额
  TOTAL_INTERNAL_CLB_QUOTA = 'TOTAL_INTERNAL_CLB_QUOTA',
  // 一个CLB下的监听器配额
  TOTAL_LISTENER_QUOTA = 'TOTAL_LISTENER_QUOTA',
  // 一个监听器下的转发规则配额
  TOTAL_LISTENER_RULE_QUOTA = 'TOTAL_LISTENER_RULE_QUOTA',
  // 一条转发规则下可绑定设备的配额
  TOTAL_TARGET_BIND_QUOTA = 'TOTAL_TARGET_BIND_QUOTA',
  // 一个CLB实例下跨地域2.0的SNAT IP配额
  TOTAL_SNAP_IP_QUOTA = 'TOTAL_SNAP_IP_QUOTA',
  // 用户当前地域下的三网CLB配额
  TOTAL_ISP_CLB_QUOTA = 'TOTAL_ISP_CLB_QUOTA',
}

// 腾讯云账号的负载均衡配额信息
export interface ClbQuota {
  // 配额名称
  quota_id: CLB_QUOTA_NAME;
  // 当前使用数量，为 null 时表示无意义
  quota_current: number;
  // 配额数量
  quota_limit: number;
}

// response - 腾讯云账号的负载均衡配额信息
export type ClbQuotasResp = IQueryResData<ClbQuota[]>;

// 负载均衡价格
export interface LbPrice {
  // 网络价格信息，对于标准账户，网络在cvm上计费，该选项为空
  bandwidth_price: LbPriceItem;
  // 实例价格信息
  instance_price: LbPriceItem;
  // lcu 价格信息
  lcu_price: LbPriceItem;
}

// 负载均衡价格项
interface LbPriceItem {
  // 后续计价单元，HOUR、GB
  charge_unit: string;
  // 折扣 ，如20.0代表2折
  discount: number;
  // 预支费用的折扣价，单位：元
  discount_price: number;
  // 预支费用的原价，单位：元
  original_price: number;
  // 后付费单价，单位：元
  unit_price: number;
  // 后付费的折扣单价，单位:元
  unit_price_discount: number;
}

// response - 负载均衡价格信息
export type LbPriceInquiryResp = IQueryResData<LbPrice>;

// 异步任务Flow详情
export interface AsyncTaskDetail {
  id: string;
  name: string;
  state: string;
  reason: string;
  creator: string;
  reviser: string;
  created_at: string;
  updated_at: string;
}

// response - 异步任务Flow详情
export type AsyncTaskDetailResp = IQueryResData<AsyncTaskDetail>;

// CLB 规格类型
export type CLBSpecType =
  | 'shared'
  | 'clb.c1.small'
  | 'clb.c2.medium'
  | 'clb.c3.small'
  | 'clb.c3.medium'
  | 'clb.c4.small'
  | 'clb.c4.medium'
  | 'clb.c4.large'
  | 'clb.c4.xlarge';

// 协议类型
export type Protocol = 'TCP' | 'UDP' | 'HTTP' | 'HTTPS';
