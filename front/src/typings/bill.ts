import { IListResData, IPageQuery, IQueryResData } from './common';
import { FilterType } from './resource';

// 账单汇总
export type BillsSummarySumReqParams = { bill_year: number; bill_month: number; filter: FilterType };
type BillsSummaryListBaseReqParams = { bill_year: number; bill_month: number; page: IPageQuery };
export type BillsSummaryListReqParams = BillsSummaryListBaseReqParams & { filter: FilterType };
export type BillsSummaryListReqParamsWithBizs = BillsSummaryListBaseReqParams & { bk_biz_id: number[] };

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
  id: string;
  root_account_id: string;
  root_account_name: string;
  vendor: string;
  bill_year: number;
  bill_month: number;
  last_synced_version: number;
  current_version: number;
  currency: string;
  last_month_cost_synced: string;
  last_month_rmb_cost_synced: string;
  current_month_cost_synced: string;
  current_month_rmb_cost_synced: string;
  month_on_month_value: number;
  current_month_cost: string;
  current_month_rmb_cost: string;
  adjustment_cost: string;
  adjustment_rmb_cost: string;
  rate: number;
  state: BillsRootAccountSummaryState;
  product_num: number;
  created_at: string;
  updated_at: string;
}
// 一级账号账单汇总状态
export enum BillsRootAccountSummaryState {
  accounting = 'accounting',
  accounted = 'accounted',
  confirmed = 'confirmed',
  syncing = 'syncing',
  synced = 'synced',
  stopped = 'stopped',
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

// 当月账单汇总（二级账号）拉取接口
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

// 当月账单汇总（业务）拉取接口
interface BillsBizSummary {
  bk_biz_id: number;
  bk_biz_name: string;
  last_month_cost_synced: string;
  last_month_rmb_cost_synced: string;
  current_month_cost_synced: string;
  current_month_rmb_cost_synced: string;
  current_month_cost: string;
  current_month_rmb_cost: string;
  adjustment_cost: string;
  adjustment_rmb_cost: string;
}
export type BillsBizSummaryResData = IListResData<BillsBizSummary[]>;

// 调账明细
export interface AdjustmentItem {
  id?: string; // 调账id
  main_account_id: string; // 所属主账号id
  vendor: string; // 云厂商
  product_id: string | number; // 业务id
  bk_biz_id?: string | number; // 业务id
  bill_year: number; // 所属年份
  bill_month: number; // 所属月份
  bill_day: number; // 所属日期
  type: 'increase' | 'decrease'; // 调账类型 枚举值（increase、decrease）
  currency: string; // 币种
  cost: string; // 金额
  rmb_cost: string; // 对应人民币金额
  memo?: string; // 备注信息
}

// 账单汇总总金额
export interface BillsSummarySum {
  count: number;
  cost_map: CostMap;
}
export interface CostMap {
  USD: USD;
}
interface USD {
  Cost: string;
  RMBCost: string;
  Currency: string;
}
export type BillsSummarySumResData = IQueryResData<BillsSummarySum>;

// 账单明细-zenlayer导入预览
export interface BillImportPreview {
  items: BillImportPreviewItem[];
  cost_map: CostMap;
}
interface BillImportPreviewItem {
  root_account_id: string;
  main_account_id: string;
  vendor: string;
  product_id: number;
  bk_biz_id: number;
  bill_year: number;
  bill_month: number;
  bill_day: number;
  version_id: number;
  currency: string;
  cost: string;
  res_amount: string;
  extension: BillImportPreviewItemExtension;
}
interface BillImportPreviewItemExtension {
  bill_id: string;
  zenlayer_order: string;
  cid: string;
  group_id: string;
  currency: string;
  city: string;
  pay_content: string;
  type: string;
  acceptance_num: string;
  pay_num: string;
  unit_price_usd: string;
  total_payable: string;
  billing_period: string;
  contract_period: string;
  remarks: string;
  business_group: string;
  cpu: null | string;
  disk: null | string;
  memory: null | string;
}
export type BillImportPreviewResData = IQueryResData<BillImportPreview>;
export type BillImportPreviewItems = BillImportPreviewItem[];

// 账单导出
type BillsExportBaseReqParams = { bill_year: number; bill_month: number; export_limit: number };
export type BillsExportReqParams = BillsExportBaseReqParams & { filter: FilterType };
export type BillsExportReqParamsWithBizs = BillsExportBaseReqParams & { bk_biz_ids: number[] };
export type BillsExportResData = IQueryResData<{ download_url: string }>;
