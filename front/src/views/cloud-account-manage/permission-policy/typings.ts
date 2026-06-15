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
  version?: number; // 当前版本号
  bk_biz_ids?: number[]; // 允许使用的业务ID列表
  memo: string; // 描述
  creator?: string;
  reviser?: string;
  created_at?: string;
  updated_at?: string;
  policy_document?: string; // 当前版本的策略JSON内容
  policy_hash?: string; // 当前版本的策略HASH
  associated_account_count?: number; // 关联二级账号数
  related_accounts?: string[]; // 关联二级账号列表
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
