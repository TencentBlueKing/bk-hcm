import { VendorEnum } from '@/common/constant';
import http from '@/http';
import { FilterType, IPageQuery } from '@/typings';
import {
  AdjustmentItem,
  BillsMainAccountSummaryResData,
  BillsRootAccountSummaryHistoryResData,
  BillsRootAccountSummaryResData,
  BillsSummarySumResData,
} from '@/typings/bill';

const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

// 获取当月筛选出来的一级账号总金额
export const reqBillsRootAccountSummarySum = async (data: {
  bill_year: number;
  bill_month: number;
  filter: FilterType;
}): Promise<BillsSummarySumResData> => {
  return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/account/bills/root-account-summarys/sum`, data);
};

// 获取当月筛选出来的二级账号总金额
export const reqBillsMainAccountSummarySum = async (data: {
  bill_year: number;
  bill_month: number;
  filter: FilterType;
}): Promise<BillsSummarySumResData> => {
  return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/account/bills/main-account-summarys/sum`, data);
};

// 拉取当月一级账号账单汇总
export const reqBillsRootAccountSummaryList = async (data: {
  bill_year: number;
  bill_month: number;
  filter: FilterType;
  page: IPageQuery;
}): Promise<BillsRootAccountSummaryResData> => {
  return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/account/bills/root-account-summarys/list`, data);
};

// 拉取当月账单汇总历史版本（一级账号）
export const reqBillsRootAccountHistorySummaryList = async (data: {
  bill_year: number;
  bill_month: number;
  filter: FilterType;
  page: IPageQuery;
}): Promise<BillsRootAccountSummaryHistoryResData> => {
  return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/account/bills/root-account-summary-historys/list`, data);
};

// 拉取当月二级账号或者业务账单汇总
export const reqBillsMainAccountSummaryList = async (data: {
  bill_year: number;
  bill_month: number;
  filter: FilterType;
  page: IPageQuery;
}): Promise<BillsMainAccountSummaryResData> => {
  return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/account/bills/main-account-summarys/list`, data);
};

// 确认某个一级账号下所有账单数据
export const confirmBillsRootAccountSummary = async (data: {
  bill_year: number;
  bill_month: number;
  root_account_id: string;
}) => {
  return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/account/bills/root-account-summarys/confirm`, data);
};

// 重新核算某个一级账号下所有账单数据
export const reAccountBillsRootAccountSummary = async (data: {
  bill_year: number;
  bill_month: number;
  root_account_id: string;
}) => {
  return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/account/bills/root-account-summarys/reaccount`, data);
};

// 账单同步至 OBS
export const syncRecordsBills = async (data: { bill_year: number; bill_month: number; vendor: VendorEnum }) => {
  return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/account/bills/sync_records`, data);
};

// 查询当月账单明细
export const reqBillsItemList = async (data: {
  vendor: VendorEnum;
  bill_year: number;
  bill_month: number;
  filter: FilterType;
  page: IPageQuery;
}) => {
  return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/account/vendors/${data.vendor}/bills/items/list`, data);
};

// 批量创建调账明细
export const createBillsAdjustment = async (data: {
  root_account_id: string; // 所属根账号id
  items: AdjustmentItem[];
}) => {
  return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/account/bills/adjustment_items/create`, data);
};

// 查询调账明细
export const reqBillsAdjustmentList = async (data: { filter: FilterType; page: IPageQuery }) => {
  return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/account/bills/adjustment_items/list`, data);
};

// 编辑调账明细，已确定的调账明细不能编辑，该接口不能确认调账明细
export const updateBillsAdjustment = async (
  id: string,
  data: {
    root_account_id?: string; // 所属根账号id
    main_account_id?: string; // 所属主账号id
    vendor?: string; // 云厂商
    product_id?: number; // 业务id
    bk_biz_id?: number; // 业务id
    bill_year?: number; // 所属年份
    bill_month?: number; // 所属月份
    bill_day?: number; // 所属日期
    type?: 'increase' | 'decrease'; // 调账类型 枚举值（increase、decrease）
    currency?: string; // 币种
    cost?: string; // 金额
    rmb_cost?: string; // 对应人民币金额
    memo?: string; // 备注信息
  },
) => {
  return http.patch(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/account/bills/adjustment_items/${id}`, data);
};

// 确认调账明细，确认后不可再修改、删除。
export const confirmBillsAdjustment = async (ids: string[]) => {
  return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/account/bills/adjustment_items/confirm`, { ids });
};

// 批量删除调账明细，已确定的调账明细不能删除
export const deleteBillsAdjustment = async (ids: string[]) => {
  return http.delete(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/account/bills/adjustment_items/batch`, { data: { ids } });
};

// 账单同步记录查询接口
export const reqBillsSyncRecordList = async (data: { filter: FilterType; page: IPageQuery }) => {
  return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/account/bills/sync_records/list`, data);
};
