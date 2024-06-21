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
