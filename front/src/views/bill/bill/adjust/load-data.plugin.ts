import { RouteLocationNormalizedLoaded } from 'vue-router';

import { BillSearchRules, injectBizField } from '@/utils';
import { QueryRuleOPEnum, RulesItem } from '@/typings';
import { BILL_BIZS_KEY, BILL_MAIN_ACCOUNTS_KEY } from '@/constants';
import { billAdjustColumns } from './columns';

export const getColumns = () => {
  const columns = billAdjustColumns.slice();
  injectBizField(columns, 3);

  return columns;
};

export const mountedCallback = (route: RouteLocationNormalizedLoaded, reloadTable: (rules: RulesItem[]) => void) => {
  // 只有业务、二级账号有保存的需求
  const billSearchRules = new BillSearchRules();
  billSearchRules
    .addRule(route, BILL_BIZS_KEY, 'bk_biz_id', QueryRuleOPEnum.IN)
    .addRule(route, BILL_MAIN_ACCOUNTS_KEY, 'main_account_id', QueryRuleOPEnum.IN);
  reloadTable(billSearchRules.rules);
};
