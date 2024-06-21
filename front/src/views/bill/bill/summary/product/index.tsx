import { Ref, defineComponent, inject } from 'vue';
import Button from '../../components/button';
import Amount from '../../components/amount';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import { useTable } from '@/hooks/useTable/useTable';
import { reqBillsMainAccountSummaryList } from '@/api/bill';

export default defineComponent({
  name: 'OperationProductTabPanel',
  setup() {
    const bill_year = inject<Ref<number>>('bill_year');
    const bill_month = inject<Ref<number>>('bill_month');

    const { columns } = useColumns('billsMainAccountSummary');
    const { CommonTable } = useTable({
      searchOptions: { disabled: true },
      tableOptions: {
        columns: columns.slice(2, -1),
      },
      requestOption: {
        apiMethod: reqBillsMainAccountSummaryList,
        extension: () => ({
          bill_year: bill_year.value,
          bill_month: bill_month.value,
        }),
      },
    });

    return () => (
      <div class='full-height p24'>
        <CommonTable>
          {{
            operation: () => <Button noSyncBtn />,
            operationBarEnd: () => <Amount />,
          }}
        </CommonTable>
      </div>
    );
  },
});
