import { RESOURCES_SYNC_STATUS_MAP, RESOURCE_TYPES_MAP } from '@/common/constant';
import http from '@/http';
import { useResourceAccountStore } from '@/store/useResourceAccountStore';
import { Switcher, Table } from 'bkui-vue';
import { defineComponent, ref, watch } from 'vue';
import successStatus from '@/assets/image/success-account.png';
import failedStatus from '@/assets/image/failed-account.png';
import loadingStatus from '@/assets/image/status_loading.png';
const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

export default defineComponent({
  setup() {
    const resourceAccountStore = useResourceAccountStore();
    const statusList = ref([]);
    const tableColumns = [
      {
        label: '资源名称',
        field: 'res_name',
        render: ({ cell }: { cell: string }) => RESOURCE_TYPES_MAP[cell],
      },
      {
        label: '任务状态',
        field: 'res_status',
        render: ({ cell }: { cell: string }) => (<div>
          <img
            // eslint-disable-next-line no-nested-ternary
            src={ cell === 'sync_success' ? successStatus : cell === 'sync_failed' ? failedStatus : loadingStatus }
            class={'resource-status-icon'}
            height={8}
            width={8}
          />
          <span>
            {RESOURCES_SYNC_STATUS_MAP[cell]}
          </span>
        </div>),
      },
      {
        label: '最近同步时间',
        field: 'res_end_time',
      },
      {
        label: '是否接入',
        field: 'is_implement',
        render: () => (
          <div>
            <Switcher/>
            同步周期: 20 分钟
          </div>
        ),
        rowspan: 7,
      },
    ];
    watch(
      () => resourceAccountStore.resourceAccount,
      async (account) => {
        const res = await http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/accounts/sync_details/${account.id}`);
        statusList.value = res.data.iass_res;
      },
      {
        immediate: true,
        deep: true,
      },
    );
    return () => (
      <>
        <Table
          data={statusList.value}
          columns={tableColumns}
          border={['row', 'col', 'outer']}
        ></Table>
      </>
    );
  },
});
