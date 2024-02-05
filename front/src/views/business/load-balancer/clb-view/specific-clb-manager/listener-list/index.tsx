import { defineComponent } from 'vue';
import './index.scss';
import { useTable } from '@/hooks/useTable/useTable';
import { Button } from 'bkui-vue';
import { Plus } from 'bkui-vue/lib/icon';

export default defineComponent({
  setup() {
    const { CommonTable } = useTable({
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
          {
            label: '监听器名称',
            field: 'listenerName',
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
            label: '域名数量',
            field: 'domainCount',
          },
          {
            label: 'URL数量',
            field: 'urlCount',
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
        reviewData: [
          {
            listenerName: 'Listener001',
            protocol: 'HTTP',
            port: 80,
            balanceMode: 'RoundRobin',
            domainCount: 5,
            urlCount: 10,
            syncStatus: 'Synchronized',
            actions: 'Edit',
          },
          {
            listenerName: 'Listener002',
            protocol: 'HTTPS',
            port: 443,
            balanceMode: 'LeastConnections',
            domainCount: 3,
            urlCount: 5,
            syncStatus: 'Pending',
            actions: 'Delete',
          },
          {
            listenerName: 'Listener003',
            protocol: 'TCP',
            port: 22,
            balanceMode: 'IPHash',
            domainCount: 2,
            urlCount: 7,
            syncStatus: 'Failed',
            actions: 'Update',
          },
        ],
        extra: {
          settings: {
            fields: [],
            checked: [],
            limit: 0,
            size: '',
            sizeList: [],
            showLineHeight: false,
          },
        },
      },
      requestOption: {
        type: '',
      },
    });
    return () => (
      <div>
        <CommonTable>
          {{
            operation: () => (
              <div class={'flex-row align-item-center'}>
                <Button theme={'primary'}>
                  <Plus class={'f20'} />
                  新增监听器
                </Button>
                <Button>批量删除</Button>
              </div>
            ),
          }}
        </CommonTable>
      </div>
    );
  },
});
