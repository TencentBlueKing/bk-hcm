import { computed, defineComponent, reactive } from 'vue';
// import components
import { Button, Checkbox, Dropdown, Loading, SearchSelect, Table } from 'bkui-vue';
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
import useBatchDeleteRs from './useBatchDeleteRs';
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
      // 过滤掉lb_id为空的目标组
      const tmpSelections = selections.value.filter(({ lb_id }) => !lb_id);
      if (tmpSelections.length < 2) return true;
      const { account_id: firstAccountId, lb_id: firstLBId } = tmpSelections[0];
      return tmpSelections.every((item) => item.lb_id === firstLBId && item.account_id === firstAccountId);
    });

    // computed-property - 判断是否同属于一个vpc
    const isSelectionsBelongSameVpc = computed(() => {
      if (selections.value.length < 2) return true;
      const firstVpcId = selections.value[0].vpc_id;
      return selections.value.every((item) => item.vpc_id === firstVpcId);
    });

    // click-handler - 批量删除目标组
    const handleBatchDeleteTG = () => {
      loadBalancerStore.setCurrentScene('BatchDelete');
      isBatchDeleteTargetGroupShow.value = true;
    };
    // click-handler - 批量删除RS
    const handleBatchDeleteRs = () => {
      loadBalancerStore.setCurrentScene('BatchDeleteRs');
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
    const handleBatchAddRs = () => {
      loadBalancerStore.setCurrentScene('BatchAddRs');
      // 同一账号下的多个目标组支持批量添加rs
      const { account_id: accountId, vpc_id: vpcId } = selections.value[0];
      // 显示添加rs弹框
      bus.$emit('showAddRsDialog', { accountId, vpcId });
      // 将选中的目标组数据传递给BatchAddRsSideslider组件
      bus.$emit('setTargetGroups', selections.value);
    };
    const batchOperationList = reactive([
      { label: t('批量删除目标组'), clickHandler: handleBatchDeleteTG, enabled: true },
      {
        label: t('批量移除 RS'),
        clickHandler: handleBatchDeleteRs,
        enabled: isSelectionsBelongSameAccountAndLB,
        tips: t('传入的目标组不同属于一个负载均衡/账号, 不可进行批量移除RS操作'),
      },
      {
        label: t('批量添加 RS'),
        clickHandler: handleBatchAddRs,
        enabled: isSelectionsBelongSameAccountAndLB.value && isSelectionsBelongSameVpc.value,
        tips: isSelectionsBelongSameVpc.value
          ? t('传入的目标组不同属于一个负载均衡/账号, 不可进行批量添加RS操作')
          : t('传入的目标组不同属于一个VPC, 不可进行批量添加RS操作'),
      },
    ]);

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
                        {batchOperationList.map(({ label, clickHandler, enabled, tips }) => (
                          <DropdownItem>
                            <Button
                              text
                              onClick={clickHandler}
                              disabled={!enabled}
                              v-bk-tooltips={{ content: tips, disabled: enabled }}>
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
        <AddOrUpdateTGSideslider origin='list' getListData={getListData} />
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
          <Loading loading={isBatchDeleteRsTableLoading.value} class='has-selection'>
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
