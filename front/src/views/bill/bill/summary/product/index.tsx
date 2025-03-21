import { Ref, defineComponent, inject, onMounted, ref, watch } from 'vue';
import { useRoute } from 'vue-router';

import Amount from '../../components/amount';
import Search from '../../components/search';

import useChangeCurrency from '../../use-change-currency';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import { useTable } from '@/hooks/useTable/useTable';
import { reqBillsMainAccountSummarySum } from '@/api/bill';
import { RulesItem } from '@/typings';
import pluginHandler from '@pluginHandler/bill-manage';

export default defineComponent({
  name: 'OperationProductTabPanel',
  setup() {
    const route = useRoute();
    const bill_year = inject<Ref<number>>('bill_year');
    const bill_month = inject<Ref<number>>('bill_month');

    const searchRef = ref();
    const amountRef = ref();

    const { useProductHandler } = pluginHandler;
    const {
      selectedIds,
      columnName,
      getColumns,
      extensionKey,
      apiMethod,
      reloadSelectedIds,
      mountedCallback,
      renderOperation,
    } = useProductHandler();

    const { customRender } = useChangeCurrency({ onlyRMB: true });
    const { columns } = useColumns(columnName, false, '', { customRender });
    const { CommonTable, getListData, clearFilter, filter } = useTable({
      searchOptions: { disabled: true },
      tableOptions: {
        columns: getColumns(columns),
      },
      requestOption: {
        sortOption: {
          sort: 'current_month_rmb_cost',
          order: 'DESC',
        },
        apiMethod,
        extension: () => ({
          bill_year: bill_year.value,
          bill_month: bill_month.value,
          [extensionKey]: selectedIds.value,
          filter: undefined,
        }),
        immediate: false,
      },
    });
    const reloadTable = (rules: RulesItem[]) => {
      reloadSelectedIds(rules);
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
        <Search ref={searchRef} searchKeys={['product_id']} onSearch={reloadTable} />
        <div class='p24' style={{ height: 'calc(100% - 162px)' }}>
          <CommonTable>
            {{
              operation: () => renderOperation(bill_year.value, bill_month.value, searchRef),
              operationBarEnd: () => (
                <Amount
                  ref={amountRef}
                  onlyRMB
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
