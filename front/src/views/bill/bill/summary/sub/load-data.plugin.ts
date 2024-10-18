import { RouteLocationNormalizedLoaded } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { useBusinessMapStore } from '@/store/useBusinessMap';

import { billsMainAccountSummaryColumns } from './columns';
import { QueryRuleOPEnum, RulesItem } from '@/typings';
import { BillSearchRules } from '@/utils';
import { BILL_BIZS_KEY, BILL_MAIN_ACCOUNTS_KEY } from '@/constants';

export const getColumns = () => {
  const { t } = useI18n();
  const businessMapStore = useBusinessMapStore();

  const columns = billsMainAccountSummaryColumns.slice();
  columns.splice(4, 0, {
    label: t('业务名称'),
    field: 'bk_biz_id',
    render: ({ data }: any) => businessMapStore.businessMap.get(data.bk_biz_id) || '未分配',
  });

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
