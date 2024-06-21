import http from '@/http';
import { FilterType, IPageQuery } from '@/typings';
import {
  BillsMainAccountSummaryResData,
  BillsRootAccountSummaryHistoryResData,
  BillsRootAccountSummaryResData,
} from '@/typings/bill';

const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

// 获取当月所有一级账号总金额（分vendor）
export const reqBillsSummaryList = async (data: { bill_year: number; bill_month: number }) => {
  return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/account/bills/summarys/list`, data);
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

// 拉取当月二级账号或者运营产品账单汇总
export const reqBillsMainAccountSummaryList = async (data: {
  bill_year: number;
  bill_month: number;
  filter: FilterType;
  page: IPageQuery;
}): Promise<BillsMainAccountSummaryResData> => {
  return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/account/bills/main-account-summarys/list`, data);
};
