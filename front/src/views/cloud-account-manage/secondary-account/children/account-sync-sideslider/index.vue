<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import dayjs from 'dayjs';
import { useSecondaryAccountStore, type ISecondaryAccountItem } from '@/store/cloud-account-manage/secondary-account';
import { VendorEnum } from '@/common/constant';
// import SecondaryAccountValue from '@/views/cloud-account-manage/components/secondary-account-value.vue';
import Status from '@/components/display-value/appearance/status.vue';

type SyncStatus = 'syncing' | 'success' | 'failed';

interface ISyncRow {
  id: string;
  account: ISecondaryAccountItem;
  status: SyncStatus;
  finishTime: string;
  error?: string;
}

const model = defineModel<boolean>();

const props = defineProps<{
  accounts: ISecondaryAccountItem[];
  bizId: number;
  vendor: VendorEnum;
}>();

const emit = defineEmits<{
  finished: [results: { success: string[]; failed: { id: string; error: any }[] }];
}>();

const STATUS_VALUE_MAP: Record<SyncStatus, string> = {
  syncing: 'running',
  success: 'success',
  failed: 'failed',
};
const STATUS_TEXT_MAP: Record<SyncStatus, string> = {
  syncing: '同步中',
  success: '成功',
  failed: '失败',
};

const secondaryAccountStore = useSecondaryAccountStore();

const rows = ref<ISyncRow[]>([]);

// 启动同步：并发请求每个账号，每个回调独立更新对应行状态
const startSync = async () => {
  if (!props.accounts?.length) return;

  rows.value = props.accounts.map((account) => ({
    id: account.id,
    account,
    status: 'syncing',
    finishTime: '--',
  }));

  const results: { success: string[]; failed: { id: string; error: any }[] } = {
    success: [],
    failed: [],
  };

  await Promise.all(
    rows.value.map(async (row) => {
      try {
        await secondaryAccountStore.syncAccountResource(props.bizId, props.vendor, row.id, 'sub_account');
        row.status = 'success';
        row.finishTime = dayjs().format('YYYY-MM-DD HH:mm:ss');
        results.success.push(row.id);
      } catch (error: any) {
        row.status = 'failed';
        row.finishTime = dayjs().format('YYYY-MM-DD HH:mm:ss');
        row.error = error?.message || '同步失败';
        results.failed.push({ id: row.id, error });
      }
    }),
  );

  emit('finished', results);
};

watch(
  () => model.value,
  (val) => {
    if (val) {
      startSync();
    }
  },
);

const handleClose = () => {
  model.value = false;
};

const hasUnfinished = computed(() => rows.value.some((r) => r.status === 'syncing'));
</script>

<template>
  <bk-sideslider v-model:is-show="model" class="account-sync-sideslider" title="同步二级账号" :width="960" quick-close>
    <template #default>
      <div class="sync-content">
        <bk-table
          :data="rows"
          :border="['outer', 'row']"
          row-key="id"
          show-overflow-tooltip
          :max-height="'calc(100vh - 160px)'"
        >
          <bk-table-column label="账号名称" min-width="280">
            <template #default="{ row }: { row: ISyncRow }">
              <!-- <SecondaryAccountValue :value="row.id" :biz-id="props.bizId" :vendor="props.vendor" res-type="sub_account" /> -->
              {{ row.account.name }}({{ row.account.extension.cloud_main_account_id }})
            </template>
          </bk-table-column>

          <bk-table-column label="同步状态" width="180">
            <template #default="{ row }: { row: ISyncRow }">
              <Status
                :value="STATUS_VALUE_MAP[row.status]"
                :display-value="STATUS_TEXT_MAP[row.status]"
                :title="row.status === 'failed' ? row.error : ''"
              />
            </template>
          </bk-table-column>

          <bk-table-column label="请求完成时间" width="220" prop="finishTime">
            <template #default="{ row }: { row: ISyncRow }">
              {{ row.finishTime }}
            </template>
          </bk-table-column>
        </bk-table>
      </div>
    </template>

    <template #footer>
      <div class="sideslider-footer">
        <bk-button theme="primary" @click="handleClose">
          {{ hasUnfinished ? '关闭（同步将在后台继续）' : '关闭' }}
        </bk-button>
      </div>
    </template>
  </bk-sideslider>
</template>

<style lang="scss" scoped>
.account-sync-sideslider {
  :deep(.bk-modal-content) {
    position: relative;
    padding: 24px;
  }

  .sync-content {
    background: #fff;
  }
}

.sideslider-footer {
  line-height: 48px;
  background: #fafbfd;
  border-top: 1px solid #eaebf0;
}
</style>
