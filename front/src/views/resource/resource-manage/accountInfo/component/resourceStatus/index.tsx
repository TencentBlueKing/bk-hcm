import { RESOURCES_SYNC_STATUS_MAP, RESOURCE_TYPES_MAP } from '@/common/constant';
import http from '@/http';
import { useResourceAccountStore } from '@/store/useResourceAccountStore';
import { Loading, Table } from 'bkui-vue';
import { defineComponent, ref, watch, onBeforeUnmount, reactive } from 'vue';
import successStatus from '@/assets/image/success-account.png';
import failedStatus from '@/assets/image/failed-account.png';
import loadingStatus from '@/assets/image/status_loading.png';
import './index.scss';
import { timeFormatter } from '@/common/util';
import interval from '@/utils/interval';
const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

export default defineComponent({
  setup() {
    const resourceAccountStore = useResourceAccountStore();
    const statusList = ref([]);
    const isLoading = ref(false);
    const timeInterval = reactive({
      set: null,
      clear: null,
    });

    const tableColumns = [
      {
        label: '资源名称',
        field: 'res_name',
        render: ({ cell }: { cell: string }) => RESOURCE_TYPES_MAP[cell],
      },
      {
        label: '任务状态',
        field: 'res_status',
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
        render: ({ cell }: { cell: string }) => timeFormatter(cell),
      },
      {
        label: '同步周期',
        field: 'is_implement',
        render: () => (
          <div>
            {/* <Switcher disabled class={'mr8'}/> */}
            {/* 同步周期: 20 分钟 */}
            20 分钟
          </div>
        ),
        // rowspan: 7,
      },
    ];
    const init = () => {
      timeInterval.clear();
      timeInterval.set();
    };
    const getList = async (account: any) => {
      if (!account?.id) return;
      isLoading.value = true;
      try {
        const res = await http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/accounts/sync_details/${account.id}`);
        statusList.value = res.data.iass_res;
      } finally {
        isLoading.value = false;
      }
    };
    onBeforeUnmount(() => {
      timeInterval?.clear();
    });
    watch(
      () => resourceAccountStore.resourceAccount,
      async (account) => {
        getList(account);
        if (!timeInterval.set) {
          const { clearTimeInterval, setTimeInterval } = interval(() => getList(account), 10000, 600000);
          timeInterval.set = setTimeInterval;
          timeInterval.clear = clearTimeInterval;
        }
        init();
      },
      {
        immediate: true,
        deep: true,
      },
    );
    return () => (
      <Loading loading={isLoading.value} style={{ margin: '8px 0' }}>
        <Table data={statusList.value} columns={tableColumns} border={['row', 'outer']}></Table>
      </Loading>
    );
  },
});
