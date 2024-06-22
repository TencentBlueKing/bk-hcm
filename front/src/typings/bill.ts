import { IListResData } from './common';

// 当月账单总金额
export interface BillsSummary {
  bill_year: number;
  bill_month: number;
  vendor: string;
  currency: string;
  current_month_cost: number;
  current_month_rmb_cost: number;
  current_month_sync_cost: number;
  current_month_rmb_sync_cost: number;
  product_count: number;
}
export type BillsSummaryResData = IListResData<BillsSummary[]>;

// 当月账单汇总（一级账号）拉取接口
export interface BillsRootAccountSummary {
  root_account_id: string;
  root_account_name: string;
  vendor: string;
  state: string;
  currency: string;
  last_synced_version: number;
  current_version: number;
  last_month_cost_synced: number;
  last_month_rmb_cost_synced: number;
  current_month_cost_synced: number;
  current_month_rmb_cost_synced: number;
  month_on_month_value: string;
  current_month_cost: number;
  current_month_rmb_cost: number;
  adjustment_cost: number;
  adjustment_rmb_cost: number;
  created_at: string;
  updated_at: string;
}
export type BillsRootAccountSummaryResData = IListResData<BillsRootAccountSummary[]>;

// 当月账单汇总历史版本（一级账号）拉取接口
export interface BillsRootAccountSummaryHistory {
  root_account_id: string;
  root_account_name: string;
  state: string;
  currency: string;
  version: number;
  month_cost: number;
  month_rmb_cost: number;
  adjustment_cost: number;
  adjustment_rmb_cost: number;
  created_at: string;
  updated_at: string;
}
export type BillsRootAccountSummaryHistoryResData = IListResData<BillsRootAccountSummaryHistory[]>;

// 当月账单汇总（二级账号or运营产品）拉取接口
export interface BillsMainAccountSummary {
  root_account_id: string;
  root_account_name: string;
  main_account_id: string;
  main_account_name: string;
  vendor: string;
  product_id: number;
  product_name: string;
  bk_biz_id: number;
  bk_biz_name: string;
  state: string;
  currency: string;
  last_synced_version: number;
  current_version: number;
  last_month_cost_synced: number;
  last_month_rmb_cost_synced: number;
  current_month_cost_synced: number;
  current_month_rmb_cost_synced: number;
  month_on_month_value: string;
  current_month_cost: number;
  current_month_rmb_cost: number;
  adjustment_cost: number;
  adjustment_rmb_cost: number;
  created_at: string;
  updated_at: string;
}
export type BillsMainAccountSummaryResData = IListResData<BillsMainAccountSummary[]>;

// 调账明细
export interface AdjustmentItem {
  id?: string; // 调账id
  main_account_id: string; // 所属主账号id
  vendor: string; // 云厂商
  product_id: number; // 运营产品id
  bk_biz_id?: number; // 业务id
  bill_year: number; // 所属年份
  bill_month: number; // 所属月份
  bill_day: number; // 所属日期
  type: 'increase' | 'decrease'; // 调账类型 枚举值（increase、decrease）
  currency: string; // 币种
  cost: string; // 金额
  rmb_cost: string; // 对应人民币金额
  memo?: string; // 备注信息
}
