import { defineComponent, ref, watch } from 'vue';
// import components
import { Button, Form, Message, Tag } from 'bkui-vue';
import { BkRadioGroup, BkRadioButton } from 'bkui-vue/lib/radio';
import { Plus } from 'bkui-vue/lib/icon';
import CommonLocalTable from '@/components/CommonLocalTable';
import CommonSideslider from '@/components/common-sideslider';
import BatchOperationDialog from '@/components/batch-operation-dialog';
// import stores
import { useLoadBalancerStore } from '@/store/loadbalancer';
import { useBusinessStore } from '@/store';
// import custom hooks
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import useAddOrUpdateDomain, { OpAction } from './useAddOrUpdateDomain';
import { useI18n } from 'vue-i18n';
import './index.scss';
import Confirm from '@/components/confirm';

const { FormItem } = Form;

export default defineComponent({
  name: 'DomainList',
  setup() {
    // use hooks
    const { t } = useI18n();
    // use stores
    const loadBalancerStore = useLoadBalancerStore();
    const businessStore = useBusinessStore();

    const formInstance = ref();

    const isLoading = ref(false);
    // 搜索相关
    const searchData = [
      {
        name: t('域名'),
        id: 'domain',
      },
      {
        name: t('URL数量'),
        id: 'url_count',
      },
      {
        name: t('同步状态'),
        id: 'sync_status',
      },
    ];
    // 表格相关
    const { columns, settings } = useColumns('domain');
    const tableColumns = [
      {
        type: 'selection',
        width: 32,
        minWidth: 32,
        align: 'right',
      },
      {
        label: t('域名'),
        field: 'domain',
        isDefaultShow: true,
        render: ({ data, cell }: { data: any; cell: string }) => {
          return (
            <div class='set-default-operation-wrap'>
              <span class='cell-value'>{cell}</span>
              {data?.is_default ? (
                <Tag theme='info' class='default-tag'>
                  {t('默认')}
                </Tag>
              ) : (
                <Button text theme='primary' class='set-default-btn'>
                  {t('设为默认')}
                </Button>
              )}
            </div>
          );
        },
      },
      ...columns,
      {
        label: t('操作'),
        width: 120,
        render({ data }: any) {
          return (
            <div class='operate-groups'>
              <Button text theme='primary' onClick={() => handleDomainSidesliderShow(data)}>
                {t('编辑')}
              </Button>
              <Button
                text
                theme='primary'
                onClick={() => {
                  const listenerId = loadBalancerStore.currentSelectedTreeNode.id;
                  Confirm('请确定删除域名', `将删除域名【${data.name}】`, async () => {
                    await businessStore.deleteRules(listenerId, {
                      lbl_id: listenerId,
                      domain: data.domain,
                    });
                    Message({
                      message: '删除成功',
                      theme: 'success',
                    });
                    getDomainList(listenerId);
                  });
                }}>
                {t('删除')}
              </Button>
            </div>
          );
        },
      },
    ];
    const domainList = ref([]); // 域名列表

    // 获取域名列表
    const getDomainList = async (id: string) => {
      isLoading.value = true;
      try {
        const res = await businessStore.getDomainListByListenerId(id);
        domainList.value = res.data.domain_list;
      } finally {
        isLoading.value = false;
      }
    };

    watch(
      () => loadBalancerStore.currentSelectedTreeNode,
      (val) => {
        const { id, type, protocol } = val;
        if (type !== 'listener') return;
        // 只有 type='listener', 并且不为7层时, 才去请求对应 listener 下的 domain 列表
        if (['TCP', 'UDP'].includes(protocol)) return;
        getDomainList(id);
      },
      {
        immediate: true,
      },
    );

    // use custom hooks
    const {
      isShow: isDomainSidesliderShow,
      action,
      formItemOptions,
      handleShow: handleDomainSidesliderShow,
      handleSubmit: handleDomainSidesliderSubmit,
      formData: formModel,
    } = useAddOrUpdateDomain(() => getDomainList(loadBalancerStore.currentSelectedTreeNode.id));

    // 批量删除
    const isBatchDeleteDialogShow = ref(false);
    const radioGroupValue = ref(false);
    const tableProps = {
      columns: [
        {
          label: t('监听器名称'),
          field: 'listenerName',
        },
        {
          label: t('监听器ID'),
          field: 'listenerID',
        },
        {
          label: t('协议'),
          field: 'protocol',
          filter: true,
        },
        {
          label: t('端口'),
          field: 'port',
          filter: true,
        },
        {
          label: t('均衡方式'),
          field: 'balanceMode',
          filter: true,
        },
        {
          label: t('是否绑定目标组'),
          field: 'isBoundToTargetGroup',
          filter: true,
          render: ({ cell }: { cell: boolean }) => {
            return cell ? <Tag theme='success'>已绑定</Tag> : <Tag>未绑定</Tag>;
          },
        },
        {
          label: t('RS权重为O'),
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
          name: t('监听器名称'),
          id: 'listenerName',
        },
        {
          name: t('监听器ID'),
          id: 'listenerID',
        },
        {
          name: t('协议'),
          id: 'protocol',
        },
        {
          name: t('端口'),
          id: 'port',
        },
        {
          name: t('均衡方式'),
          id: 'balanceMode',
        },
        {
          name: t('是否绑定目标组'),
          id: 'isBoundToTargetGroup',
        },
        {
          name: t('RS权重为O'),
          id: 'rsWeight',
        },
      ],
    };
    const handleBatchDelete = () => {
      // batch delete handler
    };
    return () => (
      <div class='domain-list-page has-selection'>
        <CommonLocalTable
          loading={isLoading.value}
          tableOptions={{
            rowKey: 'domain',
            columns: tableColumns,
            extra: {
              settings: settings.value,
              'row-class': ({ sync_status }: { sync_status: string }) => {
                if (sync_status === '绑定中') {
                  return 'binding-row';
                }
              },
            },
          }}
          tableData={domainList.value}
          searchOptions={{
            searchData,
          }}>
          {{
            operation: () => (
              <>
                <Button theme='primary' onClick={() => handleDomainSidesliderShow()} class='mr12'>
                  <Plus class='f20' />
                  {t('新增域名')}
                </Button>
                <Button onClick={() => (isBatchDeleteDialogShow.value = true)}>{t('批量删除')}</Button>
              </>
            ),
          }}
        </CommonLocalTable>
        {/* domain 操作 dialog */}
        <CommonSideslider
          class='domain-sideslider'
          title={`${action.value === OpAction.ADD ? '新增' : '编辑'}域名`}
          width={640}
          v-model:isShow={isDomainSidesliderShow.value}
          onHandleSubmit={() => {
            handleDomainSidesliderSubmit(formInstance);
          }}>
          <p class='readonly-info'>
            <span class='label'>监听器名称</span>:
            <span class='value'>{loadBalancerStore.currentSelectedTreeNode.name}</span>
          </p>
          <p class='readonly-info'>
            <span class='label'>协议端口</span>:
            <span class='value'>
              {loadBalancerStore.currentSelectedTreeNode.protocol}:{loadBalancerStore.currentSelectedTreeNode.port}
            </span>
          </p>
          <Form formType='vertical' ref={formInstance} model={formModel}>
            {formItemOptions.value
              .filter(({ hidden }) => !hidden)
              .map(({ label, required, property, content }) => {
                return (
                  <FormItem label={label} required={required} key={property}>
                    {content()}
                  </FormItem>
                );
              })}
          </Form>
        </CommonSideslider>
        <BatchOperationDialog
          v-model:isShow={isBatchDeleteDialogShow.value}
          title='批量删除域名'
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
