import { RouteLocationNormalizedLoaded } from 'vue-router';

import { billsMainAccountSummaryColumns } from './columns';
import { QueryRuleOPEnum, RulesItem } from '@/typings';
import { BillSearchRules, injectBizField } from '@/utils';
import { BILL_BIZS_KEY, BILL_MAIN_ACCOUNTS_KEY } from '@/constants';

export const getColumns = () => {
  const columns = billsMainAccountSummaryColumns.slice();
  injectBizField(columns, 4);

  return columns;
};

// mounted 时, 根据初始条件加载表格数据
export const mountedCallback = (route: RouteLocationNormalizedLoaded, reloadTable: (rules: RulesItem[]) => void) => {
  // 只有业务、二级账号有保存的需求
  const billSearchRules = new BillSearchRules();
  billSearchRules
    .addRule(route, BILL_BIZS_KEY, 'bk_biz_id', QueryRuleOPEnum.IN)
    .addRule(route, BILL_MAIN_ACCOUNTS_KEY, 'main_account_id', QueryRuleOPEnum.IN);
  reloadTable(billSearchRules.rules);
};
