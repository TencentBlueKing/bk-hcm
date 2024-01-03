import { useTable } from '@/hooks/useTable/useTable';
import { defineComponent, ref } from 'vue';
import './index.scss';
import { Button } from 'bkui-vue';
import { Plus } from 'bkui-vue/lib/icon';
import BatchOperationDialog from '@/components/batch-operation-dialog';
import { BkRadioButton, BkRadioGroup } from 'bkui-vue/lib/radio';

export default defineComponent({
  setup() {
    const isBatchDeleteDialogShow = ref(false);
    const radioGroupValue = ref(false);
    const tableProps = {
      columns: [
        {
          label: 'URL路径',
          field: 'urlPath',
        },
        {
          label: '协议',
          field: 'protocol',
        },
        {
          label: '端口',
          field: 'port',
        },
        {
          label: '轮询方式',
          field: 'pollingMethod',
        },
        {
          label: '目标组',
          field: 'targetGroup',
        },
        {
          label: '同步状态',
          field: 'syncStatus',
        },
        {
          label: '操作',
          field: 'actions',
        },
      ],
      data: [
        {
          urlPath: '/home',
          protocol: 'HTTP',
          port: 80,
          pollingMethod: 'RoundRobin',
          targetGroup: 'GroupA',
          syncStatus: 'Synchronized',
          actions: 'Edit',
        },
        {
          urlPath: '/about',
          protocol: 'HTTPS',
          port: 443,
          pollingMethod: 'LeastConnections',
          targetGroup: 'GroupB',
          syncStatus: 'Pending',
          actions: 'Delete',
        },
        {
          urlPath: '/contact',
          protocol: 'TCP',
          port: 22,
          pollingMethod: 'SourceIP',
          targetGroup: 'GroupC',
          syncStatus: 'Failed',
          actions: 'Update',
        },
      ],
      searchData: [
        {
          name: 'URL路径',
          id: 'urlPath',
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
          name: '轮询方式',
          id: 'pollingMethod',
        },
        {
          name: '目标组',
          id: 'targetGroup',
        },
        {
          name: '同步状态',
          id: 'syncStatus',
        },
        {
          name: '操作',
          id: 'actions',
        },
      ],
    };
    const { CommonTable } = useTable({
      columns: [
        {
          label: 'URL路径',
          field: 'urlPath',
        },
        {
          label: '协议',
          field: 'protocol',
        },
        {
          label: '端口',
          field: 'port',
        },
        {
          label: '轮询方式',
          field: 'pollingMethod',
        },
        {
          label: '目标组',
          field: 'targetGroup',
        },
        {
          label: '同步状态',
          field: 'syncStatus',
        },
        {
          label: '操作',
          field: 'actions',
        },
      ],
      settings: {
        fields: [],
        checked: [],
        limit: 0,
        size: '',
        sizeList: [],
        showLineHeight: false,
      },
      searchData: [
        {
          name: 'URL路径',
          id: 'urlPath',
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
          name: '轮询方式',
          id: 'pollingMethod',
        },
        {
          name: '目标组',
          id: 'targetGroup',
        },
        {
          name: '同步状态',
          id: 'syncStatus',
        },
        {
          name: '操作',
          id: 'actions',
        },
      ],
      searchUrl: '',
      tableData: [
        {
          urlPath: '/home',
          protocol: 'HTTP',
          port: 80,
          pollingMethod: 'RoundRobin',
          targetGroup: 'GroupA',
          syncStatus: 'Synchronized',
          actions: 'Edit',
        },
        {
          urlPath: '/about',
          protocol: 'HTTPS',
          port: 443,
          pollingMethod: 'LeastConnections',
          targetGroup: 'GroupB',
          syncStatus: 'Pending',
          actions: 'Delete',
        },
        {
          urlPath: '/contact',
          protocol: 'TCP',
          port: 22,
          pollingMethod: 'SourceIP',
          targetGroup: 'GroupC',
          syncStatus: 'Failed',
          actions: 'Update',
        },
      ],
    });

    const handleBatchDelete = () => {};

    return () => (
      <div class={'url-list-container'}>
        <CommonTable>
        {{
          operation: () => (
              <div class={'flex-row align-item-center'}>
                <Button theme={'primary'}><Plus class={'f20'}/>新增 URL 路径</Button>
                <Button onClick={() => isBatchDeleteDialogShow.value = true}>批量删除</Button>
              </div>
          ),
        }}
        </CommonTable>
        <BatchOperationDialog
          v-model:isShow={isBatchDeleteDialogShow.value}
          title='批量删除 URL 路径'
          theme='danger'
          confirmText='删除'
          tableProps={tableProps}
          onHandleConfirm={handleBatchDelete}>
          {{
            tips: () => (
              <>
                已选择 <span class='blue'>97</span> 个URL路径，其中 <span class='red'>22</span>
              </>
            ),
            tab: () => (
              <BkRadioGroup v-model={radioGroupValue.value}>
                <BkRadioButton label={false}>权重为0</BkRadioButton>
                <BkRadioButton label={true}>权重不为0</BkRadioButton>
              </BkRadioGroup>
            ),
          }}
        </BatchOperationDialog>
      </div>
    );
  },
});
