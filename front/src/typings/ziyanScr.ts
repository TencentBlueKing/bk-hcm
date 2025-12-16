import type { IPageQuery, IListResData, IQueryResData } from '@/typings/common';
import { CvmDataDiskType } from '@/views/ziyanScr/components/cvm-data-disk/constants';
export interface IRecycleArea {
  id: string;
  name: string;
  start_time: string;
  end_time: string;
  which_stages: number;
  recycle_type: number;
  creator: string;
  reviser: string;
  created_at: string;
  updated_at: string;
}

export interface IQueryDissolveList {
  group_ids: number[];
  bk_biz_names: string[];
  module_names: string[];
  operators: string[];
}

export interface IDissolve {
  bk_biz_id: number | string;
  bk_biz_name: string;
  module_host_count: { [key: string]: number };
  total: {
    current: {
      host_count: number | string;
      cpu_count: number;
    };
    origin: {
      host_count: number | string;
      cpu_count: number;
    };
    delivered_core: number;
  };
  progress: string;
  [key: string]: any;
}

export type IDissolveList = IQueryResData<{ items: IDissolve[] }>;

export interface IDissolveHostOriginListParam {
  organizations?: string[];
  bk_biz_names?: string[];
  module_names: string[];
  operators?: string[];
  page?: IPageQuery;
}

export interface originListResult {
  server_asset_id: string;
  ip: string;
  bk_host_outerip: string;
  app_name: string;
  module: string;
  device_type: string;
  module_name: string;
  idc_unit_name: string;
  sfw_name_version: string;
  go_up_date: string;
  raid_id: number;
  logic_area: string;
  server_operator: string;
  server_bak_operator: string;
  device_layer: string;
  cpu_score: number;
  mem_score: number;
  inner_net_traffic_score: number;
  disk_io_score: number;
  disk_util_score: number;
  is_pass: string;
  mem4linux: number;
  inner_net_traffic: number;
  outer_net_traffic: number;
  disk_io: number;
  disk_util: number;
  disk_total: number;
  max_cpu_core_amount: number;
  group_name: string;
  center: string;
}

export type IDissolveHostOriginListResult = IListResData<originListResult[]>;

export interface IDissolveHostCurrentListParam {
  organizations?: string[];
  group_names?: string[];
  bk_biz_names?: string[];
  module_names: string[];
  operators?: string[];
  page?: IPageQuery;
}

export interface CurrentListParam {
  server_asset_id: string;
  ip: string;
  bk_host_outerip: string;
  app_name: string;
  module: string;
  device_type: string;
  module_name: string;
  idc_unit_name: string;
  sfw_name_version: string;
  go_up_date: string;
  raid_id: number;
  logic_area: string;
  server_operator: string;
  server_bak_operator: string;
  device_layer: string;
  cpu_score: number;
  mem_score: number;
  inner_net_traffic_score: number;
  disk_io_score: number;
  disk_util_score: number;
  is_pass: string;
  mem4linux: number;
  inner_net_traffic: number;
  outer_net_traffic: number;
  disk_io: number;
  disk_util: number;
  disk_total: number;
  max_cpu_core_amount: number;
  group_name: string;
  center: string;
  svr_type_name: string;
}

export type IDissolveHostCurrentListResult = IListResData<CurrentListParam[]>;

export interface IDissolveRecycledModuleListParam {
  op: 'and' | 'or';
  rules: {
    field: string;
    op: 'eq' | 'neq' | 'gt' | 'gte' | 'lt' | 'lte' | 'in' | 'nin' | 'cs' | 'cis';
    value: boolean | number | string | (boolean | number | string)[];
  }[];
}

export interface IApplyCrpTicketAudit {
  crp_ticket_id: string;
  crp_ticket_link: string;
  logs: Array<{
    task_no: number;
    task_name: string;
    operate_result: string;
    operator: string;
    operate_info: string;
    operate_time: string;
  }>;
  current_step: {
    current_task_no: number;
    current_task_name: string;
    status: number;
    status_desc: string;
    fail_instance_info: Array<{
      error_msg_type_en: string;
      error_type: string;
      error_msg_type_cn: string;
      request_id: string;
      error_msg: string;
      operator: string;
      error_count: number;
    }>;
  };
}
export type IApplyCrpTicketAuditLogItem = IApplyCrpTicketAudit['logs'][number];
export type IApplyCrpTicketAuditCurrentStepItem = IApplyCrpTicketAudit['current_step'];
export type IApplyCrpTicketAuditFailInfoItem = IApplyCrpTicketAudit['current_step']['fail_instance_info'][number];

export interface ICvmDeviceDetailItem {
  id: number;
  require_type: number;
  region: string;
  zone: string;
  device_type: string;
  cpu: number;
  mem: number;
  disk: number;
  remark: string;
  label: {
    device_group: string;
    device_size: string;
  };
  capacity_flag: number;
  enable_capacity: boolean;
  enable_apply: boolean;
  score: number;
  comment: string;
}

export interface ITaskApplyRecordInitItem {
  suborder_id: string;
  ip: string;
  task_id: string;
  task_link: string;
  status: number;
  message: string;
  create_at: string;
  update_at: string;
  start_at: string;
  end_at: string;
}

export interface IApplyOrderItem {
  order_id: number;
  suborder_id: string;
  bk_biz_id: number;
  bk_username: string;
  require_type: number;
  resource_type: string;
  expect_time: string;
  description: string;
  remark: string;
  spec: {
    region: string;
    zone: string;
    zones: string[];
    device_group: string;
    device_size: string;
    device_type: string;
    image_id: string;
    image: string;
    disk_size: number;
    disk_type: string;
    network_type: string;
    vpc: string;
    subnet: string;
    os_type: string;
    raid_type: string;
    isp: string;
    mount_path: string;
    cpu_provider: string;
    kernel: string;
    charge_type: string;
    charge_months: number;
    inherit_instance_id: string;
    failed_zone_ids: string[];
    res_assign: number;
  };
  anti_affinity_level: string;
  enable_disk_check: boolean;
  stage: string;
  status: string;
  origin_num: number;
  total_num: number;
  success_num: number;
  pending_num: number;
  product_num: number;
  modify_time: number;
  create_at: string;
  update_at: string;
}

export interface ICloudInstanceConfigItem {
  zone: string;
  instance_type: string;
  instance_charge_type: string;
  network_card: number;
  externals: {
    unsupport_networks: string[];
    storage_block_attr: {
      max_size: number;
      min_size: number;
      type: CvmDataDiskType;
    };
  };
  cpu: number;
  memory: number;
  instance_family: string;
  type_name: string;
  local_disk_type_list: {
    max_size: number;
    min_size: number;
    partition_type: string;
    required: string;
    type: string;
  }[];
  status: string;
  instance_bandwidth: number;
  instance_pps: number;
  storage_block_amount: number;
  cpu_type: string;
  gpu: number;
  fpga: number;
  remark: string;
  gpu_count: number;
  frequency: string;
  status_category: string;
}
