<script setup lang="ts">
import { ref, watch } from 'vue';
import type { IAppliedAccountItem } from '../../typings';
import { PolicyApplyStatus } from '../../typings';
import { ENABLE_MOCK, MOCK_APPLIED_ACCOUNTS } from '../../constants';
import usePage from '@/hooks/use-page';

const props = defineProps<{
  policyId: string;
  appliedCount: number;
}>();

// 表格数据
const tableData = ref<IAppliedAccountItem[]>([]);
const isLoading = ref(false);
const { pagination } = usePage();

// 已选账号（用于批量更新）
const selectedAccounts = ref<IAppliedAccountItem[]>([]);
const tableRef = ref();

// 状态映射
const statusMap: Record<string, { text: string; dotClass: string }> = {
  [PolicyApplyStatus.APPLIED]: { text: '已应用', dotClass: 'status-applied' },
  [PolicyApplyStatus.PENDING]: { text: '待应用', dotClass: 'status-pending' },
  [PolicyApplyStatus.DATA_MISMATCH]: { text: '已应用(数据不一致)', dotClass: 'status-mismatch' },
};

const getStatusInfo = (status: PolicyApplyStatus) => statusMap[status] || { text: status, dotClass: '' };

// 加载已应用账号列表
const loadAppliedAccounts = async () => {
  isLoading.value = true;
  try {
    if (ENABLE_MOCK) {
      tableData.value = [...MOCK_APPLIED_ACCOUNTS];
      pagination.count = MOCK_APPLIED_ACCOUNTS.length;
      return;
    }
    // TODO: 替换为真实 API
    tableData.value = [];
    pagination.count = 0;
  } catch (error) {
    console.error('加载已应用账号列表失败:', error);
    tableData.value = [];
  } finally {
    isLoading.value = false;
  }
};

// 选择变化
const handleSelectionChange = ({ row, checked }: { row: IAppliedAccountItem; checked: boolean }) => {
  if (checked) {
    if (!selectedAccounts.value.find((item) => item.account_id === row.account_id)) {
      selectedAccounts.value.push(row);
    }
  } else {
    selectedAccounts.value = selectedAccounts.value.filter((item) => item.account_id !== row.account_id);
  }
};

const handleSelectAll = ({ checked }: { checked: boolean }) => {
  if (checked) {
    selectedAccounts.value = [...tableData.value];
  } else {
    selectedAccounts.value = [];
  }
};

// 分页处理
const handlePageChange = (page: number) => {
  pagination.current = page;
  // TODO: 真实分页时需要重新请求
};

const handlePageSizeChange = (limit: number) => {
  pagination.limit = limit;
  pagination.current = 1;
};

// 监听策略ID变化
watch(
  () => props.policyId,
  () => {
    if (props.policyId) {
      loadAppliedAccounts();
    }
  },
  { immediate: true },
);

// 暴露已选数据给父组件（defineExpose 必须是 <script setup> 的最后语句）
defineExpose({
  getSelectedAccounts: () => selectedAccounts.value,
});
</script>

<template>
  <div class="update-applied-table">
    <div class="section-label">选择要同步的账号</div>
    <bk-loading :loading="isLoading">
      <bk-table
        ref="tableRef"
        :data="tableData"
        :pagination="pagination"
        :max-height="400"
        :border="['outer', 'row']"
        row-key="account_id"
        show-overflow-tooltip
        @select="handleSelectionChange"
        @select-all="handleSelectAll"
        @page-value-change="handlePageChange"
        @page-limit-change="handlePageSizeChange"
      >
        <bk-table-column type="selection" align="center" />
        <bk-table-column label="二级账号" min-width="180">
          <template #default="{ row }">{{ row.account_id }} ({{ row.alias }})</template>
        </bk-table-column>
        <bk-table-column label="云上模版名称" prop="cloud_template_name" min-width="180" />
        <bk-table-column label="云模版同步时间" prop="cloud_sync_time" min-width="160" />
        <bk-table-column label="策略库应用版本" prop="applied_version" width="120" />
        <bk-table-column label="策略库应用状态" min-width="150">
          <template #default="{ row }">
            <span class="status-cell">
              <span :class="['status-dot', getStatusInfo(row.apply_status).dotClass]"></span>
              <span :class="{ 'text-danger': row.apply_status === PolicyApplyStatus.DATA_MISMATCH }">
                {{ getStatusInfo(row.apply_status).text }}
              </span>
            </span>
          </template>
        </bk-table-column>
        <bk-table-column label="策略库应用时间" prop="apply_time" min-width="160" />
        <bk-table-column label="操作" width="160" fixed="right">
          <template #default="{ row }">
            <bk-button theme="primary" text class="mr8">模版详情</bk-button>
            <bk-button
              v-if="
                row.apply_status === PolicyApplyStatus.PENDING || row.apply_status === PolicyApplyStatus.DATA_MISMATCH
              "
              theme="primary"
              text
            >
              策略对比
            </bk-button>
          </template>
        </bk-table-column>
      </bk-table>
    </bk-loading>
  </div>
</template>

<style lang="scss" scoped>
.update-applied-table {
  .section-label {
    font-size: 12px;
    color: #63656e;
    margin-bottom: 8px;
  }

  .status-cell {
    display: inline-flex;
    align-items: center;

    .status-dot {
      display: inline-block;
      width: 8px;
      height: 8px;
      border-radius: 50%;
      margin-right: 6px;
      flex-shrink: 0;

      &.status-applied {
        background-color: #2dcb56;
      }

      &.status-pending {
        background-color: #ff9c01;
      }

      &.status-mismatch {
        background-color: #ea3636;
      }
    }

    .text-danger {
      color: #ea3636;
    }
  }

  .mr8 {
    margin-right: 8px;
  }
}
</style>
