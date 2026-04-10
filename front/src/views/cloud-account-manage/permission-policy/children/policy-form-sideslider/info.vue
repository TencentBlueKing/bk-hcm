<script setup lang="ts">
import { computed } from 'vue';
import type { IPermissionPolicyItem } from '../../typings';
import JSON from '@/views/cloud-account-manage/components/json.vue';

// 双向绑定控制显示状态
const model = defineModel<boolean>();

const props = defineProps<{
  policyData: IPermissionPolicyItem | null;
}>();

const emit = defineEmits<{
  success: [];
  'apply-to-account': [row: IPermissionPolicyItem];
}>();

// 基本信息字段
const baseInfoFields = computed(() => {
  if (!props.policyData) return [];
  return [
    { label: '权限策略库名称', value: props.policyData.name, id: 'name' },
    { label: '关联二级账号数', value: props.policyData.associated_account_count, id: 'associated_account_count ' },
    { label: '创建人', value: `${props.policyData.creator}（平台）`, id: 'creator' },
    { label: '创建时间', value: props.policyData.created_at, id: 'created_at' },
    { label: '更新人', value: props.policyData.reviser, id: 'reviser' },
    { label: '更新时间', value: props.policyData.updated_at, id: 'updated_at' },
  ];
});

// todo 详情页跳转待定
const handleGoToAccount = () => {
  // TODO: 替换为真实路由，跳转到三级账号页面
  const url = `${window.location.origin}/#/cloud-account-manage/secondary-account/${props.policyData.related_accounts[0].account_id}`;
  window.open(url, '_blank');
};

// 应用到二级账号
const handleApplyToAccount = () => {
  emit('apply-to-account', props.policyData);
};
</script>

<template>
  <bk-sideslider v-model:is-show="model" title="权限策略库详情" :width="800" quick-close background-color="#f5f7fa">
    <template #header>
      <div class="permission-policy-header">
        <div class="title">
          权限策略库详情
          <span class="name">| {{ props.policyData.name }}</span>
        </div>
        <bk-button theme="primary" @click="handleApplyToAccount" outline>应用到二级账号</bk-button>
      </div>
    </template>
    <template #default>
      <div class="permission-policy-info">
        <!-- 基本信息卡片 -->
        <bk-card title="基本信息" :border="false" class="info-card">
          <div class="info-grid">
            <div v-for="field in baseInfoFields" :key="field.id" class="info-item">
              <span class="info-label">{{ field.label }}：</span>
              <!--区分是不是关联二级账号数-->
              <span class="info-value" v-if="field.id === 'related_account_count' && field.value">
                <div class="relate-account-count" @click="handleGoToAccount">
                  <span class="num">{{ field.value }}</span>
                  <i class="hcm-icon bkhcm-icon-jump-fill" v-if="field.value" />
                </div>
              </span>
              <span class="info-value" v-else>{{ field.value || '--' }}</span>
            </div>
          </div>
        </bk-card>

        <!-- 权限模版 -->
        <bk-card title="权限模版" :border="false" :disable-header-style="true" class="permission-card">
          <JSON :content="props.policyData.policy_document" />
        </bk-card>
      </div>
    </template>
  </bk-sideslider>
</template>

<style lang="scss" scoped>
:deep(.permission-policy-header) {
  display: inline-flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  padding-right: 24px;

  .title {
    font-size: 16px;
    color: #313238;

    .name {
      font-size: 14px;
      color: #979ba5;
    }
  }
}

.permission-policy-info {
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
      cursor: default;
    }
  }

  .info-grid {
    display: grid;
    grid-template-columns: repeat(1, 1fr);
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

        .relate-account-count {
          color: #3a84ff;
          cursor: pointer;
        }
      }
    }
  }

  .permission-card {
    padding: 0 24px 12px;

    :deep(.bk-card-head) {
      padding: 0;
      align-items: center;
      border-bottom: 0;

      .bk-card-title {
        font-size: 14px;
        font-weight: 700;
        color: #313238;
      }
    }

    :deep(.bk-card-body) {
      padding: 12px;
      background: hsl(216, 33%, 97%);
      height: 350px;
      overflow-y: auto;
    }
  }
}
</style>
