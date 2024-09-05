import { Ref, defineComponent, inject, ref, watch } from 'vue';
import { useRouter } from 'vue-router';

import { Button } from 'bkui-vue';
import Amount from '../../components/amount';
import ConfirmBillDialog from './confirm';
import RecalculateBillDialog from './recalculate';

import { useI18n } from 'vue-i18n';
import { useTable } from '@/hooks/useTable/useTable';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import { reqBillsRootAccountSummaryList, reqBillsRootAccountSummarySum } from '@/api/bill';
import { BillsRootAccountSummaryState } from '@/typings/bill';
import { BILLS_ROOT_ACCOUNT_SUMMARY_STATE_MAP } from '@/constants';
import pluginHandler from '@pluginHandler/bill-manage';

export default defineComponent({
  name: 'PrimaryAccountTabPanel',
  setup() {
    const { t } = useI18n();
    const router = useRouter();
    const bill_year = inject<Ref<number>>('bill_year');
    const bill_month = inject<Ref<number>>('bill_month');

    const { usePrimaryHandler } = pluginHandler;
    const { renderOperation } = usePrimaryHandler();

    const { columns, settings } = useColumns('billsRootAccountSummary');

    const confirmBillDialogRef = ref();
    const recalculateBillDialogRef = ref();
    const amountRef = ref();

    const canConfirmBill = (state: BillsRootAccountSummaryState) => {
      return ![BillsRootAccountSummaryState.accounting, BillsRootAccountSummaryState.syncing].includes(state);
    };
    const handleConfirmBill = (data: any) => {
      confirmBillDialogRef.value.triggerShow(true, data);
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
                <Button
                  text
                  theme='primary'
                  class='mr4'
                  onClick={() => handleConfirmBill(data)}
                  disabled={!canConfirmBill(data.state)}
                  v-bk-tooltips={{
                    content: `${BILLS_ROOT_ACCOUNT_SUMMARY_STATE_MAP[data.state]}的账单不可进行确认操作`,
                    disabled: canConfirmBill(data.state),
                  }}>
                  {t('确认账单')}
                </Button>
                <Button
                  text
                  theme='primary'
                  class='mr4'
                  onClick={() => handleRecalculate(data)}
                  disabled={!canRecalculate(data.state)}
                  v-bk-tooltips={{
                    content: `${BILLS_ROOT_ACCOUNT_SUMMARY_STATE_MAP[data.state]}的账单不可进行重新核算操作`,
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
        sortOption: {
          sort: 'current_month_rmb_cost',
          order: 'DESC',
        },
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
    });

    watch(filter, () => {
      amountRef.value.refreshAmountInfo();
    });

    return () => (
      <div class='full-height p24'>
        <CommonTable>
          {{
            operation: () => renderOperation(bill_year.value, bill_month.value, filter),
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
                immediate
              />
            ),
          }}
        </CommonTable>
        <ConfirmBillDialog ref={confirmBillDialogRef} reloadTable={getListData} />
        <RecalculateBillDialog ref={recalculateBillDialogRef} reloadTable={getListData} />
      </div>
    );
  },
});
