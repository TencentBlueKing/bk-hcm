import { Ref, defineComponent, ref } from 'vue';
import { Button, Dropdown } from 'bkui-vue';
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

const { DropdownMenu, DropdownItem } = Dropdown;

export default defineComponent({
  name: 'AllGroupsManager',
  setup() {
    const { columns, settings } = useColumns('targetGroup');
    const tableColumns = [
      ...columns,
      {
        label: '操作',
        width: 120,
        render: () => (
          <div class='operate-groups'>
            <span>编辑</span>
            <span>删除</span>
          </div>
        ),
      },
    ];
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
    const tableData = [
      {
        target_group_name: 'TargetGroup1',
        clb_name: 'CLB1',
        listener_count: 3,
        protocol: 'HTTP',
        port: 80,
        vendor: 'Amazon',
        region: 'us-west-1',
        zone: 'us-west-1a',
        type: 'public',
        vpc_id: 'vpc-1234abcd',
        health_check_port: 8080,
        ip_type: 'ipv4',
      },
      {
        target_group_name: 'TargetGroup2',
        clb_name: 'CLB2',
        listener_count: 5,
        protocol: 'HTTPS',
        port: 443,
        vendor: 'Amazon',
        region: 'eu-central-1',
        zone: 'eu-central-1b',
        type: 'internal',
        vpc_id: 'vpc-5678efgh',
        health_check_port: 8443,
        ip_type: 'ipv6',
      },
      {
        target_group_name: 'TargetGroup3',
        clb_name: 'CLB3',
        listener_count: 2,
        protocol: 'TCP',
        port: 22,
        vendor: 'Amazon',
        region: 'ap-southeast-1',
        zone: 'ap-southeast-1c',
        type: 'public',
        vpc_id: 'vpc-90ab12cd',
        health_check_port: 8000,
        ip_type: 'ipv4',
      }, // 尾后逗号
    ];
    const { CommonTable } = useTable({
      searchOptions: {
        searchData,
      },
      tableOptions: {
        columns: tableColumns,
        reviewData: tableData,
        extra: {
          settings: settings.value,
        },
      },
      requestOption: {
        type: '',
      },
    });

    const currentScene = ref('');
    // 新建目标组 sideslider
    const isTargetGroupSidesliderShow = ref(false);
    const isDropdownShow = ref(false);
    const handleAddTargetGroup = () => {
      currentScene.value = 'addTargetGroup';
      isTargetGroupSidesliderShow.value = true;
      isDropdownShow.value = false;
    };
    const handleAddTargetGroupSubmit = () => {};

    // 添加单个RS
    const isAddRsDialogShow = ref(false);
    const handleAddRs = () => {
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
      data: tableData,
      columns: [
        {
          label: '目标组名称',
          field: 'target_group_name',
        },
        {
          label: '协议',
          field: 'protocol',
          filter: true,
        },
        {
          label: '端口',
          field: 'port',
          filter: true,
        },
        {
          label: '关联的负载均衡',
          field: 'clb_name',
        },
        {
          label: '绑定监听器数量',
          field: 'listener_count',
          sort: true,
          align: 'right',
        },
      ],
      searchData,
    };
    const batchDeleteTargetGroup = () => {};
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
                      <Button onClick={() => (isDropdownShow.value = !isDropdownShow.value)}>
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
          title='新建目标组'
          width={960}
          v-model:isShow={isTargetGroupSidesliderShow.value}
          onHandleSubmit={handleAddTargetGroupSubmit}>
          <TargetGroupSidesliderContent onShowAddRsDialog={() => (isAddRsDialogShow.value = true)} />
        </CommonSideslider>
        <CommonDialog
          v-model:isShow={isAddRsDialogShow.value}
          title='添加 RS'
          width={640}
          onHandleConfirm={handleAddRs}>
          <AddRsDialogContent />
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
                已选择 <span class='blue'>150</span> 个目标组，其中可删除 <span class='green'>138</span> 个, 不可删除{' '}
                <span class='red'>12</span> 个（已绑定了监听器的目标组不可删除）。
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
            selectedTargetGroups={tableData}
            onShowAddRsDialog={() => (isAddRsDialogShow.value = true)}
          />
        </CommonSideslider>
      </div>
    );
  },
});
