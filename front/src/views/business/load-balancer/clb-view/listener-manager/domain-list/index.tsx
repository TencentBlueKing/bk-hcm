import { defineComponent, ref } from 'vue';
import { Button } from 'bkui-vue';
import { BkRadioGroup, BkRadioButton } from 'bkui-vue/lib/radio';
import { Plus } from 'bkui-vue/lib/icon';
import { useTable } from '@/hooks/useTable/useTable';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import CommonSideslider from '../../../components/common-sideslider';
import DomainSidesliderContent from '../domain-sideslider-content';
import BatchOperationDialog from '@/components/batch-operation-dialog';
import './index.scss';

const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

export default defineComponent({
  name: 'DomainList',
  setup() {
    const { columns, settings } = useColumns('domain');
    const tableColumns = [
      ...columns,
      {
        label: '操作',
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
    const searchData: any = [];
    const searchUrl = `${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/vpcs/list`;
    const { CommonTable } = useTable({
      columns: tableColumns,
      settings: settings.value,
      searchUrl,
      searchData,
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
        },
        {
          label: '端口',
          field: 'port',
        },
        {
          label: '均衡方式',
          field: 'balanceMode',
        },
        {
          label: '是否绑定目标组',
          field: 'isBoundToTargetGroup',
        },
        {
          label: 'RS权重为O',
          field: 'rsWeight',
        },
        {
          label: '',
          width: 80,
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
          isBoundToTargetGroup: '是',
          rsWeight: 1,
        },
        {
          listenerName: '监听器B',
          listenerID: 'DEF456',
          protocol: 'HTTPS',
          port: 443,
          balanceMode: '最小连接数',
          isBoundToTargetGroup: '否',
          rsWeight: 5,
        },
        {
          listenerName: '监听器C',
          listenerID: 'GHI789',
          protocol: 'TCP',
          port: 21,
          balanceMode: '源IP',
          isBoundToTargetGroup: '是',
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
      <>
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
      </>
    );
  },
});
