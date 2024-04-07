import { computed, defineComponent, ref } from 'vue';
// import components
import { Button, Dropdown } from 'bkui-vue';
import { BkRadioGroup, BkRadioButton } from 'bkui-vue/lib/radio';
import { Plus, AngleDown } from 'bkui-vue/lib/icon';
import AddOrUpdateTGSideslider from '../components/AddOrUpdateTGSideslider';
import BatchOperationDialog from '@/components/batch-operation-dialog';
import AddRsDialog from '../components/AddRsDialog';
import BatchAddRsSideslider from './BatchAddRsSideslider';
// import stores
import { useLoadBalancerStore } from '@/store';
// import custom hooks
import useRenderTRList from './useRenderTGList';
import useBatchDeleteTR from './useBatchDeleteTR';
import { useI18n } from 'vue-i18n';
// import utils
import bus from '@/common/bus';
import './index.scss';

const { DropdownMenu, DropdownItem } = Dropdown;

export default defineComponent({
  name: 'AllGroupsManager',
  setup() {
    // use hooks
    const { t } = useI18n();
    // use stores
    const loadBalancerStore = useLoadBalancerStore();

    // 目标组list
    const { searchData, selections, CommonTable, getListData } = useRenderTRList();

    // 批量删除目标组
    const {
      isSubmitLoading,
      isBatchDeleteTargetGroupShow,
      canDeleteTargetGroup,
      batchDeleteTargetGroupTableProps,
      batchDeleteTargetGroup,
    } = useBatchDeleteTR(searchData, selections, getListData);

    // computed-property - 判断是否属于同一个账号
    const isSelectionsBelongSameAccount = computed(() => {
      if (selections.value.length < 2) return true;
      const firstAccountId = selections.value[0].account_id;
      return selections.value.every((item) => item.account_id === firstAccountId);
    });

    // click-handler - 批量删除目标组
    const handleBatchDeleteTG = () => {
      loadBalancerStore.setCurrentScene('BatchDelete');
      isBatchDeleteTargetGroupShow.value = true;
    };
    // click-handler - 批量删除RS
    const handleBatchDeleteRs = () => {
      loadBalancerStore.setCurrentScene('BatchDeleteRs');
    };
    // click-handler - 批量添加RS
    const handleBatchAddRs = () => {
      loadBalancerStore.setCurrentScene('BatchAddRs');
      // 同一账号下的多个目标组支持批量添加rs
      const { account_id } = selections.value[0];
      bus.$emit('showAddRsDialog', account_id);
      bus.$emit('setTargetGroups', selections.value);
    };
    const batchOperationList = [
      { label: t('批量删除目标组'), clickHandler: handleBatchDeleteTG, disabled: false },
      { label: t('批量移除 RS'), clickHandler: handleBatchDeleteRs, disabled: !isSelectionsBelongSameAccount.value },
      { label: t('批量添加 RS'), clickHandler: handleBatchAddRs, disabled: !isSelectionsBelongSameAccount.value },
    ];

    // todo: 批量移除 RS
    const isBatchDeleteRsShow = ref(false);
    const batchDeleteRs = () => {};

    return () => (
      <div class='common-card-wrap has-selection'>
        {/* 目标组list */}
        <CommonTable>
          {{
            operation: () => (
              <>
                <Button theme='primary' onClick={() => bus.$emit('addTargetGroup')}>
                  <Plus class='f20' />
                  {t('新建')}
                </Button>
                <Dropdown trigger='click' placement='bottom-start'>
                  {{
                    default: () => (
                      <Button disabled={!selections.value.length}>
                        {t('批量操作')} <AngleDown class='f20' />
                      </Button>
                    ),
                    content: () => (
                      <DropdownMenu>
                        {batchOperationList.map(({ label, clickHandler, disabled }) => (
                          <DropdownItem>
                            <Button text onClick={clickHandler} disabled={disabled}>
                              {label}
                            </Button>
                          </DropdownItem>
                        ))}
                      </DropdownMenu>
                    ),
                  }}
                </Dropdown>
              </>
            ),
          }}
        </CommonTable>
        {/* 新增/编辑目标组 */}
        <AddOrUpdateTGSideslider getListData={getListData} />
        {/* 添加RS */}
        <AddRsDialog />
        {/* 批量删除目标组 */}
        <BatchOperationDialog
          v-model:isShow={isBatchDeleteTargetGroupShow.value}
          isSubmitLoading={isSubmitLoading.value}
          title='批量删除目标组'
          theme='danger'
          confirmText='删除'
          tableProps={batchDeleteTargetGroupTableProps}
          onHandleConfirm={batchDeleteTargetGroup}>
          {{
            tips: () => (
              <>
                已选择 <span class='blue'>{selections.value.length}</span> 个目标组，其中可删除
                <span class='green'> {selections.value.length} </span> 个, 不可删除
                <span class='red'> {selections.value.length} </span> 个（已绑定了监听器的目标组不可删除）。
              </>
            ),
            tab: () => (
              <BkRadioGroup v-model={canDeleteTargetGroup.value}>
                <BkRadioButton label={true}>可删除</BkRadioButton>
                <BkRadioButton label={false}>不可删除</BkRadioButton>
              </BkRadioGroup>
            ),
          }}
        </BatchOperationDialog>
        {/* todo: 批量删除RS */}
        <BatchOperationDialog
          v-model:isShow={isBatchDeleteRsShow.value}
          title='批量移除RS'
          theme='danger'
          confirmText='移除 RS'
          custom
          onHandleConfirm={batchDeleteRs}>
          {{}}
        </BatchOperationDialog>
        {/* 批量添加RS */}
        <BatchAddRsSideslider />
      </div>
    );
  },
});
