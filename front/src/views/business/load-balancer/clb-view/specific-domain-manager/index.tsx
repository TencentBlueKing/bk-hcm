import { useTable } from '@/hooks/useTable/useTable';
import { defineComponent } from 'vue';
import './index.scss';

export default defineComponent({
  setup() {
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
    return () => <CommonTable />;
  },
});
