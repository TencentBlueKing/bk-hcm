import { Ref, defineComponent, inject, ref, watch } from 'vue';

import Button from '../../components/button';
import Amount from '../../components/amount';
import Search from '../../components/search';

import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import { useTable } from '@/hooks/useTable/useTable';
import { reqBillsMainAccountSummaryList, reqBillsMainAccountSummarySum } from '@/api/bill';
import { RulesItem } from '@/typings';

export default defineComponent({
  name: 'OperationProductTabPanel',
  setup() {
    const bill_year = inject<Ref<number>>('bill_year');
    const bill_month = inject<Ref<number>>('bill_month');
    const amountRef = ref();

    const { columns } = useColumns('billsMainAccountSummary');
    const { CommonTable, getListData, clearFilter } = useTable({
      searchOptions: { disabled: true },
      tableOptions: {
        columns: columns.slice(2, -1),
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
      },
    });

    const reloadTable = (rules: RulesItem[]) => {
      clearFilter();
      getListData(rules);
    };

    watch([bill_year, bill_month], () => {
      getListData();
      amountRef.value.refreshAmountInfo();
    });

    return () => (
      <>
        <Search searchKeys={['product_id']} onSearch={reloadTable} />
        <div class='p24' style={{ height: 'calc(100% - 162px)' }}>
          <CommonTable>
            {{
              operation: () => <Button noSyncBtn />,
              operationBarEnd: () => (
                <Amount
                  ref={amountRef}
                  api={reqBillsMainAccountSummarySum}
                  payload={() => ({ bill_year: bill_year.value, bill_month: bill_month.value })}
                />
              ),
            }}
          </CommonTable>
        </div>
      </>
    );
  },
});
