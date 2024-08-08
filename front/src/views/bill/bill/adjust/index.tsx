import { defineComponent, ref, inject, watch, Ref, onMounted } from 'vue';

import { Button, Message } from 'bkui-vue';
import { Plus } from 'bkui-vue/lib/icon';
import Panel from '@/components/panel';
import Search from '../components/search';
import CreateAdjustSideSlider from './create';
import Amount from '../components/amount';
import Confirm from '@/components/confirm';
import BatchOperation from './batch-operation';
import BillsExportButton from '../components/bills-export-button';

import { useI18n } from 'vue-i18n';
import { cloneDeep } from 'lodash';
import { useBusinessMapStore } from '@/store/useBusinessMap';
import { useTable } from '@/hooks/useTable/useTable';
import useSelection from '@/views/resource/resource-manage/hooks/use-selection';
import { deleteBillsAdjustment, exportBillsAdjustmentItems, reqBillsAdjustmentList } from '@/api/bill';
import { timeFormatter } from '@/common/util';
import {
  BILL_ADJUSTMENT_STATE__MAP,
  BILL_ADJUSTMENT_TYPE__MAP,
  BILL_BIZS_KEY,
  BILL_MAIN_ACCOUNTS_KEY,
  CURRENCY_MAP,
} from '@/constants';
import { DoublePlainObject, QueryRuleOPEnum, RulesItem } from '@/typings';
import useBillStore from '@/store/useBillStore';
import { computed } from '@vue/reactivity';
import { BillSearchRules, formatBillCost } from '@/utils';
import { useRoute } from 'vue-router';

export default defineComponent({
  name: 'BillAdjust',
  setup() {
    const route = useRoute();
    const { t } = useI18n();
    const bill_year = inject<Ref<number>>('bill_year');
    const bill_month = inject<Ref<number>>('bill_month');
    const businessMapStore = useBusinessMapStore();
    const billStore = useBillStore();
    const amountRef = ref();

    const searchRef = ref();
    const createAdjustSideSliderRef = ref();
    const isEdit = ref(false);
    const editData = ref({});

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
      { type: 'selection', width: 30, minWidth: 30 },
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
        label: t('业务名称'),
        field: 'product_id',
        render: ({ data }: any) => businessMapStore.businessMap.get(data.bk_biz_id) || '未分配',
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
        label: t('金额'),
        field: 'cost',
        render: ({ cell }: any) => formatBillCost(cell),
      },
      {
        label: t('币种'),
        field: 'currency',
        render: ({ cell }: any) => CURRENCY_MAP[cell] || '--',
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
              onClick={() => {
                createAdjustSideSliderRef.value.triggerShow(true);
                isEdit.value = true;
                editData.value = data;
              }}
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

    const { CommonTable, getListData, clearFilter, filter } = useTable({
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
        immediate: false,
      },
    });

    const amountFilter = computed(() => ({
      filter: {
        op: 'and',
        rules: filter.rules,
      },
    }));

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

    watch(filter, () => {
      amountRef.value.refreshAmountInfo();
    });

    onMounted(() => {
      // 只有业务、二级账号有保存的需求
      const billSearchRules = new BillSearchRules();
      billSearchRules
        .addRule(route, BILL_BIZS_KEY, 'bk_biz_id', QueryRuleOPEnum.IN)
        .addRule(route, BILL_MAIN_ACCOUNTS_KEY, 'main_account_id', QueryRuleOPEnum.IN);
      reloadTable(billSearchRules.rules);
    });

    return () => (
      <div class='bill-adjust-module'>
        <Panel>
          <Search
            ref={searchRef}
            searchKeys={['product_id', 'main_account_id', 'updated_at']}
            onSearch={reloadTable}
            autoSelectMainAccount
            style={{ padding: 0, boxShadow: 'none' }}
          />
        </Panel>
        <Panel class='mt12' style={{ height: 'calc(100% - 159px)' }}>
          <CommonTable>
            {{
              operation: () => (
                <>
                  <Button
                    onClick={() => {
                      createAdjustSideSliderRef.value.triggerShow(true);
                      isEdit.value = false;
                    }}>
                    <Plus style={{ fontSize: '22px' }} />
                    {t('新增调账')}
                  </Button>
                  <Button>{t('导入')}</Button>
                  <BillsExportButton
                    cb={() =>
                      exportBillsAdjustmentItems({
                        bill_year: bill_year.value,
                        bill_month: bill_month.value,
                        export_limit: 200000,
                        filter,
                      })
                    }
                    title={t(`账单调整`)}
                    content={t(`导出当月账单调整的数据`)}
                  />
                  <Button onClick={() => handleBatchOperation('delete')} disabled={selections.value.length === 0}>
                    {t('批量删除')}
                  </Button>
                  <Button onClick={() => handleBatchOperation('confirm')} disabled={selections.value.length === 0}>
                    {t('批量确认')}
                  </Button>
                </>
              ),
              operationBarEnd: () => (
                <Amount isAdjust api={billStore.sum_adjust_items} payload={() => amountFilter.value} ref={amountRef} />
              ),
            }}
          </CommonTable>
        </Panel>
        <CreateAdjustSideSlider
          ref={createAdjustSideSliderRef}
          onUpdate={getListData}
          edit={isEdit.value}
          editData={editData.value}
          onClearEdit={() => (editData.value = undefined)}
        />
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
