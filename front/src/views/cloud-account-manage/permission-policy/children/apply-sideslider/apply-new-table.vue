<script setup lang="ts">
import { ref, watch } from 'vue';
import type { ISelectableAccount } from '../../typings';
import { ENABLE_MOCK, MOCK_SELECTABLE_ACCOUNTS } from '../../constants';

const props = defineProps<{
  policyId: string;
}>();

// 已选账号列表
const selectedAccounts = ref<ISelectableAccount[]>([]);

// 表格数据
const tableData = ref<ISelectableAccount[]>([]);
const isLoading = ref(false);

// 表格引用
const tableRef = ref();

// 加载可选账号列表
const loadAccounts = async () => {
  isLoading.value = true;
  try {
    if (ENABLE_MOCK) {
      tableData.value = [...MOCK_SELECTABLE_ACCOUNTS];
      return;
    }
    // TODO: 替换为真实 API
    tableData.value = [];
  } catch (error) {
    console.error('加载可选账号列表失败:', error);
    tableData.value = [];
  } finally {
    isLoading.value = false;
  }
};

// 选择变化
const handleSelectionChange = ({ row, checked }: { row: ISelectableAccount; checked: boolean }) => {
  if (checked) {
    // 判断是否已存在
    if (!selectedAccounts.value.find((item) => item.account_id === row.account_id)) {
      selectedAccounts.value.push(row);
    }
  } else {
    selectedAccounts.value = selectedAccounts.value.filter((item) => item.account_id !== row.account_id);
  }
};

// 全选/取消全选
const handleSelectAll = ({ checked }: { checked: boolean }) => {
  if (checked) {
    selectedAccounts.value = [...tableData.value];
  } else {
    selectedAccounts.value = [];
  }
};

// 移除已选账号
const handleRemoveSelected = (accountId: string) => {
  selectedAccounts.value = selectedAccounts.value.filter((item) => item.account_id !== accountId);
  // 同步取消表格中的选中状态 - 通过 clearSelection 后重新选中来实现
  // bk-table 没有直接取消某行选中的 API，这里通过 ref 操作
};

// 清空已选
const handleClearAll = () => {
  selectedAccounts.value = [];
  tableRef.value?.clearSelection?.();
};

// 监听策略ID变化，重新加载
watch(
  () => props.policyId,
  () => {
    if (props.policyId) {
      loadAccounts();
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
  <div class="apply-new-table">
    <div class="section-label">选择要应用的二级账号</div>
    <bk-loading :loading="isLoading">
      <bk-table
        ref="tableRef"
        :data="tableData"
        :max-height="400"
        :border="['outer', 'row']"
        row-key="account_id"
        @select="handleSelectionChange"
        @select-all="handleSelectAll"
      >
        <bk-table-column type="selection" align="center" />
        <bk-table-column label="二级账号" min-width="1000">
          <template #default="{ row }">{{ row.account_id }} ({{ row.alias }})</template>
        </bk-table-column>
      </bk-table>
    </bk-loading>

    <!-- 底部已选标签区 -->
    <div class="selected-footer">
      <div class="selected-header">
        <div class="selected-left">
          <span class="selected-label">已选账号</span>
          <span class="selected-count">{{ selectedAccounts.length }}</span>
        </div>
        <div v-if="selectedAccounts.length > 0" class="selected-right">
          <span class="clear-btn" @click="handleClearAll">
            <i class="hcm-icon bkhcm-icon-delete mr2"></i>
            清空
          </span>
        </div>
      </div>
      <div v-if="selectedAccounts.length > 0" class="selected-tags">
        <bk-tag
          v-for="account in selectedAccounts"
          :key="account.account_id"
          closable
          @close="handleRemoveSelected(account.account_id)"
        >
          {{ account.account_id }}
        </bk-tag>
      </div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.apply-new-table {
  .section-label {
    font-size: 12px;
    color: #63656e;
    margin-bottom: 8px;
  }

  .selected-footer {
    margin-top: 16px;
    border-top: 1px solid #dcdee5;
    padding-top: 12px;

    .selected-header {
      display: flex;
      align-items: center;
      justify-content: space-between;
      margin-bottom: 8px;

      .selected-left {
        display: flex;
        align-items: center;

        .selected-label {
          font-size: 12px;
          font-weight: 700;
          color: #313238;
          margin-right: 8px;
        }

        .selected-count {
          font-size: 12px;
          color: #3a84ff;
          font-weight: 700;
        }
      }

      .selected-right {
        .clear-btn {
          display: flex;
          align-items: center;
          font-size: 12px;
          color: #979ba5;
          cursor: pointer;

          &:hover {
            color: #3a84ff;
          }

          i {
            margin-right: 4px;
            font-size: 14px;
          }
        }
      }
    }

    .selected-tags {
      display: flex;
      flex-wrap: wrap;
      gap: 6px;
    }
  }
}
</style>
