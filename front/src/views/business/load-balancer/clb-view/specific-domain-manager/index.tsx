import { defineComponent, ref, watch } from 'vue';
// import components
import { Button, Form, Input, Select } from 'bkui-vue';
import { Done, EditLine, Error, Plus, Spinner } from 'bkui-vue/lib/icon';
import { BkRadioButton, BkRadioGroup } from 'bkui-vue/lib/radio';
import BatchOperationDialog from '@/components/batch-operation-dialog';
import CommonSideslider from '@/components/common-sideslider';
// use stores
import { useLoadBalancerStore } from '@/store/loadbalancer';
// import custom hooks
import { useTable } from '@/hooks/useTable/useTable';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import { useI18n } from 'vue-i18n';
// import constants
import { SYNC_STAUS_MAP } from '@/common/constant';
// import static resources
import StatusSuccess from '@/assets/image/success-account.png';
import StatusFailure from '@/assets/image/failed-account.png';
import StatusPartialSuccess from '@/assets/image/result-waiting.png';
import './index.scss';

const { FormItem } = Form;

export default defineComponent({
  setup() {
    // use hooks
    const { t } = useI18n();
    // use stores
    const loadBalancerStore = useLoadBalancerStore();

    const isBatchDeleteDialogShow = ref(false);
    const radioGroupValue = ref(false);
    const isDomainSidesliderShow = ref(false);
    const { columns, generateColumnsSettings } = useColumns('url');
    const editingID = ref('');
    const tableColumns = [
      ...columns,
      {
        label: t('目标组'),
        field: 'cloud_target_group_id',
        isDefaultShow: true,
        render: ({ cell }: any) => (
          <div class={'flex-row align-item-center target-group-name'}>
            {editingID.value === cell ? (
              <div class={'flex-row align-item-center'}>
                <Select />
                <Done width={20} height={20} class={'submit-edition-icon'} onClick={() => (editingID.value = '')} />
                <Error width={13} height={13} class={'submit-edition-icon'} onClick={() => (editingID.value = '')} />
              </div>
            ) : (
              <span class={'flex-row align-item-center'}>
                <span class={'target-group-name-btn'}>{cell}</span>
                <EditLine class={'target-group-edit-icon'} onClick={() => (editingID.value = cell)} />
              </span>
            )}
          </div>
        ),
      },
      {
        label: t('同步状态'),
        field: 'syncStatus',
        isDefaultShow: true,
        render: ({ cell }: any) => {
          let icon = StatusFailure;
          switch (cell) {
            case 'b': {
              icon = StatusSuccess;
              break;
            }
            case 'c': {
              icon = StatusFailure;
              break;
            }
            case 'd': {
              icon = StatusPartialSuccess;
              break;
            }
          }
          return (
            <div class={'url-status-container'}>
              {cell === 'a' ? (
                <Spinner fill='#3A84FF' width={13} height={13} class={'mr6'} />
              ) : (
                <img src={icon} class='mr6' width={13} height={13} />
              )}
              <span
                class={`${cell === 'd' ? 'url-sync-partial-success-status' : ''}`}
                v-bk-tooltips={{
                  content: '成功 89 个，失败 105 个',
                  disabled: cell !== 'd',
                }}>
                {SYNC_STAUS_MAP[cell]}
              </span>
            </div>
          );
        },
      },
      {
        label: t('操作'),
        field: 'actions',
        isDefaultShow: true,
        render: () => (
          <div class='operate-groups'>
            <Button text theme='primary'>
              {t('编辑')}
            </Button>
            <Button text theme='primary'>
              {t('删除')}
            </Button>
          </div>
        ),
      },
    ];
    const tableSettings = generateColumnsSettings(tableColumns);
    const tableProps = {
      columns: tableColumns,
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
    const { CommonTable, getListData } = useTable({
      searchOptions: {
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
      },
      tableOptions: {
        columns: tableColumns,
        extra: {
          settings: tableSettings.value,
          'row-class': ({ syncStatus }: { syncStatus: string }) => {
            if (syncStatus === 'a') {
              return 'binding-row';
            }
          },
        },
      },
      requestOption: {
        type: `vendors/tcloud/listeners/${loadBalancerStore.currentSelectedTreeNode.listener_id}/rules`,
      },
    });

    watch(
      () => loadBalancerStore.currentSelectedTreeNode,
      (val) => {
        const { listener_id, type } = val;
        if (type !== 'domain') return;
        // 只有 type='domain' 时, 才去请求对应 listener+domain 下的 url 列表
        getListData([], `vendors/tcloud/listeners/${listener_id}/rules`);
      },
    );

    const handleBatchDelete = () => {};
    const handleSubmit = () => {};

    return () => (
      <div class={'url-list-container has-selection has-breadcrumb'}>
        <CommonTable>
          {{
            operation: () => (
              <div class={'flex-row align-item-center'}>
                <Button theme={'primary'} onClick={() => (isDomainSidesliderShow.value = true)}>
                  <Plus class={'f20'} />
                  {t('新增 URL 路径')}
                </Button>
                <Button onClick={() => (isBatchDeleteDialogShow.value = true)}>{t('批量删除')}</Button>
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
                <BkRadioButton label={false}>{t('权重为0')}</BkRadioButton>
                <BkRadioButton label={true}>{t('权重不为0')}</BkRadioButton>
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
            <span class={'create-url-text-item-label'}>{t('监听器名称')}：</span>
            <span class={'create-url-text-item-value'}>web站点</span>
          </p>
          <p class={'create-url-text-item'}>
            <span class={'create-url-text-item-label'}>{t('协议端口')}：</span>
            <span class={'create-url-text-item-value'}>666666</span>
          </p>
          <p class={'create-url-text-item'}>
            <span class={'create-url-text-item-label'}>{t('域名')}：</span>
            <span class={'create-url-text-item-value'}>aaaaaaaaaa</span>
          </p>
          <Form formType='vertical'>
            <FormItem label={t('URL路径')}>
              <Input />
            </FormItem>
            <FormItem label={t('均衡方式')}>
              <Select />
            </FormItem>
            <FormItem label={t('目标组')}>
              <Select />
            </FormItem>
          </Form>
        </CommonSideslider>
      </div>
    );
  },
});
