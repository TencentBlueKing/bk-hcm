import { defineComponent, ref, watch } from 'vue';
// import components
import { Button, Message, Tag } from 'bkui-vue';
import { Plus, Spinner } from 'bkui-vue/lib/icon';
import CommonLocalTable from '@/components/CommonLocalTable';
import BatchOperationDialog from '@/components/batch-operation-dialog';
import Confirm from '@/components/confirm';
// import stores
import { useBusinessStore, useResourceStore, useLoadBalancerStore } from '@/store';
// import custom hooks
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import useSelection from '@/views/resource/resource-manage/hooks/use-selection';
import { useI18n } from 'vue-i18n';
import AddOrUpdateDomainSideslider from '../../components/AddOrUpdateDomainSideslider';
// import utils
import bus from '@/common/bus';
// import constants
import { APPLICATION_LAYER_LIST } from '@/constants';
import './index.scss';

export default defineComponent({
  name: 'DomainList',
  props: { id: String, type: String, protocol: String },
  setup(props) {
    // use hooks
    const { t } = useI18n();
    const isBatchDeleteLoading = ref(false);
    // use stores
    const loadBalancerStore = useLoadBalancerStore();
    const businessStore = useBusinessStore();
    const resourceStore = useResourceStore();
    const defaultDomain = ref('');
    const isCheckDomainLoading = ref(false);
    const { selections, handleSelectionChange } = useSelection();

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
              {data.domain === defaultDomain.value ? (
                <Tag theme='info' class='default-tag'>
                  {t('默认')}
                </Tag>
              ) : (
                <Button
                  text
                  theme='primary'
                  class='set-default-btn'
                  onClick={async () => {
                    isCheckDomainLoading.value = true;
                    try {
                      await businessStore.updateDomains(loadBalancerStore.currentSelectedTreeNode.id, {
                        ...data,
                        default_server: true,
                      });
                      Message({
                        message: '设置成功',
                        theme: 'success',
                      });
                      defaultDomain.value = data.domain;
                    } finally {
                      isCheckDomainLoading.value = false;
                    }
                  }}>
                  {isCheckDomainLoading.value ? (
                    <span>
                      <Spinner fill='#3A84FF' width={16} height={16} />
                      &nbsp;设置中...
                    </span>
                  ) : (
                    t('设为默认')
                  )}
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
              <Button text theme='primary' onClick={() => bus.$emit('showAddDomainSideslider', data)}>
                {t('编辑')}
              </Button>
              <Button
                text
                theme='primary'
                onClick={() => {
                  const listenerId = loadBalancerStore.currentSelectedTreeNode.id;
                  Confirm('请确定删除域名', `将删除域名【${data.domain}】`, async () => {
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
        defaultDomain.value = res.data.default_domain;
        domainList.value = res.data.domain_list;
      } finally {
        isLoading.value = false;
      }
    };

    // 获取监听器详情
    const getListenerDetail = async (id: string) => {
      const { data } = await resourceStore.detail('listeners', id);
      loadBalancerStore.setCurrentSelectedTreeNode(data);
    };

    watch(
      [() => props.id, () => props.type],
      ([id, type]) => {
        // 当id或type变更时, 重新请求数据
        const { protocol } = props;
        if (id && type === 'list') {
          // 刷新或第一次访问页面时, 请求监听器详情
          !loadBalancerStore.currentSelectedTreeNode?.id && getListenerDetail(id);
          APPLICATION_LAYER_LIST.includes(protocol) && getDomainList(id);
        }
      },
      { immediate: true },
    );

    // 批量删除
    const isBatchDeleteDialogShow = ref(false);
    const tableProps = {
      columns: [
        {
          label: t('域名'),
          field: 'domain',
        },
        ...columns,
      ],
      data: selections.value,
      searchData,
    };
    const handleBatchDelete = async () => {
      isBatchDeleteLoading.value = true;
      try {
        await businessStore.batchDeleteDomains({
          lbl_id: loadBalancerStore.currentSelectedTreeNode.id,
          domains: selections.value.map(({ domain }) => domain),
        });
        Message({
          message: '删除成功',
          theme: 'success',
        });
        isBatchDeleteDialogShow.value = false;
        getDomainList(loadBalancerStore.currentSelectedTreeNode.id);
      } finally {
        isBatchDeleteLoading.value = false;
      }
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
              onSelectionChange: (selections: any) => handleSelectionChange(selections, () => true),
              onSelectAll: (selections: any) => handleSelectionChange(selections, () => true, true),
            },
          }}
          tableData={domainList.value}
          searchOptions={{
            searchData,
          }}>
          {{
            operation: () => (
              <>
                <Button theme='primary' onClick={() => bus.$emit('showAddDomainSideslider')} class='mr12'>
                  <Plus class='f20' />
                  {t('新增域名')}
                </Button>
                <Button onClick={() => (isBatchDeleteDialogShow.value = true)} disabled={!selections.value.length}>
                  {t('批量删除')}
                </Button>
              </>
            ),
          }}
        </CommonLocalTable>
        {/* domain 操作 dialog */}
        <AddOrUpdateDomainSideslider
          originPage='listener'
          getListData={() => getDomainList(loadBalancerStore.currentSelectedTreeNode.id)}
        />
        <BatchOperationDialog
          isSubmitLoading={isBatchDeleteLoading.value}
          v-model:isShow={isBatchDeleteDialogShow.value}
          title='批量删除域名'
          theme='danger'
          confirmText='删除'
          tableProps={tableProps}
          onHandleConfirm={handleBatchDelete}></BatchOperationDialog>
      </div>
    );
  },
});
