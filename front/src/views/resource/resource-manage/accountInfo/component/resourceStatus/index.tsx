import { RESOURCES_SYNC_STATUS_MAP, RESOURCE_TYPES_MAP } from '@/common/constant';
import http from '@/http';
import { useResourceAccountStore } from '@/store/useResourceAccountStore';
import { Loading, Table } from 'bkui-vue';
import { defineComponent, ref, watch } from 'vue';
import successStatus from '@/assets/image/success-account.png';
import failedStatus from '@/assets/image/failed-account.png';
import loadingStatus from '@/assets/image/status_loading.png';
import './index.scss';
import { timeFormatter } from '@/common/util';
import useTimeoutPoll from '@/hooks/use-timeout-poll';
const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

export default defineComponent({
  setup() {
    const resourceAccountStore = useResourceAccountStore();
    const statusList = ref([]);
    const isLoading = ref(false);

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
    // 账号只能在调用时现取：轮询回调只会创建一次，参数传入的账号会被闭包固定成首次的那个
    const getList = async () => {
      const accountId = resourceAccountStore.resourceAccount?.id;
      if (!accountId) return;
      isLoading.value = true;
      try {
        const res = await http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/accounts/sync_details/${accountId}`);
        statusList.value = res.data.iass_res;
      } finally {
        isLoading.value = false;
      }
    };
    // 10s 一轮、最多 60 轮（沿用原先 10 分钟的轮询上限）
    const { reset, resume } = useTimeoutPoll(getList, 10000, { max: 60 });
    watch(
      () => resourceAccountStore.resourceAccount,
      () => {
        getList();
        // 换账号后重置轮次，不把上一个账号已消耗的轮询预算算进来
        reset();
        resume();
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
