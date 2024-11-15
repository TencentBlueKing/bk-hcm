import { defineComponent, watch, onUnmounted } from 'vue';
// import components
import { Button, Message } from 'bkui-vue';
import { BkRadioButton, BkRadioGroup } from 'bkui-vue/lib/radio';
import { Plus } from 'bkui-vue/lib/icon';
import BatchOperationDialog from '@/components/batch-operation-dialog';
import Confirm from '@/components/confirm';
// import stores
import { useBusinessStore } from '@/store';
// import custom hooks
import { useTable } from '@/hooks/useTable/useTable';
import { useI18n } from 'vue-i18n';
import { useWhereAmI, Senarios } from '@/hooks/useWhereAmI';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import useSelection from '@/views/resource/resource-manage/hooks/use-selection';
import useBatchDeleteListener from './useBatchDeleteListener';
import AddOrUpdateListenerSideslider from '../../components/AddOrUpdateListenerSideslider';
// import utils
import { getTableNewRowClass } from '@/common/util';
import bus from '@/common/bus';
// import types
import { DoublePlainObject } from '@/typings';
import './index.scss';

export default defineComponent({
  props: { id: String },
  setup(props) {
    // use hooks
    const { t } = useI18n();
    const { whereAmI } = useWhereAmI();
    const { selections, handleSelectionChange, resetSelections } = useSelection();
    let timer: string | number | NodeJS.Timeout;
    let counter = 0; // 初始化计数器
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

    // use stores
    const businessStore = useBusinessStore();

    // listener - table
    const { columns, settings } = useColumns('listener');
    const { CommonTable, getListData } = useTable({
      searchOptions: {
        searchData: [
          {
            name: '监听器名称',
            id: 'name',
          },
          {
            name: '协议',
            id: 'protocol',
          },
          {
            name: '端口',
            id: 'port',
          },
          {
            name: '均衡方式',
            id: 'scheduler',
          },
          {
            name: '域名数量',
            id: 'domain_num',
          },
          {
            name: 'URL数量',
            id: 'url_num',
          },
          {
            name: '同步状态',
            id: 'binding_status',
          },
        ],
      },
      tableOptions: {
        columns: [
          { type: 'selection', width: 30, minWidth: 30, onlyShowOnList: true },
          ...columns,
          {
            label: t('操作'),
            field: 'actions',
            render: ({ data }: any) => (
              <div class='operate-groups'>
                <Button text theme='primary' onClick={() => bus.$emit('showEditListenerSideslider', data.id)}>
                  {t('编辑')}
                </Button>
                <Button text theme='primary' onClick={() => handleDeleteListener(data)}>
                  {t('删除')}
                </Button>
              </div>
            ),
          },
        ],
        extra: {
          settings: settings.value,
          isRowSelectEnable,
          onSelectionChange: (selections: any) => handleSelectionChange(selections, isCurRowSelectEnable),
          onSelectAll: (selections: any) => handleSelectionChange(selections, isCurRowSelectEnable, true),
          // new标识
          rowClass: getTableNewRowClass(),
        },
      },
      requestOption: {
        type: `load_balancers/${props.id}/listeners`,
        sortOption: { sort: 'created_at', order: 'DESC' },
        async resolveDataListCb(dataList: any[], getListData) {
          if (dataList.length === 0) return;
          const ids = dataList.filter((item) => item.binding_status === 'binding').map((item) => item.id);
          if (ids.length) {
            timer = setTimeout(() => {
              counter = counter + 1;
              if (counter < 10) {
                getListData();
              } else {
                counter = 0;
                clearTimeout(timer);
              }
            }, 30000);
          }
          // 设置监听器的 rs 权重
          setRsWeight(dataList);
          return dataList;
        },
      },
    });
    const setRsWeight = async (dataList: any[]) => {
      // 为所有监听器设置rs权重初始值
      dataList.forEach((item: any) => Object.assign(item, { rs_weight_zero_num: 0, rs_weight_non_zero_num: 0 }));
      // 绑定了目标组的监听器, 需要查询目标组权重
      const listenersWithTargetGroup = dataList.filter(({ target_group_id }) => !!target_group_id);
      const target_group_ids = listenersWithTargetGroup.map(({ target_group_id }) => target_group_id);
      if (target_group_ids.length) {
        const { data } = await businessStore.reqStatTargetGroupRsWeight(target_group_ids);
        data.forEach((item: any, index: number) => Object.assign(listenersWithTargetGroup[index], item));
      }
    };

    onUnmounted(() => {
      timer && clearTimeout(timer);
    });

    watch(
      () => props.id,
      (id) => {
        // 清空选中项
        resetSelections();
        id && getListData([], `load_balancers/${id}/listeners`);
      },
    );

    // 删除监听器
    const handleDeleteListener = (data: any) => {
      Confirm('请确定删除监听器', `将删除监听器【${data.name}】`, () => {
        businessStore.deleteBatch('listeners', { ids: [data.id] }).then(() => {
          Message({ theme: 'success', message: '删除成功' });
          getListData();
        });
      });
    };

    // 批量删除监听器
    const {
      isSubmitLoading,
      isSubmitDisabled,
      isBatchDeleteDialogShow,
      radioGroupValue,
      tableProps,
      handleBatchDeleteListener,
      handleBatchDeleteSubmit,
      computedListenersList,
    } = useBatchDeleteListener(columns, selections, resetSelections, getListData);
    return () => (
      <div class='listener-list-page'>
        {/* 监听器list */}
        <CommonTable>
          {{
            operation: () => (
              <div class={'flex-row align-item-center'}>
                <Button theme={'primary'} onClick={() => bus.$emit('showAddListenerSideslider')}>
                  <Plus class={'f20'} />
                  {t('新增监听器')}
                </Button>
                <Button disabled={selections.value.length === 0} onClick={handleBatchDeleteListener}>
                  {t('批量删除')}
                </Button>
              </div>
            ),
          }}
        </CommonTable>

        {/* 新增/编辑监听器 */}
        <AddOrUpdateListenerSideslider originPage='lb' getListData={getListData} />

        {/* 批量删除监听器 */}
        <BatchOperationDialog
          class='batch-delete-listener-dialog'
          v-model:isShow={isBatchDeleteDialogShow.value}
          title={t('批量删除监听器')}
          theme='danger'
          confirmText='删除'
          isSubmitLoading={isSubmitLoading.value}
          isSubmitDisabled={isSubmitDisabled.value}
          tableProps={tableProps}
          list={computedListenersList.value}
          onHandleConfirm={handleBatchDeleteSubmit}>
          {{
            tips: () => (
              <>
                已选择<span class='blue'>{tableProps.data.length}</span>个监听器，其中
                <span class='red'>
                  {tableProps.data.filter(({ rs_weight_non_zero_num }) => rs_weight_non_zero_num > 0).length}
                </span>
                个监听器RS的权重均不为0，在删除监听器前，请确认是否有流量转发，仔细核对后，再提交删除。
              </>
            ),
            tab: () => (
              <BkRadioGroup v-model={radioGroupValue.value}>
                <BkRadioButton label={true}>{t('权重为0')}</BkRadioButton>
                <BkRadioButton label={false}>{t('权重不为0')}</BkRadioButton>
              </BkRadioGroup>
            ),
          }}
        </BatchOperationDialog>
      </div>
    );
  },
});
