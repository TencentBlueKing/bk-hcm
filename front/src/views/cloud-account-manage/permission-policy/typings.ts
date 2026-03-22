/**
 * 权限策略库相关类型定义
 */

// 关联二级账号信息
export interface IRelatedAccount {
  account_id: string; // 二级账号ID
  alias?: string; // 二级账号别名
}

// 权限策略库列表项
export interface IPermissionPolicyItem {
  id: string;
  name: string; // 策略库名称
  related_account_count?: number; // 关联二级账号数
  related_accounts?: IRelatedAccount[]; // 关联二级账号列表
  description: string; // 策略库描述
  creator?: string; // 创建人
  created_at?: string; // 创建时间
  reviser?: string; // 更新人
  updated_at?: string; // 更新时间
  vendor?: string; // 云厂商
  bk_biz_id?: number; // 业务ID
  usage_biz_ids?: number[]; // 允许使用的业务ID
  json?: string; // 权限策略库JSON
}

// 操作类型枚举
export enum ApplyOperationType {
  APPLY_NEW = 'apply_new', // 应用到新账号
  UPDATE_APPLIED = 'update_applied', // 更新已应用账号
}

// 可选二级账号项（用于"应用到新账号"表格）
export interface ISelectableAccount {
  account_id: string; // 二级账号ID
  alias: string; // 二级账号别名
}

// 策略库应用状态枚举
export enum PolicyApplyStatus {
  APPLIED = 'applied', // 已应用
  PENDING = 'pending', // 待应用
  DATA_MISMATCH = 'data_mismatch', // 已应用(数据不一致)
}

// 已应用账号项（用于"更新已应用账号"表格）
export interface IAppliedAccountItem {
  account_id: string; // 二级账号ID
  alias: string; // 二级账号别名
  cloud_template_name: string; // 云上模版名称
  cloud_sync_time: string; // 云模版同步时间
  applied_version: string; // 策略库应用版本
  apply_status: PolicyApplyStatus; // 策略库应用状态
  apply_time: string; // 策略库应用时间
}
