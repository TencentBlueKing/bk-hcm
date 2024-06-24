import { Ref, defineComponent, inject, ref, watch } from 'vue';
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
import { reqBillsRootAccountSummaryList, reqBillsRootAccountSummarySum } from '@/api/bill';
import { BillsRootAccountSummaryState } from '@/typings/bill';

export default defineComponent({
  name: 'PrimaryAccountTabPanel',
  setup() {
    const { t } = useI18n();
    const router = useRouter();
    const bill_year = inject<Ref<number>>('bill_year');
    const bill_month = inject<Ref<number>>('bill_month');

    const { columns, settings } = useColumns('billsRootAccountSummary');

    const billSyncDialogRef = ref();
    const confirmBillDialogRef = ref();
    const recalculateBillDialogRef = ref();
    const amountRef = ref();

    const handleConfirmBill = () => {
      confirmBillDialogRef.value.triggerShow(true);
    };
    const canRecalculate = (state: BillsRootAccountSummaryState) => {
      return [
        BillsRootAccountSummaryState.accounted,
        BillsRootAccountSummaryState.confirmed,
        BillsRootAccountSummaryState.synced,
      ].includes(state);
    };
    const handleRecalculate = (data: any) => {
      recalculateBillDialogRef.value.triggerShow(true, data);
    };

    const { CommonTable, getListData, filter } = useTable({
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
            render: ({ data }: any) => (
              <>
                <Button text theme='primary' class='mr4' onClick={handleConfirmBill}>
                  {t('确认账单')}
                </Button>
                <Button
                  text
                  theme='primary'
                  class='mr4'
                  onClick={() => handleRecalculate(data)}
                  disabled={!canRecalculate(data.state)}
                  v-bk-tooltips={{
                    content: '只有已核算、已确认、已同步的账单才能进行重新核算操作',
                    disabled: canRecalculate(data.state),
                  }}>
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
        apiMethod: reqBillsRootAccountSummaryList,
        extension: () => ({
          bill_year: bill_year.value,
          bill_month: bill_month.value,
        }),
      },
    });

    const goOperationRecord = () => {
      router.push({ name: 'billSummaryOperationRecord' });
    };

    watch([bill_year, bill_month], () => {
      getListData();
      amountRef.value.refreshAmountInfo();
    });

    watch(filter, () => {
      amountRef.value.refreshAmountInfo();
    });

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
            tableToolbar: () => (
              <Amount
                ref={amountRef}
                class='mt16 mb16'
                api={reqBillsRootAccountSummarySum}
                payload={() => ({ bill_year: bill_year.value, bill_month: bill_month.value, filter })}
              />
            ),
          }}
        </CommonTable>
        <BillSyncDialog ref={billSyncDialogRef} />
        <ConfirmBillDialog ref={confirmBillDialogRef} />
        <RecalculateBillDialog ref={recalculateBillDialogRef} reloadTable={getListData} />
      </div>
    );
  },
});
