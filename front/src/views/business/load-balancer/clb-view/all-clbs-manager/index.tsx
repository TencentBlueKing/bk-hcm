import { defineComponent } from 'vue';
import { useRouter } from 'vue-router';
// import components
import { Button } from 'bkui-vue';
import { BkRadioButton, BkRadioGroup } from 'bkui-vue/lib/radio';
import BatchOperationDialog from '@/components/batch-operation-dialog';
// import hooks
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import useSelection from '@/views/resource/resource-manage/hooks/use-selection';
import { useTable } from '@/hooks/useTable/useTable';
import { useI18n } from 'vue-i18n';
import { useWhereAmI, Senarios } from '@/hooks/useWhereAmI';
import useBatchDeleteLB from './useBatchDeleteLB';
// import utils
import { getTableRowClassOption } from '@/common/util';
import { asyncGetListenerCount } from '@/utils';
// import types
import { DoublePlainObject } from '@/typings';
import './index.scss';

export default defineComponent({
  name: 'AllClbsManager',
  setup() {
    // use hooks
    const router = useRouter();
    const { t } = useI18n();
    const { whereAmI } = useWhereAmI();
    const { selections, handleSelectionChange, resetSelections } = useSelection();

    const isRowSelectEnable = ({ row, isCheckAll }: DoublePlainObject) => {
      if (isCheckAll) return true;
      return isCurRowSelectEnable(row);
    };
    const isCurRowSelectEnable = (row: any) => {
      if (whereAmI.value === Senarios.business) return true;
      if (row.id) {
        return row.bk_biz_id === -1;
      }
    };

    const { columns, settings } = useColumns('lb');
    const { CommonTable, getListData } = useTable({
      searchOptions: {
        searchData: [],
      },
      tableOptions: {
        columns: [
          ...columns,
          {
            label: '操作',
            width: 120,
            render: () => <span class='operate-text-btn'>删除</span>,
          },
        ],
        extra: {
          settings: settings.value,
          isRowSelectEnable,
          onSelectionChange: (selections: any) => handleSelectionChange(selections, isCurRowSelectEnable),
          onSelectAll: (selections: any) => handleSelectionChange(selections, isCurRowSelectEnable, true),
          ...getTableRowClassOption(),
        },
      },
      requestOption: {
        type: 'load_balancers',
        sortOption: { sort: 'created_at', order: 'DESC' },
        callback(dataList: any[]) {
          return asyncGetListenerCount(dataList);
        },
      },
    });

    const handleApply = () => {
      router.push({
        path: '/business/service/service-apply/clb',
      });
    };

    // 批量删除负载均衡
    const {
      isBatchDeleteDialogShow,
      isSubmitLoading,
      radioGroupValue,
      tableProps,
      handleRemoveSelection,
      handleClickBatchDelete,
      handleBatchDeleteSubmit,
    } = useBatchDeleteLB(
      [
        ...columns.slice(1, 7),
        {
          label: '',
          width: 50,
          minWidth: 50,
          render: ({ data }: any) => (
            <Button text onClick={() => handleRemoveSelection(data.id)}>
              <i class='hcm-icon bkhcm-icon-minus-circle-shape'></i>
            </Button>
          ),
        },
      ],
      selections,
      resetSelections,
      getListData,
    );

    return () => (
      <div class='common-card-wrap has-selection'>
        {/* 负载均衡list */}
        <CommonTable>
          {{
            operation: () => (
              <>
                <Button class='mw64' theme='primary' onClick={handleApply}>
                  购买
                </Button>
                <Button class='mw88' onClick={handleClickBatchDelete}>
                  批量删除
                </Button>
              </>
            ),
          }}
        </CommonTable>
        {/* 批量删除负载均衡 */}
        <BatchOperationDialog
          class='batch-delete-lb-dialog'
          v-model:isShow={isBatchDeleteDialogShow.value}
          title={t('批量删除监听器')}
          theme='danger'
          confirmText='删除'
          isSubmitLoading={isSubmitLoading.value}
          tableProps={tableProps}
          onHandleConfirm={handleBatchDeleteSubmit}>
          {{
            tips: () => (
              <>
                已选择<span class='blue'>{tableProps.data.length}</span>个负载均衡，其中
                <span class='red'>{tableProps.data.filter(({ listenerNum }) => listenerNum > 0).length}</span>
                个存在监听器不可删除。
              </>
            ),
            tab: () => (
              <BkRadioGroup v-model={radioGroupValue.value}>
                <BkRadioButton label={false}>{t('可删除')}</BkRadioButton>
                <BkRadioButton label={true}>{t('不可删除')}</BkRadioButton>
              </BkRadioGroup>
            ),
          }}
        </BatchOperationDialog>
      </div>
    );
  },
});
