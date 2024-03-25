import { Ref, defineComponent, ref } from 'vue';
import { Button, Dropdown, InfoBox, Message } from 'bkui-vue';
import { BkRadioGroup, BkRadioButton } from 'bkui-vue/lib/radio';
import { Plus, AngleDown } from 'bkui-vue/lib/icon';
import { ISearchItem } from 'bkui-vue/lib/search-select/utils';
import { useTable } from '@/hooks/useTable/useTable';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import CommonSideslider from '@/components/common-sideslider';
import TargetGroupSidesliderContent from './target-group-sideslider-content';
import CommonDialog from '@/components/common-dialog';
import AddRsDialogContent from './add-rs-dialog-content';
import BatchOperationDialog from '@/components/batch-operation-dialog';
import RsSidesliderContent from './rs-sideslider-content';
import './index.scss';
import { useAccountStore, useBusinessStore } from '@/store';
import useSelection from '@/views/resource/resource-manage/hooks/use-selection';
import { useLoadBalancerStore } from '@/store/loadbalancer';

const { DropdownMenu, DropdownItem } = Dropdown;

export default defineComponent({
  name: 'AllGroupsManager',
  setup() {
    const { columns, settings } = useColumns('targetGroup');
    const businessStore = useBusinessStore();
    const accountStore = useAccountStore();
    const loadBalancerStore = useLoadBalancerStore();
    const { selections, handleSelectionChange } = useSelection();
    const rsList = ref([]);
    const rsSelections = ref([]);
    const rsData = ref([]);
    const submitData = ref([]);
    const isEdit = ref(false);
    const editRecord: Ref<{ id: string }> = ref(null);
    const tableColumns = [
      ...columns,
      {
        label: '操作',
        width: 120,
        render: ({ data }: any) => (
          <div>
            <Button
              text
              theme={'primary'}
              onClick={() => {
                isTargetGroupSidesliderShow.value = true;
                isEdit.value = true;
                editRecord.value = data;
              }}>
              编辑
            </Button>
            <span
              v-bk-tooltips={{
                content: '已绑定了监听器的目标组不可删除',
                disabled: data.listener_num === 0,
              }}>
              <Button
                text
                theme={'primary'}
                disabled={data.listener_num > 0}
                class={'ml16'}
                onClick={() => {
                  handleDeleteTargetGroup(data.id, data.name);
                }}>
                删除
              </Button>
            </span>
          </div>
        ),
      },
    ];

    // 删除单个目标组
    const handleDeleteTargetGroup = (id: string, name: string) => {
      InfoBox({
        title: '请确认是否删除',
        subTitle: `将删除【${name}】`,
        headerAlign: 'center',
        footerAlign: 'center',
        contentAlign: 'center',
        infoType: 'warning',
        onConfirm: async () => {
          await businessStore.deleteTargetGroups({
            bk_biz_id: accountStore.bizs,
            ids: [id],
          });
          getListData();
          Message({
            message: '删除成功',
            theme: 'success',
          });
        },
      });
    };

    const searchData: ISearchItem[] = [
      {
        id: 'target_group_name',
        name: '目标组名称',
      },
      {
        id: 'clb_id',
        name: 'CLB ID',
      },
      {
        id: 'listener_id',
        name: '监听器ID',
      },
      {
        id: 'vip_address',
        name: 'VIP地址',
      },
      {
        id: 'vip_domain',
        name: 'VIP域名',
      },
      {
        id: 'port',
        name: '端口',
      },
      {
        id: 'protocol',
        name: '协议',
      },
      {
        id: 'rs_ip',
        name: 'RS的IP',
      },
    ];
    const { CommonTable, getListData } = useTable({
      searchOptions: {
        searchData,
      },
      tableOptions: {
        columns: tableColumns,
        extra: {
          settings: settings.value,
          onSelect: (selections: any) => {
            handleSelectionChange(selections, () => true, false);
          },
          onSelectAll: (selections: any) => {
            handleSelectionChange(selections, () => true, true);
          },
        },
      },
      requestOption: {
        type: 'target_groups',
      },
    });

    const currentScene = ref('');
    // 新建目标组 sideslider
    const isTargetGroupSidesliderShow = ref(false);
    const isDropdownShow = ref(false);
    const handleAddTargetGroup = () => {
      isEdit.value = false;
      currentScene.value = 'addTargetGroup';
      isTargetGroupSidesliderShow.value = true;
      isDropdownShow.value = false;
    };
    const handleAddTargetGroupSubmit = async () => {
      const promise = isEdit.value
        ? businessStore.editTargetGroups(editRecord.value.id, submitData.value)
        : businessStore.createTargetGroups(submitData.value);
      await promise;
      Message({
        message: isEdit.value ? '编辑成功' : '新建成功',
        theme: 'success',
      });
      isTargetGroupSidesliderShow.value = false;
      getListData();
    };

    // 添加单个RS
    const isAddRsDialogShow = ref(false);
    const handleAddRs = () => {
      rsData.value = rsSelections.value;
      if (currentScene.value === 'addTargetGroup') {
        // todo
      } else {
        // todo
        isBatchAddRsShow.value = true;
      }
    };

    // batch operation
    const isBatchDeleteTargetGroupShow = ref(false);
    const canDeleteTargetGroup = ref(false);
    const isBatchDeleteRsShow = ref(false);
    const isBatchAddRsShow = ref(false);
    const batchOperationList = [
      { scene: 'batchDeleteTargetGroup', isShow: isBatchDeleteTargetGroupShow, label: '批量删除目标组' },
      { scene: 'batchDeleteRs', isShow: isBatchDeleteRsShow, label: '批量移除 RS' },
      { scene: 'batchAddRs', isShow: isAddRsDialogShow, label: '批量添加 RS' },
    ];
    // batch operation item 的点击事件处理函数
    const handleBatchOperationItemClick = (scene: string, isShow: Ref<boolean>) => {
      currentScene.value = scene;
      isDropdownShow.value = false;
      isShow.value = true;
    };
    // 批量删除目标组
    const batchDeleteTargetGroupTableProps = {
      data: selections.value,
      columns: [
        {
          label: '目标组名称',
          field: 'name',
        },
        {
          label: '协议',
          field: 'protocol',
          filter: true,
          render({ cell }: any) {
            return cell.trim() || '--';
          },
        },
        {
          label: '端口',
          field: 'port',
          filter: true,
        },
        {
          label: '关联的负载均衡',
          field: 'lb_name',
          render({ cell }: any) {
            return cell.trim() || '--';
          },
        },
        {
          label: '绑定监听器数量',
          field: 'listener_num',
          sort: true,
          align: 'right',
        },
      ],
      searchData,
    };
    const batchDeleteTargetGroup = async () => {
      await businessStore.deleteTargetGroups({
        bk_biz_id: accountStore.bizs,
        ids: selections.value.map(({ id }) => id),
      });
      Message({
        message: '批量删除成功',
        theme: 'success',
      });
      loadBalancerStore.getTargetGroupList();
    };
    // 批量移除 RS
    const batchDeleteRs = () => {};
    // 批量添加 RS
    const handleBatchAddRsSubmit = () => {};

    return () => (
      <div class='common-card-wrap has-selection'>
        <CommonTable>
          {{
            operation: () => (
              <>
                <Button theme='primary' onClick={handleAddTargetGroup}>
                  <Plus class='f20' />
                  新建
                </Button>
                <Dropdown trigger='manual' isShow={isDropdownShow.value} placement='bottom-start'>
                  {{
                    default: () => (
                      <Button
                        onClick={() => (isDropdownShow.value = !isDropdownShow.value)}
                        disabled={!selections.value.length}>
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
        <CommonSideslider
          title={isEdit.value ? '编辑目标组' : '新建目标组'}
          width={960}
          v-model:isShow={isTargetGroupSidesliderShow.value}
          onHandleSubmit={handleAddTargetGroupSubmit}>
          <TargetGroupSidesliderContent
            isEdit={isEdit.value}
            editData={editRecord.value}
            onShowAddRsDialog={(payload) => {
              isAddRsDialogShow.value = true;
              rsList.value = payload;
            }}
            onChange={(data: any) => {
              submitData.value = data;
            }}
            rsTableData={rsData.value}
          />
        </CommonSideslider>
        <CommonDialog
          v-model:isShow={isAddRsDialogShow.value}
          title='添加 RS'
          width={640}
          onHandleConfirm={handleAddRs}>
          <AddRsDialogContent
            rsList={rsList.value}
            onSelect={(selections: any) => (rsSelections.value = selections)}
            rsTableData={rsData.value}
          />
        </CommonDialog>
        <BatchOperationDialog
          v-model:isShow={isBatchDeleteTargetGroupShow.value}
          title='批量删除目标组'
          theme='danger'
          confirmText='删除'
          tableProps={batchDeleteTargetGroupTableProps}
          onHandleConfirm={batchDeleteTargetGroup}>
          {{
            tips: () => (
              <>
                已选择 <span class='blue'>{selections.value.length}</span> 个目标组，其中可删除{' '}
                <span class='green'> {selections.value.length} </span> 个, 不可删除{' '}
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
        <BatchOperationDialog
          v-model:isShow={isBatchDeleteRsShow.value}
          title='批量移除RS'
          theme='danger'
          confirmText='移除 RS'
          custom
          onHandleConfirm={batchDeleteRs}>
          {{}}
        </BatchOperationDialog>
        <CommonSideslider
          title='批量添加 RS'
          width={960}
          v-model:isShow={isBatchAddRsShow.value}
          onHandleSubmit={handleBatchAddRsSubmit}>
          <RsSidesliderContent
            selectedTargetGroups={loadBalancerStore.allTargetGroupList}
            onShowAddRsDialog={() => (isAddRsDialogShow.value = true)}
          />
        </CommonSideslider>
      </div>
    );
  },
});
