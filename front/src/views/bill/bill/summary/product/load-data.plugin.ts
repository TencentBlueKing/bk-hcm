import { ref } from 'vue';
import { RouteLocationNormalizedLoaded } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { useBusinessMapStore } from '@/store/useBusinessMap';

import { billsProductSummaryColumns } from './columns';
import { reqBillsBizSummaryList } from '@/api/bill';
import { QueryRuleOPEnum, RulesItem } from '@/typings';
import { BillSearchRules } from '@/utils';
import { BILL_BIZS_KEY } from '@/constants';

export const getColumns = () => {
  const { t } = useI18n();
  const businessMapStore = useBusinessMapStore();

  const columns = billsProductSummaryColumns.slice();
  columns.splice(1, 0, {
    label: t('业务名称'),
    field: 'bk_biz_id',
    render: ({ data }: any) => businessMapStore.businessMap.get(data.bk_biz_id) || '未分配',
  });

  return columns;
};

export const apiMethod: (...args: any) => Promise<any> = reqBillsBizSummaryList;
export const extensionKey = 'bk_biz_ids';

// mounted 时, 根据初始条件加载表格数据
export const mountedCallback = (route: RouteLocationNormalizedLoaded, reloadTable: (rules: RulesItem[]) => void) => {
  // 只有业务有保存的需求
  const billSearchRules = new BillSearchRules();
  billSearchRules.addRule(route, BILL_BIZS_KEY, 'bk_biz_id', QueryRuleOPEnum.IN);
  reloadTable(billSearchRules.rules);
};

export const useSelectionIds = () => {
  const selectedIds = ref<number[]>([]);

  // reloadTable 时, 重置选中项
  const reloadSelectedIds = (rules: RulesItem[]) => {
    selectedIds.value = (rules.find((rule) => rule.field === 'bk_biz_id')?.value as number[]) || [];
  };

  return { selectedIds, reloadSelectedIds };
};
