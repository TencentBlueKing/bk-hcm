<script setup lang="ts">
import { computed, ref } from 'vue';
import { ApplyOperationType } from '../../typings';
import { type IApplyResultItem } from '@/store/cloud-account-manage/permission-policy';
import usePage from '@/hooks/use-page';
import SecondaryAccountValue from '@/views/cloud-account-manage/components/secondary-account-value.vue';
import { SecondaryAccountResourceTypeEnum } from '@/common/constant';

// 双向绑定控制显示状态
const model = defineModel<boolean>();

const props = defineProps<{
  data: IApplyResultItem[];
  type: ApplyOperationType;
  bizId: number;
}>();

const tableData = computed(() => [...props.data]);
const resType = computed(() => SecondaryAccountResourceTypeEnum.PERMISSION);

const isLoading = ref(false);
const { pagination } = usePage();

// 状态映射
const statusMap: Record<string, { text: string; dotClass: string }> = {
  success: { text: '成功', dotClass: 'status-applied' },
  failed: { text: '失败', dotClass: 'status-mismatch' },
};

const getStatusInfo = (status: 'success' | 'failed') => statusMap[status] || { text: status, dotClass: '' };

// 取消
const handleCancel = () => {
  model.value = false;
};
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
        >
          <bk-table-column :label="type === ApplyOperationType.APPLY_NEW ? '二级账号' : '权限模板ID'">
            <template #default="{ row }">
              <SecondaryAccountValue
                v-if="type === ApplyOperationType.APPLY_NEW"
                :value="row.account_id"
                :res-type="resType"
              />
              <span v-else>{{ row.permission_template_id }}</span>
            </template>
          </bk-table-column>
          <bk-table-column label="策略库应用状态">
            <template #default="{ row }">
              <span class="status-cell">
                <span :class="['status-dot', getStatusInfo(row.status).dotClass]"></span>
                <span>
                  {{ getStatusInfo(row.status).text }}
                </span>
              </span>
            </template>
          </bk-table-column>
          <bk-table-column label="失败原因" prop="reason" show-overflow-tooltip />
        </bk-table>
      </bk-loading>
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
