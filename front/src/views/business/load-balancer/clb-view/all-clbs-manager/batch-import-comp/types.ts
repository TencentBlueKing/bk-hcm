import { VendorEnum } from '@/common/constant';
import { IQueryResData } from '@/typings';

export enum Action {
  CREATE_LISTENER_OR_URL_RULE,
  BIND_RS,
}

export enum Operation {
  create_layer4_listener = 'create_layer4_listener',
  create_layer7_listener = 'create_layer7_listener',
  create_layer7_rule = 'create_layer7_rule',
  binding_layer4_rs = 'binding_layer4_rs',
  binding_layer7_rs = 'binding_layer7_rs',
}

// 负载均衡批量导入 - 基本信息
export interface LbBatchImportBaseInfo {
  account_id: string;
  vendor: VendorEnum;
  region_ids: string[];
  operation_type: Operation;
}

export enum Status {
  executable = 'executable', // 可执行
  not_executable = 'not_executable', // 不可执行
  existing = 'existing', // 已存在
}

// 负载均衡批量导入预览 - 预览item
interface BaseLbBatchImportPreviewItem {
  clb_vip_domain: string;
  cloud_clb_id: string;
  protocol: string;
  listener_port: number[];
  user_remark: string;
  status: Status; // 校验结果状态
  validate_result: string[]; // 参数校验详情, 当状态为不可执行时, 会有具体的报错原因
}
interface CreateLayer4ListenerPreviewItem extends BaseLbBatchImportPreviewItem {
  scheduler: string;
  session: number;
  health_check: boolean;
  user_remark: string;
}
interface CreateLayer7ListenerPreviewItem extends BaseLbBatchImportPreviewItem {
  ssl_mode: string;
  cert_cloud_ids: string[];
  ca_cloud_id: string;
}
interface CreateUrlRulePreviewItem extends BaseLbBatchImportPreviewItem {
  domain: string;
  default_domain: boolean;
  url_path: string;
  scheduler: string;
  session: number;
  health_check: boolean;
}
interface Layer4ListenerBindRsPreviewItem extends BaseLbBatchImportPreviewItem {
  inst_type: string;
  rs_ip: string;
  rs_port: number;
  weight: number;
}
interface Layer7ListenerBindRsPreviewItem extends BaseLbBatchImportPreviewItem {
  domain: string;
  url_path: string;
  inst_type: string;
  rs_ip: string;
  rs_port: number;
  weight: number;
}
type LbBatchImportPreviewItem =
  | CreateLayer4ListenerPreviewItem
  | CreateLayer7ListenerPreviewItem
  | CreateUrlRulePreviewItem
  | Layer4ListenerBindRsPreviewItem
  | Layer7ListenerBindRsPreviewItem;

// 负载均衡批量导入预览 - 预览list
export type LbBatchImportPreviewDetails = Array<LbBatchImportPreviewItem>;

// 负载均衡批量导入预览 - 响应体
export interface LbImportPreview {
  details: LbBatchImportPreviewDetails;
}
export type LbImportPreviewResData = IQueryResData<LbImportPreview>;
