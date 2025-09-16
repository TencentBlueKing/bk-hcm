import { VendorEnum } from '@/common/constant';

export enum DeviceTabEnum {
  LISTENER = 'listener',
  URL = 'url',
  RS = 'rs',
}

export interface ICount {
  listenerCount: number;
  urlCount: number;
  rsCount: number;
}

// 条件类型
export interface ILoadBalanceDeviceCondition {
  vendor: VendorEnum;
  account_id: string;
  lb_regions?: string[];
  lb_network_types?: string[];
  lb_ip_versions?: string[];
  cloud_lb_ids?: string[];
  lb_vips?: string[];
  lb_domains?: string[];
  lbl_protocols?: string[];
  lbl_ports?: number[];
  rule_domains?: string[];
  rule_urls?: string[];
  target_ips?: string[];
  target_ports?: number[];
  [key: string]: any;
}
