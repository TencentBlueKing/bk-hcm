import { defineComponent, inject } from 'vue';
import Search from '../../components/search';
import Button from '../../components/button';
import Amount from '../../components/amount';
import { useTable } from '@/hooks/useTable/useTable';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import { reqBillsMainAccountSummaryList } from '@/api/bill';

export default defineComponent({
  name: 'SubAccountTabPanel',
  setup() {
    const bill_year = inject<number>('bill_year');
    const bill_month = inject<number>('bill_month');

    const { columns, settings } = useColumns('billsMainAccountSummary');
    const { CommonTable, getListData } = useTable({
      searchOptions: { disabled: true },
      tableOptions: {
        columns,
        extra: {
          settings: settings.value,
        },
      },
      requestOption: {
        apiMethod: reqBillsMainAccountSummaryList as any,
        extension: () => ({
          bill_year,
          bill_month,
        }),
      },
    });

    return () => (
      <>
        <Search onSearch={getListData} />
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
