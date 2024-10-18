import { Ref, defineComponent, inject, onMounted, ref, watch } from 'vue';
import { useRoute } from 'vue-router';

import Search from '../../components/search';
import BillsExportButton from '../../components/bills-export-button';
import Amount from '../../components/amount';

import { useI18n } from 'vue-i18n';
import { useTable } from '@/hooks/useTable/useTable';
import {
  exportBillsMainAccountSummary,
  reqBillsMainAccountSummaryList,
  reqBillsMainAccountSummarySum,
} from '@/api/bill';
import { RulesItem } from '@/typings';
import { getColumns, mountedCallback } from './load-data.plugin';

export default defineComponent({
  name: 'SubAccountTabPanel',
  setup() {
    const { t } = useI18n();
    const route = useRoute();
    const bill_year = inject<Ref<number>>('bill_year');
    const bill_month = inject<Ref<number>>('bill_month');

    const searchRef = ref();
    const amountRef = ref();

    const columns = getColumns();

    const { CommonTable, getListData, clearFilter, filter } = useTable({
      searchOptions: { disabled: true },
      tableOptions: {
        columns,
      },
      requestOption: {
        sortOption: {
          sort: 'current_month_rmb_cost',
          order: 'DESC',
        },
        apiMethod: reqBillsMainAccountSummaryList,
        extension: () => ({
          bill_year: bill_year.value,
          bill_month: bill_month.value,
        }),
        immediate: false,
      },
    });

    const reloadTable = (rules: RulesItem[]) => {
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
      mountedCallback(route, reloadTable);
    });

    return () => (
      <>
        <Search
          ref={searchRef}
          searchKeys={['vendor', 'root_account_id', 'product_id', 'main_account_id']}
          onSearch={reloadTable}
          autoSelectMainAccount
        />
        <div class='p24' style={{ height: 'calc(100% - 238px)' }}>
          <CommonTable>
            {{
              operation: () => (
                <BillsExportButton
                  cb={() =>
                    exportBillsMainAccountSummary({
                      bill_year: bill_year.value,
                      bill_month: bill_month.value,
                      export_limit: 200000,
                      filter,
                    })
                  }
                  title={t('账单汇总-二级账号')}
                  content={t('导出当月二级账号的账单数据')}
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
