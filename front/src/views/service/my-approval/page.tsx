import { useAccountStore } from '@/store';
import { defineComponent, ref } from 'vue';
import './index.scss';

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
        render: ({ cell }: { cell: any }) => cell.map(({ name }: { name: string }) => name).join(','),
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
        render: ({ data }: { data: any }) => (
          <>
            <bk-button
              text
              theme='primary'
              class='mr5'
              onClick={() => handleApprove(data)}>
              通过
            </bk-button>
            <bk-button text theme='primary' onClick={() => handleRefuse(data)}>
              拒绝
            </bk-button>
          </>
        ),
      },
    ];

    const datas = ref([]);
    const filter = ref({ op: 'and', rules: [] });
    const accountStore = useAccountStore();
    const isLoading = ref(false);
    const pagination = ref({
      count: false,
      limit: 20,
      start: 0,
    });

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
      isLoading.value = true;
      const { data } = await accountStore.getApprovalList({
        filter: filter.value,
        page: pagination.value,
      });
      datas.value = data;
      isLoading.value = false;
    };

    getList();

    return () => (
      <div>
        <div class={'my-approval-selector-container'}>
          {/* <bk-search-select
            class="bg-white w280 my-approval-selector"
            placeholder={'通过单号、标题搜索'}
            data={searchData}
            v-model={searchValue.value}
          /> */}
        </div>
        <bk-loading loading={isLoading.value}>
          <bk-table
            columns={tableColumns}
            data={datas.value}
          />
        </bk-loading>
      </div>
    );
  },
});
