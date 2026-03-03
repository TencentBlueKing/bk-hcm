import { RESOURCES_SYNC_STATUS_MAP, RESOURCE_TYPES_MAP } from '@/common/constant';
import http from '@/http';
import { Loading, Table } from 'bkui-vue';
import { defineComponent, ref, watch } from 'vue';
import successStatus from '@/assets/image/success-account.png';
import failedStatus from '@/assets/image/failed-account.png';
import loadingStatus from '@/assets/image/status_loading.png';
import './resource-status.scss';
import { timeFormatter } from '@/common/util';
import useTimeoutPoll from '@/hooks/use-timeout-poll';
const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

export default defineComponent({
  props: {
    accountId: {
      type: String,
      required: true,
    },
  },
  setup(props) {
    const statusList = ref([]);
    const isLoading = ref(false);

    const tableColumns = [
      {
        label: '资源名称',
        field: 'res_name',
        width: 150,
        render: ({ cell }: { cell: string }) => RESOURCE_TYPES_MAP[cell],
      },
      {
        label: '任务状态',
        field: 'res_status',
        width: 150,
        render: ({ cell }: { cell: string }) => (
          <div class={'resource-status'}>
            <img
              // eslint-disable-next-line no-nested-ternary
              src={cell === 'sync_success' ? successStatus : cell === 'sync_failed' ? failedStatus : loadingStatus}
              class={`resource-status-icon ${cell === 'syncing' && 'loading'}`}
              height={16}
              width={16}
            />
            <span>{RESOURCES_SYNC_STATUS_MAP[cell]}</span>
          </div>
        ),
      },
      {
        label: '最近同步时间',
        field: 'res_end_time',
        width: 150,
        render: ({ cell }: { cell: string }) => timeFormatter(cell),
      },
      {
        label: '同步周期',
        field: 'is_implement',
        width: 150,
        render: () => <div>20 分钟</div>,
      },
    ];
    const getList = async () => {
      const id = props.accountId;
      if (!id) return;
      isLoading.value = true;
      try {
        const res = await http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/accounts/sync_details/${id}`);
        statusList.value = res.data.iass_res;
      } finally {
        isLoading.value = false;
      }
    };

    const { resume, reset } = useTimeoutPoll(getList, 10000, { immediate: false, max: 60 });

    watch(
      () => props.accountId,
      (id) => {
        if (!id) return;
        reset();
        getList();
        resume();
      },
      {
        immediate: true,
      },
    );
    return () => (
      <Loading loading={isLoading.value}>
        <Table data={statusList.value} columns={tableColumns}></Table>
      </Loading>
    );
  },
});
