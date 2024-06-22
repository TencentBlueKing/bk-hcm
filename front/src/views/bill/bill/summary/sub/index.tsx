import { Ref, defineComponent, inject } from 'vue';
import Search from '../../components/search';
import Button from '../../components/button';
import Amount from '../../components/amount';
import { useTable } from '@/hooks/useTable/useTable';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import { reqBillsMainAccountSummaryList } from '@/api/bill';
import { RulesItem } from '@/typings';

export default defineComponent({
  name: 'SubAccountTabPanel',
  setup() {
    const bill_year = inject<Ref<number>>('bill_year');
    const bill_month = inject<Ref<number>>('bill_month');

    const { columns } = useColumns('billsMainAccountSummary');
    const { CommonTable, getListData, clearFilter } = useTable({
      searchOptions: { disabled: true },
      tableOptions: {
        columns,
      },
      requestOption: {
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

    return () => (
      <>
        <Search onSearch={reloadTable} />
        <div class='p24' style={{ height: 'calc(100% - 162px)' }}>
          <CommonTable>
            {{
              operation: () => <Button noSyncBtn />,
              operationBarEnd: () => <Amount />,
            }}
          </CommonTable>
        </div>
      </>
    );
  },
});
