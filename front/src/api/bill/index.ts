import { VendorEnum } from '@/common/constant';
import http from '@/http';
import { FilterType, IPageQuery } from '@/typings';
import {
  AdjustmentItem,
  BillImportPreviewItems,
  BillImportPreviewResData,
  BillsExportReqParams,
  BillsExportReqParamsWithBizs,
  BillsExportResData,
  BillsMainAccountSummaryResData,
  BillsRootAccountSummaryHistoryResData,
  BillsRootAccountSummaryResData,
  BillsSummarySumReqParams,
  BillsSummaryListReqParams,
  BillsSummaryListReqParamsWithBizs,
  BillsBizSummaryResData,
  BillsSummarySumResData,
} from '@/typings/bill';

const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

// 获取当月筛选出来的一级账号总金额
export const reqBillsRootAccountSummarySum = async (
  data: BillsSummarySumReqParams,
): Promise<BillsSummarySumResData> => {
  return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/account/bills/root-account-summarys/sum`, data);
};

// 获取当月筛选出来的二级账号总金额
export const reqBillsMainAccountSummarySum = async (
  data: BillsSummarySumReqParams,
): Promise<BillsSummarySumResData> => {
  return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/account/bills/main-account-summarys/sum`, data);
};

// 拉取当月一级账号账单汇总
export const reqBillsRootAccountSummaryList = async (
  data: BillsSummaryListReqParams,
): Promise<BillsRootAccountSummaryResData> => {
  return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/account/bills/root-account-summarys/list`, data);
};

// 拉取当月账单汇总历史版本（一级账号）
export const reqBillsRootAccountHistorySummaryList = async (
  data: BillsSummaryListReqParams,
): Promise<BillsRootAccountSummaryHistoryResData> => {
  return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/account/bills/root-account-summary-historys/list`, data);
};

// 拉取当月二级账号账单汇总
export const reqBillsMainAccountSummaryList = async (
  data: BillsSummaryListReqParams,
): Promise<BillsMainAccountSummaryResData> => {
  return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/account/bills/main-account-summarys/list`, data);
};

// 拉取当月业务账单汇总
export const reqBillsBizSummaryList = async (
  data: BillsSummaryListReqParamsWithBizs,
): Promise<BillsBizSummaryResData> => {
  return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/account/bills/biz_summarys/list`, data);
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

// 查询汇率
export const reqBillsExchangeRateList = async (data: { filter: FilterType; page: IPageQuery }) => {
  return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/account/bills/exchange_rates/list`, data);
};

// 账单明细-导入预览
export const billItemsImportPreview = async (
  vendor: VendorEnum,
  data: { bill_year: number; bill_month: number; excel_file_base64: string },
): Promise<BillImportPreviewResData> => {
  return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/account/vendors/${vendor}/bills/items/import/preview`, data);
};

// 账单明细-导入
export const billItemsImport = async (
  vendor: VendorEnum,
  data: { bill_year: number; bill_month: number; items: BillImportPreviewItems },
) => {
  return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/account/vendors/${vendor}/bills/items/import`, data);
};

// 导出一级账号账单汇总数据
export const exportBillsRootAccountSummary = async (data: BillsExportReqParams): Promise<BillsExportResData> => {
  return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/account/bills/root-account-summarys/export`, data);
};

// 导出二级账号账单汇总数据
export const exportBillsMainAccountSummary = async (data: BillsExportReqParams): Promise<BillsExportResData> => {
  return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/account/bills/main-account-summarys/export`, data);
};

// 导出业务账单汇总数据
export const exportBillsBizSummary = async (data: BillsExportReqParamsWithBizs): Promise<BillsExportResData> => {
  return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/account/bills/biz_summarys/export`, data);
};

// 导出账单明细数据
export const exportBillsItems = async (vendor: VendorEnum, data: BillsExportReqParams): Promise<BillsExportResData> => {
  return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/account/vendors/${vendor}/bills/items/export`, data);
};

// 导出账单调整数据
export const exportBillsAdjustmentItems = async (data: BillsExportReqParams): Promise<BillsExportResData> => {
  return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/account/bills/adjustment_items/export`, data);
};
