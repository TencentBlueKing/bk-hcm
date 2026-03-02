<script setup lang="ts">
import { ref, computed } from 'vue';
import { Share } from 'bkui-vue/lib/icon';
import { SECRET_STATUS_MAP } from '../constants';
import type { ICloudSecretItem, SecretActionType } from '../typings';
import SecretActionDialog from './secret-action-dialog.vue';

// Props 定义
const props = defineProps<{
  modelValue: boolean;
  secretData: ICloudSecretItem | null;
  vendor: string;
}>();

// Emits 定义
const emit = defineEmits<{
  'update:modelValue': [value: boolean];
  'action-success': [];
}>();

// 内部显示状态
const isShow = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val),
});

// 操作弹窗状态
const showActionDialog = ref(false);
const currentActionType = ref<SecretActionType>('disable');

// 获取密钥状态配置
const statusConfig = computed(() => {
  if (!props.secretData) return null;
  return SECRET_STATUS_MAP[props.secretData.status];
});

// 判断是否可以禁用（已启用状态）
const canDisable = computed(() => props.secretData?.status === 'enabled');

// 判断是否可以启用（已禁用状态）
const canEnable = computed(() => props.secretData?.status === 'disabled');

// 格式化时间显示
const formatDateTime = (dateStr?: string) => {
  if (!dateStr) return '--';
  return dateStr.replace('T', ' ').replace('Z', '');
};

// 禁用/启用按钮点击
const handleToggleStatus = () => {
  currentActionType.value = canDisable.value ? 'disable' : 'enable';
  showActionDialog.value = true;
};

// 操作成功回调
const handleActionSuccess = () => {
  emit('action-success');
  isShow.value = false;
};
</script>

<template>
  <bk-sideslider v-model:is-show="isShow" title="云密钥详情" :width="640" quick-close :before-close="() => true">
    <template #header>
      <div class="slider-header">
        <span class="title">云密钥详情</span>
        <div class="header-actions">
          <bk-button v-if="canDisable" theme="primary" outline size="small" @click="handleToggleStatus">禁用</bk-button>
          <bk-button v-if="canEnable" theme="primary" outline size="small" @click="handleToggleStatus">启用</bk-button>
        </div>
      </div>
    </template>

    <div class="secret-detail-container">
      <div class="detail-content">
        <!-- 云密钥ID -->
        <div class="detail-item">
          <span class="label">云密钥ID：</span>
          <span class="value">{{ secretData?.cloud_secret_id || secretData?.extension?.cloud_secret_id || '--' }}</span>
        </div>

        <!-- 云密钥状态 -->
        <div class="detail-item">
          <span class="label">云密钥状态：</span>
          <span class="value status-value" v-if="statusConfig">
            <span :class="['status-dot', statusConfig.dotClass]"></span>
            {{ statusConfig.text }}
          </span>
          <span class="value" v-else>--</span>
        </div>

        <!-- 所属三级账号 -->
        <div class="detail-item">
          <span class="label">所属三级账号：</span>
          <span class="value link-value">
            {{ secretData?.cloud_sub_account_id || secretData?.extension?.cloud_sub_account_id || '--' }}
            <Share
              class="icon-link"
              v-if="secretData?.cloud_sub_account_id || secretData?.extension?.cloud_sub_account_id"
            />
          </span>
        </div>

        <!-- 三级账号负责人 -->
        <div class="detail-item">
          <span class="label">三级账号负责人：</span>
          <span class="value">{{ secretData?.sub_account_manager || '--' }}</span>
        </div>

        <!-- 所属二级账号 -->
        <div class="detail-item">
          <span class="label">所属二级账号：</span>
          <span class="value link-value">
            {{ secretData?.cloud_main_account_id || secretData?.extension?.cloud_main_account_id || '--' }}
            <Share
              class="icon-link"
              v-if="secretData?.cloud_main_account_id || secretData?.extension?.cloud_main_account_id"
            />
          </span>
        </div>

        <!-- 二级账号负责人 -->
        <div class="detail-item">
          <span class="label">二级账号负责人：</span>
          <span class="value">{{ secretData?.account_manager || '--' }}</span>
        </div>

        <!-- 创建时间 -->
        <div class="detail-item">
          <span class="label">创建时间：</span>
          <span class="value">{{ formatDateTime(secretData?.cloud_created_at) }}</span>
        </div>

        <!-- 更新时间 -->
        <div class="detail-item">
          <span class="label">更新时间：</span>
          <span class="value">{{ formatDateTime(secretData?.updated_at) }}</span>
        </div>

        <!-- 最近访问时间 -->
        <div class="detail-item">
          <span class="label">最近访问时间：</span>
          <span class="value">{{ formatDateTime(secretData?.last_used_time) }}</span>
        </div>
      </div>
    </div>

    <!-- 操作弹窗 -->
    <SecretActionDialog
      v-model="showActionDialog"
      :action-type="currentActionType"
      :secret-data="secretData"
      :vendor="vendor"
      @success="handleActionSuccess"
    />
  </bk-sideslider>
</template>

<style lang="scss" scoped>
.slider-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  padding-right: 40px;

  .title {
    font-size: 16px;
    font-weight: 700;
    color: #313238;
  }

  .header-actions {
    display: flex;
    gap: 8px;
  }
}

.secret-detail-container {
  padding: 24px 40px;

  .detail-content {
    .detail-item {
      display: flex;
      align-items: center;
      font-size: 12px;
      line-height: 32px;
      padding: 4px 0;

      .label {
        color: #979ba5;
        min-width: 110px;
        text-align: right;
        flex-shrink: 0;
      }

      .value {
        color: #313238;
        margin-left: 16px;

        &.status-value {
          display: flex;
          align-items: center;
        }

        &.link-value {
          display: flex;
          align-items: center;
          color: #3a84ff;
          cursor: pointer;

          .icon-link {
            margin-left: 4px;
            font-size: 14px;
          }
        }
      }
    }
  }
}

.status-dot {
  display: inline-block;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  margin-right: 6px;

  &.dot-enabled {
    background-color: #2dcb56;
  }

  &.dot-disabled {
    background-color: #979ba5;
  }
}
</style>
