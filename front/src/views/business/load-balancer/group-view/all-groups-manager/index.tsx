import { computed, ComputedRef, defineComponent, inject } from 'vue';
// import components
import { Button, Checkbox, Dropdown, Loading, SearchSelect, Table } from 'bkui-vue';
import { BkRadioGroup, BkRadioButton } from 'bkui-vue/lib/radio';
import { Plus, AngleDown } from 'bkui-vue/lib/icon';
import AddOrUpdateTGSideslider from '../components/AddOrUpdateTGSideslider';
import BatchOperationDialog from '@/components/batch-operation-dialog';
import AddRsDialog from '../components/AddRsDialog';
import BatchAddRsSideslider from './BatchAddRsSideslider';
// import stores
import { useBusinessStore, useLoadBalancerStore } from '@/store';
// import custom hooks
import useRenderTRList from './useRenderTGList';
import useBatchDeleteTR from './useBatchDeleteTR';
import useBatchDeleteRs from './useBatchDeleteRs';
import { useI18n } from 'vue-i18n';
// import utils
import bus from '@/common/bus';
import './index.scss';
import { useVerify } from '@/hooks';
import { useGlobalPermissionDialog } from '@/store/useGlobalPermissionDialog';
import { TargetGroupOperationScene } from '@/constants';

const { DropdownMenu, DropdownItem } = Dropdown;

export default defineComponent({
  name: 'AllGroupsManager',
  setup() {
    // use hooks
    const { t } = useI18n();
    const { authVerifyData, handleAuth } = useVerify();
    const globalPermissionDialogStore = useGlobalPermissionDialog();
    const createClbActionName: ComputedRef<'clb_resource_create' | 'biz_clb_resource_create'> =
      inject('createClbActionName');
    // use stores
    const loadBalancerStore = useLoadBalancerStore();
    const businessStore = useBusinessStore();

    // 目标组list
    const { searchData, selections, CommonTable, getListData } = useRenderTRList();

    // 批量删除目标组
    const {
      isSubmitLoading,
      isBatchDeleteTargetGroupShow,
      canDeleteTargetGroup,
      batchDeleteTargetGroupTableProps,
      batchDeleteTargetGroup,
      computedListenersList,
      isSubmitDisabled,
    } = useBatchDeleteTR(searchData, selections, getListData);

    // 批量移除RS
    const {
      isBatchDeleteRsShow,
      isBatchDeleteRsSubmitLoading,
      isBatchDeleteRsTableLoading,
      batchDeleteRsTableColumn,
      batchDeleteRsTableData,
      initMap,
      getRsListOfTargetGroups,
      handleChangeChecked,
      batchDeleteRsSearchData,
      batchDeleteRs,
    } = useBatchDeleteRs();

    // computed-property - 判断是否属于同一个账号&负载均衡(lb_id为空时, 不算做一种)
    const isSelectionsBelongSameAccountAndLB = computed(() => {
      if (selections.value.length < 2) return true;
      const firstAccountId = selections.value[0].account_id;
      const firstLBId = selections.value.find((item) => item.lb_id).lb_id;
      return selections.value.every((item) => {
        return (item.lb_id === firstLBId || item.lb_id === '') && item.account_id === firstAccountId;
      });
    });

    // computed-property - 判断是否同属于一个vpc
    const isSelectionsBelongSameVpc = computed(() => {
      if (selections.value.length < 2) return true;
      const firstVpcId = selections.value[0].vpc_id;
      return selections.value.every((item) => item.vpc_id === firstVpcId);
    });

    // click-handler - 批量删除目标组
    const handleBatchDeleteTG = () => {
      loadBalancerStore.setCurrentScene(TargetGroupOperationScene.BATCH_DELETE);
      isBatchDeleteTargetGroupShow.value = true;
    };
    // click-handler - 批量删除RS
    const handleBatchDeleteRs = () => {
      loadBalancerStore.setCurrentScene(TargetGroupOperationScene.BATCH_DELETE_RS);
      // init
      initMap(
        selections.value.map(({ id }) => id),
        selections.value[0].account_id,
      );
      // 请求选中目标组的rs列表
      getRsListOfTargetGroups(selections.value);
      isBatchDeleteRsShow.value = true;
    };
    // click-handler - 批量添加RS
    const handleBatchAddRs = async () => {
      loadBalancerStore.setCurrentScene(TargetGroupOperationScene.BATCH_ADD_RS);
      // 同一账号下的多个目标组支持批量添加rs
      const { account_id: accountId, lb_id } = selections.value[0];

      // 获取负载均衡的跨域信息
      let isCorsV2 = false;
      if (lb_id) {
        const res = await businessStore.getLbDetail(lb_id);
        isCorsV2 = res.data.extension?.snat_pro;
      }
      const vpcIds = [...new Set(selections.value.map((item) => item.vpc_id))];

      // 显示添加rs弹框
      bus.$emit('showAddRsDialog', { accountId, vpcIds, rsList: [], isCorsV2 });
      // 将选中的目标组数据传递给BatchAddRsSideslider组件
      bus.$emit('setTargetGroups', selections.value);
    };

    return () => (
      <div class='common-card-wrap'>
        {/* 目标组list */}
        <CommonTable>
          {{
            operation: () => (
              <>
                <Button
                  theme='primary'
                  onClick={() => {
                    if (!authVerifyData?.value?.permissionAction?.[createClbActionName.value]) {
                      handleAuth(createClbActionName.value);
                      globalPermissionDialogStore.setShow(true);
                    } else bus.$emit('addTargetGroup');
                  }}
                  class={[
                    'mr8',
                    { 'hcm-no-permision-btn': !authVerifyData?.value?.permissionAction?.[createClbActionName.value] },
                  ]}>
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
                        <DropdownItem>
                          <Button text onClick={handleBatchDeleteTG}>
                            {t('批量移除目标组')}
                          </Button>
                        </DropdownItem>
                        <DropdownItem>
                          <Button
                            text
                            onClick={handleBatchDeleteRs}
                            disabled={!isSelectionsBelongSameAccountAndLB.value}
                            v-bk-tooltips={{
                              content: '传入的目标组不同属于一个负载均衡/账号, 不可进行批量移除RS操作',
                              disabled: isSelectionsBelongSameAccountAndLB.value,
                            }}>
                            {t('批量移除 RS')}
                          </Button>
                        </DropdownItem>
                        <DropdownItem>
                          <Button
                            text
                            onClick={handleBatchAddRs}
                            disabled={!isSelectionsBelongSameAccountAndLB.value || !isSelectionsBelongSameVpc.value}
                            v-bk-tooltips={
                              !isSelectionsBelongSameAccountAndLB.value
                                ? {
                                    content: '传入的目标组不同属于一个负载均衡/账号, 不可进行批量添加RS操作',
                                    disabled: isSelectionsBelongSameAccountAndLB.value,
                                  }
                                : {
                                    content: '传入的目标组不同属于一个VPC, 不可进行批量添加RS操作',
                                    disabled: isSelectionsBelongSameVpc.value,
                                  }
                            }>
                            {t('批量添加 RS')}
                          </Button>
                        </DropdownItem>
                      </DropdownMenu>
                    ),
                  }}
                </Dropdown>
              </>
            ),
          }}
        </CommonTable>
        {/* 新增/编辑目标组 */}
        <AddOrUpdateTGSideslider origin='list' getListData={getListData} />
        {/* 添加RS */}
        <AddRsDialog />
        {/* 批量删除目标组 */}
        <BatchOperationDialog
          v-model:isShow={isBatchDeleteTargetGroupShow.value}
          isSubmitLoading={isSubmitLoading.value}
          isSubmitDisabled={isSubmitDisabled.value}
          title='批量删除目标组'
          theme='danger'
          confirmText='删除'
          tableProps={batchDeleteTargetGroupTableProps}
          list={computedListenersList.value}
          onHandleConfirm={batchDeleteTargetGroup}>
          {{
            tips: () => (
              <>
                已选择 <span class='blue'>{selections.value.length}</span> 个目标组，其中可删除
                <span class='green'>{selections.value.filter(({ listener_num }) => listener_num === 0).length}</span>
                个, 不可删除
                <span class='red'>{selections.value.filter(({ listener_num }) => listener_num > 0).length}</span>
                个（已绑定了监听器的目标组不可删除）。
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
          class='batch-delete-rs-dialog'
          isSubmitLoading={isBatchDeleteRsSubmitLoading.value}
          v-model:isShow={isBatchDeleteRsShow.value}
          title='批量移除RS'
          theme='danger'
          confirmText='移除 RS'
          custom
          onHandleConfirm={batchDeleteRs}>
          <div class='top-area'>
            <div class='tips'>
              已选择<span class='blue'>{selections.value.length}</span>
              个目标组，可选择当前目标组内需要删除的IP进行移除。
            </div>
            <SearchSelect class='w400' data={batchDeleteRsSearchData} />
          </div>
          <Loading loading={isBatchDeleteRsTableLoading.value}>
            <Table data={batchDeleteRsTableData.value} columns={batchDeleteRsTableColumn} rowHeight={32} border='none'>
              {{
                expandContent: (row: any) => (
                  <span class='i-expand-content'>
                    <span class='main'>{row.tgName}</span>
                    <span class='extension'>（共有 {row.ipCount} 个 IP）</span>
                  </span>
                ),
                expandRow: (row: any) => {
                  return row.ipList.map((item: any) => {
                    return (
                      <div class='i-expand-row'>
                        {Object.getOwnPropertyNames(item).map((key) => {
                          if (key === 'id') return null;
                          return (
                            <div class='i-expand-cell'>
                              {key === 'isChecked' && (
                                <Checkbox
                                  modelValue={item[key]}
                                  onUpdate:modelValue={(isChecked) => handleChangeChecked(row.tgId, item, isChecked)}
                                />
                              )}
                              <span class='i-cell-content'>
                                {Array.isArray(item[key]) ? item[key].join(',') : item[key]}
                              </span>
                            </div>
                          );
                        })}
                      </div>
                    );
                  });
                },
              }}
            </Table>
          </Loading>
        </BatchOperationDialog>
        {/* 批量添加RS */}
        <BatchAddRsSideslider />
      </div>
    );
  },
});
