<script setup lang="ts">
import { ref, computed, watch, inject, Ref } from 'vue';
import { Plus, Transfer } from 'bkui-vue/lib/icon';
import type { IAppliedReasonItem, IPermissionPolicyItem } from '../../typings';
import { IPermissionAppliedItem, usePermissionPolicyStore } from '@/store/cloud-account-manage/permission-policy';
import { ApplyOperationType } from '../../typings';
import ApplyNewTable from './apply-new-table.vue';
import UpdateAppliedTable from './update-applied-table.vue';
import { InfoBox } from 'bkui-vue';
import { VendorEnum } from '@/common/constant';
import { useWhereAmI } from '@/hooks/useWhereAmI';

// 双向绑定控制显示状态
const model = defineModel<boolean>();

const props = defineProps<{
  policyData: IPermissionPolicyItem | null;
}>();

const emit = defineEmits<{
  success: [row: IAppliedReasonItem[], type: ApplyOperationType];
}>();

const permissionPolicyStore = usePermissionPolicyStore();
const { isBusinessPage, getBizsId } = useWhereAmI();
const currentVendor = inject<Ref<VendorEnum>>('currentVendor', ref(VendorEnum.TCLOUD));

const appliedList = ref<IPermissionAppliedItem[]>([]);
const unAppliedList = ref<string[]>([]);

// 基本信息折叠状态
const baseInfoCollapsed = ref(true);

// 操作类型
const operationType = ref<ApplyOperationType>(ApplyOperationType.APPLY_NEW);

// 子表格引用
const applyNewTableRef = ref();
const updateAppliedTableRef = ref();
const appliedCount = ref(0);

// 提交加载状态
const submitLoading = ref(false);

// 基本信息字段
const baseInfoFields = computed(() => {
  if (!props.policyData) return [];
  return [
    { label: '策略库名称', value: props.policyData.name, type: 'string', id: 'name' },
    { label: '当前版本', value: `v${props.policyData.version}`, type: 'string', id: 'version' },
    { label: '创建人', value: `${props.policyData.creator}`, type: 'user', id: 'creator' },
    { label: '更新人', value: props.policyData.reviser, type: 'user', id: 'reviser' },
    { label: '创建时间', value: props.policyData.created_at, type: 'datetime', id: 'created_at' },
    { label: '更新时间', value: props.policyData.updated_at, type: 'datetime', id: 'updated_at' },
  ];
});
const policyData = computed(() => props.policyData);
const bizId = computed(() => (isBusinessPage ? getBizsId() : 0));

// 应用按钮是否可以点击
const applyBtnDisabled = computed(() => {
  return getSelected.value.length === 0;
});

const getSelected = computed(() => {
  if (operationType.value === ApplyOperationType.APPLY_NEW) return applyNewTableRef.value?.getSelectedAccounts() || [];
  return updateAppliedTableRef.value?.getSelectedAccounts() || [];
});

const applyMethod = computed(() => {
  if (isBusinessPage) {
    return operationType.value === ApplyOperationType.APPLY_NEW ? 'createAppliedAccountBiz' : 'updateAppliedAccountBiz';
  }
  return operationType.value === ApplyOperationType.APPLY_NEW ? 'createAppliedAccount' : 'updateAppliedAccount';
});
const applyParams = computed(() => {
  const params = { vendor: currentVendor.value, id: props.policyData.id };
  if (isBusinessPage) {
    return { bizId: bizId.value, ...params };
  }
  return params;
});

// 获取未应用和已经应用了的列表
const getList = async () => {
  const policyId = props.policyData.id;
  const accountSet = new Set();
  const [unAppliedRes, appliedRes] = await Promise.all([
    permissionPolicyStore.getUnappliedAccountIdsList(bizId.value, currentVendor.value, policyId),
    permissionPolicyStore.getAppliedAccountIdsList(bizId.value, currentVendor.value, policyId),
  ]);
  appliedList.value = [...appliedRes];
  unAppliedList.value = [...unAppliedRes];
  // 对已经应用的列表中account_id去重计算出总共有几个账号已经应用了
  appliedRes.forEach((item: IPermissionAppliedItem) => accountSet.add(item.account_id));
  appliedCount.value = accountSet.size;
  // 对已经应用的列表进行应用状态比较
  setAppliedListStatus();
};

const setAppliedListStatus = () => {
  appliedList.value.forEach((item: IPermissionAppliedItem) => {
    if (item.policy_library_version !== props.policyData.version) {
      // 如果版本号不一致，则表示还未应用
      item.apply_status = 'pending';
      return;
    }
    // 如果版本一直hash值不一致，则表示待应用
    if (item.policy_hash !== props.policyData.policy_hash) {
      item.apply_status = 'data_mismatch';
      return;
    }
    item.apply_status = 'applied';
  });
};

// 应用提交
const handleApply = async () => {
  submitLoading.value = true;
  const tips = InfoBox({
    type: 'loading',
    title: '策略应用正在提交中...',
    content: '应用过程中，请勿关闭本弹窗',
  });
  try {
    const key = operationType.value === ApplyOperationType.APPLY_NEW ? 'account_id' : 'id';
    const _selected = getSelected.value.map((item: { [key: string]: string }) => item[key]);
    const res: IAppliedReasonItem[] = [];
    const max = 100; // 每次接口selected最大数目

    while (_selected.length) {
      const list = await permissionPolicyStore[applyMethod.value]({
        ...applyParams.value,
        selectedIds: _selected.splice(0, max),
      });
      res.push(...list);
    }

    model.value = false;
    emit('success', res, operationType.value);
  } finally {
    tips.hide();
    submitLoading.value = false;
  }
};

// 取消
const handleCancel = () => {
  model.value = false;
};

watch(
  () => model.value,
  (isShow) => {
    if (isShow) {
      operationType.value = ApplyOperationType.APPLY_NEW;
      baseInfoCollapsed.value = true;
      // 获取未应用和已应用了的列表
      getList();
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
    render-directive="if"
  >
    <template #default>
      <div :class="['apply-sideslider-container', operationType === ApplyOperationType.APPLY_NEW ? 'apply' : 'update']">
        <!-- 基本信息卡片 -->
        <bk-card
          title="基本信息"
          :is-collapse="true"
          v-model:collapse-status="baseInfoCollapsed"
          :border="false"
          class="info-card"
        >
          <div class="info-grid">
            <div v-for="field in baseInfoFields" :key="field.id" class="info-item">
              <span class="info-label">{{ field.label }}：</span>
              <display-value class="info-value" :property="field" :value="field.value" />
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
              :list="unAppliedList"
              :biz-id="bizId"
              :vendor="currentVendor"
            />
            <UpdateAppliedTable
              v-else
              ref="updateAppliedTableRef"
              :list="appliedList"
              :policy-data="policyData"
              :biz-id="bizId"
              :vendor="currentVendor"
            />
          </div>
        </bk-card>
      </div>
    </template>

    <template #footer>
      <div class="sideslider-footer">
        <bk-button theme="primary" :loading="submitLoading" :disabled="applyBtnDisabled" @click="handleApply">
          应用
        </bk-button>
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

  &.apply {
    padding-bottom: 150px;
  }

  // 基本信息卡片
  .info-card {
    :deep(.bk-card-head) {
      border-bottom: 0;

      .title {
        padding-left: 5px;
      }
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
      border-bottom: 0;

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
  gap: 8px;
  position: fixed;
  bottom: 0;

  .bk-button {
    min-width: 88px;
  }
}
</style>
