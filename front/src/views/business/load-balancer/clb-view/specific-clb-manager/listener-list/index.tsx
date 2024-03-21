import { defineComponent, watch } from 'vue';
// import components
import { Button } from 'bkui-vue';
import { Plus } from 'bkui-vue/lib/icon';
// import stores
import { useLoadBalancerStore } from '@/store/loadbalancer';
// import custom hooks
import { useTable } from '@/hooks/useTable/useTable';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import { useI18n } from 'vue-i18n';
import './index.scss';

export default defineComponent({
  setup() {
    // use hooks
    const { t } = useI18n();
    // use stores
    const loadBalancerStore = useLoadBalancerStore();

    const { columns, settings } = useColumns('listener');
    const { CommonTable, getListData } = useTable({
      searchOptions: {
        searchData: [
          {
            name: '监听器名称',
            id: 'listenerName',
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
            name: '域名数量',
            id: 'domainCount',
          },
          {
            name: 'URL数量',
            id: 'urlCount',
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
        columns: [
          ...columns,
          {
            label: t('操作'),
            field: 'actions',
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
        ],
        extra: {
          settings: settings.value,
        },
      },
      requestOption: {
        type: `load_balancers/${loadBalancerStore.currentSelectedTreeNode.id}/listeners`,
      },
    });

    watch(
      () => loadBalancerStore.currentSelectedTreeNode,
      (val) => {
        const { id, type } = val;
        if (type !== 'lb') return;
        // 只有当 type='lb' 时, 才去请求对应 lb 下的 listener 列表
        getListData([], `load_balancers/${id}/listeners`);
      },
    );

    return () => (
      <div>
        <CommonTable>
          {{
            operation: () => (
              <div class={'flex-row align-item-center'}>
                <Button theme={'primary'}>
                  <Plus class={'f20'} />
                  {t('新增监听器')}
                </Button>
                <Button>{t('批量删除')}</Button>
              </div>
            ),
          }}
        </CommonTable>
      </div>
    );
  },
});
