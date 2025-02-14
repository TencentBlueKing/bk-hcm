import { Ref, ref } from 'vue';
import { RouteLocationNormalizedLoaded } from 'vue-router';

import BillsExportButton from '@/views/bill/bill/components/bills-export-button';
import BusinessSelector from '@/components/business-selector/index.vue';

import { useI18n } from 'vue-i18n';
import { exportBillsBizSummary, exportBillsRootAccountSummary, reqBillsBizSummaryList } from '@/api/bill';
import { FilterType, QueryRuleOPEnum, RulesItem } from '@/typings';
import { BillSearchRules } from '@/utils';
import { BILL_BIZS_KEY, BILL_MAIN_ACCOUNTS_KEY } from '@/constants';
import { ISearchModal } from '@/views/bill/bill/components/search';

// 账单汇总-一级账号
const usePrimaryHandler = () => {
  const renderOperation = (bill_year: number, bill_month: number, filter: FilterType) => {
    const { t } = useI18n();

    return (
      <BillsExportButton
        cb={() => exportBillsRootAccountSummary({ bill_year, bill_month, export_limit: 200000, filter })}
        title={t('账单汇总-一级账号')}
        content={t('导出当月一级账号的账单数据')}
      />
    );
  };

  return {
    renderOperation,
  };
};

// 账单汇总-二级账号
const useSubHandler = () => {
  // mounted 时, 根据初始条件加载表格数据
  const mountedCallback = (route: RouteLocationNormalizedLoaded, reloadTable: (rules: RulesItem[]) => void) => {
    // 只有业务、二级账号有保存的需求
    const billSearchRules = new BillSearchRules();
    billSearchRules
      .addRule(route, BILL_BIZS_KEY, 'bk_biz_id', QueryRuleOPEnum.IN)
      .addRule(route, BILL_MAIN_ACCOUNTS_KEY, 'main_account_id', QueryRuleOPEnum.IN);
    reloadTable(billSearchRules.rules);
  };

  return {
    mountedCallback,
  };
};

// 账单汇总-业务
const useProductHandler = () => {
  // table 相关状态
  const selectedIds = ref<number[]>([]);
  const columnName = 'billsMainAccountSummary';
  const getColumns = (columns: any[]) => columns.slice(4);
  const apiMethod: (...args: any) => Promise<any> = reqBillsBizSummaryList;
  const extensionKey = 'bk_biz_ids';

  // reloadTable 时, 重置选中项
  const reloadSelectedIds = (rules: RulesItem[]) => {
    selectedIds.value = (rules.find((rule) => rule.field === 'bk_biz_id')?.value as number[]) || [];
  };

  // mounted 时, 根据初始条件加载表格数据
  const mountedCallback = (route: RouteLocationNormalizedLoaded, reloadTable: (rules: RulesItem[]) => void) => {
    // 只有业务有保存的需求
    const billSearchRules = new BillSearchRules();
    billSearchRules.addRule(route, BILL_BIZS_KEY, 'bk_biz_id', QueryRuleOPEnum.IN);
    reloadTable(billSearchRules.rules);
  };

  // 操作栏
  const renderOperation = (bill_year: number, bill_month: number, searchRef: Ref<any>) => {
    const { t } = useI18n();

    return (
      <BillsExportButton
        cb={() =>
          exportBillsBizSummary({
            bill_year,
            bill_month,
            export_limit: 200000,
            bk_biz_ids: searchRef.value.rules.find((rule: any) => rule.field === 'bk_biz_id')?.value || [],
          })
        }
        title={t('账单汇总-业务')}
        content={t('导出当月业务的账单数据')}
      />
    );
  };

  return {
    selectedIds,
    columnName,
    getColumns,
    extensionKey,
    apiMethod,
    reloadSelectedIds,
    mountedCallback,
    renderOperation,
  };
};

// 账单调整
const useAdjustHandler = () => {
  // mounted 时, 根据初始条件加载表格数据
  const mountedCallback = (route: RouteLocationNormalizedLoaded, reloadTable: (rules: RulesItem[]) => void) => {
    // 只有业务、二级账号有保存的需求
    const billSearchRules = new BillSearchRules();
    billSearchRules
      .addRule(route, BILL_BIZS_KEY, 'bk_biz_id', QueryRuleOPEnum.IN)
      .addRule(route, BILL_MAIN_ACCOUNTS_KEY, 'main_account_id', QueryRuleOPEnum.IN);
    reloadTable(billSearchRules.rules);
  };

  return {
    mountedCallback,
  };
};

// 搜索组件
const useSearchCompHandler = () => {
  const { t } = useI18n();
  const productSearchLabel = t('业务');

  const renderProductComponent = (modal: Ref<ISearchModal>) => {
    return <BusinessSelector v-model={modal.value.bk_biz_id} clearable multiple urlKey={BILL_BIZS_KEY} />;
  };

  return {
    productSearchLabel,
    renderProductComponent,
  };
};

const pluginHandler = {
  usePrimaryHandler,
  useSubHandler,
  useProductHandler,
  useAdjustHandler,
  useSearchCompHandler,
};

export default pluginHandler;
export type PluginHandlerType = typeof pluginHandler;
