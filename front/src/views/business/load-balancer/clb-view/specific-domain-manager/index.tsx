import { useTable } from '@/hooks/useTable/useTable';
import { defineComponent, ref } from 'vue';
import './index.scss';
import { Button, Form, Input, Select } from 'bkui-vue';
import { Plus } from 'bkui-vue/lib/icon';
import BatchOperationDialog from '@/components/batch-operation-dialog';
import { BkRadioButton, BkRadioGroup } from 'bkui-vue/lib/radio';
import CommonSideslider from '@/components/common-sideslider';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
const { FormItem } = Form;

export default defineComponent({
  setup() {
    const isBatchDeleteDialogShow = ref(false);
    const radioGroupValue = ref(false);
    const isDomainSidesliderShow = ref(false);
    const { columns } = useColumns('url');
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
      columns,
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
          syncStatus: 'a',
          actions: 'Edit',
        },
        {
          urlPath: '/about',
          protocol: 'HTTPS',
          port: 443,
          pollingMethod: 'LeastConnections',
          targetGroup: 'GroupB',
          syncStatus: 'b',
          actions: 'Delete',
        },
        {
          urlPath: '/contact',
          protocol: 'TCP',
          port: 22,
          pollingMethod: 'SourceIP',
          targetGroup: 'GroupC',
          syncStatus: 'c',
          actions: 'Update',
        },
        {
          urlPath: '/contact',
          protocol: 'TCP',
          port: 22,
          pollingMethod: 'SourceIP',
          targetGroup: 'GroupC',
          syncStatus: 'd',
          actions: 'Update',
        },
      ],
    });

    const handleBatchDelete = () => {};
    const handleSubmit = () => {};

    return () => (
      <div class={'url-list-container'}>
        <CommonTable>
        {{
          operation: () => (
              <div class={'flex-row align-item-center'}>
                <Button theme={'primary'} onClick={() => isDomainSidesliderShow.value = true}><Plus class={'f20'}/>新增 URL 路径</Button>
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
        <CommonSideslider
          title='新建域名'
          width={640}
          v-model:isShow={isDomainSidesliderShow.value}
          onHandleSubmit={handleSubmit}>
            <p class={'create-url-text-item'}>
              <span class={'create-url-text-item-label'}>监听器名称：</span>
              <span class={'create-url-text-item-value'}>web站点</span>
            </p>
            <p class={'create-url-text-item'}>
              <span class={'create-url-text-item-label'}>协议端口：</span>
              <span class={'create-url-text-item-value'}>666666</span>
            </p>
            <p class={'create-url-text-item'}>
              <span class={'create-url-text-item-label'}>域名：</span>
              <span class={'create-url-text-item-value'}>aaaaaaaaaa</span>
            </p>
            <Form formType='vertical'>
              <FormItem label='URL 路径'>
                <Input />
              </FormItem>
              <FormItem label='均衡方式'>
                <Select />
              </FormItem>
              <FormItem label='目标组'>
                <Select />
              </FormItem>
            </Form>
        </CommonSideslider>
      </div>
    );
  },
});
