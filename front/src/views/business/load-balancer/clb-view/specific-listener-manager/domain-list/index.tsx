import { defineComponent, ref } from 'vue';
import { Button, Tag } from 'bkui-vue';
import { BkRadioGroup, BkRadioButton } from 'bkui-vue/lib/radio';
import { Plus } from 'bkui-vue/lib/icon';
import { useTable } from '@/hooks/useTable/useTable';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import CommonSideslider from '@/components/common-sideslider';
import DomainSidesliderContent from '../domain-sideslider-content';
import BatchOperationDialog from '@/components/batch-operation-dialog';
import './index.scss';

export default defineComponent({
  name: 'DomainList',
  setup() {
    const { columns, settings } = useColumns('domain');
    const tableColumns = [
      {
        type: 'selection',
        width: 32,
        minWidth: 32,
        align: 'right',
      },
      {
        label: '域名',
        field: 'domain',
        isDefaultShow: true,
        render: ({ data, cell }: { data: any; cell: string }) => {
          return (
            <div class='set-default-operation-wrap'>
              <span class='cell-value'>{cell}</span>
              {data?.is_default ? (
                <Tag theme='info' class='default-tag'>
                  默认
                </Tag>
              ) : (
                <Button text theme='primary' class='set-default-btn'>
                  设为默认
                </Button>
              )}
            </div>
          );
        },
      },
      ...columns,
      {
        label: '操作',
        width: 120,
        render() {
          return (
            <div class='operate-groups'>
              <span>编辑</span>
              <span>删除</span>
            </div>
          );
        },
      },
    ];
    const searchData: any = [
      {
        name: '域名',
        id: 'domain',
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
        id: 'polling_method',
      },
      {
        name: 'URL数量',
        id: 'url_count',
      },
      {
        name: '同步状态',
        id: 'sync_status',
      },
    ];
    const { CommonTable } = useTable({
      searchOptions: { searchData },
      tableOptions: {
        columns: tableColumns,
        reviewData: [
          {
            domain: 'example.com',
            protocol: 'HTTP',
            port: '80',
            polling_method: '轮询',
            url_count: '10',
            sync_status: '成功',
            is_default: false,
          },
          {
            domain: 'example.org',
            protocol: 'HTTPS',
            port: '443',
            polling_method: '加权轮询',
            url_count: '15',
            sync_status: '失败',
            is_default: false,
          },
          {
            domain: 'example.net',
            protocol: 'FTP',
            port: '21',
            polling_method: '源地址哈希',
            url_count: '5',
            sync_status: '部分成功',
            is_default: true,
          },
          {
            domain: 'example.edu',
            protocol: 'HTTP',
            port: '8080',
            polling_method: '最少连接',
            url_count: '8',
            sync_status: '绑定中',
            is_default: false,
          },
          {
            domain: 'example.biz',
            protocol: 'TCP',
            port: '22',
            polling_method: '随机',
            url_count: '12',
            sync_status: '成功',
            is_default: false,
          },
        ],
        extra: {
          settings: settings.value,
          'row-class': ({ sync_status }: { sync_status: string }) => {
            if (sync_status === '绑定中') {
              return 'binding-row';
            }
          },
        },
      },
      requestOption: {
        type: '',
      },
    });

    const isDomainSidesliderShow = ref(false);
    const handleSubmit = () => {};

    // 批量删除
    const isBatchDeleteDialogShow = ref(false);
    const radioGroupValue = ref(false);
    const tableProps = {
      columns: [
        {
          label: '监听器名称',
          field: 'listenerName',
        },
        {
          label: '监听器ID',
          field: 'listenerID',
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
          label: '均衡方式',
          field: 'balanceMode',
          filter: true,
        },
        {
          label: '是否绑定目标组',
          field: 'isBoundToTargetGroup',
          filter: true,
          render: ({ cell }: { cell: boolean }) => {
            return cell ? <Tag theme='success'>已绑定</Tag> : <Tag>未绑定</Tag>;
          },
        },
        {
          label: 'RS权重为O',
          field: 'rsWeight',
          sort: true,
          align: 'right',
        },
        {
          label: '',
          width: 80,
          align: 'right',
          render: () => <i class='hcm-icon bkhcm-icon-minus-circle-shape batch-delete-listener-icon'></i>,
        },
      ],
      data: [
        {
          listenerName: '监听器A',
          listenerID: 'ABC123',
          protocol: 'HTTP',
          port: 80,
          balanceMode: '轮询',
          isBoundToTargetGroup: true,
          rsWeight: 1,
        },
        {
          listenerName: '监听器B',
          listenerID: 'DEF456',
          protocol: 'HTTPS',
          port: 443,
          balanceMode: '最小连接数',
          isBoundToTargetGroup: false,
          rsWeight: 5,
        },
        {
          listenerName: '监听器C',
          listenerID: 'GHI789',
          protocol: 'TCP',
          port: 21,
          balanceMode: '源IP',
          isBoundToTargetGroup: true,
          rsWeight: 10,
        },
      ],
      searchData: [
        {
          name: '监听器名称',
          id: 'listenerName',
        },
        {
          name: '监听器ID',
          id: 'listenerID',
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
          id: 'balanceMode',
        },
        {
          name: '是否绑定目标组',
          id: 'isBoundToTargetGroup',
        },
        {
          name: 'RS权重为O',
          id: 'rsWeight',
        },
      ],
    };
    const handleBatchDelete = () => {
      // batch delete handler
    };
    return () => (
      <div class='domain-list-page has-selection'>
        <CommonTable>
          {{
            operation: () => (
              <>
                <Button theme='primary' onClick={() => (isDomainSidesliderShow.value = true)}>
                  <Plus class='f20' />
                  新增域名
                </Button>
                <Button onClick={() => (isBatchDeleteDialogShow.value = true)}>批量删除</Button>
              </>
            ),
          }}
        </CommonTable>
        <CommonSideslider
          title='新建域名'
          width={640}
          v-model:isShow={isDomainSidesliderShow.value}
          onHandleSubmit={handleSubmit}>
          <DomainSidesliderContent />
        </CommonSideslider>
        <BatchOperationDialog
          v-model:isShow={isBatchDeleteDialogShow.value}
          title='批量删除监听器'
          theme='danger'
          confirmText='删除'
          tableProps={tableProps}
          onHandleConfirm={handleBatchDelete}>
          {{
            tips: () => (
              <>
                已选择 <span class='blue'>97</span> 个监听器，其中 <span class='red'>22</span>{' '}
                个监听器RS的权重均不为0，在删除监听器前，请确认是否有流量转发，仔细核对后，再提交删除。
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
