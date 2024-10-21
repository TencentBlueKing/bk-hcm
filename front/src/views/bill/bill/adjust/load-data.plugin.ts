import { useI18n } from 'vue-i18n';
import { RouteLocationNormalizedLoaded } from 'vue-router';
import { useBusinessMapStore } from '@/store/useBusinessMap';

import { BillSearchRules } from '@/utils';
import { QueryRuleOPEnum, RulesItem } from '@/typings';
import { BILL_BIZS_KEY, BILL_MAIN_ACCOUNTS_KEY } from '@/constants';
import { billAdjustColumns } from './columns';

export const getColumns = () => {
  const { t } = useI18n();
  const businessMapStore = useBusinessMapStore();

  const columns = billAdjustColumns.slice();
  columns.splice(5, 0, {
    label: t('业务名称'),
    field: 'bk_biz_id',
    render: ({ data }: any) => businessMapStore.businessMap.get(data.bk_biz_id) || '未分配',
  });

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
