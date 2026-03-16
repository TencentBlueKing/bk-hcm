<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import { Plus, Transfer } from 'bkui-vue/lib/icon';
import type { IPermissionPolicyItem } from '../../typings';
import { ApplyOperationType } from '../../typings';
import ApplyNewTable from './apply-new-table.vue';
import UpdateAppliedTable from './update-applied-table.vue';

// 双向绑定控制显示状态
const model = defineModel<boolean>();

const props = defineProps<{
  policyData: IPermissionPolicyItem | null;
}>();

const emit = defineEmits<{
  success: [];
}>();

// 基本信息折叠状态
const baseInfoCollapsed = ref(true);

// 操作类型
const operationType = ref<ApplyOperationType>(ApplyOperationType.APPLY_NEW);

// 子表格引用
const applyNewTableRef = ref();
const updateAppliedTableRef = ref();

// 提交加载状态
const submitLoading = ref(false);

// 基本信息字段
const baseInfoFields = computed(() => {
  if (!props.policyData) return [];
  return [
    { label: '策略库名称', value: props.policyData.name },
    { label: '当前版本', value: 'v3' },
    { label: '创建人', value: `${props.policyData.creator}（平台）` },
    { label: '更新人', value: props.policyData.reviser },
    { label: '创建时间', value: props.policyData.created_at },
    { label: '更新时间', value: props.policyData.updated_at },
  ];
});

// 已应用账号数
const appliedCount = computed(() => props.policyData?.related_account_count || 0);

// 应用提交
const handleApply = async () => {
  submitLoading.value = true;
  try {
    // TODO: 替换为真实 API 调用
    if (operationType.value === ApplyOperationType.APPLY_NEW) {
      const _selected = applyNewTableRef.value?.getSelectedAccounts() || [];
      // TODO: 调用应用到新账号 API
    } else {
      const _selected = updateAppliedTableRef.value?.getSelectedAccounts() || [];
      // TODO: 调用更新已应用账号 API
    }
    model.value = false;
    emit('success');
  } catch (error) {
    // TODO: 使用全局消息提示替代 console
  } finally {
    submitLoading.value = false;
  }
};

// IOA 校验
const handleIOACheck = () => {
  // TODO: IOA 校验逻辑
};

// 取消
const handleCancel = () => {
  model.value = false;
};

// 重置状态
watch(
  () => model.value,
  (isShow) => {
    if (isShow) {
      operationType.value = ApplyOperationType.APPLY_NEW;
      baseInfoCollapsed.value = true;
    }
  },
);
</script>

<template>
  <bk-sideslider
    v-model:is-show="model"
    title="应用策略库到二级账号"
    :width="1200"
    quick-close
    background-color="#f5f7fa"
  >
    <template #default>
      <div class="apply-sideslider-container">
        <!-- 基本信息卡片 -->
        <bk-card
          title="基本信息"
          :is-collapse="true"
          v-model:collapse-status="baseInfoCollapsed"
          :border="false"
          :disable-header-style="true"
          class="info-card"
        >
          <div class="info-grid">
            <div v-for="(field, index) in baseInfoFields" :key="index" class="info-item">
              <span class="info-label">{{ field.label }}：</span>
              <span class="info-value">{{ field.value || '--' }}</span>
            </div>
          </div>
        </bk-card>

        <!-- 应用到二级账号卡片 -->
        <bk-card title="应用到二级账号" :border="false" :disable-header-style="true" class="apply-card">
          <!-- 操作类型切换 -->
          <div class="operation-type-label">操作类型</div>
          <div class="operation-type-cards">
            <div
              :class="['type-card', { active: operationType === ApplyOperationType.APPLY_NEW }]"
              @click="operationType = ApplyOperationType.APPLY_NEW"
            >
              <div class="card-icon apply-new-icon">
                <Plus />
              </div>
              <div class="card-content">
                <div class="card-title">应用到新账号</div>
                <div class="card-desc">首次应用此策略</div>
              </div>
            </div>
            <div
              :class="['type-card', { active: operationType === ApplyOperationType.UPDATE_APPLIED }]"
              @click="operationType = ApplyOperationType.UPDATE_APPLIED"
            >
              <div class="card-icon update-icon">
                <Transfer />
              </div>
              <div class="card-content">
                <div class="card-title">更新已应用账号</div>
                <div class="card-desc">已有 {{ appliedCount }} 个账号应用此策略</div>
              </div>
            </div>
          </div>

          <!-- 表格区域 -->
          <div class="table-area">
            <ApplyNewTable
              v-if="operationType === ApplyOperationType.APPLY_NEW"
              ref="applyNewTableRef"
              :policy-id="policyData?.id || ''"
            />
            <UpdateAppliedTable
              v-else
              ref="updateAppliedTableRef"
              :policy-id="policyData?.id || ''"
              :applied-count="appliedCount"
            />
          </div>
        </bk-card>
      </div>
    </template>

    <template #footer>
      <div class="sideslider-footer">
        <bk-button theme="primary" :loading="submitLoading" @click="handleApply">应用</bk-button>
        <bk-button @click="handleIOACheck">IOA校验</bk-button>
        <bk-button @click="handleCancel">取消</bk-button>
      </div>
    </template>
  </bk-sideslider>
</template>

<style lang="scss" scoped>
.apply-sideslider-container {
  padding: 24px;
  display: flex;
  flex-direction: column;
  gap: 16px;

  // 基本信息卡片
  .info-card {
    :deep(.bk-card-head) {
      padding: 12px 24px;
    }

    :deep(.bk-card-body) {
      padding: 0 24px 16px 40px;
    }
  }

  .info-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 12px 48px;

    .info-item {
      display: flex;
      align-items: flex-start;
      font-size: 12px;
      line-height: 20px;

      .info-label {
        min-width: 84px;
        flex-shrink: 0;
        color: #979ba5;
        text-align: right;
      }

      .info-value {
        color: #313238;
        word-break: break-all;
      }
    }
  }

  // 应用区域卡片
  .apply-card {
    :deep(.bk-card-head) {
      padding: 12px 24px;

      .bk-card-title {
        font-size: 14px;
        font-weight: 700;
        color: #313238;
      }
    }

    :deep(.bk-card-body) {
      padding: 0 24px 24px;
    }

    .operation-type-label {
      font-size: 12px;
      color: #63656e;
      margin-bottom: 8px;
    }

    // 操作类型卡片
    .operation-type-cards {
      display: flex;
      gap: 12px;
      margin-bottom: 16px;

      .type-card {
        display: flex;
        align-items: center;
        width: 220px;
        padding: 16px;
        border: 1px solid #dcdee5;
        border-radius: 2px;
        cursor: pointer;
        transition: all 0.2s;
        background: #fff;

        &:hover {
          border-color: #a3c5fd;
        }

        &.active {
          border-color: #3a84ff;
          background-color: #f0f5ff;
        }

        .card-icon {
          display: flex;
          align-items: center;
          justify-content: center;
          width: 32px;
          height: 32px;
          border-radius: 50%;
          margin-right: 12px;
          flex-shrink: 0;
          font-size: 16px;

          &.apply-new-icon {
            background-color: #e1ecff;
            color: #3a84ff;
          }

          &.update-icon {
            background-color: #e1ecff;
            color: #3a84ff;
          }
        }

        .card-content {
          .card-title {
            font-size: 14px;
            font-weight: 700;
            color: #313238;
            line-height: 22px;
          }

          .card-desc {
            font-size: 12px;
            color: #979ba5;
            line-height: 20px;
          }
        }
      }
    }

    .table-area {
      margin-top: 4px;
    }
  }
}

.sideslider-footer {
  display: flex;
  align-items: center;
  height: 48px;
  padding: 8px 24px;
  background-color: #fafbfd;
  border-top: 1px solid #eaebf0;
  gap: 8px;

  .bk-button {
    min-width: 88px;
  }
}
</style>
