import { useAccountStore } from '@/store';
import { defineComponent, ref } from 'vue';

export default defineComponent({
  setup() {
    const tableColumns = [
      {
        field: 'sn',
        label: '单号',
      },
      {
        field: 'title',
        label: '标题',
      },
      {
        field: 'current_steps',
        label: '当前步骤',
        render: ({ cell }: { cell: any}) => cell.map(({ name }: { name: string }) => name).join(','),
      },
      {
        field: 'current_processors',
        label: '当前处理人',
        width: 400,
      },
      {
        field: 'current_status_display',
        label: '状态',
      },
      {
        field: 'create_at',
        label: '提单时间',
      },
      {
        field: 'creator',
        label: '提单人',
      },
      {
        label: '操作',
        render: ({ data }) => (
          <>
            <bk-button text theme='primary' class='mr5' onClick={() => handleApprove(data)}>通过</bk-button>
            <bk-button text theme='primary' onClick={() => handleRefuse(data)}>拒绝</bk-button>
          </>
        ),
      },
    ];

    const datas = ref([]);
    const accountStore = useAccountStore();

    const handleApprove = async (data: any) => {
      await accountStore.approveTickets({
        sn: data.sn,
        state_id: data.current_steps[0].state_id,
        action: 'pass',
        memo: '',
      });
      getList();
    };

    const handleRefuse = async (data: any) => {
      await accountStore.approveTickets({
        sn: data.sn,
        state_id: data.current_steps[0].state_id,
        action: 'refuse',
        memo: '',
      });
      getList();
    };

    const getList = async () => {
      const { data } = await accountStore.getApprovalList({
        filter: { op: 'and', rules: [] },
        page: {
          count: false,
          limit: 20,
          start: 0,
        },
      });
      datas.value = data;
    };

    getList();

    return () => (
      <div>
        <bk-table
          columns={tableColumns}
          data={datas.value}
        ></bk-table>
      </div>
    );
  },
});
