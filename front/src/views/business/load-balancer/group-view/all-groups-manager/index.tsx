import { Ref, defineComponent, onMounted, onUnmounted, ref } from 'vue';
// import components
import { Button, Dropdown } from 'bkui-vue';
import { BkRadioGroup, BkRadioButton } from 'bkui-vue/lib/radio';
import { Plus, AngleDown } from 'bkui-vue/lib/icon';
import AddOrUpdateTGSideslider from './AddOrUpdateTGSideslider';
import BatchOperationDialog from '@/components/batch-operation-dialog';
import AddRsDialog from './AddRsDialog';
import BatchAddRsSideslider from './BatchAddRsSideslider';
// import stores
import { useLoadBalancerStore } from '@/store';
// import custom hooks
import useRenderTRList from './useRenderTGList';
import useBatchDeleteTR from './useBatchDeleteTR';
// import utils
import bus from '@/common/bus';
import './index.scss';

const { DropdownMenu, DropdownItem } = Dropdown;

export default defineComponent({
  name: 'AllGroupsManager',
  setup() {
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

    // todo: 批量操作等联调时再继续refactor
    const isBatchDeleteRsShow = ref(false);
    const isBatchAddRsShow = ref(false);
    const batchOperationList = [
      { scene: 'batchDeleteTargetGroup', isShow: isBatchDeleteTargetGroupShow, label: '批量删除目标组' },
      { scene: 'batchDeleteRs', isShow: isBatchDeleteRsShow, label: '批量移除 RS' },
      { scene: 'batchAddRs', isShow: null, label: '批量添加 RS' },
    ];
    // batch operation item 的点击事件处理函数
    const handleBatchOperationItemClick = (scene: string, isShow: Ref<boolean>) => {
      loadBalancerStore.setCurrentScene(scene);
      if (scene !== 'batchAddRs') {
        isShow.value = true;
      } else {
        // todo: 这里可能需要将 selections 中的每个目标组对应的 account_id 提取出来, 请求对应的 cvms 数据(目前是报错状态, 后续处理)
        bus.$emit('showAddRsDialog');
      }
    };

    // 批量移除 RS
    const batchDeleteRs = () => {};

    onMounted(() => {
      bus.$on('showBatchAddRsDialog', () => {
        isBatchAddRsShow.value = true;
      });
    });

    onUnmounted(() => {
      bus.$off('showBatchAddRsDialog');
    });

    return () => (
      <div class='common-card-wrap has-selection'>
        {/* 目标组list */}
        <CommonTable>
          {{
            operation: () => (
              <>
                <Button theme='primary' onClick={() => bus.$emit('addTargetGroup')}>
                  <Plus class='f20' />
                  新建
                </Button>
                <Dropdown trigger='click' placement='bottom-start'>
                  {{
                    default: () => (
                      <Button disabled={!selections.value.length}>
                        批量操作 <AngleDown class='f20' />
                      </Button>
                    ),
                    content: () => (
                      <DropdownMenu>
                        {batchOperationList.map(({ scene, isShow, label }) => (
                          <DropdownItem onClick={() => handleBatchOperationItemClick(scene, isShow)}>
                            {label}
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
        {/* 新增/编辑目标组 - SideSlider */}
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
        {/* 批量删除RS */}
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
