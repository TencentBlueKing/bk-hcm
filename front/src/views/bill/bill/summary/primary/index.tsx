import { defineComponent, inject, ref } from 'vue';
import { useRouter } from 'vue-router';

import { Button } from 'bkui-vue';
import IButton from '../../components/button';
import Amount from '../../components/amount';
import BillSyncDialog from './sync';
import ConfirmBillDialog from './confirm';
import RecalculateBillDialog from './recalculate';

import { useI18n } from 'vue-i18n';
import { useTable } from '@/hooks/useTable/useTable';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import { reqBillsRootAccountSummaryList } from '@/api/bill';

export default defineComponent({
  name: 'PrimaryAccountTabPanel',
  setup() {
    const { t } = useI18n();
    const router = useRouter();
    const bill_year = inject<number>('bill_year');
    const bill_month = inject<number>('bill_month');

    const { columns, settings } = useColumns('billsRootAccountSummary');

    const billSyncDialogRef = ref();
    const confirmBillDialogRef = ref();
    const recalculateBillDialogRef = ref();

    const handleConfirmBill = () => {
      confirmBillDialogRef.value.triggerShow(true);
    };
    const handleRecalculate = () => {
      recalculateBillDialogRef.value.triggerShow(true);
    };

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
                <Button text theme='primary' class='mr4' onClick={handleConfirmBill}>
                  {t('确认账单')}
                </Button>
                <Button text theme='primary' class='mr4' onClick={handleRecalculate}>
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

    const goOperationRecord = () => {
      router.push({ name: 'billSummaryOperationRecord' });
    };

    return () => (
      <div class='full-height p24'>
        <CommonTable>
          {{
            operation: () => <IButton billSyncDialogRef={billSyncDialogRef} />,
            operationBarEnd: () => (
              <Button theme='primary' text onClick={goOperationRecord}>
                <i class='hcm-icon bkhcm-icon-lishijilu mr4'></i>
                {t('操作记录')}
              </Button>
            ),
            tableToolbar: () => <Amount class='mt16 mb16' />,
          }}
        </CommonTable>
        <BillSyncDialog ref={billSyncDialogRef} />
        <ConfirmBillDialog ref={confirmBillDialogRef} />
        <RecalculateBillDialog ref={recalculateBillDialogRef} />
      </div>
    );
  },
});
