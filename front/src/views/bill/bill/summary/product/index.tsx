import { Ref, defineComponent, inject, onMounted, ref, watch } from 'vue';
import { useRoute } from 'vue-router';

import BillsExportButton from '../../components/bills-export-button';
import Amount from '../../components/amount';
import Search from '../../components/search';

import { useI18n } from 'vue-i18n';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import { useTable } from '@/hooks/useTable/useTable';
import { exportBillsBizSummary, reqBillsBizSummaryList, reqBillsMainAccountSummarySum } from '@/api/bill';
import { QueryRuleOPEnum, RulesItem } from '@/typings';
import { BILL_BIZS_KEY } from '@/constants';
import { BillSearchRules } from '@/utils';

export default defineComponent({
  name: 'OperationProductTabPanel',
  setup() {
    const { t } = useI18n();
    const route = useRoute();
    const bill_year = inject<Ref<number>>('bill_year');
    const bill_month = inject<Ref<number>>('bill_month');

    const searchRef = ref();
    const amountRef = ref();
    const selectedBkBizIds = ref([]);

    const { columns } = useColumns('billsMainAccountSummary');
    const { CommonTable, getListData, clearFilter, filter } = useTable({
      searchOptions: { disabled: true },
      tableOptions: {
        columns: columns.slice(3),
      },
      requestOption: {
        sortOption: {
          sort: 'current_month_rmb_cost',
          order: 'DESC',
        },
        apiMethod: reqBillsBizSummaryList,
        extension: () => ({
          bill_year: bill_year.value,
          bill_month: bill_month.value,
          bk_biz_ids: selectedBkBizIds.value,
          filter: undefined,
        }),
        immediate: false,
      },
    });

    const reloadTable = (rules: RulesItem[]) => {
      selectedBkBizIds.value = (rules.find((rule) => rule.field === 'bk_biz_id')?.value as number[]) || [];
      clearFilter();
      getListData(rules);
    };

    watch([bill_year, bill_month], () => {
      searchRef.value.handleSearch();
    });

    watch(filter, () => {
      amountRef.value.refreshAmountInfo();
    });

    onMounted(() => {
      // 只有业务有保存的需求
      const billSearchRules = new BillSearchRules();
      billSearchRules.addRule(route, BILL_BIZS_KEY, 'bk_biz_id', QueryRuleOPEnum.IN);
      reloadTable(billSearchRules.rules);
    });

    return () => (
      <>
        <Search ref={searchRef} searchKeys={['product_id']} onSearch={reloadTable} />
        <div class='p24' style={{ height: 'calc(100% - 162px)' }}>
          <CommonTable>
            {{
              operation: () => (
                <BillsExportButton
                  cb={() =>
                    exportBillsBizSummary({
                      bill_year: bill_year.value,
                      bill_month: bill_month.value,
                      export_limit: 200000,
                      bk_biz_ids: searchRef.value.rules.find((rule: any) => rule.field === 'bk_biz_id')?.value || [],
                    })
                  }
                  title={t('账单汇总-业务')}
                  content={t('导出当月业务的账单数据')}
                />
              ),
              operationBarEnd: () => (
                <Amount
                  ref={amountRef}
                  api={reqBillsMainAccountSummarySum}
                  payload={() => ({ bill_year: bill_year.value, bill_month: bill_month.value, filter })}
                />
              ),
            }}
          </CommonTable>
        </div>
      </>
    );
  },
});
