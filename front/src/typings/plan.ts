export interface IDemandListDetail {
  demand_id: string; // CRP需求ID
  bk_biz_id: number; // 业务ID
  bk_biz_name: string; // 业务名称
  op_product_id: number; // 运营产品ID
  op_product_name: string; // 运营产品名称
  status: string; // 需求状态，可选值：can_apply（可申请), not_ready（未到申请时间), expired（已过期), spent_all（已耗尽), locked（变更中)
  status_name: string; // 需求状态名称
  demand_class: string; // 预期需求类型
  available_year_month: string; // 需求年月
  expect_time: string; // 期望交付日期
  return_plan_time: string; // 短租返还日期
  device_class: string; // 机型类型
  device_type: string; // 机型
  total_os: string; // 总OS数量
  applied_os: string; // 已申请OS数量
  remained_os: string; // 剩余OS数量
  total_cpu_core: number; // 总CPU核数
  applied_cpu_core: number; // 已申请CPU核数
  remained_cpu_core: number; // 剩余CPU核数
  total_memory: number; // 总内存大小
  applied_memory: number; // 已申请内存大小
  remained_memory: number; // 剩余内存大小
  total_disk_size: number; // 总云盘大小
  applied_disk_size: number; // 已申请云盘大小
  remained_disk_size: number; // 剩余云盘大小
  region_id: string; // 地区ID
  region_name: string; // 地区/楼层名称
  zone_id: string; // 可用区ID
  zone_name: string; // 可用区名称
  plan_type: string; // 预测类型
  obs_project: string; // OBS项目类型
  generation_type: string; // 机型类型
  device_family: string; // 机型族
  disk_type: string; // 云盘类型
  disk_type_name: string; // 云盘类型名称
  disk_io: number; // 云盘IO
  adjustType?: AdjustType;
  demand_source: string; // 变更原因
  res_mode: string; // 资源模式
  delay_os?: string;
  demand_res_type: string;
}

export enum AdjustType {
  config = 'update', // 修改配置
  time = 'delay', // 修改时间
  none = 'none', // 未做修改
}

export interface IListConfigCvmChargeTypeDeviceTypeParams {
  bk_biz_id: number; // CC业务ID
  require_type: number; // 需求类型。1: 常规项目; 2: 春节保障; 3: 机房裁撤
  region: string; // 地域
  zone?: string; // 可用区，若为空则查询地域下所有可用区支持的机型
}

export interface DeviceType {
  device_type: string; // 机型
  available: boolean; // 是否可用
}

export interface Info {
  charge_type: string; // 计费模式 (PREPAID:包年包月, POSTPAID_BY_HOUR:按量计费)
  available: boolean; // 是否可用
  device_types: DeviceType[]; // 机型配置信息列表
}

export interface IListConfigCvmChargeTypeDeviceTypeData {
  count: number; // 当前规则匹配到的总记录条数
  info: Info[]; // 机型详情列表
}

// 规格参数
export interface IDemandSpec {
  region: string; // 地域
  zone: string; // 可用区
  zones: string[]; // 可用区
  device_type: string; // 机型
  image_id: string; // 镜像ID
  disk_size: number; // 数据盘盘大小，单位G
  disk_type: string; // 数据盘盘类型, "CLOUD_SSD":SSD云硬盘, "CLOUD_PREMIUM":高性能云盘
  network_type: string; // 网络类型, "ONETHOUSAND":千兆, "TENTHOUSAND":万兆
  vpc: string; // 私有网络, 默认为空
  subnet: string; // 子网, 默认为空
  charge_type: string; // 计费模式(PREPAID:包年包月, POSTPAID_BY_HOUR:按量计费), 默认:包年包月
  charge_months: number; // 计费模式为包年包月时，该字段必传
}

// 子订单参数
export interface IDemandSuborder {
  resource_type: string; // 资源类型, "QCLOUDCVM":腾讯云虚拟机, "IDCPM":IDC物理机, "QCLOUDDVM":Qcloud富容器, "IDCDVM":IDC富容器
  replicas: number; // 需求实例数量
  anti_affinity_level?: string; // 反亲和策略, 默认值为"ANTI_NONE", "ANTI_NONE":无要求, "ANTI_CAMPUS":分Campus, "ANTI_MODULE":分Module, "ANTI_RACK":分机架
  remark?: string; // 备注
  spec: IDemandSpec; // 资源需求声明
}

// 输入参数
export interface IVerifyResourceDemandParams {
  bk_biz_id: number; // CC业务ID
  require_type: number; // 需求类型。1:常规项目; 2:春节保障; 3:机房裁撤
  suborders: IDemandSuborder[]; // 资源申请子需求信息
}

export interface IDemandVerification {
  verify_result: string; // 预测校验结果，PASS:通过，FAILED:未通过，NOT_INVOLVED:不涉及
  reason: string; // 预测校验结果原因
}

export interface IVerifyResourceDemandData {
  verifications: IDemandVerification[]; // 资源申请子需求单的预测校验信息，校验信息顺序与需求入参的资源申请子需求单顺序一致
}

// CVM 信息
export interface CVMInfo {
  res_mode: string; // 资源模式 (枚举值: 按机器模式, 按机型族)
  device_type: string; // 机型规格
  os: number; // OS数，单位: 台
  cpu_core: number; // CPU核数，单位: 核
  memory: number; // 内存大小，单位: GB
}

// CBS 信息
export interface CBSInfo {
  disk_type: string; // 云盘类型 (枚举值: CLOUD_PREMIUM(高性能云硬盘), CLOUD_SSD(SSD云硬盘))
  disk_io: number; // 磁盘IO配额，单位: MB/s
  disk_size: number; // 云盘大小，单位: GB
}

// 调整信息
export interface AdjustInfo {
  obs_project: string; // 归属项目
  expect_time: string; // 期望交付时间，格式为YYYY-MM-DD，例如2024-01-01
  region_id: string; // 地区/城市ID
  zone_id?: string; // 可用区ID (可选)
  demand_res_types: string[]; // 预期资源类型列表 (枚举值: CVM, CBS)
  cvm?: CVMInfo; // 申请CVM的信息 (可选)
  cbs?: CBSInfo; // 申请CBS的信息 (可选)
  demand_source: string; // 变更原因（默认为指标变化）
  remark: string; // 备注
}

// 调整项
export interface IAdjust {
  demand_id: string; // CRP需求ID
  adjust_type: string; // 调整类型 (枚举值: update (常规修改), delay (加急延期))
  demand_source: string; // 调整来源
  original_info: AdjustInfo; // 原始信息
  updated_info: AdjustInfo; // 更新后的信息
  expect_time?: string; // 期望交付时间，adjust_type为delay时必填，格式为YYYY-MM-DD，例如2024-01-01 (可选)
  delay_reason?: string; // 延期原因，adjust_type为delay时必填 (可选)
  delay_os?: string; // 延期OS数
}

// 数据
export interface IAdjustParams {
  adjusts: IAdjust[]; // 调整列表
}

export interface IAdjustData {
  id: string; // 预测单据ID
}

export interface IYearMonthWeek {
  year: number; // 需要年
  month: number; // 需要月
  week: number; // 需要月内的第几周
}

export interface IDateRange {
  start: string; // 起始时间, 不能晚于当前时间，格式为YYYY-MM-DD，例如2024-01-01
  end: string; // 结束时间, 不能早于start时间，格式为YYYY-MM-DD，例如2024-01-01
}

export interface IExceptTimeRange {
  year_month_week: IYearMonthWeek; // 需求年月周
  date_range_in_week: IDateRange; // 需求年月周天范围
  date_range_in_month: IDateRange; // 需求年月天范围
}

export enum ChargeType {
  PREPAID = 'PREPAID',
  POSTPAID_BY_HOUR = 'POSTPAID_BY_HOUR',
}

export const ChargeTypeMap = {
  [ChargeType.PREPAID]: '包年包月',
  [ChargeType.POSTPAID_BY_HOUR]: '按量计费',
};

export interface ITimeRange {
  start: string;
  end: string;
}
