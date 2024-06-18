import { defineComponent, inject } from 'vue';
import { Button } from 'bkui-vue';
import IButton from '../../components/button';
import Amount from '../../components/amount';
import { useTable } from '@/hooks/useTable/useTable';
import { reqBillsRootAccountSummaryList } from '@/api/bill';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import { useI18n } from 'vue-i18n';

export default defineComponent({
  name: 'PrimaryAccountTabPanel',
  setup() {
    const { t } = useI18n();
    const bill_year = inject<number>('bill_year');
    const bill_month = inject<number>('bill_month');

    const { columns, settings } = useColumns('billsRootAccountSummary');

    const { CommonTable } = useTable({
      searchOptions: {
        searchData: [
          { name: '一级账号ID', id: 'root_account_id' },
          { name: '一级账号名称', id: 'root_account_name' },
        ],
      },
      tableOptions: {
        columns: [
          ...columns,
          {
            label: '操作',
            width: 150,
            fixed: 'right',
            render: () => (
              <>
                <Button text theme='primary' class='mr4'>
                  {t('确认账单')}
                </Button>
                <Button text theme='primary' class='mr4'>
                  {t('重新核算')}
                </Button>
              </>
            ),
          },
        ],
        extra: {
          settings: settings.value,
        },
      },
      requestOption: {
        apiMethod: reqBillsRootAccountSummaryList as any,
        extension: () => ({
          bill_year,
          bill_month,
        }),
      },
    });

    const goOperationRecord = () => {};

    return () => (
      <div class='full-height p24'>
        <CommonTable>
          {{
            operation: () => <IButton />,
            operationBarEnd: () => (
              <Button theme='primary' text onClick={goOperationRecord}>
                <i class='hcm-icon bkhcm-icon-lishijilu mr4'></i>
                {t('操作记录')}
              </Button>
            ),
            tableToolbar: () => <Amount class='mt16 mb16' />,
          }}
        </CommonTable>
      </div>
    );
  },
});
