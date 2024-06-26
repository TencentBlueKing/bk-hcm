import { defineComponent, ref, inject, watch, Ref } from 'vue';

import { Button, Message } from 'bkui-vue';
import { Plus } from 'bkui-vue/lib/icon';
import Panel from '@/components/panel';
import Search from '../components/search';
import CreateAdjustSideSlider from './create';
import Amount from '../components/amount';
import Confirm from '@/components/confirm';
import BatchOperation from './batch-operation';

import { useI18n } from 'vue-i18n';
import { cloneDeep } from 'lodash';
import { useTable } from '@/hooks/useTable/useTable';
import useSelection from '@/views/resource/resource-manage/hooks/use-selection';
import { deleteBillsAdjustment, reqBillsAdjustmentList } from '@/api/bill';
import { timeFormatter } from '@/common/util';
import { BILL_ADJUSTMENT_STATE__MAP, BILL_ADJUSTMENT_TYPE__MAP } from '@/constants';
import { DoublePlainObject, QueryRuleOPEnum, RulesItem } from '@/typings';

export default defineComponent({
  name: 'BillAdjust',
  setup() {
    const { t } = useI18n();
    const bill_year = inject<Ref<number>>('bill_year');
    const bill_month = inject<Ref<number>>('bill_month');
    const searchRef = ref();
    const createAdjustSideSliderRef = ref();

    const isRowSelectEnable = ({ row, isCheckAll }: DoublePlainObject) => {
      if (isCheckAll) return true;
      return isCurRowSelectEnable(row);
    };
    const isCurRowSelectEnable = (row: any) => row.state === 'unconfirmed';
    const { selections, handleSelectionChange, resetSelections } = useSelection();

    const handleDelete = (id: string) => {
      Confirm('删除调账明细', `将删除调账明细: ${id}`, async () => {
        await deleteBillsAdjustment([id]);
        Message({ theme: 'success', message: '删除成功' });
        getListData();
        resetSelections();
      });
    };

    const columns = [
      {
        label: '',
        type: 'selection',
        width: 32,
        minWidth: 32,
      },
      {
        label: t('更新时间'),
        field: 'updated_at',
        render: ({ cell }: any) => timeFormatter(cell),
      },
      {
        label: t('调账ID'),
        field: 'id',
      },
      {
        label: '业务',
        field: 'product_id',
      },
      {
        label: t('二级账号名称'),
        field: 'main_account_cloud_id',
      },
      {
        label: t('调账类型'),
        field: 'type',
        render: ({ cell }: any) => (
          <bk-tag theme={cell === 'increase' ? 'success' : 'danger'}>{BILL_ADJUSTMENT_TYPE__MAP[cell]}</bk-tag>
        ),
      },
      {
        label: t('操作人'),
        field: 'operator',
      },
      {
        label: t('人民币（元）'),
        field: 'rmb_cost',
      },
      {
        label: t('美金（美元）'),
        field: 'cost',
      },
      {
        label: t('调账状态'),
        field: 'state',
        width: 100,
        render: ({ cell }: any) => (
          <bk-tag theme={cell === 'confirmed' ? 'success' : undefined}>{BILL_ADJUSTMENT_STATE__MAP[cell]}</bk-tag>
        ),
      },
      {
        label: t('操作'),
        render: ({ data }: any) => (
          <>
            <Button
              text
              theme='primary'
              class='mr8'
              onClick={() => createAdjustSideSliderRef.value.triggerShow(true)}
              disabled={data.state !== 'unconfirmed'}
              v-bk-tooltips={{ content: t('当前调账单已确认，无法编辑'), disabled: data.state === 'unconfirmed' }}>
              {t('编辑')}
            </Button>
            <Button
              text
              theme='primary'
              onClick={() => handleDelete(data.id)}
              disabled={data.state !== 'unconfirmed'}
              v-bk-tooltips={{ content: t('当前调账单已确认，无法删除'), disabled: data.state === 'unconfirmed' }}>
              {t('删除')}
            </Button>
          </>
        ),
      },
    ];

    const action = ref();
    const batchOperationRef = ref();
    const handleBatchOperation = (actionType: 'delete' | 'confirm') => {
      action.value = actionType;
      batchOperationRef.value.changeData(cloneDeep(selections.value));
      batchOperationRef.value.triggerShow(true);
    };

    const { CommonTable, getListData, clearFilter } = useTable({
      searchOptions: {
        disabled: true,
      },
      tableOptions: {
        columns,
        extra: {
          isRowSelectEnable,
          onSelectionChange: (selections: any) => handleSelectionChange(selections, isCurRowSelectEnable),
          onSelectAll: (selections: any) => handleSelectionChange(selections, isCurRowSelectEnable, true),
        },
      },
      requestOption: {
        apiMethod: reqBillsAdjustmentList,
        filterOption: {
          rules: [
            { field: 'bill_year', op: QueryRuleOPEnum.EQ, value: bill_year.value },
            { field: 'bill_month', op: QueryRuleOPEnum.EQ, value: bill_month.value },
          ],
        },
      },
    });

    const reloadTable = (rules: RulesItem[]) => {
      clearFilter();
      getListData(() => [
        ...rules,
        { field: 'bill_year', op: QueryRuleOPEnum.EQ, value: bill_year.value },
        { field: 'bill_month', op: QueryRuleOPEnum.EQ, value: bill_month.value },
      ]);
    };

    watch([bill_year, bill_month], () => {
      searchRef.value.handleSearch();
    });

    return () => (
      <div class='bill-adjust-module'>
        <Panel>
          <Search
            ref={searchRef}
            searchKeys={['product_id', 'main_account_id', 'updated_at']}
            onSearch={reloadTable}
            style={{ padding: 0, boxShadow: 'none' }}
          />
        </Panel>
        <Panel class='mt12'>
          <CommonTable>
            {{
              operation: () => (
                <>
                  <Button onClick={() => createAdjustSideSliderRef.value.triggerShow(true)}>
                    <Plus style={{ fontSize: '22px' }} />
                    {t('新增调账')}
                  </Button>
                  <Button>{t('导入')}</Button>
                  <Button>{t('导出')}</Button>
                  <Button onClick={() => handleBatchOperation('delete')} disabled={selections.value.length === 0}>
                    {t('批量删除')}
                  </Button>
                  <Button onClick={() => handleBatchOperation('confirm')} disabled={selections.value.length === 0}>
                    {t('批量确认')}
                  </Button>
                </>
              ),
              operationBarEnd: () => <Amount isAdjust />,
            }}
          </CommonTable>
        </Panel>
        <CreateAdjustSideSlider ref={createAdjustSideSliderRef} />
        <BatchOperation
          ref={batchOperationRef}
          action={action.value}
          getListData={getListData}
          resetSelections={resetSelections}
        />
      </div>
    );
  },
});
