import { VendorEnum } from '@/common/constant';
import { LoadBalancerBatchImportOperationType } from '@/views/load-balancer/constants';
import { IQueryResData } from '@/typings';

export interface ILoadBalancerBatchImportModel {
  account_id: string;
  vendor: VendorEnum;
  region_ids: string[];
  operation_type: LoadBalancerBatchImportOperationType;
}

export enum Status {
  executable = 'executable', // 可执行
  not_executable = 'not_executable', // 不可执行
  existing = 'existing', // 已存在
}
// 负载均衡批量导入预览 - 预览item
interface ILoadBalancerBatchImportPreviewBaseItem {
  clb_vip_domain: string;
  cloud_clb_id: string;
  protocol: string;
  listener_port: number[];
  user_remark: string;
  status: Status; // 校验结果状态
  validate_result: string[]; // 参数校验详情, 当状态为不可执行时, 会有具体的报错原因
}
interface ICreateLayer4ListenerPreviewItem extends ILoadBalancerBatchImportPreviewBaseItem {
  scheduler: string;
  session: number;
  health_check: boolean;
  user_remark: string;
}
interface ICreateLayer7ListenerPreviewItem extends ILoadBalancerBatchImportPreviewBaseItem {
  ssl_mode: string;
  cert_cloud_ids: string[];
  ca_cloud_id: string;
}
interface ICreateUrlRulePreviewItem extends ILoadBalancerBatchImportPreviewBaseItem {
  domain: string;
  default_domain: boolean;
  url_path: string;
  scheduler: string;
  session: number;
  health_check: boolean;
}
interface ILayer4ListenerBindRsPreviewItem extends ILoadBalancerBatchImportPreviewBaseItem {
  inst_type: string;
  rs_ip: string;
  rs_port: number;
  weight: number;
}
interface ILayer7ListenerBindRsPreviewItem extends ILoadBalancerBatchImportPreviewBaseItem {
  domain: string;
  url_path: string;
  inst_type: string;
  rs_ip: string;
  rs_port: number;
  weight: number;
}
type LoadBalancerBatchImportPreviewItem =
  | ICreateLayer4ListenerPreviewItem
  | ICreateLayer7ListenerPreviewItem
  | ICreateUrlRulePreviewItem
  | ILayer4ListenerBindRsPreviewItem
  | ILayer7ListenerBindRsPreviewItem;

// 负载均衡批量导入预览 - 预览list
export type LoadBalancerBatchImportPreviewDetails = Array<LoadBalancerBatchImportPreviewItem>;

// 负载均衡批量导入预览 - 响应体
export interface ILoadBalancerImportPreview {
  details: LoadBalancerBatchImportPreviewDetails;
}
export type LoadBalancerImportPreviewResData = IQueryResData<ILoadBalancerImportPreview>;
