import { useAccountStore } from '@/store';
import { defineComponent, ref } from 'vue';
import './index.scss';
import { timeFormatter } from '@/common/util';

enum ApprovalType {
  pass = 'pass',
  refuse = 'refuse',
}

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
        render: ({ cell }: any) => cell.slice(1, -1),
      },
      {
        field: 'current_status_display',
        label: '状态',
      },
      {
        field: 'create_at',
        label: '提单时间',
        render: ({ cell }: { cell: string }) => timeFormatter(cell),
      },
      {
        field: 'creator',
        label: '提单人',
      },
      {
        label: '操作',
        render: ({ data }: { data: any }) => (
          <>
            <bk-button text theme='primary' class='mr5' onClick={() => handleApprove(data)}>
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
    const isDialogShow = ref(false);
    const detail = ref<any>({});
    const approvalType = ref(ApprovalType.refuse);
    const memo = ref('');
    const pagination = ref({
      limit: 20,
      start: 0,
      current: 1,
    });
    const paginationCount = ref(0);

    const handleApprove = (data: any) => {
      isDialogShow.value = true;
      detail.value = data;
      approvalType.value = ApprovalType.pass;
    };

    const handleRefuse = (data: any) => {
      isDialogShow.value = true;
      detail.value = data;
      approvalType.value = ApprovalType.refuse;
    };

    const handleConfirm = async () => {
      await accountStore.approveTickets({
        sn: detail.value.sn,
        state_id: detail.value.current_steps[0].state_id,
        action: approvalType.value,
        memo: memo.value,
      });
      isDialogShow.value = false;
      memo.value = '';
      getList();
    };

    const handlePageValueChange = (val: number) => {
      pagination.value.current = val;
      pagination.value.start = pagination.value.limit * (val - 1);
      getList();
    };

    const handlePageLimitChange = (val: number) => {
      pagination.value.limit = val;
      pagination.value.current = 0;
      pagination.value.start = 0;
      getList();
    };

    const getList = async () => {
      isLoading.value = true;
      const { data } = await accountStore.getApprovalList({
        filter: filter.value,
        page: {
          limit: pagination.value.limit,
          start: pagination.value.start,
        },
      });
      const { count, details } = data;
      paginationCount.value = count;
      datas.value = details;
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
            pagination={{
              ...pagination.value,
              count: paginationCount.value,
            }}
            onPageValueChange={handlePageValueChange}
            onPageLimitChange={handlePageLimitChange}
            remote-pagination
          />
        </bk-loading>

        <bk-dialog
          isShow={isDialogShow.value}
          title={'审批'}
          theme={'primary'}
          quick-close
          width={560}
          onClosed={() => {
            isDialogShow.value = false;
            memo.value = '';
          }}
          onConfirm={handleConfirm}>
          <bk-form>
            <bk-form-item label={'审批意见'}>
              {approvalType.value === ApprovalType.pass ? (
                <bk-tag type='stroke' theme='success'>
                  通过
                </bk-tag>
              ) : (
                <bk-tag type='stroke' theme='danger'>
                  拒绝
                </bk-tag>
              )}
            </bk-form-item>
            <bk-form-item label={'备注'}>
              <bk-input type='textarea' v-model={memo.value}></bk-input>
            </bk-form-item>
          </bk-form>
        </bk-dialog>
      </div>
    );
  },
});
