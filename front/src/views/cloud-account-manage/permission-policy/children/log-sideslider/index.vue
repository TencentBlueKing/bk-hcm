<script setup lang="ts">
import { ref, watch, reactive } from 'vue';
import type { IAppliedAccountItem } from '../../typings';
import { PolicyApplyStatus } from '../../typings';
import { ENABLE_MOCK, MOCK_APPLIED_ACCOUNTS } from '../../constants';
import usePage from '@/hooks/use-page';
import LogDetaiSideslider from '@/views/cloud-account-manage/permission-policy/children/log-sideslider/log-detail.vue';

// 双向绑定控制显示状态
const model = defineModel<boolean>();

// const props = defineProps<{
//   // policyId: string;
//   // appliedCount: number;
// }>();

// 表格数据
const tableData = ref<IAppliedAccountItem[]>([]);
const isLoading = ref(false);
const { pagination } = usePage();

// 策略内容对比数据
const LogDetailInfo = reactive({
  show: false,
  id: '',
  accountId: '',
});

const tableRef = ref();

// 状态映射
const statusMap: Record<string, { text: string; dotClass: string }> = {
  [PolicyApplyStatus.APPLIED]: { text: '成功', dotClass: 'status-applied' },
  [PolicyApplyStatus.PENDING]: { text: '待应用', dotClass: 'status-pending' },
  [PolicyApplyStatus.DATA_MISMATCH]: { text: '失败', dotClass: 'status-mismatch' },
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

// 分页处理
const handlePageChange = (page: number) => {
  pagination.current = page;
  // TODO: 真实分页时需要重新请求
};
const handlePageSizeChange = (limit: number) => {
  pagination.limit = limit;
  pagination.current = 1;
};

// 取消
const handleCancel = () => {
  model.value = false;
};

// 查看详细日志
const handleLogDetail = () => {
  LogDetailInfo.show = true;
  // TODO: 真实数据需要调接口
};

// 监听策略ID变化
watch(
  () => model.value,
  (value: boolean) => {
    if (value) {
      loadAppliedAccounts();
    }
  },
  { immediate: true },
);
</script>

<template>
  <bk-sideslider
    v-model:is-show="model"
    class="log-list-sideslider"
    title="应用策略到二级账号"
    :width="1200"
    quick-close
    background-color="#f5f7fa"
  >
    <template #default>
      <bk-loading :loading="isLoading">
        <bk-table
          ref="tableRef"
          :data="tableData"
          :pagination="pagination"
          :max-height="580"
          :border="['outer', 'row']"
          row-key="account_id"
          show-overflow-tooltip
          @page-value-change="handlePageChange"
          @page-limit-change="handlePageSizeChange"
        >
          <bk-table-column label="二级账号" min-width="180">
            <template #default="{ row }">{{ row.account_id }} ({{ row.alias }})</template>
          </bk-table-column>
          <bk-table-column label="云模版同步时间" prop="cloud_sync_time" min-width="160" />
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
            <template #default>
              <bk-button theme="primary" text class="mr8" @click="handleLogDetail">日志</bk-button>
            </template>
          </bk-table-column>
        </bk-table>
      </bk-loading>

      <!--执行日志-->
      <LogDetaiSideslider v-bind="LogDetailInfo" @close="LogDetailInfo.show = false" />
    </template>

    <template #footer>
      <div class="sideslider-footer">
        <bk-button @click="handleCancel">关闭</bk-button>
      </div>
    </template>
  </bk-sideslider>
</template>

<style lang="scss" scoped>
.log-list-sideslider {
  :deep(.bk-modal-content) {
    position: relative;
  }

  :deep(.bk-modal-content) {
    padding: 24px;
  }

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

  .sideslider-footer {
    position: fixed;
    bottom: 0;
    right: 0;
    left: calc(100vw - 1200px);
    background: #f5f7fa;
    padding-left: 36px;
    line-height: 48px;
    border-top: 1px solid #eaebf0;
  }
}
</style>
